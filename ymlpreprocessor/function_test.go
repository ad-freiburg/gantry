package ymlpreprocessor_test

import (
	"testing"

	"github.com/ad-freiburg/gantry/ymlpreprocessor"
)

func TestFunctionCheck(t *testing.T) {
	f := ymlpreprocessor.Function{}
	i := ymlpreprocessor.Instruction{}
	if err := f.Check(i); err != nil {
		t.Errorf("unexpected error: %s", err)
	}

	f.NeedsVariable = true
	if err := f.Check(i); err == nil {
		t.Errorf("expected error, got nil")
	}

	i.Variable = "var"
	if err := f.Check(i); err != nil {
		t.Errorf("unexpected error: %s", err)
	}

	f.NumArgsMin = 1
	f.NumArgsMax = 99
	if err := f.Check(i); err == nil {
		t.Errorf("expected error, got nil")
	}

	i.Arguments = []string{"arg0", "arg1"}
	if err := f.Check(i); err != nil {
		t.Errorf("unexpected error: %s", err)
	}

	f.NumArgsMax = 1
	if err := f.Check(i); err == nil {
		t.Errorf("expected error, got nil")
	}

	i.Arguments = []string{"arg0"}
	if err := f.Check(i); err != nil {
		t.Errorf("unexpected error: %s", err)
	}
}

func TestFunctionUsage(t *testing.T) {
	cases := []struct {
		function ymlpreprocessor.Function
		result   string
	}{
		{
			ymlpreprocessor.Function{},
			"",
		},
		{
			ymlpreprocessor.Function{
				Names: []string{"FOO"},
			},
			"#! FOO",
		},
		{
			ymlpreprocessor.Function{
				Names: []string{"FOO", "bar"},
			},
			`#! FOO
#! bar`,
		},
		{
			ymlpreprocessor.Function{
				Names:         []string{"FOO"},
				NeedsVariable: true,
			},
			"#! FOO ${VAR}",
		},
		{
			ymlpreprocessor.Function{
				Names:         []string{"FOO"},
				NeedsVariable: true,
				NumArgsMin:    1,
				NumArgsMax:    1,
			},
			"#! FOO ${VAR} ARG0",
		},
		{
			ymlpreprocessor.Function{
				Names:         []string{"FOO"},
				NeedsVariable: true,
				NumArgsMin:    1,
				NumArgsMax:    2,
			},
			"#! FOO ${VAR} ARG0 [ ARG1 ]",
		},
		{
			ymlpreprocessor.Function{
				Names:         []string{"FOO"},
				NeedsVariable: true,
				NumArgsMin:    0,
				NumArgsMax:    2,
			},
			"#! FOO ${VAR} [ ARG0 ARG1 ]",
		},
		{
			ymlpreprocessor.Function{
				Names:         []string{"FOO"},
				NeedsVariable: true,
				NumArgsMin:    2,
				NumArgsMax:    2,
			},
			"#! FOO ${VAR} ARG0 ARG1",
		},
		{
			ymlpreprocessor.Function{
				Names:       []string{"FOO"},
				Description: "Lorem ipsum dolor.",
			},
			`#! FOO
Lorem ipsum dolor.`,
		},
	}

	for i, c := range cases {
		r := c.function.Usage()
		if r != c.result {
			t.Errorf("incorrect result @%d, got: '%s', wanted: '%s'", i, r, c.result)
		}
	}
}
