package gantry

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"text/template"

	"github.com/ad-freiburg/gantry/types"
	"github.com/ghodss/yaml"
)

// PipelineEnvironment stores additional data for pipelines.
type PipelineEnvironment struct {
	Machines    []Machine               `json:"machines"`
	LogSettings LogSettings             `json:"log"`
	Environment types.MappingWithEquals `json:"environment"`
	TempDirPath string                  `json:"tempdir"`
	tempFiles   []string
	tempPaths   map[string]string
}

func NewPipelineEnvironment() *PipelineEnvironment {
	e := &PipelineEnvironment{}
	e.tempPaths = make(map[string]string, 0)
	return e
}

func (e *PipelineEnvironment) Load(path string) error {
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
	return yaml.Unmarshal(data, e)
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
	fm["TempDir"] = func(args ...interface{}) (string, error) {
		parts := make([]string, len(args))
		for i, v := range args {
			parts[i] = fmt.Sprint(v)
		}
		return e.getOrCreateTempDir(strings.Join(parts, "_"))
	}
	return template.New("PipelineEnvironment").Funcs(fm)
}

func (e *PipelineEnvironment) ApplyTo(def *PipelineDefinition) error {
	templateParser := e.createTemplateParser()
	if err := e.applyToVolumes(def, templateParser); err != nil {
		return err
	}
	return nil
}

func (e *PipelineEnvironment) applyToVolumes(def *PipelineDefinition, tp *template.Template) error {
	pipelines, err := def.Pipelines()
	if err != nil {
		return err
	}
	for _, s := range pipelines.AllSteps() {
		for i, volumePath := range s.Volumes {
			var b bytes.Buffer
			bw := bufio.NewWriter(&b)
			t, err := tp.Parse(volumePath)
			if err != nil {
				return err
			}
			err = t.Execute(bw, nil)
			if err != nil {
				return err
			}
			bw.Flush()
			s.Volumes[i] = b.String()
		}
	}
	return nil
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
