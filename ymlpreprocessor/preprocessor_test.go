package ymlpreprocessor_test

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"testing"

	"github.com/ad-freiburg/gantry"
	"github.com/ad-freiburg/gantry/types"
	"github.com/ad-freiburg/gantry/ymlpreprocessor"
)

const barValue string = "Bar"

func TestPreprocessorRegister(t *testing.T) {
	p, err := ymlpreprocessor.NewPreprocessor()
	if err != nil {
		t.Error(err)
		return
	}
	smallF1 := &ymlpreprocessor.Function{
		Names: []string{"f1"},
	}
	bigF1 := &ymlpreprocessor.Function{
		Names: []string{"F1"},
	}
	bigF2andBigF1 := &ymlpreprocessor.Function{
		Names: []string{"F2", "F1"},
	}
	if err := p.Register(smallF1); err != nil {
		t.Errorf("unexpected error: %s", err)
	}
	if err := p.Register(bigF1); err != nil {
		t.Errorf("unexpected error: %s", err)
	}
	if err := p.Register(bigF2andBigF1); err == nil {
		t.Errorf("expected error, got nil")
	}
}

func TestPreprocessorProcess(t *testing.T) {
	bar := barValue
	cases := []struct {
		in            string
		out           string
		env           string
		substitutions types.StringMap
	}{
		{
			"${NOT_DECLARED}",
			"",
			"",
			types.StringMap{},
		},
		{
			"${NOT_DECLARED}",
			"",
			`substitutions:
  Foo: Baz`,
			types.StringMap{},
		},
		{
			"${EMPTY}",
			"",
			"",
			types.StringMap{"EMPTY": nil},
		},
		{
			"${Foo}",
			barValue,
			"",
			types.StringMap{"Foo": &bar},
		},
		{
			"${Foo}",
			"Baz",
			`substitutions:
  Foo: Baz`,
			types.StringMap{},
		},
		{
			"${Foo}",
			barValue,
			`substitutions:
  Foo: Baz`,
			types.StringMap{"Foo": &bar},
		},
		{
			`${Foo}
${EMPTY}`,
			fmt.Sprintf("%s\n", barValue),
			`substitutions:
  Foo: Baz`,
			types.StringMap{"Foo": &bar},
		},
	}
	preprocessor, err := ymlpreprocessor.NewPreprocessor()
	if err != nil {
		t.Error(err)
		return
	}
	for _, c := range cases {
		path := ""
		if len(c.env) > 0 {
			tmpEnv, err := ioutil.TempFile("", "envWithoutIgnore")
			if err != nil {
				log.Fatal(err)
			}
			defer os.Remove(tmpEnv.Name())
			err = ioutil.WriteFile(tmpEnv.Name(), []byte(c.env), 0644)
			if err != nil {
				log.Fatal(err)
			}
			path = tmpEnv.Name()
		}
		e, err := gantry.NewPipelineEnvironment(path, c.substitutions, types.StringSet{}, types.StringSet{})
		if err != nil {
			if os.IsNotExist(err) && len(c.env) < 1 {
				// No env provided, error is expected
			} else {
				log.Fatal(err)
			}
		}
		resBytes, err := preprocessor.Process([]byte(c.in), e)
		if err != nil {
			t.Error(err)
		}
		resString := string(resBytes)
		if resString != c.out {
			t.Errorf("incorrect transformation of '%s': got: '%s', wanted: '%s'", c.in, resString, c.out)
		}
	}
}

func TestPreprocessorProcessErrors(t *testing.T) {
	cases := []struct {
		in            string
		out           string
		err           string
		env           string
		substitutions types.StringMap
	}{
		{
			"#! UNKNOWN",
			"",
			"unknown preprocessor directive: 'UNKNOWN'",
			"",
			types.StringMap{},
		},
		{
			"#! DEFECTIVE_CHECK",
			"",
			"missing argument(s) in DEFECTIVE_CHECK for , wanted: 1, got: 0",
			"",
			types.StringMap{},
		},
	}
	preprocessor, err := ymlpreprocessor.NewPreprocessor()
	if err != nil {
		t.Error(err)
		return
	}
	err = preprocessor.Register(&ymlpreprocessor.Function{
		Names:      []string{"DEFECTIVE_CHECK"},
		NumArgsMin: 1,
	})
	if err != nil {
		t.Error(err)
		return
	}
	for _, c := range cases {
		path := ""
		if len(c.env) > 0 {
			tmpEnv, err := ioutil.TempFile("", "envWithoutIgnore")
			if err != nil {
				log.Fatal(err)
			}
			defer os.Remove(tmpEnv.Name())
			err = ioutil.WriteFile(tmpEnv.Name(), []byte(c.env), 0644)
			if err != nil {
				log.Fatal(err)
			}
			path = tmpEnv.Name()
		}
		e, err := gantry.NewPipelineEnvironment(path, c.substitutions, types.StringSet{}, types.StringSet{})
		if err != nil {
			if os.IsNotExist(err) && len(c.env) < 1 {
				// No env provided, error is expected
			} else {
				log.Fatal(err)
			}
		}
		resBytes, err := preprocessor.Process([]byte(c.in), e)
		if err == nil {
			t.Error("expected error, got nil")
		}
		if err.Error() != c.err {
			t.Errorf("incorrect error for '%s': got: '%s', wanted: '%s'", c.in, err.Error(), c.err)
		}
		resString := string(resBytes)
		if resString != "" {
			t.Errorf("incorrect result for '%s': got: '%s', wanted: '%s'", c.in, resString, c.out)
		}
	}
}

func TestPreprocessorFunctions(t *testing.T) {
	preprocessor, err := ymlpreprocessor.NewPreprocessor()
	if err != nil {
		t.Error(err)
		return
	}
	numFunctions := len(preprocessor.Functions())
	err = preprocessor.Register(&ymlpreprocessor.Function{
		Names:      []string{"DEFECTIVE_CHECK"},
		NumArgsMin: 1,
	})
	if err != nil {
		t.Error(err)
		return
	}
	updatedNumFunctions := len(preprocessor.Functions())
	if numFunctions+1 != updatedNumFunctions {
		t.Errorf("incorrect number of functions, got: %d, wanted: %d", updatedNumFunctions, numFunctions+1)
	}
}
