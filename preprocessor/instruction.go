package preprocessor

import (
	"fmt"
	"strings"
)

// Instruction is a parsed line
type Instruction struct {
	Function          string
	Variable          string
	Arguments         []string
	CurrentValue      *string
	CurrentValueFound bool
}

// NewInstruction parses a line and looks up the current value from the environment
func NewInstruction(line string, env Environment) (Instruction, error) {
	result := Instruction{}
	parts := strings.Split(line, " ")
	if len(parts[0]) < 1 {
		return result, fmt.Errorf("empty preprocessor line found")
	}
	result.Function = parts[0]
	if len(parts) < 2 {
		return result, nil
	}
	// Remove ${ } from variable
	if !strings.HasPrefix(parts[1], "${") || !strings.HasSuffix(parts[1], "}") {
		return result, fmt.Errorf("invalid variable in: '%s'", line)
	}
	result.Variable = parts[1][2 : len(parts[1])-1]
	// Store arguments
	if len(parts) > 2 {
		result.Arguments = parts[2:]
	}
	// Lookup substitution value
	result.CurrentValue, result.CurrentValueFound = env.GetSubstitution(result.Variable)
	return result, nil
}
