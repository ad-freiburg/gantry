package gantry

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/ad-freiburg/gantry/types"
	"github.com/ghodss/yaml"
)

type pipelineEnvironmentJson struct {
	Version            string                  `json:"version"`
	Machines           []Machine               `json:"machines"`
	LogSettings        LogSettings             `json:"log"`
	Substitutions      types.MappingWithEquals `json:"substitutions"`
	TempDirPath        string                  `json:"tempdir"`
	TempDirNoAutoClean bool                    `json:"tempdir_no_autoclean"`
	Services           ServiceMetaList         `json:"services"`
	Steps              ServiceMetaList         `json:"steps"`
}

// PipelineEnvironment stores additional data for pipelines and steps.
type PipelineEnvironment struct {
	Version            string
	Machines           []Machine
	LogSettings        LogSettings
	Substitutions      types.MappingWithEquals
	TempDirPath        string
	TempDirNoAutoClean bool
	Steps              ServiceMetaList
	tempFiles          []string
	tempPaths          map[string]string
}

// UnmarshalJSON loads a PipelineDefinition from json using the pipelineJson struct.
func (r *PipelineEnvironment) UnmarshalJSON(data []byte) error {
	result := PipelineEnvironment{
		Steps:     ServiceMetaList{},
		tempFiles: []string{},
		tempPaths: map[string]string{},
	}
	storage := pipelineEnvironmentJson{}
	if err := json.Unmarshal(data, &storage); err != nil {
		return err
	}
	result.Version = storage.Version
	result.Machines = storage.Machines
	result.LogSettings = storage.LogSettings
	result.Substitutions = storage.Substitutions
	result.TempDirPath = storage.TempDirPath
	result.TempDirNoAutoClean = storage.TempDirNoAutoClean
	for name, meta := range storage.Services {
		meta.Type = ServiceTypeService
		result.Steps[name] = meta
	}
	for name, meta := range storage.Steps {
		if _, found := result.Steps[name]; found {
			return fmt.Errorf("Duplicate step/service '%s'", name)
		}
		meta.Type = ServiceTypeStep
		meta.KeepAlive = KeepAliveNo
		result.Steps[name] = meta
	}
	*r = result
	return nil
}

// NewPipelineEnvironment builds a new environment merging the current
// environment, the environment given by path and the user provided steps to
// ignore.
func NewPipelineEnvironment(path string, substitutions types.MappingWithEquals, ignoredSteps types.StringSet, selectedSteps types.StringSet) (*PipelineEnvironment, error) {
	// Set defaults
	e := &PipelineEnvironment{
		tempPaths:     make(map[string]string, 0),
		Substitutions: types.MappingWithEquals{},
		Steps:         ServiceMetaList{},
	}
	e.updateSubstitutions(substitutions)
	e.updateStepsMeta(ignoredSteps, selectedSteps)

	// Import settings from file
	dir, err := os.Getwd()
	if err != nil {
		return e, err
	}
	defaultPath := filepath.Join(dir, GantryEnv)
	if _, err := os.Stat(defaultPath); path == "" && err == nil {
		path = defaultPath
	}
	file, err := os.Open(path)
	if err != nil {
		return e, err
	}
	defer file.Close()

	data, err := ioutil.ReadAll(file)
	if err != nil {
		return e, err
	}
	e.Steps = nil
	err = yaml.Unmarshal(data, e)
	if err != nil {
		return e, err
	}
	// Reimport defaults
	e.updateSubstitutions(substitutions)
	e.updateStepsMeta(ignoredSteps, selectedSteps)
	return e, nil
}

func (e *PipelineEnvironment) updateSubstitutions(substitutions types.MappingWithEquals) {
	for k, v := range substitutions {
		e.Substitutions[k] = v
	}
}

func (e *PipelineEnvironment) updateStepsMeta(ignoredSteps types.StringSet, selectedSteps types.StringSet) {
	for name := range ignoredSteps {
		if _, found := e.Steps[name]; !found {
			e.Steps[name] = ServiceMeta{}
		}
	}
	for name := range selectedSteps {
		if _, found := e.Steps[name]; !found {
			e.Steps[name] = ServiceMeta{}
		}
	}
	// Update defined steps and serives
	for name, stepMeta := range e.Steps {
		if val, ignored := ignoredSteps[name]; ignored {
			stepMeta.Ignore = val
		}
		if val, selected := selectedSteps[name]; selected {
			stepMeta.Selected = val
		}
		e.Steps[name] = stepMeta
	}
}

func (e *PipelineEnvironment) exportSubstitutions(path string) error {
	data, err := yaml.Marshal(e)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(path, data, 0644)
}

func (e *PipelineEnvironment) createTemplateParser() *template.Template {
	fm := template.FuncMap{}

	// {{ Key }}
	// Required substitution value, if not defined it will not be found as
	// function and raise an error.
	for k, v := range e.Substitutions {
		if v == nil {
			// If no explicit value is set, return the empty string.
			fm[k] = func() string {
				return ""
			}
		} else {
			// Ensure that each key uses it's own function, as otherwise all
			// keys would report the last defined value.
			fm[k] = func(v string) func() string {
				return func() string {
					return v
				}
			}(*v)

		}
	}

	// {{ Env "Key" ["Default"] }}
	// Usable as optional environment variable, can provide default value if not defined.
	fm["Env"] = func(args ...interface{}) (string, error) {
		if len(args) < 1 {
			return "", errors.New(fmt.Sprintf("Env: missing argument(s). Need atleast 1 argument"))
		}
		if len(args) > 2 {
			return "", errors.New(fmt.Sprintf("Env: too many arguments. Got %d want <=2", len(args)))
		}
		parts := make([]string, len(args))
		for i, v := range args {
			parts[i] = fmt.Sprint(v)
		}
		val, ok := e.Substitutions[parts[0]]
		if !ok {
			if len(parts) < 2 {
				return "", errors.New(fmt.Sprintf("Env '%s' not defined, no fallback provided", parts[0]))
			}
			return parts[1], nil
		}
		return *val, nil
	}

	// {{ EnvDir "Key" ["Default"] }}
	// Get Path from environment, converts to absolute path using filepath.Abs.
	fm["EnvDir"] = func(args ...interface{}) (string, error) {
		if len(args) < 1 {
			return "", errors.New(fmt.Sprintf("EnvDir: missing argument(s). Need atleast 1 argument"))
		}
		if len(args) > 2 {
			return "", errors.New(fmt.Sprintf("EnvDir: too many arguments. Got %d want <=2", len(args)))
		}
		parts := make([]string, len(args))
		for i, v := range args {
			parts[i] = fmt.Sprint(v)
		}
		var path string
		val, ok := e.Substitutions[parts[0]]
		if ok {
			path = *val
		} else {
			if len(parts) < 2 {
				return "", errors.New(fmt.Sprintf("EnvDir '%s' not defined, no fallback provided", parts[0]))
			}
			path = parts[1]
		}
		path, err := filepath.Abs(path)
		if err != nil {
			return path, err
		}
		if _, err := os.Stat(path); os.IsNotExist(err) {
			return path, err
		}
		return path, nil
	}

	// {{ TempDir ["optional" ["optional" ... ]] }}
	// Calls to TempDir with equivalent arguments result in the same directory.
	// This allows to share temporary directories between steps/services.
	fm["TempDir"] = func(args ...interface{}) (string, error) {
		parts := make([]string, len(args))
		for i, v := range args {
			parts[i] = fmt.Sprint(v)
		}
		return e.getOrCreateTempDir(strings.Join(parts, "_"))
	}
	return template.New("PipelineEnvironment").Funcs(fm)
}

// ApplyTo executes the environment template parser on the provided data.
func (e *PipelineEnvironment) ApplyTo(rawFile []byte) ([]byte, error) {
	var b bytes.Buffer
	bw := bufio.NewWriter(&b)
	t, err := e.createTemplateParser().Parse(string(rawFile))
	if err != nil {
		return []byte(""), err
	}
	err = t.Execute(bw, e)
	bw.Flush()
	return b.Bytes(), err
}

// CleanUp tries to remove all managed temporary files and directories.
func (e *PipelineEnvironment) CleanUp(signal os.Signal) {
	for _, file := range e.tempFiles {
		if err := os.Remove(file); err != nil {
			log.Print(err)
		}
	}
	for _, path := range e.tempPaths {
		if err := os.RemoveAll(path); err != nil {
			log.Print(err)
		}
	}
}

func (e *PipelineEnvironment) getOrCreateTempDir(prefix string) (string, error) {
	val, ok := e.tempPaths[prefix]
	if ok {
		return val, nil
	}
	return e.tempDir(prefix)
}

func (e *PipelineEnvironment) tempDir(prefix string) (string, error) {
	path, err := ioutil.TempDir(e.TempDirPath, prefix)
	if err == nil {
		e.tempPaths[prefix] = path
	}
	return path, os.Chmod(path, 0777)
}

func (e *PipelineEnvironment) tempFile(pattern string) (*os.File, error) {
	file, err := ioutil.TempFile(e.TempDirPath, pattern)
	if err == nil {
		e.tempFiles = append(e.tempFiles, file.Name())
	}
	return file, err
}
