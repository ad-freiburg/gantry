package gantry

import (
	"testing"

	"github.com/ad-freiburg/gantry/types"
)

func TestPipelineDefinitionPipelines(t *testing.T) {
	cases := []struct {
		definition PipelineDefinition
		err        string
		result     Pipelines
	}{
		{PipelineDefinition{ignoredSteps: types.StringSet{"a": true}, Steps: StepList{"b": Step{Service: Service{Name: "b"}, After: map[string]bool{"a": true}}}, Services: ServiceList{"c": Step{Service: Service{Name: "c", DependsOn: map[string]bool{"a": true}}}}}, "", [][]Step{{{}}, {{}}}},
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
