package gantry

import (
	"testing"
)

func TestGetContainerExecutable(t *testing.T) {
	ForceWharfer = true
	if r := getContainerExecutable(); r != wharfer {
		t.Errorf("incorrect result with ForceWharfer=true, got: %s, wanted: %s", r, wharfer)
	}
	ForceWharfer = false
	if r := getContainerExecutable(); r != wharfer && r != docker {
		t.Errorf("incorrect result with ForceWharfer=false, got: %s", r)
	}
}

func TestNoopRunnerCopy(t *testing.T) {
	s := NewNoopRunner(true)
	c, ok := s.Copy().(*NoopRunner)
	if !ok {
		t.Errorf("incorrect return type")
		return
	}
	if s.silent != c.silent {
		t.Errorf("incorrect value in copy, got: %T, wanted: %T", c.silent, s.silent)
	}

	s = NewNoopRunner(false)
	c, ok = s.Copy().(*NoopRunner)
	if !ok {
		t.Errorf("incorrect return type")
		return
	}
	if s.silent != c.silent {
		t.Errorf("incorrect value in copy, got: %T, wanted: %T", c.silent, s.silent)
	}
}
