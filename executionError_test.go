package gantry

import (
	"fmt"
	"os/exec"
	"testing"
)

func TestExecutionErrorError(t *testing.T) {
	msg := "I am an Error"
	e := ExecutionError{
		err:      fmt.Errorf(msg),
		override: 1,
	}
	if e.Error() != msg {
		t.Errorf("incorrect error message, got: %s wanted: %s", e.Error(), msg)
	}
}

func TestExecutionErrorExitCodeSimpleOverride(t *testing.T) {
	msg := "I am an Error"
	e := ExecutionError{
		err:      fmt.Errorf(msg),
		override: 0,
	}
	if e.ExitCode() != 1 {
		t.Errorf("incorrect exit code, got: %d wanted: 1", e.ExitCode())
	}
	e.override = 42
	if e.ExitCode() != 42 {
		t.Errorf("incorrect exit code, got: %d wanted: 42", e.ExitCode())
	}
}

func TestExecutionErrorExitCodeExitError(t *testing.T) {
	e := ExecutionError{
		err:      &exec.ExitError{},
		override: 0,
	}
	if e.ExitCode() != -1 {
		t.Errorf("incorrect exit code, got: %d wanted: -1", e.ExitCode())
	}
}
