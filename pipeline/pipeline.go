package pipeline // import "github.com/ad-freiburg/gantry/pipeline"

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/ghodss/yaml"
)

type Pipeline struct {
	Definition  PipelineDefinition
	Environment PipelineEnvironment
}

type PipelineDefinition struct {
	Steps []Step `json:"steps"`
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
	yaml.Unmarshal(data, &p.Definition)
	return nil
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
	yaml.Unmarshal(data, &p.Environment)
	return nil
}

func (p *Pipeline) Check() error {
	roleProvider := make(map[string][]Machine)
	for _, machine := range p.Environment.Machines {
		for role, _ := range machine.Roles {
			roleProvider[role] = append(roleProvider[role], machine)
		}
	}
	for _, step := range p.Definition.Steps {
		if len(roleProvider[step.Role]) < 1 {
			return fmt.Errorf("No machine for role '%s'", step.Role)
		}
	}
	return nil
}
