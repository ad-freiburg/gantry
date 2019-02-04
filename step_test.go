package gantry_test

import (
	"reflect"
	"testing"

	"github.com/ad-freiburg/gantry"
	"github.com/ad-freiburg/gantry/types"
)

func TestStepDependencies(t *testing.T) {
	stepA := gantry.Step{}
	stepA.SetName("a")
	stepB := gantry.Step{}
	stepB.SetName("b")
	stepB.After = map[string]bool{"a": true}
	stepC := gantry.Step{}
	stepC.SetName("c")
	stepC.DependsOn = map[string]bool{"a": true}
	stepD := gantry.Step{}
	stepD.SetName("d")
	stepD.After = map[string]bool{"c": true}
	stepD.DependsOn = map[string]bool{"d": true}

	cases := []struct {
		step   gantry.Step
		result *types.StringSet
	}{
		{stepA, &types.StringSet{}},
		{stepB, &types.StringSet{"a": true}},
		{stepC, &types.StringSet{"a": true}},
		{stepD, &types.StringSet{"c": true, "d": true}},
	}

	for _, c := range cases {
		r, err := c.step.Dependencies()
		if err != nil {
			t.Errorf("%v for %v", err, c.step)
		}
		if !reflect.DeepEqual(r, c.result) {
			t.Errorf("Incorrect result for '%v', got: %#v, wanted %#v", c.step, r, c.result)
		}
	}
}

func TestStepImageName(t *testing.T) {
	stepA := gantry.Step{Service: gantry.Service{}}
	stepA.SetName("a Step")
	stepB := gantry.Step{Service: gantry.Service{Image: "b"}}
	stepC := gantry.Step{Service: gantry.Service{Image: "c"}}
	stepC.SetName("c Step")

	cases := []struct {
		step   gantry.Step
		result string
	}{
		{stepA, "a_step"},
		{stepB, "b"},
		{stepC, "c"},
	}

	for _, c := range cases {
		r := c.step.ImageName()
		if r != c.result {
			t.Errorf("Incorrect result for '%v', got: %s, wanted %s", c.step, r, c.result)
		}
	}
}
