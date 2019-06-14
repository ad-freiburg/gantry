package gantry // import "github.com/ad-freiburg/gantry"

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
)

const (
	KeepAlive_Yes ServiceKeepAlive = iota
	KeepAlive_No
	KeepAlive_Replace

	Log_Stdout ServiceLogHandler = iota
	Log_File
	Log_Both
	Log_Discard
)

type ServiceMetaList map[string]ServiceMeta

// UnmarshalJSON sets *r to a copy of data.
func (r *ServiceMetaList) UnmarshalJSON(data []byte) error {
	fmt.Printf("\n%s\n\n", string(data))
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
	Ignore    bool             `json:"ignore"`
	KeepAlive ServiceKeepAlive `json:"keep-alive"`
	Stdout    ServiceLog       `json:"stdout"`
	Stderr    ServiceLog       `json:"stderr"`
}

func (m *ServiceMeta) Init() error {
	if m.KeepAlive == 0 {
		m.KeepAlive = KeepAlive_Yes
	}
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

type ServiceKeepAlive int

func (d *ServiceKeepAlive) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	switch strings.ToLower(s) {
	default:
		*d = KeepAlive_Yes
	case "no":
		*d = KeepAlive_No
	case "replace":
		*d = KeepAlive_Replace
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

type ServiceLog struct {
	Handler ServiceLogHandler `json:"handler"`
	Path    string            `json:"path"`
	std     *os.File
	file    *os.File
}

func (l *ServiceLog) Init(std *os.File) error {
	l.std = std
	if l.Handler == 0 {
		l.Handler = Log_Stdout
	}
	if l.Handler != Log_Stdout && l.Handler != Log_Discard {
		if l.Path == "" {
			return errors.New("Missing 'path'")
		}
		p, err := filepath.Abs(l.Path)
		if err != nil {
			log.Fatal(err)
		}
		f, err := os.Create(p)
		if err != nil {
			log.Fatal(err)
		}
		l.file = f
	}
	return nil
}

func (l ServiceLog) Write(p []byte) (int, error) {
	var n1, n2 int
	var err1, err2 error
	if l.Handler == Log_Stdout || l.Handler == Log_Both {
		n1, err1 = l.std.Write(p)
	}
	if l.Handler == Log_File || l.Handler == Log_Both {
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
