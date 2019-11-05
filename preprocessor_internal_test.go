package gantry

import (
	"fmt"
	"os"
	"testing"
)

func TestExtractPreprocessorStatements(t *testing.T) {
	input := []string{
		"#! PRE",
		"#!WITHOUT SPACE",
		"outer:",
		"  inner1:",
		"    # COMMENT",
		"    value1",
		"    value2",
		"    #! INNER",
		"  inner2:",
		"    #!  MULTIPLE SPACES",
		"    value3",
		"#! POST",
	}
	statements := []string{
		"PRE",
		"WITHOUT SPACE",
		"INNER",
		"MULTIPLE SPACES",
		"POST",
	}
	lines := []string{
		"outer:",
		"  inner1:",
		"    value1",
		"    value2",
		"  inner2:",
		"    value3",
	}
	rs, rl := extractPreprocessorLines(input)
	if len(rs) != len(statements) {
		t.Errorf("incorrect number of statements, got: %d, wanted: %d", len(rs), len(statements))
		return
	}
	for i, s := range rs {
		if s != statements[i] {
			t.Errorf("incorrect statement @%d, got: '%s', wanted: '%s'", i, s, statements[i])
		}
	}
	if len(rl) != len(lines) {
		t.Errorf("incorrect number of normal lines, got: %d, wanted: %d", len(rl), len(lines))
		return
	}
	for i, l := range rl {
		if l != lines[i] {
			t.Errorf("incorrect line @%d, got: '%s', wanted: '%s'", i, l, lines[i])
		}
	}
}

func TestProcessPreprocessorLines(t *testing.T) {
	empty := ""
	tempDir := os.TempDir()
	cases := []struct {
		line         string
		errorMessage string
	}{
		{
			"",
			"empty preprocessor line found!",
		},
		{
			"FUNCTION_WITHOUT_VAR_OR_ARG",
			"unknown preprocessor directive: 'FUNCTION_WITHOUT_VAR_OR_ARG'",
		},
		{
			"FUNCTION INVALID_VAR",
			"invalid variable in: 'FUNCTION INVALID_VAR'",
		},
		{
			"SET_IF_EMPTY ${X} Foo",
			"",
		},
		{
			"CHECK_IF_DIR_EXISTS ${TEMPDIR}",
			"",
		},
	}
	for i, c := range cases {
		env := &PipelineEnvironment{
			Substitutions: map[string]*string{
				"NIL":     nil,
				"EMPTY":   &empty,
				"TEMPDIR": &tempDir,
			},
		}
		err := processPreprocessorLines([]string{c.line}, env)
		if len(c.errorMessage) > 0 {
			if err == nil {
				t.Errorf("expected error @%d, got nil", i)
				continue
			}
			if err.Error() != c.errorMessage {
				t.Errorf("incorrect error @%d, got: '%s', wanted: '%s'", i, err, c.errorMessage)
			}
		} else if err != nil {
			t.Errorf("unexpected error @%d: '%s'", i, err)
		}
	}
}

func TestParsePreprocessorLine(t *testing.T) {
	empty := ""
	tempDir := os.TempDir()
	cases := []struct {
		line         string
		inst         *preprocessorInstruction
		errorMessage string
	}{
		{
			"FUNCTION_WITHOUT_VAR_OR_ARG",
			&preprocessorInstruction{
				function: "FUNCTION_WITHOUT_VAR_OR_ARG",
			},
			"",
		},
		{
			"FUNCTION VARIABLE",
			&preprocessorInstruction{
				function: "FUNCTION",
				variable: "VARIABLE",
			},
			"invalid variable in: 'FUNCTION VARIABLE'",
		},
		{
			"FUNCTION ${VARIABLE}",
			&preprocessorInstruction{
				function:          "FUNCTION",
				variable:          "VARIABLE",
				currentValueFound: false,
			},
			"",
		},
		{
			"FUNCTION ${NIL}",
			&preprocessorInstruction{
				function:          "FUNCTION",
				variable:          "NIL",
				currentValue:      nil,
				currentValueFound: true,
			},
			"",
		},
		{
			"FUNCTION ${VARIABLE} ARG0",
			&preprocessorInstruction{
				function:  "FUNCTION",
				variable:  "VARIABLE",
				arguments: []string{"ARG0"},
			},
			"",
		},
		{
			"FUNCTION ${VARIABLE} ARG0 ARG1 ARG2",
			&preprocessorInstruction{
				function:  "FUNCTION",
				variable:  "VARIABLE",
				arguments: []string{"ARG0", "ARG1", "ARG2"},
			},
			"",
		},
	}
	for i, c := range cases {
		env := &PipelineEnvironment{
			Substitutions: map[string]*string{
				"NIL":     nil,
				"EMPTY":   &empty,
				"TEMPDIR": &tempDir,
			},
		}
		inst, err := parsePreprocessorLine(c.line, env)
		if len(c.errorMessage) > 0 {
			if err == nil {
				t.Errorf("expected error @%d, got nil", i)
				continue
			}
			if err.Error() != c.errorMessage {
				t.Errorf("incorrect error @%d, got: '%s', wanted: '%s'", i, err, c.errorMessage)
			}
			continue
		} else if err != nil {
			t.Errorf("unexpected error @%d: '%s'", i, err)
			continue
		}
		if inst.function != c.inst.function {
			t.Errorf("incorrect inst.function, got: %s, wanted: %s", inst.function, c.inst.function)
		}
		if inst.variable != c.inst.variable {
			t.Errorf("incorrect inst.variable, got: %s, wanted: %s", inst.variable, c.inst.variable)
		}
		if inst.currentValue != c.inst.currentValue {
			t.Errorf("incorrect inst.currentValue, got: %v, wanted: %v", inst.currentValue, c.inst.currentValue)
		}
		if inst.currentValueFound != c.inst.currentValueFound {
			t.Errorf("incorrect inst.currentValueFound, got: %t, wanted: %t", inst.currentValueFound, c.inst.currentValueFound)
		}
		if len(inst.arguments) != len(c.inst.arguments) {
			t.Errorf("incorrect inst.arguments length, got: %d, wanted: %d", len(inst.arguments), len(c.inst.arguments))
			continue
		}
		for j, arg := range inst.arguments {
			if arg != c.inst.arguments[j] {
				t.Errorf("incorrect inst.arguments @%d, got: %s, wanted: %s", j, arg, c.inst.arguments[j])
			}
		}
	}
}

func TestProcessCheckIfDirExists(t *testing.T) {
	empty := ""
	iDoNotExist := "/iDoNotExist"
	notAPath := os.Args[0]
	tempDir := os.TempDir()
	cases := []struct {
		instruction  *preprocessorInstruction
		errorMessage string
	}{
		{
			&preprocessorInstruction{
				function: "FUNCTION",
			},
			"missing variable in FUNCTION",
		},
		{
			&preprocessorInstruction{
				function:          "FUNCTION",
				variable:          "NIL",
				currentValue:      nil,
				currentValueFound: true,
			},
			"empty variable in FUNCTION for NIL",
		},
		{
			&preprocessorInstruction{
				function:          "FUNCTION",
				variable:          "EMPTY",
				currentValue:      &empty,
				currentValueFound: true,
			},
			"empty variable in FUNCTION for EMPTY",
		},
		{
			&preprocessorInstruction{
				function:          "FUNCTION",
				variable:          "I_DO_NOT_EXIST",
				currentValue:      &iDoNotExist,
				currentValueFound: true,
			},
			"path error in FUNCTION for I_DO_NOT_EXIST: err: 'stat /iDoNotExist: no such file or directory'",
		},
		{
			&preprocessorInstruction{
				function:          "FUNCTION",
				variable:          "NOT_A_PATH",
				currentValue:      &notAPath,
				currentValueFound: true,
			},
			fmt.Sprintf("path error in FUNCTION for NOT_A_PATH: not a directory '%s'", notAPath),
		},
		{
			&preprocessorInstruction{
				function:          "FUNCTION",
				variable:          "TEMPDIR",
				currentValue:      &tempDir,
				currentValueFound: true,
			},
			"",
		},
		{
			&preprocessorInstruction{
				function:          "FUNCTION",
				variable:          "TEMPDIR",
				arguments:         []string{"ARGUMENT"},
				currentValue:      &tempDir,
				currentValueFound: true,
			},
			"too many arguments in FUNCTION for TEMPDIR",
		},
	}
	for i, c := range cases {
		env := &PipelineEnvironment{
			Substitutions: map[string]*string{
				"NIL":            nil,
				"EMPTY":          &empty,
				"TEMPDIR":        &tempDir,
				"I_DO_NOT_EXIST": &iDoNotExist,
				"NOT_A_PATH":     &notAPath,
			},
		}
		err := processCheckIfDirExists(c.instruction, env)
		if len(c.errorMessage) > 0 {
			if err == nil {
				t.Errorf("expected error @%d, got nil", i)
				continue
			}
			if err.Error() != c.errorMessage {
				t.Errorf("incorrect error @%d, got: '%s', wanted: '%s'", i, err, c.errorMessage)
			}
		} else if err != nil {
			t.Errorf("unexpected error @%d: '%s'", i, err)
		}
	}
}

func TestProcessPreprocessorStatementsSetIfEmpty(t *testing.T) {
	empty := ""
	tempDir := os.TempDir()

	// missing variable
	env := &PipelineEnvironment{
		Substitutions: map[string]*string{
			"TEMPDIR": &tempDir,
		},
	}
	err := processSetIfEmpty(&preprocessorInstruction{
		function: "FUNCTION",
	}, env)
	if err == nil {
		t.Errorf("expected error, got nil")
	}
	e := "missing variable in FUNCTION"
	if err.Error() != e {
		t.Errorf("incorrect error, got: '%s', wanted: '%s'", err, e)
	}

	// missing arguments
	err = processSetIfEmpty(&preprocessorInstruction{
		function:     "FUNCTION",
		variable:     "TEMPDIR",
		currentValue: &tempDir,
	}, env)
	if err == nil {
		t.Errorf("expected error, got nil")
	}
	e = "missing argument in FUNCTION for TEMPDIR"
	if err.Error() != e {
		t.Errorf("incorrect error, got: '%s', wanted: '%s'", err, e)
	}

	// too many arguments
	err = processSetIfEmpty(&preprocessorInstruction{
		function:     "FUNCTION",
		variable:     "TEMPDIR",
		arguments:    []string{"foo", "bar"},
		currentValue: &tempDir,
	}, env)
	if err == nil {
		t.Errorf("expected error, got nil")
	}
	e = "too many arguments in FUNCTION for TEMPDIR"
	if err.Error() != e {
		t.Errorf("incorrect error, got: '%s', wanted: '%s'", err, e)
	}

	// non-empty variable
	err = processSetIfEmpty(&preprocessorInstruction{
		function:          "FUNCTION",
		variable:          "TEMPDIR",
		arguments:         []string{"foo"},
		currentValue:      &tempDir,
		currentValueFound: true,
	}, env)
	if err != nil {
		t.Errorf("unexpected error, got: %s", err)
	}
	if env.Substitutions["TEMPDIR"] != &tempDir {
		t.Errorf("unexpected value change, got: '%s', wanted: '%s'", *env.Substitutions["TEMPDIR"], tempDir)
	}

	// nil variable
	env = &PipelineEnvironment{
		Substitutions: map[string]*string{
			"NIL": nil,
		},
	}
	err = processSetIfEmpty(&preprocessorInstruction{
		function:          "FUNCTION",
		variable:          "NIL",
		arguments:         []string{"foo"},
		currentValue:      nil,
		currentValueFound: true,
	}, env)
	if err != nil {
		t.Errorf("unexpected error, got: %s", err)
	}
	if val, ok := env.Substitutions["NIL"]; !ok {
		t.Errorf("key error, 'NIL' lost")
	} else {
		if val == nil {
			t.Errorf("incorrect value, got: 'nil', wanted: 'foo'")
		}
		if *val != "foo" {
			t.Errorf("incorrect value, got: '%s', wanted: 'foo'", *val)
		}
	}

	// empty variable
	env = &PipelineEnvironment{
		Substitutions: map[string]*string{
			"EMPTY": &empty,
		},
	}
	err = processSetIfEmpty(&preprocessorInstruction{
		function:     "FUNCTION",
		variable:     "EMPTY",
		arguments:    []string{"foo"},
		currentValue: &empty,
	}, env)
	if err != nil {
		t.Errorf("unexpected error, got: %s", err)
	}
	if *env.Substitutions["EMPTY"] == "" {
		t.Errorf("incorrect value change, got: '', wanted: 'foo'")
	} else if *env.Substitutions["EMPTY"] != "foo" {
		t.Errorf("incorrect value change, got: '%s', wanted: 'foo'", *env.Substitutions["EMPTY"])
	}
	if val, ok := env.Substitutions["EMPTY"]; !ok {
		t.Errorf("key error, 'EMPTY' lost")
	} else {
		if val == nil {
			t.Errorf("incorrect value, got: 'nil', wanted: 'foo'")
		}
		if *val != "foo" {
			t.Errorf("incorrect value, got: '%s', wanted: 'foo'", *val)
		}
	}

	// undefined variable
	env = &PipelineEnvironment{
		Substitutions: map[string]*string{},
	}
	err = processSetIfEmpty(&preprocessorInstruction{
		function:     "FUNCTION",
		variable:     "UNDEFINED",
		arguments:    []string{"foo"},
		currentValue: nil,
	}, env)
	if err != nil {
		t.Errorf("unexpected error, got: %s", err)
	}
	if val, ok := env.Substitutions["UNDEFINED"]; !ok {
		t.Errorf("variable creation error, 'UNDEFINED' not created")
	} else {
		if val == nil {
			t.Errorf("incorrect value, got: 'nil', wanted: 'foo'")
		}
		if *val != "foo" {
			t.Errorf("incorrect value, got: '%s', wanted: 'foo'", *val)
		}
	}
}

func TestExpandVariables(t *testing.T) {
	baz := "baz"
	env := &PipelineEnvironment{
		Substitutions: map[string]*string{
			"Foo": nil,
			"BAR": &baz,
		},
	}
	cases := []struct {
		input    []string
		expected []string
	}{
		{
			[]string{},
			[]string{},
		},
		{
			[]string{
				"",
			},
			[]string{
				"",
			},
		},
		{
			[]string{
				"multiple",
				"static",
				"lines",
			},
			[]string{
				"multiple",
				"static",
				"lines",
			},
		},
		{
			[]string{
				"static",
				"${X}",
				"$X",
			},
			[]string{
				"static",
				"",
				"$X",
			},
		},
		{
			[]string{
				"${Foo}",
				" ${Foo}",
				"${Foo} ",
			},
			[]string{
				"",
				" ",
				" ",
			},
		},
		{
			[]string{
				"${FOO}",
				" ${FOO}",
				"${FOO} ",
			},
			[]string{
				"",
				" ",
				" ",
			},
		},
		{
			[]string{
				"${BAR}",
				" ${BAR}",
				"${BAR} ",
			},
			[]string{
				"baz",
				" baz",
				"baz ",
			},
		},
		{
			[]string{
				"${Bar}",
				" ${Bar}",
				"${Bar} ",
			},
			[]string{
				"",
				" ",
				" ",
			},
		},
	}
	for i, c := range cases {
		r := expandVariables(c.input, env)
		if len(r) != len(c.expected) {
			t.Errorf("incorrect result size @%d, got: %d, wanted: %d", i, len(r), len(c.expected))
			continue
		}
	}
}
