package gantry // import "github.com/ad-freiburg/gantry"

import (
	"encoding/json"
)

// ServiceList stores docker-compose service definitions as steps.
type ServiceList map[string]Step

// UnmarshalJSON sets *r to a copy of data.
func (r *ServiceList) UnmarshalJSON(data []byte) error {
	parsedJSON := make(map[string]Service)
	steps := make(map[string]Step)
	err := json.Unmarshal(data, &parsedJSON)
	if err != nil {
		return err
	}
	for name, step := range parsedJSON {
		step.Name = name
		step.InitColor()
		step.Meta = ServiceMeta{
			Type: ServiceTypeService,
		}
		steps[name] = Step{
			Service: step,
			Detach:  true,
		}
	}
	*r = steps
	return nil
}

// StepList stores gantry steps as steps.
type StepList map[string]Step

// UnmarshalJSON sets *r to a copy of data.
func (r *StepList) UnmarshalJSON(data []byte) error {
	parsedJSON := make(map[string]Step)
	err := json.Unmarshal(data, &parsedJSON)
	if err != nil {
		return err
	}
	for name, step := range parsedJSON {
		step.Name = name
		step.InitColor()
		step.Meta = ServiceMeta{
			Type: ServiceTypeStep,
		}
		parsedJSON[name] = step
	}
	*r = parsedJSON
	return nil
}
