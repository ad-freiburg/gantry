package gantry_test

import (
	"testing"

	"github.com/ad-freiburg/gantry"
)

func TestNewTarjan(t *testing.T) {
	// Diamond
	stepA := gantry.Step{Service: gantry.Service{Name: "a"}}
	stepB := gantry.Step{Service: gantry.Service{Name: "b"}, After: map[string]bool{"a": true}}
	stepC := gantry.Step{Service: gantry.Service{Name: "c"}, After: map[string]bool{"a": true}}
	stepD := gantry.Step{Service: gantry.Service{Name: "d"}, After: map[string]bool{"b": true, "c": true}}
	// Cycle
	stepE := gantry.Step{Service: gantry.Service{Name: "e"}, After: map[string]bool{"g": true}}
	stepF := gantry.Step{Service: gantry.Service{Name: "f"}, After: map[string]bool{"e": true}}
	stepG := gantry.Step{Service: gantry.Service{Name: "g"}, After: map[string]bool{"f": true}}

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
		{
			map[string]gantry.Step{
				"e": stepE,
				"f": stepF,
				"g": stepG,
			}, gantry.Pipelines{
				[]gantry.Step{stepE, stepF, stepG},
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
		for i := range result {
			if len(result[i]) != len(c.result[i]) {
				t.Errorf("Incorrect length for '%d'@'%d', got %d, wanted %d", ci, i, len(result[i]), len(c.result[i]))
			}
			// Tarjan is not deterministic on each level, thus no further comparisons
		}
	}
}

func TestNewTarjanMissingDependency(t *testing.T) {
	stepB := gantry.Step{Service: gantry.Service{Name: "b"}, After: map[string]bool{"a": true}}

	input := map[string]gantry.Step{"b": stepB}
	_, err := gantry.NewTarjan(input)
	if err == nil {
		t.Errorf("Got no error for: '%#v'", input)
	}
}
