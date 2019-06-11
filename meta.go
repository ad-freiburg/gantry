package gantry // import "github.com/ad-freiburg/gantry"

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

const (
	KeepAlive_No ServiceKeepAlive = iota
	KeepAlive_Yes
	KeepAlive_Replace

	Log_Stdout ServiceLogDestination = iota
	Log_File
	Log_Both
	Log_Discard
)

type ServiceMetaList map[string]ServiceMeta

// UnmarshalJSON sets *r to a copy of data.
func (r *ServiceMetaList) UnmarshalJSON(data []byte) error {
	storage := make(map[string]ServiceMeta, 0)
	err := json.Unmarshal(data, &storage)
	if err != nil {
		return err
	}
	for name, meta := range storage {
		err := meta.Check()
		if err != nil {
			return errors.New(fmt.Sprintf("Error in '%s': %s", name, err))
		}
	}
	*r = storage
	return nil
}

type ServiceMeta struct {
	KeepRunning   ServiceKeepAlive      `json:"keep-running"`
	Log           ServiceLogDestination `json:"log"`
	LogFileStderr string                `json:"log-file-stderr"`
	LogFileStdout string                `json:"log-file-stdout"`
}

func (m ServiceMeta) Check() error {
	if m.Log != Log_Stdout && m.Log != Log_Discard {
		if m.LogFileStderr == "" {
			return errors.New("Missing 'log-file-stderr'")
		}
		if m.LogFileStdout == "" {
			return errors.New("Missing 'log-file-stdout'")
		}
	}
	return nil
}

type ServiceKeepAlive int

func (d *ServiceKeepAlive) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	switch strings.ToLower(s) {
	default:
		*d = KeepAlive_No
	case "yes":
		*d = KeepAlive_Yes
	case "replace":
		*d = KeepAlive_Replace
	}
	return nil
}

type ServiceLogDestination int

func (d *ServiceLogDestination) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	switch strings.ToLower(s) {
	default:
		*d = Log_Stdout
	case "file":
		*d = Log_File
	case "both":
		*d = Log_Both
	case "discard":
		*d = Log_Discard
	}
	return nil
}
