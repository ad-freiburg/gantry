package gantry // import "github.com/ad-freiburg/gantry"

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
)

const (
	// KeepAliveYes signals that the service will keep running after gantry exits.
	KeepAliveYes ServiceKeepAlive = iota
	// KeepAliveNo signals that the service will be killed when gantry stops.
	KeepAliveNo
	// KeepAliveReplace signals that the service is killed prior to replacement.
	KeepAliveReplace
)
const (
	// LogHandlerStdout signals that the standard location (stdout or stderr) is used.
	LogHandlerStdout ServiceLogHandler = iota
	// LogHandlerFile signals that the log shall be stored in a file.
	LogHandlerFile
	// LogHandlerBoth signals that the log is printed into a file and to the standard location.
	LogHandlerBoth
	// LogHandlerDiscard signals that the log shall be discarded.
	LogHandlerDiscard
)
const (
	// ServiceTypeService signals a normal docker service.
	ServiceTypeService ServiceType = iota
	// ServiceTypeStep signals a gantry step.
	ServiceTypeStep
)

// ServiceMetaList stores ServiceMeta as a map[stepname]Meta.
type ServiceMetaList map[string]ServiceMeta

// ServiceMeta stores all metainformation for a step.
type ServiceMeta struct {
	Ignore        bool             `json:"ignore"`
	IgnoreFailure bool             `json:"ignore_failure"`
	KeepAlive     ServiceKeepAlive `json:"keep_alive"`
	Stdout        ServiceLog       `json:"stdout"`
	Stderr        ServiceLog       `json:"stderr"`
	Selected      bool
	Type          ServiceType
}

// Open handles output initialisation by setting defaults.
func (m *ServiceMeta) Open() error {
	if err := m.Stdout.Open(os.Stdout); err != nil {
		return err
	}
	if err := m.Stderr.Open(os.Stderr); err != nil {
		return err
	}
	return nil
}

// Close closes stderr and stdout writers.
func (m *ServiceMeta) Close() {
	m.Stdout.Close()
	m.Stderr.Close()
}

// ServiceType stores the type of the service.
type ServiceType int

// ServiceKeepAlive stores the KeepAlive state of the service.
type ServiceKeepAlive int

// UnmarshalJSON sets ServiceKeepAlive d.
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

// ServiceLogHandler stores the logging configuration of the service.
type ServiceLogHandler int

// UnmarshalJSON sets ServiceLogHandler d.
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

// ServiceLog stores log configuration for an output stream.
type ServiceLog struct {
	Handler ServiceLogHandler `json:"handler"`
	Path    string            `json:"path"`
	std     *os.File
	file    *os.File
}

// Open handles output initialisation by setting defaults and creating files.
func (l *ServiceLog) Open(std *os.File) error {
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

// Write writes data to an output stream or discards if LogHandlerDiscard is configured.
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

// Close closes the output file if one is used.
func (l *ServiceLog) Close() {
	if l.file != nil {
		l.file.Close()
	}
}
