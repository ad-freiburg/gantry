package gantry

import (
	"os/exec"
)

// ExecutionError is an error which stores an additional exit code.
type ExecutionError struct {
	err              error
	exitCodeOverride int
}

// Error returns the string representation of the error.
func (e ExecutionError) Error() string {
	return e.err.Error()
}

// ExitCode returns the integer with which the programm shall exit.
func (e ExecutionError) ExitCode() int {
	if e.exitCodeOverride != 0 {
		return e.exitCodeOverride
	}
	if err, ok := e.err.(*exec.ExitError); ok {
		return err.ExitCode()
	}
	return 1
}
