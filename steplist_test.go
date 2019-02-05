package gantry_test

import (
	"testing"

	"github.com/ad-freiburg/gantry"
)

func TestServiceListUnmarshalJSON(t *testing.T) {
	cases := []struct {
		json   string
		err    string
		result gantry.ServiceList
	}{
		{"{", "unexpected end of JSON input", gantry.ServiceList{}},
		{"{}", "", gantry.ServiceList{}},
		{"{\"a\": {}}", "", gantry.ServiceList{"a": gantry.Step{}}},
	}

	for _, c := range cases {
		r := gantry.ServiceList{}
		err := r.UnmarshalJSON([]byte(c.json))
		if (err != nil && c.err == "") || (err == nil && c.err != "") {
			t.Errorf("Incorrect Error for '%s', got: '%s', wanted '%s'", c.json, err.Error(), c.err)
		}
		if len(r) != len(c.result) {
			t.Errorf("Incorrect list length for '%s', got: '%d', wanted: '%d'", c.json, len(r), len(c.result))
		}
	}
}

func TestStepListUnmarshalJSON(t *testing.T) {
	cases := []struct {
		json   string
		err    string
		result gantry.StepList
	}{
		{"{", "unexpected end of JSON input", gantry.StepList{}},
		{"{}", "", gantry.StepList{}},
		{"{\"a\": {}}", "", gantry.StepList{"a": gantry.Step{}}},
	}

	for _, c := range cases {
		r := gantry.StepList{}
		err := r.UnmarshalJSON([]byte(c.json))
		if (err != nil && c.err == "") || (err == nil && c.err != "") {
			t.Errorf("Incorrect Error for %s, got: %s, wanted %s", c.json, err.Error(), c.err)
		}
		if len(r) != len(c.result) {
			t.Errorf("Incorrect list length for '%s', got: '%d', wanted: '%d'", c.json, len(r), len(c.result))
		}
	}
}
