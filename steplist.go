package gantry // import "github.com/ad-freiburg/gantry"

import (
	"encoding/json"
)

type ServiceList map[string]Step

func (l *ServiceList) UnmarshalJSON(data []byte) error {
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
	*l = stepStorage
	return nil
}

type StepList map[string]Step

func (l *StepList) UnmarshalJSON(data []byte) error {
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
	*l = storage
	return nil
}
