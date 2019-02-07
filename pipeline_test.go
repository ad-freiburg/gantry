package gantry_test

import (
	"reflect"
	"testing"

	"github.com/ad-freiburg/gantry"
)

func TestPipelinesAllSteps(t *testing.T) {
	cases := []struct {
		input  gantry.Pipelines
		result []gantry.Step
	}{
		{gantry.Pipelines{}, []gantry.Step{}},
		{gantry.Pipelines{[]gantry.Step{gantry.Step{}}}, []gantry.Step{gantry.Step{}}},
	}

	for _, c := range cases {
		r := c.input.AllSteps()
		if !reflect.DeepEqual(r, c.result) {
			t.Errorf("Incorrect result for '%v', got: %#v, wanted %#v", c.input, r, c.result)
		}
	}
}

func TestPipelinesCheck(t *testing.T) {
	stepA := gantry.Step{}
	stepA.SetName("a")
	stepA.After = map[string]bool{"b": true}
	stepB := gantry.Step{}
	stepB.SetName("b")
	stepB.After = map[string]bool{"a": true}
	stepA2 := gantry.Step{}
	stepA2.SetName("a")
	stepC := gantry.Step{}
	stepC.SetName("c")
	stepC.After = map[string]bool{"b": true, "a": true}

	cases := []struct {
		input  gantry.Pipelines
		result string
	}{
		{gantry.Pipelines{}, ""},
		{gantry.Pipelines{[]gantry.Step{stepA, stepB}}, "cyclic component found in (sub)pipeline: '%s'"},
		{[][]gantry.Step{[]gantry.Step{stepA2}, []gantry.Step{stepB}, []gantry.Step{stepC}}, ""},
	}

	for _, c := range cases {
		r := c.input.Check()
		if (r == nil && c.result != "") || (r != nil && c.result == "") {
			t.Errorf("Incorrect result for '%v', got: %s, wanted %s", c.input, r, c.result)
		}
	}
}
