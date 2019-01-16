package gantry // import "github.com/ad-freiburg/gantry"

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"time"

	"github.com/ghodss/yaml"
)

type Pipeline struct {
	Definition  PipelineDefinition
	Environment PipelineEnvironment
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
	p := &Pipeline{}
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
		log.Println("Could not open pipeline definition.")
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

func (p Pipeline) PrepareImages(force bool) error {
	pipelines, err := p.Definition.Pipelines()
	if err != nil {
		return err
	}
	log.Printf("Prepare Images:")
	images := 0
	start := time.Now()
	durations := make(map[string]time.Duration)

	for _, step := range pipelines.AllSteps() {
		log.Printf("- %s", step.Name)
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
				return err
			}
			duration += duration2
		}
		durations[step.Name] = duration
		images++
	}

	log.Printf("Prepared %d images in %s", images, time.Since(start))
	var totalElapsedTime time.Duration = 0
	for _, duration := range durations {
		totalElapsedTime += duration
	}
	log.Printf("Total time spend preparing images: %s", totalElapsedTime)
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

func (p Pipeline) ExecuteSteps() error {
	pipelines, err := p.Definition.Pipelines()
	if err != nil {
		return err
	}
	steps := 0
	start := time.Now()
	durations := make(map[string]time.Duration)
	for _, pipeline := range *pipelines {
		for _, step := range pipeline {
			log.Printf("- Starting: %s", step.Name)
			duration, err := executeF(NewContainerRunner(step))
			log.Printf("- Finished %s after %s", step.Name, duration)
			if err != nil {
				return err
			}
			durations[step.Name] = duration
			steps++
		}
	}
	log.Printf("Executed %d steps in %s", steps, time.Since(start))
	var totalElapsedTime time.Duration = 0
	for _, duration := range durations {
		totalElapsedTime += duration
	}
	log.Printf("Total time spend inside steps: %s", totalElapsedTime)
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
	r := NewLocalRunner(s.ContainerName(), os.Stdout, os.Stderr)
	return r
}
