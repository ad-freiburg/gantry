package gantry // import "github.com/ad-freiburg/gantry"

import (
	"encoding/json"
	"strings"
)

type ServiceKeepAlive int

const (
	KeepAlive_No ServiceKeepAlive = iota
	KeepAlive_Yes
	KeepAlive_Replace
)

func (k *ServiceKeepAlive) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	switch strings.ToLower(s) {
	default:
		*k = KeepAlive_No
	case "yes":
		*k = KeepAlive_Yes
	case "replace":
		*k = KeepAlive_Replace
	}
	return nil
}

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

type ServiceMetaList map[string]ServiceMeta

// UnmarshalJSON sets *r to a copy of data.
func (r *ServiceMetaList) UnmarshalJSON(data []byte) error {
	storage := make(map[string]ServiceMeta, 0)
	err := json.Unmarshal(data, &storage)
	if err != nil {
		return err
	}
	*r = storage
	return nil
}

type ServiceMeta struct {
	KeepRunning ServiceKeepAlive `json:"keep-running"`
}
