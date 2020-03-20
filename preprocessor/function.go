package preprocessor

import (
	"fmt"
)

// Function is a function executable by the preprocessor
type Function struct {
	Func          func(Instruction, Environment, bool) error
	Names         []string
	Description   string
	NumArgsMin    int
	NumArgsMax    int
	NeedsVariable bool
}

// Check performs basic checks, e.g. to enforce correct number of arguments
func (f Function) Check(i Instruction) error {
	if f.NeedsVariable && len(i.Variable) == 0 {
		return fmt.Errorf("missing variable in %s", i.Function)
	}
	if len(i.Arguments) < f.NumArgsMin {
		return fmt.Errorf("missing argument(s) in %s for %s, wanted: %d, got: %d", i.Function, i.Variable, f.NumArgsMin, len(i.Arguments))
	}
	if len(i.Arguments) > f.NumArgsMax {
		return fmt.Errorf("too many arguments in %s for %s, wanted: %d, got: %d", i.Function, i.Variable, f.NumArgsMax, len(i.Arguments))
	}
	return nil
}

// Execute executes the function for the given instruction and environment
func (f Function) Execute(i Instruction, e Environment, dryRun bool) error {
	if err := f.Check(i); err != nil {
		return err
	}
	return f.Func(i, e, dryRun)
}

// Usage returns this usage information of the function.
func (f Function) Usage() string {
	// Name1 ${VAR} ARG [OPT_ARG]
	// Name1 ${VAR} ARG [OPT_ARG]
	// Description
	argline := ""
	if f.NeedsVariable {
		argline = fmt.Sprintf(" ${VAR}")
	}
	argc := 0
	if f.NumArgsMin > 0 {
		for i := 0; i < f.NumArgsMin; i++ {
			argline = fmt.Sprintf("%s ARG%d", argline, argc)
			argc++
		}
	}
	if f.NumArgsMax > f.NumArgsMin {
		argline = fmt.Sprintf("%s [", argline)
		for i := 0; i < f.NumArgsMax-f.NumArgsMin; i++ {
			argline = fmt.Sprintf("%s ARG%d", argline, argc)
			argc++
		}
		argline = fmt.Sprintf("%s ]", argline)
	}

	result := ""
	for i, name := range f.Names {
		if i > 0 {
			result = fmt.Sprintf("%s\n", result)
		}
		result = fmt.Sprintf("%s#! %s%s", result, name, argline)
	}
	if len(f.Description) > 0 {
		if len(result) > 0 {
			result = fmt.Sprintf("%s\n", result)
		}
		result = fmt.Sprintf("%s%s", result, f.Description)
	}
	return result
}
