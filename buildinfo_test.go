package gantry_test

import (
	"reflect"
	"testing"

	"github.com/ad-freiburg/gantry"
)

func TestBuildInfoUnmarshalJSON(t *testing.T) {
	cases := []struct {
		json   string
		result gantry.BuildInfo
	}{
		{"{}", gantry.BuildInfo{}},
		{"{\"context\": \"test\"}", gantry.BuildInfo{Context: "test"}},
		{"{\"dockerfile\": \"file\"}", gantry.BuildInfo{Dockerfile: "file"}},
		{"{\"args\": [\"key1=val1\", \"key2\"]}", gantry.BuildInfo{Args: map[string]string{"key1": "val1", "key2": ""}}},
		{"{\"args\": {\"key1\": \"val1\", \"key2\": \"\"}}", gantry.BuildInfo{Args: map[string]string{"key1": "val1", "key2": ""}}},
	}

	for _, c := range cases {
		b := gantry.BuildInfo{}
		err := b.UnmarshalJSON([]byte(c.json))
		if err != nil {
			t.Errorf("%v for %s", err, c.json)
		}
		if b.Context != c.result.Context {
			t.Errorf("Incorrect Context for '%s', got: %s, wanted %s", c.json, b.Context, c.result.Context)
		}
		if b.Dockerfile != c.result.Dockerfile {
			t.Errorf("Incorrect Dockerfile for '%s', got: %s, wanted %s", c.json, b.Dockerfile, c.result.Dockerfile)
		}
		if !reflect.DeepEqual(b.Args, c.result.Args) {
			t.Errorf("Incorrect Args for '%s', got: %#v, wanted %#v", c.json, b.Args, c.result.Args)
		}
	}
}
