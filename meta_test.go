package gantry_test

import (
	"encoding/json"
	"testing"

	"github.com/ad-freiburg/gantry"
)

func TestMetaServiceKeepAlive(t *testing.T) {
	cases := []struct {
		input  string
		result gantry.ServiceKeepAlive
	}{
		{`""`, gantry.KeepAliveYes},
		{`"yes"`, gantry.KeepAliveYes},
		{`"no"`, gantry.KeepAliveNo},
		{`"replace"`, gantry.KeepAliveReplace},
		{`"iUseTheDefault"`, gantry.KeepAliveYes},
	}

	for _, c := range cases {
		var r gantry.ServiceKeepAlive
		if err := json.Unmarshal([]byte(c.input), &r); err != nil {
			t.Error(err)
		}
		if r != c.result {
			t.Errorf("Incorrect ServiceKeepAlive for '%s', got: '%d', wanted: '%d'", c.input, r, c.result)
		}
	}
}

func TestMetaServiceLogHandler(t *testing.T) {
	cases := []struct {
		input  string
		result gantry.ServiceLogHandler
	}{
		{`""`, gantry.LogHandlerStdout},
		{`"file"`, gantry.LogHandlerFile},
		{`"both"`, gantry.LogHandlerBoth},
		{`"discard"`, gantry.LogHandlerDiscard},
		{`"iUseTheDefault"`, gantry.LogHandlerStdout},
	}

	for _, c := range cases {
		var r gantry.ServiceLogHandler
		if err := json.Unmarshal([]byte(c.input), &r); err != nil {
			t.Error(err)
		}
		if r != c.result {
			t.Errorf("Incorrect ServiceLogHandler for '%s', got: '%d', wanted: '%d'", c.input, r, c.result)
		}
	}
}

func TestMetaServiceLog(t *testing.T) {
	cases := []struct {
		input  string
		result gantry.ServiceLog
	}{
		{`{}`, gantry.ServiceLog{Handler: 0, Path: ""}},
		{`{"handler": "both", "path": "/some/path"}`, gantry.ServiceLog{Handler: gantry.LogHandlerBoth, Path: "/some/path"}},
		{`{"handler": "discard"}`, gantry.ServiceLog{Handler: gantry.LogHandlerDiscard, Path: ""}},
		{`{"path": "/some/path"}`, gantry.ServiceLog{Handler: 0, Path: "/some/path"}},
		{`{"handler": "iUseTheDefault"}`, gantry.ServiceLog{Handler: gantry.LogHandlerStdout, Path: ""}},
	}

	for _, c := range cases {
		var r gantry.ServiceLog
		if err := json.Unmarshal([]byte(c.input), &r); err != nil {
			t.Error(err)
		}
		if r.Handler != c.result.Handler {
			t.Errorf("Incorrect ServiceLog.Handler for '%s', got: '%d', wanted: '%d'", c.input, r.Handler, c.result.Handler)
		}
		if r.Path != c.result.Path {
			t.Errorf("Incorrect ServiceLog.Path for '%s', got: '%s', wanted: '%s'", c.input, r.Path, c.result.Path)
		}
	}
}

func TestMetaServiceMeta(t *testing.T) {
	cases := []struct {
		input  string
		result gantry.ServiceMeta
	}{
		{`{}`, gantry.ServiceMeta{Ignore: false, IgnoreFailure: false, KeepAlive: 0, Stdout: gantry.ServiceLog{Handler: 0, Path: ""}, Stderr: gantry.ServiceLog{Handler: 0, Path: ""}}},
		{`{"ignore": true}`, gantry.ServiceMeta{Ignore: true, IgnoreFailure: false, KeepAlive: 0, Stdout: gantry.ServiceLog{Handler: 0, Path: ""}, Stderr: gantry.ServiceLog{Handler: 0, Path: ""}}},
		{`{"ignore_failure": true}`, gantry.ServiceMeta{Ignore: false, IgnoreFailure: true, KeepAlive: 0, Stdout: gantry.ServiceLog{Handler: 0, Path: ""}, Stderr: gantry.ServiceLog{Handler: 0, Path: ""}}},
		{`{"keep_alive": "replace"}`, gantry.ServiceMeta{Ignore: false, IgnoreFailure: false, KeepAlive: gantry.KeepAliveReplace, Stdout: gantry.ServiceLog{Handler: 0, Path: ""}, Stderr: gantry.ServiceLog{Handler: 0, Path: ""}}},
		{`{"stdout": {"handler": "discard"}}`, gantry.ServiceMeta{Ignore: false, IgnoreFailure: false, KeepAlive: 0, Stdout: gantry.ServiceLog{Handler: gantry.LogHandlerDiscard, Path: ""}, Stderr: gantry.ServiceLog{Handler: 0, Path: ""}}},
		{`{"stderr": {"handler": "discard"}}`, gantry.ServiceMeta{Ignore: false, IgnoreFailure: false, KeepAlive: 0, Stdout: gantry.ServiceLog{Handler: 0, Path: ""}, Stderr: gantry.ServiceLog{Handler: gantry.LogHandlerDiscard, Path: ""}}},
	}

	for _, c := range cases {
		var r gantry.ServiceMeta
		if err := json.Unmarshal([]byte(c.input), &r); err != nil {
			t.Error(err)
		}
		if r.Ignore != c.result.Ignore {
			t.Errorf("Incorrect ServiceMeta.Ignore for '%s', got: '%t', wanted: '%t'", c.input, r.Ignore, c.result.Ignore)
		}
		if r.IgnoreFailure != c.result.IgnoreFailure {
			t.Errorf("Incorrect ServiceMeta.IgnoreFailure for '%s', got: '%t', wanted: '%t'", c.input, r.IgnoreFailure, c.result.IgnoreFailure)
		}
		if r.KeepAlive != c.result.KeepAlive {
			t.Errorf("Incorrect ServiceMeta.KeepAlive for '%s', got: '%d', wanted: '%d'", c.input, r.KeepAlive, c.result.KeepAlive)
		}
		if r.Stderr.Handler != c.result.Stderr.Handler {
			t.Errorf("Incorrect ServiceMeta.Stderr.Handler for '%s', got: '%d', wanted: '%d'", c.input, r.Stderr.Handler, c.result.Stderr.Handler)
		}
		if r.Stderr.Path != c.result.Stderr.Path {
			t.Errorf("Incorrect ServiceMeta.Stderr.Path for '%s', got: '%s', wanted: '%s'", c.input, r.Stderr.Path, c.result.Stderr.Path)
		}
		if r.Stdout.Handler != c.result.Stdout.Handler {
			t.Errorf("Incorrect ServiceMeta.Stdout.Handler for '%s', got: '%d', wanted: '%d'", c.input, r.Stdout.Handler, c.result.Stdout.Handler)
		}
		if r.Stdout.Path != c.result.Stdout.Path {
			t.Errorf("Incorrect ServiceMeta.Stdout.Path for '%s', got: '%s', wanted: '%s'", c.input, r.Stdout.Path, c.result.Stdout.Path)
		}
	}
}

func TestMetaServiceMetaList(t *testing.T) {
	cases := []struct {
		input  string
		result gantry.ServiceMetaList
	}{
		{`{}`, gantry.ServiceMetaList{}},
		{`{"a": {}}`, gantry.ServiceMetaList{"a": gantry.ServiceMeta{}}},
		{`{"a": {}, "b": {}}`, gantry.ServiceMetaList{"a": gantry.ServiceMeta{}, "b": gantry.ServiceMeta{}}},
		{`{"a": {}, "a": {}}`, gantry.ServiceMetaList{"a": gantry.ServiceMeta{}}},
		{`{"a": {}, "b": {"keep_alive": "yes"}}`, gantry.ServiceMetaList{"a": gantry.ServiceMeta{}, "b": gantry.ServiceMeta{KeepAlive: gantry.KeepAliveYes}}},
		{`{"a": {}, "b": {"keep_alive": "no"}}`, gantry.ServiceMetaList{"a": gantry.ServiceMeta{}, "b": gantry.ServiceMeta{KeepAlive: gantry.KeepAliveNo}}},
	}

	for _, c := range cases {
		var r gantry.ServiceMetaList
		if err := json.Unmarshal([]byte(c.input), &r); err != nil {
			t.Error(err)
		}
		if len(r) != len(c.result) {
			t.Errorf("Incorrect number of entries in ServiceMetaList for '%s', got: '%d', wanted: '%d'", c.input, len(r), len(c.result))
		}
		for key, meta := range r {
			if meta.KeepAlive != c.result[key].KeepAlive {
				t.Errorf("Incorrect KeepAlive for '%s', got: '%d', wanted: '%d'", c.input, meta.KeepAlive, c.result[key].KeepAlive)
			}
		}
	}
}
