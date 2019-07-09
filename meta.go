package gantry // import "github.com/ad-freiburg/gantry"

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	KeepAliveYes ServiceKeepAlive = iota
	KeepAliveNo
	KeepAliveReplace
)
const (
	LogHandlerStdout ServiceLogHandler = iota
	LogHandlerFile
	LogHandlerBoth
	LogHandlerDiscard
)
const (
	ServiceTypeService ServiceType = iota
	ServiceTypeStep
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
		err := meta.Init()
		if err != nil {
			return errors.New(fmt.Sprintf("Error in '%s': %s", name, err))
		}
		storage[name] = meta
	}
	*r = storage
	return nil
}

type ServiceMeta struct {
	Ignore        bool             `json:"ignore"`
	IgnoreFailure bool             `json:"ignore_failure"`
	KeepAlive     ServiceKeepAlive `json:"keep_alive"`
	Stdout        ServiceLog       `json:"stdout"`
	Stderr        ServiceLog       `json:"stderr"`
	Selected      bool
	Type          ServiceType
}

// Init handles initialisation by setting defaults.
func (m *ServiceMeta) Init() error {
	if err := m.Stdout.Init(os.Stdout); err != nil {
		return err
	}
	if err := m.Stderr.Init(os.Stderr); err != nil {
		return err
	}
	return nil
}

func (m *ServiceMeta) Close() {
	m.Stdout.Close()
	m.Stderr.Close()
}

type ServiceType int

type ServiceKeepAlive int

// UnmarshalJSON sets *r to a copy of data.
func (d *ServiceKeepAlive) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	switch strings.ToLower(s) {
	default:
		*d = KeepAliveYes
	case "no":
		*d = KeepAliveNo
	case "replace":
		*d = KeepAliveReplace
	}
	return nil
}

type ServiceLogHandler int

func (d *ServiceLogHandler) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	switch strings.ToLower(s) {
	default:
		*d = LogHandlerStdout
	case "file":
		*d = LogHandlerFile
	case "both":
		*d = LogHandlerBoth
	case "discard":
		*d = LogHandlerDiscard
	}
	return nil
}

type ServiceLog struct {
	Handler ServiceLogHandler `json:"handler"`
	Path    string            `json:"path"`
	std     *os.File
	file    *os.File
}

// Init handles initialisation by setting defaults and creating files.
func (l *ServiceLog) Init(std *os.File) error {
	l.std = std
	if l.Handler != LogHandlerStdout && l.Handler != LogHandlerDiscard {
		if l.Path == "" {
			return errors.New("Missing 'path'")
		}
		p, err := filepath.Abs(l.Path)
		if err != nil {
			return err
		}
		f, err := os.Create(p)
		if err != nil {
			return err
		}
		l.file = f
	}
	return nil
}

func (l ServiceLog) Write(p []byte) (int, error) {
	var n1, n2 int
	var err1, err2 error
	if l.Handler == LogHandlerStdout || l.Handler == LogHandlerBoth {
		n1, err1 = l.std.Write(p)
	}
	if l.Handler == LogHandlerFile || l.Handler == LogHandlerBoth {
		n2, err2 = l.file.Write(p)
	}
	if err1 != nil {
		return n1, err1
	}
	if err2 != nil {
		return n2, err2
	}
	return len(p), nil
}

func (l *ServiceLog) Close() {
	if l.file != nil {
		l.file.Close()
	}
}
