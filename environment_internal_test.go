package gantry

import (
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/ad-freiburg/gantry/types"
)

func TestPipelineEnvironmentUpdateSubstitutions(t *testing.T) {
	checkKeyAndValue := func(m types.StringMap, key string, value *string) error {
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
	e, err := NewPipelineEnvironment("", types.StringMap{}, types.StringSet{}, types.StringSet{})
	if err != nil && !os.IsNotExist(err) {
		log.Fatal(err)
	}

	// Assert empty Substitutions
	if len(e.Substitutions) > 0 {
		t.Errorf("Incorrect number of Substitutions entries, got: '%d', wanted: '%d'", len(e.Substitutions), 0)
	}

	// Add foo -> baz
	e.updateSubstitutions(types.StringMap{"foo": &baz})
	if len(e.Substitutions) != 1 {
		t.Errorf("Incorrect number of Substitutions entries, got: '%d', wanted: '%d'", len(e.Substitutions), 1)
	}
	if err := checkKeyAndValue(e.Substitutions, "foo", &baz); err != nil {
		t.Error(err)
	}

	// Update foo -> bar
	e.updateSubstitutions(types.StringMap{"foo": &bar})
	if len(e.Substitutions) != 1 {
		t.Errorf("Incorrect number of Substitutions entries, got: '%d', wanted: '%d'", len(e.Substitutions), 1)
	}
	if err := checkKeyAndValue(e.Substitutions, "foo", &bar); err != nil {
		t.Error(err)
	}

	// Update foo -> nil
	e.updateSubstitutions(types.StringMap{"foo": nil})
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
	e, err := NewPipelineEnvironment("", types.StringMap{}, types.StringSet{}, types.StringSet{})
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
