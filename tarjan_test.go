package gantry_test

import (
	"testing"

	"github.com/ad-freiburg/gantry"
)

func TestNewTarjan(t *testing.T) {
	stepA := gantry.Step{}
	stepA.SetName("a")
	stepB := gantry.Step{}
	stepB.SetName("b")
	stepB.After = map[string]bool{"a": true}
	stepC := gantry.Step{}
	stepC.SetName("c")
	stepC.After = map[string]bool{"a": true}
	stepD := gantry.Step{}
	stepD.SetName("d")
	stepD.After = map[string]bool{"b": true, "c": true}
	cases := []struct {
		input  map[string]gantry.Step
		result gantry.Pipelines
	}{
		// Test diamond
		{
			map[string]gantry.Step{},
			gantry.Pipelines{},
		},
		{
			map[string]gantry.Step{
				"a": stepA,
			},
			gantry.Pipelines{
				[]gantry.Step{stepA},
			},
		},
		{
			map[string]gantry.Step{
				"a": stepA,
				"b": stepB,
			},
			gantry.Pipelines{
				[]gantry.Step{stepA},
				[]gantry.Step{stepB},
			},
		},
		{
			map[string]gantry.Step{
				"a": stepA,
				"b": stepB,
				"c": stepC,
			},
			gantry.Pipelines{
				[]gantry.Step{stepA},
				[]gantry.Step{stepB},
				[]gantry.Step{stepC},
			},
		},
		{
			map[string]gantry.Step{
				"a": stepA,
				"b": stepB,
				"c": stepC,
				"d": stepD,
			}, gantry.Pipelines{
				[]gantry.Step{stepA},
				[]gantry.Step{stepB},
				[]gantry.Step{stepC},
				[]gantry.Step{stepD},
			},
		},
	}
	for ci, c := range cases {
		r, err := gantry.NewTarjan(c.input)
		result := *r
		if err != nil {
			t.Errorf("Got error: %v", err)
		}
		if len(result) != len(c.result) {
			t.Errorf("Incorrect length for '%d', got %d, wanted %d", ci, len(result), len(c.result))
		}
		for i, _ := range result {
			if len(result[i]) != len(c.result[i]) {
				t.Errorf("Incorrect length for '%d'@'%d', got %d, wanted %d", ci, i, len(result[i]), len(c.result[i]))
			}
			// Tarjan is not deterministic on each level, thus no further comparisons
		}
	}
}
