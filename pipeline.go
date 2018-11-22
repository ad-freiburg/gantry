package gantry // import "github.com/ad-freiburg/gantry"

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/ghodss/yaml"
)

type Pipeline struct {
	Definition  PipelineDefinition
	Environment PipelineEnvironment
}

type PipelineDefinition struct {
	Steps    StepList    `json:"steps"`
	Services ServiceList `json:"services"`
}

func (p PipelineDefinition) Pipelines() (*pipelines, error) {
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
	return &result, nil
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
	file, err := os.Open(path)
	if err != nil {
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

func (p Pipeline) PrepareImages() error {
	pipelines, err := p.Definition.Pipelines()
	if err != nil {
		return err
	}
	for _, step := range pipelines.AllSteps() {
		fmt.Printf("\n Prepare Image: %s\n", step.Name)
		err := NewImageExistenceChecker(step)()
		if err == nil {
			continue
		}

		if step.Image != "" {
			err := NewImagePuller(step)()
			if err == nil {
				continue
			}
		}
		err = NewImageBuilder(step)()
		if err != nil {
			return err
		}
	}
	return nil
}

func (p Pipeline) ExecuteSteps() error {
	pipelines, err := p.Definition.Pipelines()
	if err != nil {
		return err
	}
	for _, pipeline := range *pipelines {
		for _, step := range pipeline {
			fmt.Printf("\n Starting: %s\n", step.Name)
			NewContainerKiller(step)()
			NewOldContainerRemover(step)()
			err := NewContainerRunner(step)()
			if err != nil {
				return err
			}
		}
	}
	return nil
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
	r := NewLocalRunner()
	return r
}
