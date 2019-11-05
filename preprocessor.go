package gantry

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func extractPreprocessorStatements(lines []string) ([]string, []string) {
	preprocessorStatements := []string{}
	normalLines := []string{}
	for _, line := range lines {
		trimmedLine := strings.TrimSpace(line)
		if len(trimmedLine) < 2 || trimmedLine[0] != '#' {
			normalLines = append(normalLines, line)
			continue
		}
		if trimmedLine[1] != '!' {
			continue
		}
		preprocessorStatements = append(preprocessorStatements, strings.TrimSpace(trimmedLine[2:]))
	}
	return preprocessorStatements, normalLines
}

func processPreprocessorStatements(statements []string, env *PipelineEnvironment) error {
	for _, statement := range statements {
		parts := strings.Split(statement, " ")
		if len(parts[0]) < 1 {
			return fmt.Errorf("empty preprocessor statement found!")
		}
		if len(parts) < 2 {
			return fmt.Errorf("missing variable name for: '%s'", statement)
		}
		// Lookup substitution value
		key := parts[1][2 : len(parts[1])-1]
		val, ok := env.Substitutions[key]
		log.Printf("%#v", parts)
		switch parts[0] {
		case "CHECK_IF_DIR_EXISTS":
			// If environment variable is empty or not set, abort!
			if !ok || val == nil || len(*val) < 1 {
				return fmt.Errorf("empty variable for CHECK_IF_DIR_EXISTS: '%s'", key)
			}
			path, err := filepath.Abs(*val)
			if err != nil {
				return fmt.Errorf("path error for CHECK_IF_DIR_EXISTS '%s': err: '%s'", key, err)
			}
			fi, err := os.Stat(path)
			if os.IsNotExist(err) {
				return fmt.Errorf("path error for CHECK_IF_DIR_EXISTS '%s': err: '%s'", key, err)
			}
			if !fi.Mode().IsDir() {
				return fmt.Errorf("path error for CHECK_IF_DIR_EXISTS '%s': not a directory '%s'", key, path)
			}
		case "SET_IF_EMPTY":
			if len(parts) < 3 {
				return fmt.Errorf("invalid syntax SET_IF_EMPTY '%s': missing value", key)
			}
			// If environment variable set and not empty, do not create it!
			if ok && val != nil && len(*val) > 0 {
				continue
			}
			env.Substitutions[key] = &parts[2]
		case "TEMP_DIR":
			// If environment variable for temp dir is set, do not create it!
			if ok && val != nil && len(*val) > 0 {
				path, err := filepath.Abs(*val)
				if err != nil {
					return fmt.Errorf("path error for TEMP_DIR '%s': err: '%s'", key, err)
				}
				fi, err := os.Stat(path)
				if os.IsNotExist(err) {
					return fmt.Errorf("path error for TEMP_DIR '%s': err: '%s'", key, err)
				}
				if !fi.Mode().IsDir() {
					return fmt.Errorf("path error for CHECK_IF_DIR_EXISTS '%s': not a directory", key)
				}
			}
			// We have an empty value or a new variable: create the directory.
			path, err := env.getOrCreateTempDir(key)
			if err != nil {
				return err
			}
			env.Substitutions[key] = &path
		default:
			return fmt.Errorf("unknown preprocessor directive: '%s'", parts[0])
		}
	}
	return nil
}

func expandVariables(lines []string, env *PipelineEnvironment) []string {
	expandFunc := func(placeholder string) string {
		val, ok := env.Substitutions[placeholder]
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

func PreprocessYAML(rawFile []byte, env *PipelineEnvironment) ([]byte, error) {
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
	statements, normalLines := extractPreprocessorStatements(lines)
	if err := processPreprocessorStatements(statements, env); err != nil {
		return []byte(""), err
	}
	lines = expandVariables(normalLines, env)
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
