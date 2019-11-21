package preprocessor

import (
	"fmt"
	"io/ioutil"
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

type testDefectiveEnv map[string]*string

func (e testDefectiveEnv) GetSubstitution(key string) (*string, bool) {
	value, ok := e[key]
	return value, ok
}

func (e testDefectiveEnv) SetSubstitution(key string, value *string) {
	e[key] = value
}

func (e testDefectiveEnv) GetOrCreateTempDir(key string) (string, error) {
	return key, fmt.Errorf("error")
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

func TestNewInstruction(t *testing.T) {
	empty := ""
	tempDir, _ := ioutil.TempDir("", "newInstruction")
	defer os.RemoveAll(tempDir)

	cases := []struct {
		line         string
		inst         *Instruction
		errorMessage string
	}{
		{
			"FUNCTION_WITHOUT_VAR_OR_ARG",
			&Instruction{
				Function: "FUNCTION_WITHOUT_VAR_OR_ARG",
			},
			"",
		},
		{
			"FUNCTION VARIABLE",
			&Instruction{
				Function: "FUNCTION",
				Variable: "VARIABLE",
			},
			"invalid variable in: 'FUNCTION VARIABLE'",
		},
		{
			"FUNCTION ${VARIABLE}",
			&Instruction{
				Function:          "FUNCTION",
				Variable:          "VARIABLE",
				CurrentValueFound: false,
			},
			"",
		},
		{
			"FUNCTION ${NIL}",
			&Instruction{
				Function:          "FUNCTION",
				Variable:          "NIL",
				CurrentValue:      nil,
				CurrentValueFound: true,
			},
			"",
		},
		{
			"FUNCTION ${VARIABLE} ARG0",
			&Instruction{
				Function:  "FUNCTION",
				Variable:  "VARIABLE",
				Arguments: []string{"ARG0"},
			},
			"",
		},
		{
			"FUNCTION ${VARIABLE} ARG0 ARG1 ARG2",
			&Instruction{
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
		inst, err := NewInstruction(c.line, &env)
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

func TestCheckIfDirExists(t *testing.T) {
	empty := ""
	iDoNotExist := "/iDoNotExist"
	tempFile, err := ioutil.TempFile("", "tempDirIfEmptyFile")
	if err != nil {
		t.Error(err)
		return
	}
	defer os.Remove(tempFile.Name())
	notAPath := tempFile.Name()
	tempDir, _ := ioutil.TempDir("", "checkIfDirExists")
	defer os.RemoveAll(tempDir)

	cases := []struct {
		instruction  Instruction
		errorMessage string
	}{
		{
			Instruction{
				Function:          "FUNCTION",
				Variable:          "I_DO_NOT_EXIST",
				CurrentValue:      &iDoNotExist,
				CurrentValueFound: true,
			},
			"path error in FUNCTION for I_DO_NOT_EXIST: err: 'stat /iDoNotExist: no such file or directory'",
		},
		{
			Instruction{
				Function:          "FUNCTION",
				Variable:          "NOT_A_PATH",
				CurrentValue:      &notAPath,
				CurrentValueFound: true,
			},
			fmt.Sprintf("path error in FUNCTION for NOT_A_PATH: not a directory '%s'", notAPath),
		},
		{
			Instruction{
				Function:          "FUNCTION",
				Variable:          "TEMPDIR",
				CurrentValue:      &tempDir,
				CurrentValueFound: true,
			},
			"",
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
		err := checkIfDirExists(c.instruction, env)
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

func TestSetIfEmpty(t *testing.T) {
	empty := ""
	tempDir, _ := ioutil.TempDir("", "setIfEmpty")
	defer os.RemoveAll(tempDir)

	env := testEnv{
		"TEMPDIR": &tempDir,
	}
	// non-empty variable
	err := setIfEmpty(Instruction{
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
	err = setIfEmpty(Instruction{
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
	err = setIfEmpty(Instruction{
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
	err = setIfEmpty(Instruction{
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

func TestProcessPreprocessorLines(t *testing.T) {
	empty := ""
	tempDir, _ := ioutil.TempDir("", "processPreprocessorLines")
	defer os.RemoveAll(tempDir)

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
		{
			"DEFECTIVE_CHECK",
			"missing argument(s) in DEFECTIVE_CHECK for , wanted: 1, got: 0",
		},
	}
	preprocessor, err := NewPreprocessor()
	if err != nil {
		t.Error(err)
		return
	}
	err = preprocessor.Register(&Function{
		Names:      []string{"DEFECTIVE_CHECK"},
		NumArgsMin: 1,
	})
	if err != nil {
		t.Error(err)
		return
	}
	for i, c := range cases {
		env := testEnv{
			"NIL":     nil,
			"EMPTY":   &empty,
			"TEMPDIR": &tempDir,
		}
		err := preprocessor.processPreprocessorLines([]string{c.line}, &env)
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

func TestTempDirIfEmpty(t *testing.T) {
	empty := ""
	iDoNotExist := "/iDoNotExist"
	tempFile, err := ioutil.TempFile("", "tempDirIfEmptyFile")
	if err != nil {
		t.Error(err)
		return
	}
	defer os.Remove(tempFile.Name())
	notAPath := tempFile.Name()
	tempDir, err := ioutil.TempDir("", "tempDirIfEmpty")
	if err != nil {
		t.Error(err)
		return
	}
	defer os.RemoveAll(tempDir)

	cases := []struct {
		instruction  Instruction
		errorMessage string
	}{
		{
			Instruction{
				Function:          "FUNCTION",
				Variable:          "EMPTY",
				CurrentValue:      &empty,
				CurrentValueFound: true,
			},
			"",
		},
		{
			Instruction{
				Function:          "FUNCTION",
				Variable:          "TEMPDIR",
				CurrentValue:      &tempDir,
				CurrentValueFound: true,
			},
			"",
		},
		{
			Instruction{
				Function:          "FUNCTION",
				Variable:          "NIL",
				CurrentValue:      nil,
				CurrentValueFound: true,
			},
			"",
		},
		{
			Instruction{
				Function:          "FUNCTION",
				Variable:          "UNDEFINED",
				CurrentValue:      nil,
				CurrentValueFound: false,
			},
			"",
		},
		{
			Instruction{
				Function:          "FUNCTION",
				Variable:          "I_DO_NOT_EXIST",
				CurrentValue:      &iDoNotExist,
				CurrentValueFound: true,
			},
			"path error in FUNCTION for I_DO_NOT_EXIST: err: 'stat /iDoNotExist: no such file or directory'",
		},
		{
			Instruction{
				Function:          "FUNCTION",
				Variable:          "NOT_A_PATH",
				CurrentValue:      &notAPath,
				CurrentValueFound: true,
			},
			fmt.Sprintf("path error in FUNCTION for NOT_A_PATH: not a directory '%s'", notAPath),
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
		err := tempDirIfEmpty(c.instruction, env)
		if len(c.errorMessage) > 0 {
			if err == nil {
				t.Errorf("expected error @%d, got nil", i)
				continue
			}
			if err.Error() != c.errorMessage {
				t.Errorf("incorrect error @%d, got: '%s', wanted: '%s'", i, err, c.errorMessage)
				continue
			}
		} else if err != nil {
			t.Errorf("unexpected error @%d: '%s'", i, err)
		}
	}
	{
		env := testDefectiveEnv{
			"NIL":            nil,
			"EMPTY":          &empty,
			"TEMPDIR":        &tempDir,
			"I_DO_NOT_EXIST": &iDoNotExist,
			"NOT_A_PATH":     &notAPath,
		}
		c := cases[0]
		err = tempDirIfEmpty(c.instruction, env)
		if err == nil {
			t.Error("expected error, got nil")
		}
		if err.Error() != "error" {
			t.Errorf("unexpected error, got: %s, wanted: error", err.Error())
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
