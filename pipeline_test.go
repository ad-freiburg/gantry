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
	}

	for _, c := range cases {
		r := c.input.AllSteps()
		if !reflect.DeepEqual(r, c.result) {
			t.Errorf("Incorrect result for '%v', got: %#v, wanted %#v", c.input, r, c.result)
		}
	}
}

func TestPipelinesCheck(t *testing.T) {
	cases := []struct {
		input  gantry.Pipelines
		result error
	}{
		{gantry.Pipelines{}, nil},
	}

	for _, c := range cases {
		r := c.input.Check()
		if r != c.result {
			t.Errorf("Incorrect result for '%v', got: %#v, wanted %#v", c.input, r, c.result)
		}
	}
}
