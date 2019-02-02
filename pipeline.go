package gantry // import "github.com/ad-freiburg/gantry"

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/ad-freiburg/gantry/types"
	"github.com/ghodss/yaml"
)

var (
	pipelineLogger *PrefixedLogger
)

func init() {
	pipelineLogger = NewPrefixedLogger(
		ApplyStyle("pipeline", STYLE_BOLD),
		log.New(os.Stderr, "", log.LstdFlags),
	)
}

type Pipeline struct {
	Definition  PipelineDefinition
	Environment PipelineEnvironment
	NetworkName string
}

type PipelineDefinition struct {
	Steps        StepList    `json:"steps"`
	Services     ServiceList `json:"services"`
	pipelines    *pipelines
	ignoredSteps types.StringSet
}

func (p *PipelineDefinition) Pipelines() (*pipelines, error) {
	if p.pipelines == nil {
		steps := make(map[string]Step, 0)
		for name, step := range p.Steps {
			if _, ignore := p.ignoredSteps[name]; ignore {
				continue
			}
			if val, ok := steps[name]; ok {
				return nil, fmt.Errorf("Redeclaration of step '%s'", val.Name())
			}
			for ignored, _ := range p.ignoredSteps {
				delete(step.After, ignored)
				delete(step.DependsOn, ignored)
			}
			steps[name] = step
		}
		for name, step := range p.Services {
			if _, ignore := p.ignoredSteps[name]; ignore {
				continue
			}
			if val, ok := steps[name]; ok {
				return nil, fmt.Errorf("Redeclaration of step '%s'", val.Name())
			}
			for ignored, _ := range p.ignoredSteps {
				delete(step.After, ignored)
				delete(step.DependsOn, ignored)
			}
			steps[name] = step
		}

		t, err := NewTarjan(steps)
		if err != nil {
			return nil, err
		}
		p.pipelines, err = t.Parse()
		if err != nil {
			return nil, err
		}
	}
	return p.pipelines, nil
}

type pipelines [][]Step

func (p pipelines) AllSteps() []Step {
	result := make([]Step, 0)
	for _, pipeline := range p {
		for _, step := range pipeline {
			result = append(result, step)
		}
	}
	return result
}

type PipelineEnvironment struct {
	Machines []Machine `json:"machines"`
}

func NewPipeline(definitionPath, environmentPath string, ignoredSteps types.StringSet) (*Pipeline, error) {
	p := &Pipeline{}
	err := p.loadPipelineDefinition(definitionPath)
	if err != nil {
		return nil, err
	}
	p.Definition.ignoredSteps = ignoredSteps
	if p.checkRequireEnvironment() {
		err = p.setPipelineEnvironment(environmentPath)
		if err != nil {
			return nil, err
		}
	}
	return p, nil
}

func (p *Pipeline) loadPipelineDefinition(path string) error {
	if _, err := os.Stat(GantryDef); path == "" && !os.IsNotExist(err) {
		path = GantryDef
	}
	if _, err := os.Stat(DockerCompose); path == "" && !os.IsNotExist(err) {
		path = DockerCompose
	}
	file, err := os.Open(path)
	if err != nil {
		pipelineLogger.Println("Could not open pipeline definition.")
		return err
	}
	defer file.Close()

	data, err := ioutil.ReadAll(file)
	if err != nil {
		return err
	}
	return yaml.Unmarshal(data, &p.Definition)
}

func (p *Pipeline) checkRequireEnvironment() bool {
	return false
}

func (p *Pipeline) setPipelineEnvironment(path string) error {
	if _, err := os.Stat(GantryEnv); path == "" && !os.IsNotExist(err) {
		path = GantryEnv
	}
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	data, err := ioutil.ReadAll(file)
	if err != nil {
		return err
	}
	return yaml.Unmarshal(data, &p.Environment)
}

func (p Pipeline) Check() error {
	roleProvider := make(map[string][]Machine)
	for _, machine := range p.Environment.Machines {
		for role, _ := range machine.Roles {
			roleProvider[role] = append(roleProvider[role], machine)
		}
	}

	pipelines, err := p.Definition.Pipelines()
	if err != nil {
		return err
	}
	for _, step := range pipelines.AllSteps() {
		if step.Role != "" && len(roleProvider[step.Role]) < 1 {
			return fmt.Errorf("No machine for role '%s' in '%s'", step.Role, step.ColoredName())
		}
		if step.Image == "" && step.BuildInfo.Context == "" && step.BuildInfo.Dockerfile == "" {
			return fmt.Errorf("No container information for '%s'", step.ColoredName())
		}
	}
	return nil
}

func runParallelBuildImage(step Step, pull bool, durations *sync.Map, wg *sync.WaitGroup, s chan struct{}) {
	defer wg.Done()
	<-s

	duration, err := executeF(NewImageBuilder(step, pull))
	if err != nil {
		pipelineLogger.Println(err)
	}
	durations.Store(step.name, duration)
}

func (p Pipeline) BuildImages(force bool) error {
	pipelines, err := p.Definition.Pipelines()
	if err != nil {
		return err
	}
	runChannel := make(chan struct{})
	var wg sync.WaitGroup
	images := 0
	durations := &sync.Map{}

	for _, step := range pipelines.AllSteps() {
		if step.BuildInfo.Dockerfile == "" && step.BuildInfo.Context == "" {
			continue
		}
		wg.Add(1)
		go runParallelBuildImage(step, force, durations, &wg, runChannel)
		images++
	}

	if Verbose {
		pipelineLogger.Printf("Build Images:")
	}
	start := time.Now()
	close(runChannel)
	wg.Wait()

	if Verbose {
		pipelineLogger.Printf("Build %d images in %s", images, time.Since(start))
	}
	var totalElapsedTime time.Duration = 0
	durations.Range(func(key, value interface{}) bool {
		duration, ok := value.(time.Duration)
		if ok {
			totalElapsedTime += duration
		}
		return ok
	})
	if Verbose {
		pipelineLogger.Printf("Total time spent building images: %s", totalElapsedTime)
	}
	return nil
}

func runParallelPullImage(step Step, force bool, durations *sync.Map, wg *sync.WaitGroup, s chan struct{}) {
	defer wg.Done()
	<-s

	duration, err := executeF(NewImageExistenceChecker(step))
	if err != nil || force {
		duration2, err := executeF(NewImagePuller(step))
		if err != nil {
			pipelineLogger.Println(err)
		}
		duration += duration2
	}
	durations.Store(step.name, duration)
}

func (p Pipeline) PullImages(force bool) error {
	pipelines, err := p.Definition.Pipelines()
	if err != nil {
		return err
	}
	runChannel := make(chan struct{})
	var wg sync.WaitGroup
	images := 0
	durations := &sync.Map{}

	for _, step := range pipelines.AllSteps() {
		if step.BuildInfo.Dockerfile != "" || step.BuildInfo.Context != "" {
			continue
		}
		wg.Add(1)
		go runParallelPullImage(step, force, durations, &wg, runChannel)
		images++
	}

	if Verbose {
		pipelineLogger.Printf("Pull Images:")
	}
	start := time.Now()
	close(runChannel)
	wg.Wait()

	if Verbose {
		pipelineLogger.Printf("Pulled %d images in %s", images, time.Since(start))
	}
	var totalElapsedTime time.Duration = 0
	durations.Range(func(key, value interface{}) bool {
		duration, ok := value.(time.Duration)
		if ok {
			totalElapsedTime += duration
		}
		return ok
	})
	if Verbose {
		pipelineLogger.Printf("Total time spent pulling images: %s", totalElapsedTime)
	}
	return nil
}

func (p Pipeline) KillContainers() error {
	pipelines, err := p.Definition.Pipelines()
	if err != nil {
		return err
	}
	for _, pipeline := range *pipelines {
		for _, step := range pipeline {
			NewContainerKiller(step)()
		}
	}
	return nil
}

func (p Pipeline) RemoveContainers() error {
	pipelines, err := p.Definition.Pipelines()
	if err != nil {
		return err
	}
	for _, pipeline := range *pipelines {
		for _, step := range pipeline {
			NewOldContainerRemover(step)()
		}
	}
	return nil
}

func (p Pipeline) CreateNetwork() error {
	NewNetworkCreator(p)()
	return nil
}

func (p Pipeline) RemoveNetwork() error {
	NewNetworkRemover(p)()
	return nil
}

func (p Pipeline) Runner() Runner {
	r := NewLocalRunner("local", os.Stdout, os.Stderr)
	return r
}

func runParallelStep(step Step, pipeline Pipeline, durations *sync.Map, wg *sync.WaitGroup, preconditions []chan struct{}, o chan struct{}) {
	defer wg.Done()
	defer close(o)
	for i, c := range preconditions {
		if Verbose {
			pipelineLogger.Printf("%s waiting for %d preconditions", step.ColoredContainerName(), len(preconditions)-i)
		}
		<-c
		if Verbose {
			pipelineLogger.Printf("Precondition for %s satisfied %d remaining", step.ColoredContainerName(), len(preconditions)-i-1)
		}
	}
	pipelineLogger.Printf("- Starting: %s", step.ColoredContainerName())
	duration, err := executeF(NewContainerRunner(step, pipeline.NetworkName))
	pipelineLogger.Printf("- Finished %s after %s", step.ColoredContainerName(), duration)
	if err != nil {
		pipelineLogger.Println(err)
	}
	durations.Store(step.name, duration)
}

func (p Pipeline) ExecuteSteps() error {
	pipelines, err := p.Definition.Pipelines()
	if err != nil {
		return err
	}
	var wg sync.WaitGroup
	steps := 0
	durations := &sync.Map{}
	runChannel := make(chan struct{})
	channels := make(map[string]chan struct{})
	for _, pipeline := range *pipelines {
		for _, step := range pipeline {
			channels[step.name] = make(chan struct{})
			preChannels := make([]chan struct{}, 0)
			preChannels = append(preChannels, runChannel)
			dependencies, _ := step.Dependencies()
			for pre, _ := range *dependencies {
				if Verbose {
					pipelineLogger.Printf("Adding %s as precondition for %s", ApplyStyle(pre, STYLE_BOLD), step.ColoredContainerName())
				}
				val, ok := channels[pre]
				if !ok {
					log.Fatalf("Unknown precondition: %s", pre)
				}
				preChannels = append(preChannels, val)
			}
			wg.Add(1)
			go runParallelStep(step, p, durations, &wg, preChannels, channels[step.name])
			steps++
		}
	}

	pipelineLogger.Printf("Execute:")
	start := time.Now()
	close(runChannel)
	wg.Wait()
	pipelineLogger.Printf("Executed %d steps in %s", steps, time.Since(start))
	var totalElapsedTime time.Duration = 0
	durations.Range(func(key, value interface{}) bool {
		duration, ok := value.(time.Duration)
		if ok {
			totalElapsedTime += duration
		}
		return ok
	})
	pipelineLogger.Printf("Total time spent inside steps: %s", totalElapsedTime)
	return nil
}

func executeF(f func() error) (time.Duration, error) {
	start := time.Now()
	err := f()
	elapsed := time.Since(start)
	return elapsed, err
}

type Machine struct {
	Host  string
	Roles types.StringSet
	Paths Paths
}

type Paths struct {
	Input   map[string]string
	Output  map[string]string
	Scratch string
}

type BuildInfo struct {
	Context    string                       `json:"context"`
	Dockerfile string                       `json:"dockerfile"`
	Args       types.StringMapOrStringSlice `json:"args"`
}

type Service struct {
	BuildInfo   BuildInfo                 `json:"build"`
	Command     types.StringOrStringSlice `json:"command"`
	Entrypoint  types.StringOrStringSlice `json:"entrypoint"`
	Image       string                    `json:"image"`
	Ports       []string                  `json:"ports"`
	Volumes     []string                  `json:"volumes"`
	Environment []string                  `json:"environment"`
	DependsOn   types.StringSet           `json:"depends_on"`
	name        string
	color       int
}

type Step struct {
	Service
	Role   string          `json:"role"`
	After  types.StringSet `json:"after"`
	Detach bool            `json:"detach"`
}

func (s Step) Dependencies() (*types.StringSet, error) {
	r := types.StringSet{}
	for dep, _ := range s.After {
		r[dep] = true
	}
	for dep, _ := range s.DependsOn {
		r[dep] = true
	}
	return &r, nil
}

func (s *Service) SetName(name string) {
	s.name = name
	s.color = GetNextFriendlyColor()
}

func (s Service) Name() string {
	return s.name
}

func (s Service) ColoredName() string {
	return ApplyStyle(s.name, s.color)
}

func (s Service) ColoredContainerName() string {
	return ApplyStyle(s.ContainerName(), s.color)
}

func (s Service) ImageName() string {
	if s.Image != "" {
		return s.Image
	}
	return strings.Replace(strings.ToLower(s.name), " ", "_", -1)
}

func (s Service) RawContainerName() string {
	return strings.Replace(strings.ToLower(s.name), " ", "_", -1)
}

func (s Service) ContainerName() string {
	return fmt.Sprintf("%s_%s", ProjectName, strings.Replace(strings.ToLower(s.name), " ", "_", -1))
}

func (s Service) Runner() Runner {
	r := NewLocalRunner(s.ColoredContainerName(), os.Stdout, os.Stderr)
	return r
}
