package gantry // import "github.com/ad-freiburg/gantry"

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"sync"
	"time"

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
	Steps     StepList    `json:"steps"`
	Services  ServiceList `json:"services"`
	pipelines *pipelines
}

func (p PipelineDefinition) Pipelines() (*pipelines, error) {
	if p.pipelines == nil {
		steps := make(map[string]Step, 0)
		for name, step := range p.Steps {
			if val, ok := steps[name]; ok {
				return nil, fmt.Errorf("Redeclaration of step '%s'", val.Name)
			}
			steps[name] = step
		}
		for name, step := range p.Services {
			if val, ok := steps[name]; ok {
				return nil, fmt.Errorf("Redeclaration of step '%s'", val.Name)
			}
			steps[name] = step
		}

		t, err := NewTarjan(steps)
		if err != nil {
			return nil, err
		}
		res, err := t.Parse()
		if err != nil {
			return nil, err
		}
		result := pipelines(*res)
		p.pipelines = &result
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

func NewPipeline(definitionPath, environmentPath string) (*Pipeline, error) {
	p := &Pipeline{NetworkName: "gantry"}
	err := p.loadPipelineDefinition(definitionPath)
	if err != nil {
		return nil, err
	}
	// Environment is only needed for steps
	if len(p.Definition.Steps) > 0 {
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
			return fmt.Errorf("No machine for role '%s' in '%s'", step.Role, step.Name)
		}
		if step.Image == "" && step.BuildInfo.Context == "" {
			return fmt.Errorf("No container information for '%s'", step.Name)
		}
		if step.Command != "" && len(step.Args) > 0 {
			return fmt.Errorf("Only command or args allowed for '%s'", step.Name)
		}
	}
	return nil
}

func runParallelPrepareImage(step Step, force bool, durations *sync.Map, wg *sync.WaitGroup, s chan struct{}) {
	defer wg.Done()
	<-s

	pipelineLogger.Printf("- Preparing %s", step.Name)
	duration, err := executeF(NewImageExistenceChecker(step))
	exists := err == nil

	var f func() error
	if step.Image != "" {
		f = NewImagePuller(step)
	} else {
		f = NewImageBuilder(step)
	}
	if !exists || force {
		err := f()
		duration2, err := executeF(f)
		if err != nil {
			pipelineLogger.Println(err)
		}
		duration += duration2
	}
	durations.Store(step.Name, duration)
	pipelineLogger.Printf("- Prepared %s after %s", step.Name, duration)
}

func (p Pipeline) PrepareImages(force bool) error {
	pipelines, err := p.Definition.Pipelines()
	if err != nil {
		return err
	}
	runChannel := make(chan struct{})
	var wg sync.WaitGroup
	images := 0
	durations := &sync.Map{}

	for _, step := range pipelines.AllSteps() {
		wg.Add(1)
		go runParallelPrepareImage(step, force, durations, &wg, runChannel)
		images++
	}

	pipelineLogger.Printf("Prepare Images:")
	start := time.Now()
	close(runChannel)
	wg.Wait()

	pipelineLogger.Printf("Prepared %d images in %s", images, time.Since(start))
	var totalElapsedTime time.Duration = 0
	durations.Range(func(key, value interface{}) bool {
		duration, ok := value.(time.Duration)
		if ok {
			totalElapsedTime += duration
		}
		return ok
	})
	pipelineLogger.Printf("Total time spend preparing images: %s", totalElapsedTime)
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

func runParallelStep(step Step, pipeline Pipeline, durations *sync.Map, wg *sync.WaitGroup, p []chan struct{}, o chan struct{}) {
	defer wg.Done()
	defer close(o)
	for x := range p {
		<-p[x]
	}
	pipelineLogger.Printf("- Starting: %s", step.Name)
	duration, err := executeF(NewContainerRunner(step, pipeline.NetworkName))
	pipelineLogger.Printf("- Finished %s after %s", step.Name, duration)
	if err != nil {
		pipelineLogger.Println(err)
	}
	durations.Store(step.Name, duration)
}

func (p Pipeline) ExecuteSteps() error {
	pipelines, err := p.Definition.Pipelines()
	if err != nil {
		return err
	}
	wgs := make([]sync.WaitGroup, len(*pipelines))
	steps := 0
	durations := &sync.Map{}
	runChannel := make(chan struct{})
	channels := make(map[string]chan struct{})
	for pi, pipeline := range *pipelines {
		for _, step := range pipeline {
			channels[step.Name] = make(chan struct{})
			preChannels := make([]chan struct{}, 0)
			preChannels = append(preChannels, runChannel)
			for pre, _ := range step.After {
				preChannels = append(preChannels, channels[pre])
			}
			wgs[pi].Add(1)
			go runParallelStep(step, p, durations, &wgs[pi], preChannels, channels[step.Name])
			steps++
		}
	}

	pipelineLogger.Printf("Execute:")
	start := time.Now()
	close(runChannel)
	for pi, _ := range *pipelines {
		wgs[pi].Wait()
	}
	pipelineLogger.Printf("Executed %d steps in %s", steps, time.Since(start))
	var totalElapsedTime time.Duration = 0
	durations.Range(func(key, value interface{}) bool {
		duration, ok := value.(time.Duration)
		if ok {
			totalElapsedTime += duration
		}
		return ok
	})
	pipelineLogger.Printf("Total time spend inside steps: %s", totalElapsedTime)
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
	Roles StringSet
	Paths Paths
}

type Paths struct {
	Input   map[string]string
	Output  map[string]string
	Scratch string
}

type BuildInfo struct {
	Context    string `json:"context"`
	Dockerfile string `json:"Dockerfile"`
}

type Service struct {
	BuildInfo   BuildInfo `json:"build"`
	Command     string    `json:"command"`
	Image       string    `json:"image"`
	Ports       []string  `json:"ports"`
	Volumes     []string  `json:"volumes"`
	Environment []string  `json:"environment"`
	DependsOn   StringSet `json:"depends_on"`
	Name        string
}

type Step struct {
	Service
	Role   string    `json:"role"`
	Args   []string  `json:"args"`
	After  StringSet `json:"after"`
	Detach bool      `json:"detach"`
	prefix string
}

func (s Step) Dependencies() (*StringSet, error) {
	r := StringSet{}
	for dep, _ := range s.After {
		r[dep] = true
	}
	for dep, _ := range s.DependsOn {
		r[dep] = true
	}
	return &r, nil
}

func (s Step) ImageName() string {
	if s.Image != "" {
		return s.Image
	}
	return strings.Replace(strings.ToLower(s.Name), " ", "_", -1)
}

func (s Step) ContainerName() string {
	return strings.Replace(strings.ToLower(s.Name), " ", "_", -1)
}

func (s Step) Runner() Runner {
	if s.prefix == "" {
		s.prefix = ApplyStyle(s.ContainerName(), GetNextFriendlyColor())
	}
	r := NewLocalRunner(s.prefix, os.Stdout, os.Stderr)
	return r
}
