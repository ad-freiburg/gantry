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
		ApplyAnsiStyle("pipeline", AnsiStyleBold),
		log.New(os.Stderr, "", log.LstdFlags),
	)
}

// Pipeline stores all definitions and settings regarding a deployment.
type Pipeline struct {
	Definition  *PipelineDefinition
	Environment *PipelineEnvironment
	NetworkName string
}

// NewPipeline creates a new Pipeline from given files which ignores the
// existence of steps with names provided in ignoreSteps.
func NewPipeline(definitionPath, environmentPath string, ignoredSteps types.StringSet) (*Pipeline, error) {
	p := &Pipeline{}
	def, err := NewPipelineDefinition(definitionPath)
	if err != nil {
		return nil, err
	}
	def.IgnoredSteps = ignoredSteps
	p.Definition = def
	p.Environment = NewPipelineEnvironment()
	if p.checkRequireEnvironment() {
		err := p.Environment.Load(environmentPath)
		if err != nil {
			return nil, err
		}
	}
	if err = p.Environment.ApplyTo(def); err != nil {
		return nil, err
	}
	return p, nil
}

func (p *Pipeline) checkRequireEnvironment() bool {
	return false
}

// CleanUp removes temporarie data.
func (p *Pipeline) CleanUp(signal os.Signal) {
	p.Definition.CleanUp(signal)
	if p.Environment != nil {
		p.Environment.CleanUp(signal)
	}
}

// Check validates Pipeline p, checks if all required information is present.
func (p Pipeline) Check() error {
	roleProvider := make(map[string][]Machine)
	if p.Environment != nil {
		for _, machine := range p.Environment.Machines {
			for role := range machine.Roles {
				roleProvider[role] = append(roleProvider[role], machine)
			}
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

// Pipelines stores parallel and dependent steps/services.
type Pipelines [][]Step

// AllSteps returns all in p defined steps without ordering information.
func (p Pipelines) AllSteps() []Step {
	result := make([]Step, 0)
	for _, pipeline := range p {
		for _, step := range pipeline {
			result = append(result, step)
		}
	}
	return result
}

// Check performs checks for cyclic dependencies and requirements fulfillment.
func (p *Pipelines) Check() error {
	result := make(Pipelines, 0)
	// walk reverse order, if all requirements are found the next step is a new
	// component
	resultIndex := 0
	requirements := make(map[string]bool, 0)
	for i := len(*p) - 1; i >= 0; i-- {
		steps := (*p)[i]
		if len(steps) > 1 {
			names := make([]string, len(steps))
			for i, step := range steps {
				names[i] = step.Name
			}
			return fmt.Errorf("cyclic component found in (sub)pipeline: '%s'", strings.Join(names, ", "))
		}
		var step = steps[0]
		for r := range *step.Dependencies() {
			requirements[r] = true
		}
		delete(requirements, step.Name)
		if len(result)-1 < resultIndex {
			result = append(result, make([]Step, 0))
		}
		result[resultIndex] = append([]Step{step}, result[resultIndex]...)
		if len(requirements) == 0 {
			resultIndex++
		}
	}
	*p = result
	return nil
}

// PipelineDefinition stores docker-compose services and gantry steps.
type PipelineDefinition struct {
	Steps        StepList    `json:"steps"`
	Services     ServiceList `json:"services"`
	pipelines    *Pipelines
	IgnoredSteps types.StringSet
}

func NewPipelineDefinition(path string) (*PipelineDefinition, error) {
	if _, err := os.Stat(GantryDef); path == "" && !os.IsNotExist(err) {
		path = GantryDef
	}
	if _, err := os.Stat(DockerCompose); path == "" && !os.IsNotExist(err) {
		path = DockerCompose
	}
	file, err := os.Open(path)
	if err != nil {
		pipelineLogger.Println("Could not open pipeline definition.")
		return nil, err
	}
	defer file.Close()

	data, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, err
	}
	d := &PipelineDefinition{}
	err = yaml.Unmarshal(data, d)
	return d, err
}

// CleanUp removes temporarie data.
func (p *PipelineDefinition) CleanUp(signal os.Signal) {
}

// Pipelines calculates and verifies dependencies and ordering for steps
// defined in the PipelineDefinition p.
func (p *PipelineDefinition) Pipelines() (*Pipelines, error) {
	if p.pipelines == nil {
		steps := make(map[string]Step, 0)
		// Verfiy steps
		for name, step := range p.Steps {
			if _, ignore := p.IgnoredSteps[name]; ignore {
				continue
			}
			if val, ok := steps[name]; ok {
				return nil, fmt.Errorf("Redeclaration of step '%s'", val.Name)
			}
			for ignored := range p.IgnoredSteps {
				delete(step.After, ignored)
				delete(step.DependsOn, ignored)
			}
			steps[name] = step
		}
		// Verify services
		for name, step := range p.Services {
			if _, ignore := p.IgnoredSteps[name]; ignore {
				continue
			}
			if val, ok := steps[name]; ok {
				return nil, fmt.Errorf("Redeclaration of step '%s'", val.Name)
			}
			for ignored := range p.IgnoredSteps {
				delete(step.After, ignored)
				delete(step.DependsOn, ignored)
			}
			steps[name] = step
		}

		// Calculate order and indepenence
		pipelines, err := NewTarjan(steps)
		if err != nil {
			return nil, err
		}
		// Verify pipelines
		err = pipelines.Check()
		if err != nil {
			return nil, err
		}
		p.pipelines = pipelines
	}
	return p.pipelines, nil
}

func runParallelBuildImage(step Step, pull bool, durations *sync.Map, wg *sync.WaitGroup, s chan struct{}) {
	defer wg.Done()
	<-s

	duration, err := executeF(NewImageBuilder(step, pull))
	if err != nil {
		pipelineLogger.Println(err)
	}
	durations.Store(step.Name, duration)
}

// BuildImages builds all buildable images of Pipeline p in parallel.
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
	var totalElapsedTime time.Duration
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
	durations.Store(step.Name, duration)
}

// PullImages pulls all pullable images of Pipeline p in parallel.
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
	var totalElapsedTime time.Duration
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

// KillContainers kills all running containers of Pipeline p.
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

// RemoveContainers removes all stopped containers of Pipeline p.
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

// CreateNetwork creates a network using the NetworkName of the Pipeline p.
func (p Pipeline) CreateNetwork() error {
	NewNetworkCreator(p)()
	return nil
}

// RemoveNetwork removes the network of Pipeline p.
func (p Pipeline) RemoveNetwork() error {
	NewNetworkRemover(p)()
	return nil
}

// Runner returns a runner for the pipeline itself. Currently only localhost.
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
	durations.Store(step.Name, duration)
}

// ExecuteSteps runs all not ignored steps/services in the order defined by
// there dependencies. Each step/service is run as soon as possible.
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
			channels[step.Name] = make(chan struct{})
			preChannels := make([]chan struct{}, 0)
			preChannels = append(preChannels, runChannel)
			for pre := range *step.Dependencies() {
				if Verbose {
					pipelineLogger.Printf("Adding %s as precondition for %s", ApplyAnsiStyle(pre, AnsiStyleBold), step.ColoredContainerName())
				}
				val, ok := channels[pre]
				if !ok {
					log.Fatalf("Unknown precondition: %s", pre)
				}
				preChannels = append(preChannels, val)
			}
			wg.Add(1)
			go runParallelStep(step, p, durations, &wg, preChannels, channels[step.Name])
			steps++
		}
	}

	pipelineLogger.Printf("Execute:")
	start := time.Now()
	close(runChannel)
	wg.Wait()
	pipelineLogger.Printf("Executed %d steps in %s", steps, time.Since(start))
	var totalElapsedTime time.Duration
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
