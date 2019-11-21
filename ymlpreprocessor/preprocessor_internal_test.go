package ymlpreprocessor

import (
	"fmt"
	"os"
	"testing"
)

type testEnv map[string]*string

func (e testEnv) GetSubstitution(key string) (*string, bool) {
	value, ok := e[key]
	return value, ok
}

func (e testEnv) SetSubstitution(key string, value *string) {
	e[key] = value
}

func (e testEnv) GetOrCreateTempDir(key string) (string, error) {
	return key, nil
}

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
		env := testEnv{
			"NIL":     nil,
			"EMPTY":   &empty,
			"TEMPDIR": &tempDir,
		}
		err := processPreprocessorLines([]string{c.line}, &env)
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

func TestNewPreprocessorInstruction(t *testing.T) {
	empty := ""
	tempDir := os.TempDir()
	cases := []struct {
		line         string
		inst         *PreprocessorInstruction
		errorMessage string
	}{
		{
			"FUNCTION_WITHOUT_VAR_OR_ARG",
			&PreprocessorInstruction{
				Function: "FUNCTION_WITHOUT_VAR_OR_ARG",
			},
			"",
		},
		{
			"FUNCTION VARIABLE",
			&PreprocessorInstruction{
				Function: "FUNCTION",
				Variable: "VARIABLE",
			},
			"invalid variable in: 'FUNCTION VARIABLE'",
		},
		{
			"FUNCTION ${VARIABLE}",
			&PreprocessorInstruction{
				Function:          "FUNCTION",
				Variable:          "VARIABLE",
				CurrentValueFound: false,
			},
			"",
		},
		{
			"FUNCTION ${NIL}",
			&PreprocessorInstruction{
				Function:          "FUNCTION",
				Variable:          "NIL",
				CurrentValue:      nil,
				CurrentValueFound: true,
			},
			"",
		},
		{
			"FUNCTION ${VARIABLE} ARG0",
			&PreprocessorInstruction{
				Function:  "FUNCTION",
				Variable:  "VARIABLE",
				Arguments: []string{"ARG0"},
			},
			"",
		},
		{
			"FUNCTION ${VARIABLE} ARG0 ARG1 ARG2",
			&PreprocessorInstruction{
				Function:  "FUNCTION",
				Variable:  "VARIABLE",
				Arguments: []string{"ARG0", "ARG1", "ARG2"},
			},
			"",
		},
	}
	for i, c := range cases {
		env := testEnv{
			"NIL":     nil,
			"EMPTY":   &empty,
			"TEMPDIR": &tempDir,
		}
		inst, err := NewPreprocessorInstruction(c.line, &env)
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
		if inst.Function != c.inst.Function {
			t.Errorf("incorrect inst.Function, got: %s, wanted: %s", inst.Function, c.inst.Function)
		}
		if inst.Variable != c.inst.Variable {
			t.Errorf("incorrect inst.Variable, got: %s, wanted: %s", inst.Variable, c.inst.Variable)
		}
		if inst.CurrentValue != c.inst.CurrentValue {
			t.Errorf("incorrect inst.CurrentValue, got: %v, wanted: %v", inst.CurrentValue, c.inst.CurrentValue)
		}
		if inst.CurrentValueFound != c.inst.CurrentValueFound {
			t.Errorf("incorrect inst.CurrentValueFound, got: %t, wanted: %t", inst.CurrentValueFound, c.inst.CurrentValueFound)
		}
		if len(inst.Arguments) != len(c.inst.Arguments) {
			t.Errorf("incorrect inst.Arguments length, got: %d, wanted: %d", len(inst.Arguments), len(c.inst.Arguments))
			continue
		}
		for j, arg := range inst.Arguments {
			if arg != c.inst.Arguments[j] {
				t.Errorf("incorrect inst.Arguments @%d, got: %s, wanted: %s", j, arg, c.inst.Arguments[j])
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
		instruction  *PreprocessorInstruction
		errorMessage string
	}{
		{
			&PreprocessorInstruction{
				Function: "FUNCTION",
			},
			"missing variable in FUNCTION",
		},
		{
			&PreprocessorInstruction{
				Function:          "FUNCTION",
				Variable:          "NIL",
				CurrentValue:      nil,
				CurrentValueFound: true,
			},
			"empty variable in FUNCTION for NIL",
		},
		{
			&PreprocessorInstruction{
				Function:          "FUNCTION",
				Variable:          "EMPTY",
				CurrentValue:      &empty,
				CurrentValueFound: true,
			},
			"empty variable in FUNCTION for EMPTY",
		},
		{
			&PreprocessorInstruction{
				Function:          "FUNCTION",
				Variable:          "I_DO_NOT_EXIST",
				CurrentValue:      &iDoNotExist,
				CurrentValueFound: true,
			},
			"path error in FUNCTION for I_DO_NOT_EXIST: err: 'stat /iDoNotExist: no such file or directory'",
		},
		{
			&PreprocessorInstruction{
				Function:          "FUNCTION",
				Variable:          "NOT_A_PATH",
				CurrentValue:      &notAPath,
				CurrentValueFound: true,
			},
			fmt.Sprintf("path error in FUNCTION for NOT_A_PATH: not a directory '%s'", notAPath),
		},
		{
			&PreprocessorInstruction{
				Function:          "FUNCTION",
				Variable:          "TEMPDIR",
				CurrentValue:      &tempDir,
				CurrentValueFound: true,
			},
			"",
		},
		{
			&PreprocessorInstruction{
				Function:          "FUNCTION",
				Variable:          "TEMPDIR",
				Arguments:         []string{"ARGUMENT"},
				CurrentValue:      &tempDir,
				CurrentValueFound: true,
			},
			"too many arguments in FUNCTION for TEMPDIR",
		},
	}
	for i, c := range cases {
		env := testEnv{
			"NIL":            nil,
			"EMPTY":          &empty,
			"TEMPDIR":        &tempDir,
			"I_DO_NOT_EXIST": &iDoNotExist,
			"NOT_A_PATH":     &notAPath,
		}
		err := processCheckIfDirExists(c.instruction, &env)
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
	env := testEnv{
		"TEMPDIR": &tempDir,
	}
	err := processSetIfEmpty(&PreprocessorInstruction{
		Function: "FUNCTION",
	}, &env)
	if err == nil {
		t.Errorf("expected error, got nil")
	}
	e := "missing variable in FUNCTION"
	if err.Error() != e {
		t.Errorf("incorrect error, got: '%s', wanted: '%s'", err, e)
	}

	// missing arguments
	err = processSetIfEmpty(&PreprocessorInstruction{
		Function:     "FUNCTION",
		Variable:     "TEMPDIR",
		CurrentValue: &tempDir,
	}, &env)
	if err == nil {
		t.Errorf("expected error, got nil")
	}
	e = "missing argument in FUNCTION for TEMPDIR"
	if err.Error() != e {
		t.Errorf("incorrect error, got: '%s', wanted: '%s'", err, e)
	}

	// too many arguments
	err = processSetIfEmpty(&PreprocessorInstruction{
		Function:     "FUNCTION",
		Variable:     "TEMPDIR",
		Arguments:    []string{"foo", "bar"},
		CurrentValue: &tempDir,
	}, &env)
	if err == nil {
		t.Errorf("expected error, got nil")
	}
	e = "too many arguments in FUNCTION for TEMPDIR"
	if err.Error() != e {
		t.Errorf("incorrect error, got: '%s', wanted: '%s'", err, e)
	}

	// non-empty variable
	err = processSetIfEmpty(&PreprocessorInstruction{
		Function:          "FUNCTION",
		Variable:          "TEMPDIR",
		Arguments:         []string{"foo"},
		CurrentValue:      &tempDir,
		CurrentValueFound: true,
	}, env)
	if err != nil {
		t.Errorf("unexpected error, got: %s", err)
	}
	if env["TEMPDIR"] != &tempDir {
		t.Errorf("unexpected value change, got: '%s', wanted: '%s'", *env["TEMPDIR"], tempDir)
	}

	// nil variable
	env = testEnv{
		"NIL": nil,
	}
	err = processSetIfEmpty(&PreprocessorInstruction{
		Function:          "FUNCTION",
		Variable:          "NIL",
		Arguments:         []string{"foo"},
		CurrentValue:      nil,
		CurrentValueFound: true,
	}, &env)
	if err != nil {
		t.Errorf("unexpected error, got: %s", err)
	}
	if val, ok := env["NIL"]; !ok {
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
	env = testEnv{
		"EMPTY": &empty,
	}
	err = processSetIfEmpty(&PreprocessorInstruction{
		Function:     "FUNCTION",
		Variable:     "EMPTY",
		Arguments:    []string{"foo"},
		CurrentValue: &empty,
	}, &env)
	if err != nil {
		t.Errorf("unexpected error, got: %s", err)
	}
	if *env["EMPTY"] == "" {
		t.Errorf("incorrect value change, got: '', wanted: 'foo'")
	} else if *env["EMPTY"] != "foo" {
		t.Errorf("incorrect value change, got: '%s', wanted: 'foo'", *env["EMPTY"])
	}
	if val, ok := env["EMPTY"]; !ok {
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
	env = testEnv{}
	err = processSetIfEmpty(&PreprocessorInstruction{
		Function:     "FUNCTION",
		Variable:     "UNDEFINED",
		Arguments:    []string{"foo"},
		CurrentValue: nil,
	}, &env)
	if err != nil {
		t.Errorf("unexpected error, got: %s", err)
	}
	if val, ok := env["UNDEFINED"]; !ok {
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
	env := testEnv{
		"Foo": nil,
		"BAR": &baz,
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
		r := expandVariables(c.input, &env)
		if len(r) != len(c.expected) {
			t.Errorf("incorrect result size @%d, got: %d, wanted: %d", i, len(r), len(c.expected))
			continue
		}
	}
}
