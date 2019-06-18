package gantry

import (
	"bufio"
	"bytes"
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

// PipelineEnvironment stores additional data for pipelines and steps.
type PipelineEnvironment struct {
	Version     string                  `json:"version"`
	Machines    []Machine               `json:"machines"`
	LogSettings LogSettings             `json:"log"`
	Environment types.MappingWithEquals `json:"environment"`
	TempDirPath string                  `json:"tempdir"`
	Services    ServiceMetaList         `json:"services"`
	Steps       ServiceMetaList         `json:"steps"`
	tempFiles   []string
	tempPaths   map[string]string
}

// NewPipelineEnvironment builds a new environment merging the current
// environment, the environment given by path and the user provided steps to
// ignore.
func NewPipelineEnvironment(path string, environment types.MappingWithEquals, ignoredSteps types.StringSet) (*PipelineEnvironment, error) {
	// Set defaults
	e := &PipelineEnvironment{
		tempPaths:   make(map[string]string, 0),
		Environment: types.MappingWithEquals{},
	}
	e.updateEnvironment(environment)
	e.updateIgnoredSteps(ignoredSteps)

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
	// TODO(lehmann): apply current environment to .env.yml ?
	e.Services = nil
	e.Steps = nil
	err = yaml.Unmarshal(data, e)
	if err != nil {
		return e, err
	}
	// Reimport defaults
	e.updateEnvironment(environment)
	e.updateIgnoredSteps(ignoredSteps)
	return e, nil
}

func (e *PipelineEnvironment) updateEnvironment(environment types.MappingWithEquals) {
	for k, v := range environment {
		e.Environment[k] = v
	}
}

func (e *PipelineEnvironment) updateIgnoredSteps(ignoredSteps types.StringSet) {
	if e.Services == nil {
		e.Services = ServiceMetaList{}
	}
	if e.Steps == nil {
		e.Steps = ServiceMetaList{}
	}
	// Update defined steps and serives
	for name, stepMeta := range e.Steps {
		if val, ignored := ignoredSteps[name]; ignored {
			stepMeta.Ignore = val
			e.Steps[name] = stepMeta
		}
	}
	for name, stepMeta := range e.Services {
		if val, ignored := ignoredSteps[name]; ignored {
			stepMeta.Ignore = val
			e.Steps[name] = stepMeta
		}
	}
	for name, val := range ignoredSteps {
		stepMeta := ServiceMeta{Ignore: val}
		if _, found := e.Steps[name]; !found {
			e.Steps[name] = stepMeta
		}
		if _, found := e.Services[name]; !found {
			e.Services[name] = stepMeta
		}
	}
}

func (e *PipelineEnvironment) exportEnvironment(path string) error {
	data, err := yaml.Marshal(e)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(path, data, 0644)
}

func (e *PipelineEnvironment) createTemplateParser() *template.Template {
	fm := template.FuncMap{}
	// {{ Key }}
	// Required environment value, if not defined it will not be found as function and raise an error
	for k, v := range e.Environment {
		// Convert string pointer to static string
		fm[k] = func(v string) func() (string, error) {
			return func() (string, error) {
				return v, nil
			}
		}(*v)
	}
	// {{ Env "Key" ["Default"] }}
	// Usable as optional environment variable, can provide default value if not defined
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
		val, ok := e.Environment[parts[0]]
		if !ok {
			if len(parts) < 2 {
				return "", errors.New(fmt.Sprintf("Env '%s' not defined, no fallback provided", parts[0]))
			}
			return parts[1], nil
		}
		return *val, nil
	}
	// {{ EnvDir "Key" ["Default"] }}
	// Get Path from environment, converts to absolute path using filepath.Abs
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
		val, ok := e.Environment[parts[0]]
		if ok {
			path = *val
		} else {
			if len(parts) < 2 {
				return "", errors.New(fmt.Sprintf("EnvDir '%s' not defined, no fallback provided", parts[0]))
			}
			path = parts[1]
		}
		return filepath.Abs(path)
	}
	// {{ TempDir ["optional" ["optional" ... ]] }}
	// TempDir with the same arguments point to the same directory
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
func (e *PipelineEnvironment) ApplyTo(data []byte) ([]byte, error) {
	var b bytes.Buffer
	bw := bufio.NewWriter(&b)
	t, err := e.createTemplateParser().Parse(string(data))
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
