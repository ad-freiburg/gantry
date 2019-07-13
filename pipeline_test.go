package gantry_test

import (
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/ad-freiburg/gantry"
	"github.com/ad-freiburg/gantry/types"
)

func TestPipelinesAllSteps(t *testing.T) {
	cases := []struct {
		input  gantry.Pipelines
		result []gantry.Step
	}{
		{gantry.Pipelines{}, []gantry.Step{}},
		{gantry.Pipelines{[]gantry.Step{{}}}, []gantry.Step{{}}},
	}

	for _, c := range cases {
		r := c.input.AllSteps()
		if !reflect.DeepEqual(r, c.result) {
			t.Errorf("Incorrect result for '%v', got: '%#v', wanted '%#v'", c.input, r, c.result)
		}
	}
}

func TestPipelinesCheck(t *testing.T) {
	stepA := gantry.Step{Service: gantry.Service{Name: "a"}, After: map[string]bool{"b": true}}
	stepB := gantry.Step{Service: gantry.Service{Name: "b"}, After: map[string]bool{"a": true}}
	stepA2 := gantry.Step{Service: gantry.Service{Name: "a"}}
	stepC := gantry.Step{Service: gantry.Service{Name: "c"}, After: map[string]bool{"b": true, "a": true}}

	cases := []struct {
		input  gantry.Pipelines
		result string
	}{
		{gantry.Pipelines{}, ""},
		{gantry.Pipelines{[]gantry.Step{stepA, stepB}}, "cyclic component found in (sub)pipeline: '%s'"},
		{[][]gantry.Step{{stepA2}, {stepB}, {stepC}}, ""},
	}

	for _, c := range cases {
		r := c.input.Check()
		if (r == nil && c.result != "") || (r != nil && c.result == "") {
			t.Errorf("Incorrect result for '%v', got: '%s', wanted '%s'", c.input, r, c.result)
		}
	}
}

func TestPipelineDefinitionPipelines(t *testing.T) {
	cases := []struct {
		definition gantry.PipelineDefinition
		err        string
		result     gantry.Pipelines
	}{
		{gantry.PipelineDefinition{}, "", gantry.Pipelines{}},
		{gantry.PipelineDefinition{Steps: gantry.StepList{"a": gantry.Step{}}}, "", gantry.Pipelines{[]gantry.Step{{}}}},
	}

	for _, c := range cases {
		r, err := c.definition.Pipelines()
		if (err == nil && c.err != "") || (err != nil && c.err == "") {
			t.Errorf("Incorrect error for '%v', got: '%s', wanted '%s'", c.definition, err, c.err)
		}
		if err != nil {
			continue
		}
		for i, ri := range *r {
			if len(ri) != len(c.result[i]) {
				t.Errorf("Incorrect length for '%v'@'%d', got: '%d', wanted: '%d'", c.definition, i, len(ri), len(c.result[i]))
			}
		}
	}
}

func TestPipelineIgnoreStepsFromMetaAndArgument(t *testing.T) {
	tmpDef, err := ioutil.TempFile("", "def")
	if err != nil {
		log.Fatal(err)
	}
	defer os.Remove(tmpDef.Name())
	diamond, err := ioutil.ReadFile(filepath.Join(".", "examples", "diamond", "gantry.def.yml"))
	if err != nil {
		log.Fatal(err)
	}
	err = ioutil.WriteFile(tmpDef.Name(), diamond, 0644)
	if err != nil {
		log.Fatal(err)
	}
	tmpEnvWithoutIgnore, err := ioutil.TempFile("", "envWithoutIgnore")
	if err != nil {
		log.Fatal(err)
	}
	defer os.Remove(tmpEnvWithoutIgnore.Name())
	err = ioutil.WriteFile(tmpEnvWithoutIgnore.Name(), []byte(`steps:
  b:
    stdout:
      handler: discard
`), 0644)
	if err != nil {
		log.Fatal(err)
	}
	tmpEnvWithIgnore, err := ioutil.TempFile("", "envWithIgnore")
	if err != nil {
		log.Fatal(err)
	}
	defer os.Remove(tmpEnvWithIgnore.Name())
	err = ioutil.WriteFile(tmpEnvWithIgnore.Name(), []byte(`steps:
  b:
    stdout:
      handler: discard
    ignore: yes
`), 0644)
	if err != nil {
		log.Fatal(err)
	}

	cases := []struct {
		def         string
		env         string
		environment types.MappingWithEquals
		ignore      types.StringSet
		selected    types.StringSet
		err         string
		numIgnore   int
	}{
		{tmpDef.Name(), tmpEnvWithoutIgnore.Name(), types.MappingWithEquals{}, types.StringSet{}, types.StringSet{}, "", 0},
		{tmpDef.Name(), tmpEnvWithoutIgnore.Name(), types.MappingWithEquals{}, types.StringSet{"a": true}, types.StringSet{}, "", 1},
		{tmpDef.Name(), tmpEnvWithIgnore.Name(), types.MappingWithEquals{}, types.StringSet{}, types.StringSet{}, "", 1},
		{tmpDef.Name(), tmpEnvWithIgnore.Name(), types.MappingWithEquals{}, types.StringSet{"a": true}, types.StringSet{}, "", 2},
	}

	for _, c := range cases {
		r, err := gantry.NewPipeline(c.def, c.env, c.environment, c.ignore, c.selected)
		if (err == nil && c.err != "") || (err != nil && c.err == "") {
			t.Errorf("Incorrect error for '%v','%v','%v',%v', got: '%s', wanted '%s'", c.def, c.env, c.environment, c.ignore, err, c.err)
		}
		if err != nil {
			continue
		}
		pipelines, err := r.Definition.Pipelines()
		if err != nil {
			t.Error(err)
		}
		ignoreCount := 0
		for _, step := range pipelines.AllSteps() {
			if step.Meta.Ignore {
				ignoreCount += 1
			}
		}
		if ignoreCount != c.numIgnore {
			t.Errorf("Incorrect number of ignored steps for '%v','%v','%v', got: '%d', wanted '%d'", c.def, c.env, c.ignore, ignoreCount, c.numIgnore)
		}
	}
}
