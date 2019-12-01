package gantry

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/ad-freiburg/gantry/types"
	"github.com/ghodss/yaml"
)

const undefinedArgumentFormat string = "argmuent '%s' not defined for '%s', no fallback provided"
const missingArgumentFormat string = "missing argument(s) for '%s'. Need atleast %d argument"
const tooManyArgumentsFormat string = "too many arguments for '%s'. Got %d want <= %d"

type pipelineEnvironmentJSON struct {
	Version            string          `json:"version"`
	Substitutions      types.StringMap `json:"substitutions"`
	TempDirPath        string          `json:"tempdir"`
	TempDirNoAutoClean bool            `json:"tempdir_no_autoclean"`
	Services           ServiceMetaList `json:"services"`
	Steps              ServiceMetaList `json:"steps"`
	ProjectName        string          `json:"project_name"`
}

// PipelineEnvironment stores additional data for pipelines and steps.
type PipelineEnvironment struct {
	Version            string
	Substitutions      types.StringMap
	TempDirPath        string
	TempDirNoAutoClean bool
	Steps              ServiceMetaList
	ProjectName        string
	tempFiles          []string
	tempPaths          map[string]string
}

// UnmarshalJSON loads a PipelineDefinition from json using the pipelineJSON struct.
func (e *PipelineEnvironment) UnmarshalJSON(data []byte) error {
	result := PipelineEnvironment{
		Steps:     ServiceMetaList{},
		tempFiles: []string{},
		tempPaths: map[string]string{},
	}
	parsedJSON := pipelineEnvironmentJSON{}
	if err := json.Unmarshal(data, &parsedJSON); err != nil {
		return err
	}
	result.Version = parsedJSON.Version
	result.Substitutions = parsedJSON.Substitutions
	result.TempDirPath = parsedJSON.TempDirPath
	result.TempDirNoAutoClean = parsedJSON.TempDirNoAutoClean
	result.ProjectName = parsedJSON.ProjectName
	if result.Substitutions == nil {
		result.Substitutions = types.StringMap{}
	}
	for name, meta := range parsedJSON.Services {
		meta.Type = ServiceTypeService
		result.Steps[name] = meta
	}
	for name, meta := range parsedJSON.Steps {
		if _, found := result.Steps[name]; found {
			return fmt.Errorf("duplicate step/service '%s'", name)
		}
		meta.Type = ServiceTypeStep
		meta.KeepAlive = KeepAliveNo
		result.Steps[name] = meta
	}
	*e = result
	return nil
}

// NewPipelineEnvironment builds a new environment merging the current
// environment, the environment given by path and the user provided steps to
// ignore.
func NewPipelineEnvironment(path string, substitutions types.StringMap, ignoredSteps types.StringSet, selectedSteps types.StringSet) (*PipelineEnvironment, error) {
	// Set defaults
	e := &PipelineEnvironment{
		tempPaths:     make(map[string]string),
		Substitutions: types.StringMap{},
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

func (e *PipelineEnvironment) updateSubstitutions(substitutions types.StringMap) {
	for k, v := range substitutions {
		e.Substitutions[k] = v
	}
}

func (e *PipelineEnvironment) GetSubstitution(key string) (*string, bool) {
	value, ok := e.Substitutions[key]
	return value, ok
}

func (e *PipelineEnvironment) SetSubstitution(key string, value *string) {
	e.Substitutions[key] = value
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

// CleanUp tries to remove all managed temporary files and directories.
func (e *PipelineEnvironment) CleanUp(signal os.Signal) error {
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
	return nil
}

func (e *PipelineEnvironment) GetOrCreateTempDir(prefix string) (string, error) {
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
