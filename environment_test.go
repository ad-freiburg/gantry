package gantry_test

import (
	"io/ioutil"
	"log"
	"os"
	"testing"

	"github.com/ad-freiburg/gantry"
	"github.com/ad-freiburg/gantry/types"
)

const barValue string = "Bar"

func TestPipelineEnvironmentApplyTo(t *testing.T) {
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
		resBytes, err := e.ApplyTo([]byte(c.in))
		if err != nil {
			t.Error(err)
		}
		resString := string(resBytes)
		if resString != c.out {
			t.Errorf("Incorrect transformation of '%s': got: '%s', wanted: '%s'", c.in, resString, c.out)
		}
	}
}
