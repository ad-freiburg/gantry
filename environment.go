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

// PipelineEnvironment stores additional data for pipelines.
type PipelineEnvironment struct {
	Machines      []Machine               `json:"machines"`
	LogSettings   LogSettings             `json:"log"`
	Environment   types.MappingWithEquals `json:"environment"`
	TempDirPath   string                  `json:"tempdir"`
	tempFiles     []string
	tempPaths     map[string]string
	envFilePath   string
	envFileLoaded bool
}

func NewPipelineEnvironment(path string) *PipelineEnvironment {
	e := &PipelineEnvironment{
		tempPaths:   make(map[string]string, 0),
		envFilePath: path,
	}
	e.importCurrentEnv()
	return e
}

func (e *PipelineEnvironment) load(path string) error {
	if _, err := os.Stat(GantryEnv); path == "" && os.IsExist(err) {
		path = GantryEnv
	}
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	data, err := ioutil.ReadAll(file)
	if err != nil {
		return err
	}
	err = yaml.Unmarshal(data, e)
	if err != nil {
		return err
	}
	e.envFileLoaded = true
	e.importCurrentEnv()
	return nil
}

func (e *PipelineEnvironment) importCurrentEnv() {
	// Import current environment
	if e.Environment == nil {
		e.Environment = types.MappingWithEquals{}
	}
	for _, pair := range os.Environ() {
		parts := strings.SplitN(pair, "=", 2)
		if old, exists := e.Environment[parts[0]]; exists && *old != parts[1] {
			log.Printf("Replacing Environment '%s': '%s' with '%s'", parts[0], old, parts[1])
		}
		e.Environment[parts[0]] = &parts[1]
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
	for k, v := range e.Environment {
		fm[k] = func() string { return *v }
	}
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
	fm["TempDir"] = func(args ...interface{}) (string, error) {
		parts := make([]string, len(args))
		for i, v := range args {
			parts[i] = fmt.Sprint(v)
		}
		return e.getOrCreateTempDir(strings.Join(parts, "_"))
	}
	return template.New("PipelineEnvironment").Funcs(fm)
}

func (e *PipelineEnvironment) ApplyTo(data []byte) ([]byte, error) {
	var b bytes.Buffer
	bw := bufio.NewWriter(&b)
	t, err := e.createTemplateParser().Parse(string(data))
	if err != nil {
		return []byte(""), err
	}
	err = t.Execute(bw, e)
	bw.Flush()
	if err != nil && !e.envFileLoaded {
		if e.load(e.envFilePath) == nil {
			return e.ApplyTo(data)
		}
	}
	return b.Bytes(), err
}

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

func (e *PipelineEnvironment) TempFile(pattern string) (*os.File, error) {
	file, err := ioutil.TempFile(e.TempDirPath, pattern)
	if err == nil {
		e.tempFiles = append(e.tempFiles, file.Name())
	}
	return file, err
}
