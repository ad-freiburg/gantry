package ymlpreprocessor

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// Environment represents substitutions and tmp dirs
type Environment interface {
	GetSubstitution(string) (*string, bool)
	SetSubstitution(string, *string)
	GetOrCreateTempDir(string) (string, error)
}

func checkIfDirExists(i Instruction, e Environment) error {
	path, err := filepath.Abs(*i.CurrentValue)
	if err != nil {
		return fmt.Errorf("path error in %s for %s: err: '%s'", i.Function, i.Variable, err)
	}
	fi, err := os.Stat(path)
	if os.IsNotExist(err) {
		return fmt.Errorf("path error in %s for %s: err: '%s'", i.Function, i.Variable, err)
	}
	if !fi.Mode().IsDir() {
		return fmt.Errorf("path error in %s for %s: not a directory '%s'", i.Function, i.Variable, path)
	}
	return nil

}

func setIfEmpty(i Instruction, e Environment) error {
	// If environment variable set and not empty, do not create it!
	if i.CurrentValueFound && i.CurrentValue != nil && len(*i.CurrentValue) > 0 {
		return nil
	}
	e.SetSubstitution(i.Variable, &i.Arguments[0])
	return nil
}

func tempDirIfEmpty(i Instruction, e Environment) error {
	if i.CurrentValueFound && i.CurrentValue != nil && len(*i.CurrentValue) > 0 {
		if err := checkIfDirExists(i, e); err != nil {
			return err
		}
	}
	// We have an empty value or a new variable: create the directory.
	path, err := e.GetOrCreateTempDir(i.Variable)
	if err != nil {
		return err
	}
	e.SetSubstitution(i.Variable, &path)
	return nil
}

// Preprocessor preprocesses yml files and manipulates the environment.
type Preprocessor struct {
	mapping   map[string]*Function
	functions []*Function
}

// NewPreprocessor returns a new Preprocessor with basic functions preregistered.
func NewPreprocessor() (Preprocessor, error) {
	p := Preprocessor{
		mapping:   map[string]*Function{},
		functions: []*Function{},
	}
	if err := p.Register(&Function{
		Names: []string{
			"CHECK_IF_DIR_EXISTS",
			"check_if_dir_exists",
		},
		NeedsVariable: true,
		Func:          checkIfDirExists,
		Description:   "Checks if ${VAR} points to a directory, aborts execution on failure.",
	}); err != nil {
		return p, err
	}
	if err := p.Register(&Function{
		Names: []string{
			"SET_IF_EMPTY",
			"set_if_empty",
		},
		NeedsVariable: true,
		NumArgsMin:    1,
		NumArgsMax:    1,
		Func:          setIfEmpty,
		Description:   "Sets ${VAR} to ARG0 if ${VAR} is empty or not set.",
	}); err != nil {
		return p, err
	}
	if err := p.Register(&Function{
		Names: []string{
			"TEMP_DIR_IF_EMPTY",
			"temp_dir_if_empty",
			"mktemp",
		},
		NeedsVariable: true,
		Func:          tempDirIfEmpty,
	}); err != nil {
		return p, err
	}
	return p, nil
}

// Register adds a PreprocessorFunction, raises error if a name is already in use.
func (p *Preprocessor) Register(f *Function) error {
	for _, name := range f.Names {
		if v, found := p.mapping[name]; found {
			return fmt.Errorf("name %s for function %s already defined by %s", name, f.Names, v.Names)
		}
		p.mapping[name] = f
	}
	p.functions = append(p.functions, f)
	return nil
}

// Functions returns all registered functions.
func (p Preprocessor) Functions() []*Function {
	return p.functions
}

// Process processes a raw file with a given environment.
func (p Preprocessor) Process(rawFile []byte, env Environment) ([]byte, error) {
	// Parse bytes as lines
	var lines []string
	var lineBytesBuffer bytes.Buffer
	r := bufio.NewReader(bytes.NewReader(rawFile))
	for {
		lineBytes, prefix, err := r.ReadLine()
		if err != nil {
			if err == io.EOF {
				break
			}
			return []byte(""), err
		}
		lineBytesBuffer.Write(lineBytes)
		// Line continues, continue reading before storing
		if prefix {
			continue
		}
		lines = append(lines, lineBytesBuffer.String())
		lineBytesBuffer.Reset()
	}
	// Run preprocessor steps
	preprocessor, normal := extractPreprocessorLines(lines)
	if err := p.processPreprocessorLines(preprocessor, env); err != nil {
		return []byte(""), err
	}
	lines = expandVariables(normal, env)
	// Reconvert to byte slice
	var b bytes.Buffer
	bw := bufio.NewWriter(&b)
	for i, line := range lines {
		if i > 0 {
			if _, err := bw.WriteString("\n"); err != nil {
				return []byte(""), err
			}
		}
		if _, err := bw.WriteString(line); err != nil {
			return []byte(""), err
		}
	}
	bw.Flush()
	return b.Bytes(), nil
}

// processPreprocessorLines executes each `#!` line
func (p Preprocessor) processPreprocessorLines(lines []string, env Environment) error {
	for _, line := range lines {
		instruction, err := NewInstruction(line, env)
		if err != nil {
			return err
		}
		if f, ok := p.mapping[instruction.Function]; ok {
			if err := f.Execute(instruction, env); err != nil {
				return err
			}
		} else {
			return fmt.Errorf("unknown preprocessor directive: '%s'", instruction.Function)
		}
	}
	return nil
}

// extractPreprocessorLines splits lines in two lists, preprocessor instructions and normal lines.
func extractPreprocessorLines(lines []string) ([]string, []string) {
	preprocessor := []string{}
	normal := []string{}
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if len(trimmed) < 2 || trimmed[0] != '#' {
			normal = append(normal, line)
			continue
		}
		if trimmed[1] != '!' {
			continue
		}
		preprocessor = append(preprocessor, strings.TrimSpace(trimmed[2:]))
	}
	return preprocessor, normal
}

// expandVariables expands variables in all lines
func expandVariables(lines []string, env Environment) []string {
	expandFunc := func(placeholder string) string {
		val, ok := env.GetSubstitution(placeholder)
		if ok && val != nil {
			return *val
		}
		// No real substitution found, return empty string
		return ""
	}
	result := make([]string, len(lines))
	for i, l := range lines {
		result[i] = os.Expand(l, expandFunc)
	}
	return result
}
