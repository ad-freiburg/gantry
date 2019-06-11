package gantry // import "github.com/ad-freiburg/gantry"

import (
	"encoding/json"
)

// ServiceList stores docker-compose service definitions as steps.
type ServiceList map[string]Step

// UnmarshalJSON sets *r to a copy of data.
func (r *ServiceList) UnmarshalJSON(data []byte) error {
	serviceStorage := make(map[string]Service, 0)
	stepStorage := make(map[string]Step, 0)
	err := json.Unmarshal(data, &serviceStorage)
	if err != nil {
		return err
	}
	for name, step := range serviceStorage {
		step.Name = name
		step.InitColor()
		stepStorage[name] = Step{
			Service: step,
			Detach:  true,
		}
	}
	*r = stepStorage
	return nil
}

// StepList stores gantry steps as steps.
type StepList map[string]Step

// UnmarshalJSON sets *r to a copy of data.
func (r *StepList) UnmarshalJSON(data []byte) error {
	storage := make(map[string]Step, 0)
	err := json.Unmarshal(data, &storage)
	if err != nil {
		return err
	}
	for name, step := range storage {
		step.Name = name
		step.InitColor()
		storage[name] = step
	}
	*r = storage
	return nil
}
