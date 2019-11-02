package gantry

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

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

func (e *PipelineEnvironment) preprocess(lines []string) ([]string, error) {
	var newLines []string
	for i, line := range lines {
		// Only return non comment lines
		if len(line) < 1 {
			continue
		}
		if line[0] != '#' {
			newLines = append(newLines, line)
		}
		// If comment does not start with a shebang (#!) ignore it
		if line[1] != '!' {
			continue
		}
		parts := strings.Split(string(line[3:]), " ")
		log.Print(parts)
		if len(parts) < 2 {
			log.Printf("INVALID DIRECTIVE: %s in line %d: %s\n", parts[0], i+1, line)
			continue
		}
		key := parts[1][2 : len(parts[1])-1]
		val, ok := e.Substitutions[key]
		switch parts[0] {
		case "CHECK_IF_DIR_EXISTS":
			// If environment variable is empty or not set, abort!
			if !ok || val == nil || len(*val) < 1 {
				return []string{}, fmt.Errorf("empty variable for CHECK_IF_DIR_EXISTS: %s", key)
			}
			path, err := filepath.Abs(*val)
			if err != nil {
				return []string{}, fmt.Errorf("path error for CHECK_IF_DIR_EXISTS %s: err: %s", key, err)
			}
			if _, err := os.Stat(path); os.IsNotExist(err) {
				return []string{}, fmt.Errorf("path error for CHECK_IF_DIR_EXISTS %s: err: %s", key, err)
			}
		case "SET_IF_EMTPY":
			if len(parts) < 3 {
				return []string{}, fmt.Errorf("invalid syntax SET_IF_EMPTY %s: missing argument", key)
			}
			// If environment variable set and not empty, do not create it!
			log.Print(*e.Substitutions[key])
			if ok && val != nil && len(*val) > 0 {
				continue
			}
			e.Substitutions[key] = &parts[2]
		case "TEMP_DIR":
			// If environment variable for temp dir is set, do not create it!
			if ok && val != nil && len(*val) > 0 {
				continue
			}
			// We have an empty value or a new variable: create the directory.
			path, err := e.getOrCreateTempDir(key)
			if err != nil {
				return []string{}, err
			}
			e.Substitutions[key] = &path
		default:
			log.Printf("UNKNOWN DIRECTIVE: %s in line %d: %s\n", parts[0], i+1, line)
		}
	}
	return newLines, nil
}

func (e *PipelineEnvironment) expandVariablesInLine(line string) string {
	expandFunc := func(placeholder string) string {
		val, ok := e.Substitutions[placeholder]
		if ok && val != nil {
			return *val
		}
		// No real substitution found, return empty string
		return ""
	}
	return os.Expand(line, expandFunc)
}

func (e *PipelineEnvironment) expandVariables(lines []string) []string {
	var newLines []string
	for _, line := range lines {
		newLines = append(newLines, e.expandVariablesInLine(line))
	}
	return newLines
}

// ApplyTo executes the environment template parser on the provided data.
func (e *PipelineEnvironment) ApplyTo(rawFile []byte) ([]byte, error) {
	// Parse bytes as lines
	var lines []string
	var lineBytesBuffer bytes.Buffer
	r := bufio.NewReader(bytes.NewReader(rawFile))
	for {
		lineBytes, prefix, err := r.ReadLine()
		if err != nil {
			if err == io.EOF {
				break
			}
			return []byte(""), err
		}
		lineBytesBuffer.Write(lineBytes)
		// Line continues, continue reading before storing
		if prefix {
			continue
		}
		lines = append(lines, lineBytesBuffer.String())
		lineBytesBuffer.Reset()
	}
	// Run preprocessor steps
	lines, err := e.preprocess(lines)
	if err != nil {
		return []byte(""), err
	}
	lines = e.expandVariables(lines)
	// Reconvert to byte slice
	var b bytes.Buffer
	bw := bufio.NewWriter(&b)
	for i, line := range lines {
		if i > 0 {
			if _, err := bw.WriteString("\n"); err != nil {
				return []byte(""), err
			}
		}
		if _, err := bw.WriteString(line); err != nil {
			return []byte(""), err
		}
	}
	bw.Flush()
	return b.Bytes(), nil
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
