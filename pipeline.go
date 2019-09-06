package gantry // import "github.com/ad-freiburg/gantry"

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
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
	Network     Network
	localRunner Runner
	noopRunner  Runner
}

// NewPipeline creates a new Pipeline from given files which ignores the
// existence of steps with names provided in ignoreSteps.
func NewPipeline(definitionPath, environmentPath string, environment types.StringMap, ignoredSteps types.StringSet, selectedSteps types.StringSet) (*Pipeline, error) {
	p := &Pipeline{}
	var err error
	// Load environment
	p.Environment, err = NewPipelineEnvironment(environmentPath, environment, ignoredSteps, selectedSteps)
	if err != nil {
		// As environment files are optional, handle if non is accessible
		if e, ok := err.(*os.PathError); ok && e.Err == syscall.ENOENT {
			log.Print("No environment file is used")
		} else {
			return nil, err
		}
	}
	// Load definition
	p.Definition, err = NewPipelineDefinition(definitionPath, p.Environment)
	p.localRunner = NewLocalRunner("pipeline", os.Stdout, os.Stderr)
	p.noopRunner = NewNoopRunner(false)
	return p, err
}

// CleanUp removes containers and temporary data.
func (p *Pipeline) CleanUp(signal os.Signal) error {
	var keepNetworkAlive bool
	// Stop all services which are not marked as keep-running
	pipelines, err := p.Definition.Pipelines()
	if err != nil {
		return err
	}
	for _, pipeline := range *pipelines {
		for _, step := range pipeline {
			// If services are still running, keep the network
			if step.Meta.Type == ServiceTypeService && step.Meta.KeepAlive != KeepAliveNo {
				if Verbose {
					log.Printf("Keeping network as '%s' can be still alive", step.ColoredName())
				}
				keepNetworkAlive = true
			}
			// Remove all steps and services marked as not to keep alive
			if step.Meta.Type == ServiceTypeStep || step.Meta.KeepAlive == KeepAliveNo {
				runner := p.GetRunnerForMeta(step.Meta)
				if _, err := runner.ContainerKiller(step)(); err != nil {
					pipelineLogger.Printf("Error killing %s: %s", step.ColoredName(), err)
				}
				if err := runner.ContainerRemover(step)(); err != nil {
					pipelineLogger.Printf("Error removing %s: %s", step.ColoredName(), err)
				}
			}
			step.Meta.Close()
		}
	}
	// If we are allowed, start a cleanup container to delete all files in the
	// temporary directories as deletion from outside will fail when
	// user-namespaces are used.
	if !p.Environment.TempDirNoAutoClean {
		if err := p.RemoveTempDirData(); err != nil {
			pipelineLogger.Printf("Error removing temporary directories: %s", err)
		}
	}
	// Remove network if not needed anymore
	if !keepNetworkAlive {
		if err := p.RemoveNetwork(); err != nil {
			pipelineLogger.Printf("Error removing network %s: %s", string(p.Network), err)
		}
	}
	return p.Environment.CleanUp(signal)
}

// Check validates Pipeline p, checks if all required information is present.
func (p Pipeline) Check() error {
	pipelines, err := p.Definition.Pipelines()
	if err != nil {
		return err
	}
	for _, step := range pipelines.AllSteps() {
		if err := step.Check(); err != nil {
			return err
		}
	}
	return nil
}

// GetRunnerForMeta selects a suitable runner given a ServiceMeta instance.
func (p Pipeline) GetRunnerForMeta(meta ServiceMeta) Runner {
	if meta.Ignore {
		return p.noopRunner.Copy()
	}
	return p.localRunner.Copy()
}

// Pipelines stores parallel and dependent steps/services.
type Pipelines [][]Step

// AllSteps returns all in p defined steps without ordering information.
func (p Pipelines) AllSteps() []Step {
	result := make([]Step, 0)
	for _, pipeline := range p {
		result = append(result, pipeline...)
	}
	return result
}

// Check performs checks for cyclic dependencies and requirements fulfillment.
func (p *Pipelines) Check() error {
	result := make(Pipelines, 0)
	// walk reverse order, if all requirements are found the next step is a new
	// component
	resultIndex := 0
	requirements := make(map[string]bool)
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
		for r := range step.Dependencies() {
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

type pipelineDefinitionJSON struct {
	Version  string
	Steps    StepList
	Services ServiceList
}

// PipelineDefinition stores docker-compose services and gantry steps.
type PipelineDefinition struct {
	Version   string
	Steps     StepList
	pipelines *Pipelines
}

// UnmarshalJSON loads a PipelineDefinition from json using the pipelineJSON struct.
func (p *PipelineDefinition) UnmarshalJSON(data []byte) error {
	result := PipelineDefinition{
		Steps: StepList{},
	}
	parsedJSON := pipelineDefinitionJSON{}
	if err := json.Unmarshal(data, &parsedJSON); err != nil {
		return err
	}
	result.Version = parsedJSON.Version
	for name, service := range parsedJSON.Services {
		service.Meta = ServiceMeta{
			Type: ServiceTypeService,
		}
		result.Steps[name] = service
	}
	for name, step := range parsedJSON.Steps {
		if _, found := result.Steps[name]; found {
			return fmt.Errorf("duplicate step/service '%s'", name)
		}
		step.Meta = ServiceMeta{
			Type:      ServiceTypeStep,
			KeepAlive: KeepAliveNo,
		}
		result.Steps[name] = step
	}
	*p = result
	return nil
}

// NewPipelineDefinition generates a pipeline definition from a path and an environment.
func NewPipelineDefinition(path string, env *PipelineEnvironment) (*PipelineDefinition, error) {
	dir, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	defaultPath := filepath.Join(dir, GantryDef)
	if _, err := os.Stat(defaultPath); path == "" && err == nil {
		path = defaultPath
	}
	defaultPath = filepath.Join(dir, DockerCompose)
	if _, err := os.Stat(defaultPath); path == "" && err == nil {
		path = defaultPath
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
	// Apply environment to yaml
	data, err = env.ApplyTo(data)
	if err != nil {
		return nil, err
	}
	d := &PipelineDefinition{}
	err = yaml.Unmarshal(data, d)
	// Update with specific meta if defined
	for name, meta := range env.Steps {
		s, ok := d.Steps[name]
		if ok {
			meta.Type = s.Meta.Type
			s.Meta = meta
			if meta.Type == ServiceTypeStep {
				s.Meta.KeepAlive = KeepAliveNo
			}
			d.Steps[name] = s
		} else if !meta.Ignore {
			log.Printf("Metadata: unknown step '%s'", name)
		}
	}
	// Open output files for container logs
	for n, step := range d.Steps {
		if err = step.Meta.Open(); err != nil {
			pipelineLogger.Printf("Error creating log output of %s: %s", step.ColoredName(), err)
		}
		d.Steps[n] = step
	}
	return d, err
}

// Pipelines calculates and verifies dependencies and ordering for steps
// defined in the PipelineDefinition p.
func (p *PipelineDefinition) Pipelines() (*Pipelines, error) {
	if p.pipelines == nil {
		// Collect ignored and selected step names
		ignoredSteps := types.StringSet{}
		selectedSteps := types.StringSet{}
		for name, step := range p.Steps {
			if step.Meta.Ignore {
				ignoredSteps[name] = true
			}
			if step.Meta.Selected {
				selectedSteps[name] = true
				if step.Meta.Ignore {
					return nil, fmt.Errorf("instructed to ignore selected step '%s'", step.Name)
				}
			}
		}

		// If steps or services are marked es selected, expand the selection
		queue := make([]string, 0)
		for name := range selectedSteps {
			queue = append(queue, name)
		}
		for len(queue) > 0 {
			name := queue[0]
			queue = queue[1:]
			if s, ok := p.Steps[name]; ok {
				if s.Meta.Ignore {
					continue
				}
				for dep := range s.Dependencies() {
					queue = append(queue, dep)
				}
				if s.Meta.Selected {
					continue
				}
				s.Meta.Selected = true
				selectedSteps[name] = true
				p.Steps[name] = s
			}
		}
		if len(selectedSteps) > 0 {
			// Ignore all not selected steps
			for name, step := range p.Steps {
				if step.Meta.Selected {
					continue
				}
				step.Meta.Ignore = true
				p.Steps[name] = step
				ignoredSteps[name] = true
			}
		}

		// Build list of active steps
		steps := make(map[string]Step)
		for name, step := range p.Steps {
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

func runParallelBuildImage(runner Runner, step Step, pull bool, durations *sync.Map, wg *sync.WaitGroup, s chan struct{}) {
	defer wg.Done()
	<-s

	duration, err := executeF(runner.ImageBuilder(step, pull))
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
		if step.IsPullable() {
			continue
		}
		wg.Add(1)
		go runParallelBuildImage(p.GetRunnerForMeta(step.Meta), step, force, durations, &wg, runChannel)
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

func runParallelPullImage(runner Runner, step Step, force bool, durations *sync.Map, wg *sync.WaitGroup, s chan struct{}) {
	defer wg.Done()
	<-s

	duration, err := executeF(runner.ImageExistenceChecker(step))
	if err != nil || force {
		duration2, err := executeF(runner.ImagePuller(step))
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
		if step.IsBuildable() {
			continue
		}
		wg.Add(1)
		go runParallelPullImage(p.GetRunnerForMeta(step.Meta), step, force, durations, &wg, runChannel)
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
func (p Pipeline) KillContainers(preRun bool) error {
	pipelines, err := p.Definition.Pipelines()
	if err != nil {
		return err
	}
	for _, pipeline := range *pipelines {
		for _, step := range pipeline {
			if preRun && step.Meta.KeepAlive == KeepAliveReplace {
				continue
			}
			runner := p.GetRunnerForMeta(step.Meta)
			if _, err := runner.ContainerKiller(step)(); err != nil {
				pipelineLogger.Printf("Error killing %s: %s", step.ColoredName(), err)
			}
			if err := runner.ContainerRemover(step)(); err != nil {
				pipelineLogger.Printf("Error removing %s: %s", step.ColoredName(), err)
			}
		}
	}
	return nil
}

// RemoveContainers removes all stopped containers of Pipeline p.
func (p Pipeline) RemoveContainers(preRun bool) error {
	pipelines, err := p.Definition.Pipelines()
	if err != nil {
		return err
	}
	for _, pipeline := range *pipelines {
		for _, step := range pipeline {
			if preRun && step.Meta.KeepAlive == KeepAliveReplace {
				continue
			}
			if err := p.GetRunnerForMeta(step.Meta).ContainerRemover(step)(); err != nil {
				pipelineLogger.Printf("Error removing %s: %s", step.ColoredName(), err)
			}
		}
	}
	return nil
}

// CreateNetwork creates a network using the NetworkName of the Pipeline p.
func (p Pipeline) CreateNetwork() error {
	return p.localRunner.NetworkCreator(p.Network)()
}

// RemoveNetwork removes the network of Pipeline p.
func (p Pipeline) RemoveNetwork() error {
	return p.localRunner.NetworkRemover(p.Network)()
}

// RemoveTempDirData deletes all data stored in temporary directories.
func (p Pipeline) RemoveTempDirData() error {
	if len(p.Environment.tempPaths) < 1 {
		return nil
	}
	step := Step{
		Service: Service{
			Name:       "TempDirCleanUp",
			Image:      "alpine",
			Entrypoint: []string{"/bin/sh"},
			Command:    []string{"-c", "rm -rf /data/*/*"},
			Meta: ServiceMeta{
				Stdout: ServiceLog{
					Handler: LogHandlerStdout,
				},
				Stderr: ServiceLog{
					Handler: LogHandlerStdout,
				},
			},
		},
	}
	if err := step.Meta.Open(); err != nil {
		pipelineLogger.Printf("Error creating log output of %s: %s", step.ColoredName(), err)
	}
	step.InitColor()
	// Mount all temporary directories as /data/i
	i := 0
	for _, v := range p.Environment.tempPaths {
		step.Volumes = append(step.Volumes, fmt.Sprintf("%s:/data/%d", v, i))
		i++
	}
	runner := p.GetRunnerForMeta(step.Meta)
	if _, err := runner.ContainerKiller(step)(); err != nil {
		pipelineLogger.Printf("Error killing %s: %s", step.ColoredName(), err)
	}
	if err := runner.ContainerRemover(step)(); err != nil {
		pipelineLogger.Printf("Error removing %s: %s", step.ColoredName(), err)
	}
	pipelineLogger.Printf("- Starting: %s", step.ColoredName())
	duration, err := executeF(runner.ContainerRunner(step, p.Network))
	if err != nil {
		pipelineLogger.Printf("  %s: %s", step.ColoredName(), err)
	}
	pipelineLogger.Printf("- Finished %s after %s", step.ColoredName(), duration)
	if err := runner.ContainerRemover(step)(); err != nil {
		pipelineLogger.Printf("Error removing %s: %s", step.ColoredName(), err)
	}
	step.Meta.Close()
	return err
}

func runParallelStep(runner Runner, step Step, pipeline Pipeline, durations *sync.Map, wg *sync.WaitGroup, preconditions []chan struct{}, o chan struct{}, abort chan struct{}) {
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
	// If an error was encountered previusly, skip the rest
	select {
	case <-abort:
		pipelineLogger.Printf("- Skipping %s: an error occurred previously", step.ColoredContainerName())
		return
	default:
	}
	// Kill old container if KeepAlive_Replace
	count, err := runner.ContainerKiller(step)()
	if err != nil {
		pipelineLogger.Printf("Error killing %s: %s", step.ColoredName(), err)
	}
	if count > 0 {
		pipelineLogger.Printf("- Killed: %s", step.ColoredContainerName())
	}
	if err := runner.ContainerRemover(step)(); err != nil {
		pipelineLogger.Printf("Error removing %s: %s", step.ColoredName(), err)
	}
	pipelineLogger.Printf("- Starting: %s", step.ColoredContainerName())
	duration, err := executeF(runner.ContainerRunner(step, pipeline.Network))
	if err != nil {
		pipelineLogger.Printf("  %s: %s", step.ColoredContainerName(), err)
		if !step.Meta.IgnoreFailure {
			select {
			case <-abort:
			default:
				close(abort)
			}
		} else {
			pipelineLogger.Printf("  Ignoring error of: %s", step.ColoredContainerName())
		}
	}
	pipelineLogger.Printf("- Finished %s after %s", step.ColoredContainerName(), duration)
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
	abort := make(chan struct{})
	runChannel := make(chan struct{})
	channels := make(map[string]chan struct{})
	for _, pipeline := range *pipelines {
		for _, step := range pipeline {
			channels[step.Name] = make(chan struct{})
			preChannels := make([]chan struct{}, 0)
			preChannels = append(preChannels, runChannel)
			for pre := range step.Dependencies() {
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
			go runParallelStep(p.GetRunnerForMeta(step.Meta), step, p, durations, &wg, preChannels, channels[step.Name], abort)
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
