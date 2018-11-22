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

func (p PipelineDefinition) StepList() (*StepList, error) {
	steps := make(map[string]Step, 0)
	for _, pipeline := range p.Steps {
		for _, step := range pipeline {
			if val, ok := steps[step.Name]; ok {
				return nil, fmt.Errorf("Redeclaration of step '%s'", val.Name)
			}
			steps[step.Name] = step
		}
	}
	for _, pipeline := range p.Services {
		for _, step := range pipeline {
			if val, ok := steps[step.Name]; ok {
				return nil, fmt.Errorf("Redeclaration of step '%s'", val.Name)
			}
			steps[step.Name] = step
		}
	}

	t, err := NewTarjan(steps)
	if err != nil {
		return nil, err
	}
	res, err := t.Parse()
	result := &StepList{}
	*result = *res
	return result, err
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
	err = p.setPipelineEnvironment(environmentPath)
	if err != nil {
		return nil, err
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

	steplist, err := p.Definition.StepList()
	if err != nil {
		return err
	}
	for _, step := range steplist.All() {
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
	steplist, err := p.Definition.StepList()
	if err != nil {
		return err
	}
	for _, step := range steplist.All() {
		fmt.Printf("\n Prepare Image: %s\n", step.Name)
		existence := NewImageExistenceChecker(step)
		err := existence.Exec()
		if err == nil {
			continue
		}

		if step.Image != "" {
			r := NewImagePuller(step)
			err := r.Exec()
			if err == nil {
				continue
			}
		}
		r := NewImageBuilder(step)
		err = r.Exec()
		if err != nil {
			return err
		}
	}
	return nil
}

func (p Pipeline) ExecuteSteps() error {
	steplist, err := p.Definition.StepList()
	if err != nil {
		return err
	}
	for _, steps := range *steplist {
		for _, step := range steps {
			fmt.Printf("\n Starting: %s\n", step.Name)
			r := NewImageRunner(step)
			err := r.Exec()
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

func (s Step) Dependencies() (StringSet, error) {
	return StringSet{}, nil
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
