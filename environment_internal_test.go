package gantry

import (
	"bufio"
	"bytes"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"text/template"

	"github.com/ad-freiburg/gantry/types"
)

func TestPipelineEnvironmentCreatedTemplateParserDirectSubstitution(t *testing.T) {
	cases := []struct {
		in            string
		out           string
		err           bool
		substitutions types.MappingWithEquals
	}{
		{"{{ Foo }}", "", true, types.MappingWithEquals{}},
		{"{{ Foo }}", "", false, types.MappingWithEquals{"Foo": nil}},
	}

	for i, c := range cases {
		var tpl *template.Template
		e, err := NewPipelineEnvironment("", c.substitutions, types.StringSet{}, types.StringSet{})
		if err != nil && !os.IsNotExist(err) {
			log.Fatal(err)
		}

		// Create parser and parse template
		tpl, err = e.createTemplateParser().Parse(c.in)
		if err != nil {
			if !c.err {
				t.Errorf("Unexpected error in case '%d', got: '%s'", i, err)
			}
			continue
		}

		// Execute and check resutl
		var b bytes.Buffer
		bw := bufio.NewWriter(&b)
		err = tpl.Execute(bw, e)
		bw.Flush()
		if err != nil {
			if !c.err {
				t.Errorf("Unexpected error in case '%d', got: '%s'", i, err)
			}
			continue
		}
		res := string(b.Bytes())
		if res != c.out {
			t.Errorf("Incorrect result in case '%d', got: '%s', wanted: '%s'", i, res, c.out)
		}
		e.CleanUp(syscall.Signal(0))
	}
}

func TestPipelineEnvironmentCreatedTemplateParserEnv(t *testing.T) {
	baz := "Baz"
	cases := []struct {
		in            string
		out           string
		err           bool
		substitutions types.MappingWithEquals
	}{
		{"{{ Env }}", "", true, types.MappingWithEquals{}},
		{"{{ Env \"Foo\" }}", "", true, types.MappingWithEquals{}},
		{"{{ Env \"Foo\" }}", baz, false, types.MappingWithEquals{"Foo": &baz}},
		{"{{ Env \"Foo\" \"Bar\" }}", "Bar", false, types.MappingWithEquals{}},
		{"{{ Env \"Foo\" \"Bar\" }}", baz, false, types.MappingWithEquals{"Foo": &baz}},
		{"{{ Env \"Foo\" \"Bar\" \"X\" }}", "", true, types.MappingWithEquals{}},
	}

	for i, c := range cases {
		var tpl *template.Template
		e, err := NewPipelineEnvironment("", c.substitutions, types.StringSet{}, types.StringSet{})
		if err != nil && !os.IsNotExist(err) {
			log.Fatal(err)
		}

		// Create parser and parse template
		tpl, err = e.createTemplateParser().Parse(c.in)
		if err != nil {
			if !c.err {
				t.Errorf("Unexpected error in case '%d', got: '%s'", i, err)
			}
			continue
		}

		// Execute and check resutl
		var b bytes.Buffer
		bw := bufio.NewWriter(&b)
		err = tpl.Execute(bw, e)
		bw.Flush()
		if err != nil {
			if !c.err {
				t.Errorf("Unexpected error in case '%d', got: '%s'", i, err)
			}
			continue
		}
		res := string(b.Bytes())
		if res != c.out {
			t.Errorf("Incorrect result in case '%d', got: '%s', wanted: '%s'", i, res, c.out)
		}
		e.CleanUp(syscall.Signal(0))
	}
}

func TestPipelineEnvironmentCreatedTemplateParserEnvDir(t *testing.T) {
	cwd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	cases := []struct {
		in            string
		out           string
		err           bool
		substitutions types.MappingWithEquals
	}{
		{"{{ EnvDir }}", "", true, types.MappingWithEquals{}},
		{"{{ EnvDir }}", "", true, types.MappingWithEquals{}},
		{"{{ EnvDir \"Foo\" }}", "", true, types.MappingWithEquals{}},
		{"{{ EnvDir \"Foo\" }}", cwd, false, types.MappingWithEquals{"Foo": &cwd}},
		{"{{ EnvDir \"Foo\" \"I_WILL_NEVER_EXIST\" }}", "", true, types.MappingWithEquals{}},
		{"{{ EnvDir \"Foo\" \"/tmp\" }}", "/tmp", false, types.MappingWithEquals{}},
		{"{{ EnvDir \"Foo\" \"/tmp\" }}", cwd, false, types.MappingWithEquals{"Foo": &cwd}},
		{"{{ EnvDir \"Foo\" \"/tmp\" \"X\" }}", "", true, types.MappingWithEquals{}},
	}

	for i, c := range cases {
		var tpl *template.Template
		e, err := NewPipelineEnvironment("", c.substitutions, types.StringSet{}, types.StringSet{})
		if err != nil && !os.IsNotExist(err) {
			log.Fatal(err)
		}

		// Create parser and parse template
		tpl, err = e.createTemplateParser().Parse(c.in)
		if err != nil {
			if !c.err {
				t.Errorf("Unexpected error in case '%d', got: '%s'", i, err)
			}
			continue
		}

		// Execute and check resutl
		var b bytes.Buffer
		bw := bufio.NewWriter(&b)
		err = tpl.Execute(bw, e)
		bw.Flush()
		if err != nil {
			if !c.err {
				t.Errorf("Unexpected error in case '%d', got: '%s'", i, err)
			}
			continue
		}
		res := string(b.Bytes())
		if res != c.out {
			t.Errorf("Incorrect result in case '%d', got: '%s', wanted: '%s'", i, res, c.out)
		}
		e.CleanUp(syscall.Signal(0))
	}
}

func TestPipelineEnvironmentCreatedTemplateParserTmpDir(t *testing.T) {
	cases := []struct {
		in            string
		prefix        string
		err           bool
		substitutions types.MappingWithEquals
	}{
		{"{{ TempDir }}", "", false, types.MappingWithEquals{}},
		{"{{ TempDir \"a\" }}", "a", false, types.MappingWithEquals{}},
		{"{{ TempDir \"a\" \"b\" }}", "a_b", false, types.MappingWithEquals{}},
	}

	for i, c := range cases {
		var tpl *template.Template
		e, err := NewPipelineEnvironment("", c.substitutions, types.StringSet{}, types.StringSet{})
		if err != nil && !os.IsNotExist(err) {
			log.Fatal(err)
		}

		// Create parser and parse template
		tpl, err = e.createTemplateParser().Parse(c.in)
		if err != nil {
			if !c.err {
				t.Errorf("Unexpected error in case '%d', got: '%s'", i, err)
			}
			continue
		}

		// Execute and check resutl
		var b bytes.Buffer
		bw := bufio.NewWriter(&b)
		err = tpl.Execute(bw, e)
		bw.Flush()
		if err != nil {
			if !c.err {
				t.Errorf("Unexpected error in case '%d', got: '%s'", i, err)
			}
			continue
		}
		res := filepath.Base(string(b.Bytes()))
		if !strings.HasPrefix(res, c.prefix) {
			t.Errorf("Incorrect result in case '%d', got: '%s', wanted prefix: '%s'", i, res, c.prefix)
		}
		e.CleanUp(syscall.Signal(0))
	}
}

func TestPipelineEnvironmentUpdateSubstitutions(t *testing.T) {
	checkKeyAndValue := func(m types.MappingWithEquals, key string, value *string) error {
		val, found := m[key]
		if !found {
			return fmt.Errorf("Unknown Substitution '%s'", key)
		}
		if val != value {
			return fmt.Errorf("Incorrect Substitution for '%s', got: '%s', wanted: '%s'", key, *val, *value)
		}
		return nil
	}
	bar := "bar"
	baz := "baz"
	e, err := NewPipelineEnvironment("", types.MappingWithEquals{}, types.StringSet{}, types.StringSet{})
	if err != nil && !os.IsNotExist(err) {
		log.Fatal(err)
	}

	// Assert empty Substitutions
	if len(e.Substitutions) > 0 {
		t.Errorf("Incorrect number of Substitutions entries, got: '%d', wanted: '%d'", len(e.Substitutions), 0)
	}

	// Add foo -> baz
	e.updateSubstitutions(types.MappingWithEquals{"foo": &baz})
	if len(e.Substitutions) != 1 {
		t.Errorf("Incorrect number of Substitutions entries, got: '%d', wanted: '%d'", len(e.Substitutions), 1)
	}
	if err := checkKeyAndValue(e.Substitutions, "foo", &baz); err != nil {
		t.Error(err)
	}

	// Update foo -> bar
	e.updateSubstitutions(types.MappingWithEquals{"foo": &bar})
	if len(e.Substitutions) != 1 {
		t.Errorf("Incorrect number of Substitutions entries, got: '%d', wanted: '%d'", len(e.Substitutions), 1)
	}
	if err := checkKeyAndValue(e.Substitutions, "foo", &bar); err != nil {
		t.Error(err)
	}

	// Update foo -> nil
	e.updateSubstitutions(types.MappingWithEquals{"foo": nil})
	if len(e.Substitutions) != 1 {
		t.Errorf("Incorrect number of Substitutions entries, got: '%d', wanted: '%d'", len(e.Substitutions), 1)
	}
	if err := checkKeyAndValue(e.Substitutions, "foo", nil); err != nil {
		t.Error(err)
	}
}

func TestPipelineEnvironmentUpdateStepsMeta(t *testing.T) {
	checkIgnore := func(m ServiceMetaList, key string, value bool) error {
		val, found := m[key]
		if !found {
			return fmt.Errorf("Unknown Step/Service '%s'", key)
		}
		if val.Ignore != value {
			return fmt.Errorf("Incorrect .Ignore for '%s', got: '%t', wanted: '%t'", key, val.Ignore, value)
		}
		return nil
	}
	checkSelected := func(m ServiceMetaList, key string, value bool) error {
		val, found := m[key]
		if !found {
			return fmt.Errorf("Unknown Step/Service '%s'", key)
		}
		if val.Selected != value {
			return fmt.Errorf("Incorrect .Selected for '%s', got: '%t', wanted: '%t'", key, val.Selected, value)
		}
		return nil
	}
	e, err := NewPipelineEnvironment("", types.MappingWithEquals{}, types.StringSet{}, types.StringSet{})
	if err != nil && !os.IsNotExist(err) {
		log.Fatal(err)
	}

	// Assert empty ServiceMetaList
	if len(e.Steps) > 0 {
		t.Errorf("Incorrect number of ServiceMetaList entries, got: '%d', wanted: '%d'", len(e.Steps), 0)
	}

	// Add ignored step
	e.updateStepsMeta(types.StringSet{"ignored": true}, types.StringSet{})
	if len(e.Steps) != 1 {
		t.Errorf("Incorrect number of ServiceMetaList entries, got: '%d', wanted: '%d'", len(e.Steps), 1)
	}
	if err := checkIgnore(e.Steps, "ignored", true); err != nil {
		t.Error(err)
	}

	// Update ignored step value
	e.updateStepsMeta(types.StringSet{"ignored": false}, types.StringSet{})
	if len(e.Steps) != 1 {
		t.Errorf("Incorrect number of ServiceMetaList entries, got: '%d', wanted: '%d'", len(e.Steps), 1)
	}
	if err := checkIgnore(e.Steps, "ignored", false); err != nil {
		t.Error(err)
	}

	// Add selected step
	e.updateStepsMeta(types.StringSet{}, types.StringSet{"selected": false})
	if len(e.Steps) != 2 {
		t.Errorf("Incorrect number of ServiceMetaList entries, got: '%d', wanted: '%d'", len(e.Steps), 2)
	}
	if err := checkSelected(e.Steps, "selected", false); err != nil {
		t.Error(err)
	}

	// Update ignored step value
	e.updateStepsMeta(types.StringSet{}, types.StringSet{"selected": true})
	if len(e.Steps) != 2 {
		t.Errorf("Incorrect number of ServiceMetaList entries, got: '%d', wanted: '%d'", len(e.Steps), 2)
	}
	if err := checkSelected(e.Steps, "selected", true); err != nil {
		t.Error(err)
	}

	// Add combined step, ignored first
	e.updateStepsMeta(types.StringSet{"combined_1": true}, types.StringSet{})
	if len(e.Steps) != 3 {
		t.Errorf("Incorrect number of ServiceMetaList entries, got: '%d', wanted: '%d'", len(e.Steps), 3)
	}
	if err := checkIgnore(e.Steps, "combined_1", true); err != nil {
		t.Error(err)
	}
	e.updateStepsMeta(types.StringSet{}, types.StringSet{"combined_1": true})
	if len(e.Steps) != 3 {
		t.Errorf("Incorrect number of ServiceMetaList entries, got: '%d', wanted: '%d'", len(e.Steps), 3)
	}
	if err := checkSelected(e.Steps, "combined_1", true); err != nil {
		t.Error(err)
	}

	// Add combined step, selected first
	e.updateStepsMeta(types.StringSet{}, types.StringSet{"combined_2": true})
	if len(e.Steps) != 4 {
		t.Errorf("Incorrect number of ServiceMetaList entries, got: '%d', wanted: '%d'", len(e.Steps), 4)
	}
	if err := checkSelected(e.Steps, "combined_2", true); err != nil {
		t.Error(err)
	}
	e.updateStepsMeta(types.StringSet{"combined_2": true}, types.StringSet{})
	if len(e.Steps) != 4 {
		t.Errorf("Incorrect number of ServiceMetaList entries, got: '%d', wanted: '%d'", len(e.Steps), 4)
	}
	if err := checkIgnore(e.Steps, "combined_2", true); err != nil {
		t.Error(err)
	}
}
