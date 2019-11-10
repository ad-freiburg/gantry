package gantry

import (
	"fmt"
	"os/exec"
	"testing"
)

func TestExecutionErrorError(t *testing.T) {
	msg := "I am an Error"
	e := ExecutionError{
		err:              fmt.Errorf(msg),
		exitCodeOverride: 1,
	}
	if e.Error() != msg {
		t.Errorf("incorrect error message, got: %s wanted: %s", e.Error(), msg)
	}
}

func TestExecutionErrorExitCodeSimpleExitCodeOverride(t *testing.T) {
	msg := "I am an Error"
	e := ExecutionError{
		err:              fmt.Errorf(msg),
		exitCodeOverride: 0,
	}
	if e.ExitCode() != 1 {
		t.Errorf("incorrect exit code, got: %d wanted: 1", e.ExitCode())
	}
	e.exitCodeOverride = 42
	if e.ExitCode() != 42 {
		t.Errorf("incorrect exit code, got: %d wanted: 42", e.ExitCode())
	}
}

func TestExecutionErrorExitCodeExitError(t *testing.T) {
	e := ExecutionError{
		err:              &exec.ExitError{},
		exitCodeOverride: 0,
	}
	if e.ExitCode() != -1 {
		t.Errorf("incorrect exit code, got: %d wanted: -1", e.ExitCode())
	}
}
