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

type Environment interface {
	GetSubstitution(string) (*string, bool)
	SetSubstitution(string, *string)
	GetOrCreateTempDir(string) (string, error)
}

const setIfEmpty = "SET_IF_EMPTY"
const checkIfDirExists = "CHECK_IF_DIR_EXISTS"
const tempDirIfEmpty = "TEMP_DIR_IF_EMPTY"

type Preprocessor struct {
}

func NewPreprocessor() *Preprocessor {
	return &Preprocessor{}
}

func (*Preprocessor) Register() error {
	return nil
}

type PreprocessorInstruction struct {
	Function          string
	Variable          string
	Arguments         []string
	CurrentValue      *string
	CurrentValueFound bool
}

func NewPreprocessorInstruction(line string, env Environment) (*PreprocessorInstruction, error) {
	result := &PreprocessorInstruction{}
	parts := strings.Split(line, " ")
	if len(parts[0]) < 1 {
		return result, fmt.Errorf("empty preprocessor line found!")
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

func processCheckIfDirExists(inst *PreprocessorInstruction, env Environment) error {
	if len(inst.Variable) == 0 {
		return fmt.Errorf("missing variable in %s", inst.Function)
	}
	if len(inst.Arguments) > 0 {
		return fmt.Errorf("too many arguments in %s for %s", inst.Function, inst.Variable)
	}
	if !inst.CurrentValueFound || inst.CurrentValue == nil || len(*inst.CurrentValue) < 1 {
		return fmt.Errorf("empty variable in %s for %s", inst.Function, inst.Variable)
	}
	path, err := filepath.Abs(*inst.CurrentValue)
	if err != nil {
		return fmt.Errorf("path error in %s for %s: err: '%s'", inst.Function, inst.Variable, err)
	}
	fi, err := os.Stat(path)
	if os.IsNotExist(err) {
		return fmt.Errorf("path error in %s for %s: err: '%s'", inst.Function, inst.Variable, err)
	}
	if !fi.Mode().IsDir() {
		return fmt.Errorf("path error in %s for %s: not a directory '%s'", inst.Function, inst.Variable, path)
	}
	return nil
}

func processSetIfEmpty(inst *PreprocessorInstruction, env Environment) error {
	if len(inst.Variable) == 0 {
		return fmt.Errorf("missing variable in %s", inst.Function)
	}
	if len(inst.Arguments) == 0 {
		return fmt.Errorf("missing argument in %s for %s", inst.Function, inst.Variable)
	}
	if len(inst.Arguments) > 1 {
		return fmt.Errorf("too many arguments in %s for %s", inst.Function, inst.Variable)
	}
	// If environment variable set and not empty, do not create it!
	if inst.CurrentValueFound && inst.CurrentValue != nil && len(*inst.CurrentValue) > 0 {
		return nil
	}
	env.SetSubstitution(inst.Variable, &inst.Arguments[0])
	return nil
}

func processTempDirIfEmpty(inst *PreprocessorInstruction, env Environment) error {
	if len(inst.Variable) == 0 {
		return fmt.Errorf("missing variable in %s", inst.Function)
	}
	if len(inst.Arguments) > 0 {
		return fmt.Errorf("too many arguments in %s for %s", inst.Function, inst.Variable)
	}
	if inst.CurrentValueFound && inst.CurrentValue != nil && len(*inst.CurrentValue) > 0 {
		path, err := filepath.Abs(*inst.CurrentValue)
		if err != nil {
			return fmt.Errorf("path error in %s for %s: err: '%s'", inst.Function, inst.Variable, err)
		}
		fi, err := os.Stat(path)
		if os.IsNotExist(err) {
			return fmt.Errorf("path error in %s for %s: err: '%s'", inst.Function, inst.Variable, err)
		}
		if !fi.Mode().IsDir() {
			return fmt.Errorf("path error in %s for %s: not a directory '%s'", inst.Function, inst.Variable, path)
		}
	}
	// We have an empty value or a new variable: create the directory.
	path, err := env.GetOrCreateTempDir(inst.Variable)
	if err != nil {
		return err
	}
	env.SetSubstitution(inst.Variable, &path)
	return nil
}

func processPreprocessorLines(lines []string, env Environment) error {
	for _, line := range lines {
		instruction, err := NewPreprocessorInstruction(line, env)
		if err != nil {
			return err
		}
		switch instruction.Function {
		case checkIfDirExists:
			return processCheckIfDirExists(instruction, env)
		case setIfEmpty:
			return processSetIfEmpty(instruction, env)
		case tempDirIfEmpty:
			return processTempDirIfEmpty(instruction, env)
		default:
			return fmt.Errorf("unknown preprocessor directive: '%s'", instruction.Function)
		}
	}
	return nil
}

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

func PreprocessYAML(rawFile []byte, env Environment) ([]byte, error) {
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
	if err := processPreprocessorLines(preprocessor, env); err != nil {
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
