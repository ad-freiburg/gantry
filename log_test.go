package gantry_test

import (
	"bufio"
	"bytes"
	"fmt"
	"log"
	"testing"

	"github.com/ad-freiburg/gantry"
)

func TestColorStoreNextColor(t *testing.T) {
	var cases = []struct {
		store   *gantry.ColorStore
		results []int
	}{
		{gantry.NewColorStore([]int{42}), []int{42, 42, 42, 42, 42}},
		{gantry.NewColorStore([]int{21, 42}), []int{21, 42, 21, 42, 21}},
	}

	for _, c := range cases {
		for i, v := range c.results {
			color := c.store.NextColor()
			if color != v {
				t.Errorf("Incorrect color for index %d, got: %d, wanted: %d", i, color, v)
			}
		}
	}
}

func TestPrefixedWriterWrite(t *testing.T) {
	input := []byte("Hello World")
	buf := bytes.NewBuffer([]byte(""))
	writer := bufio.NewWriter(buf)
	pw := gantry.NewPrefixedWriter("prefix", writer)
	n, err := pw.Write(input)
	if err != nil {
		t.Errorf("Got unexpected errror: %#v", err)
	}
	if n != len(input) {
		t.Errorf("Incorrect number of bytes written, got: '%d', wanted: '%d'", n, len(input))
	}
	writer.Flush()
	expected := fmt.Sprintf(gantry.PrefixedWriterFormat, "prefix", string(input))
	result := buf.String()
	if result != expected {
		t.Errorf("Incorrect buffer contents, got: '%s', wanted: '%s'", result, expected)
	}
	input2 := []byte("!\nSecond line")
	n, err = pw.Write(input2)
	writer.Flush()
	if err != nil {
		t.Errorf("Got unexpected errror: %#v", err)
	}
	if n != len(input2) {
		t.Errorf("Incorrect number of bytes written, got: '%d', wanted: '%d'", n, len(input2))
	}
	expected = fmt.Sprintf(
		"%s%s%s",
		fmt.Sprintf(gantry.PrefixedWriterFormat, "prefix", string(input)),
		fmt.Sprintf(gantry.PrefixedWriterFormat, "prefix", "!\n"),
		fmt.Sprintf(gantry.PrefixedWriterFormat, "prefix", "Second line"))
	result = buf.String()
	if result != expected {
		t.Errorf("Incorrect buffer contents, got: '%s', wanted: '%s'", result, expected)
	}
}

func TestPrefixedLoggerPrintf(t *testing.T) {
	buf := bytes.NewBuffer([]byte(""))
	writer := bufio.NewWriter(buf)
	// Create new logger without date and time
	logger := gantry.NewPrefixedLogger("prefix", log.New(writer, "", 0))
	logger.Printf("%s:%d", "Answer", 42)
	writer.Flush()
	result := buf.String()
	expected := fmt.Sprintf(gantry.PrefixedWriterFormat, "prefix", "Answer:42") + "\n"
	if result != expected {
		t.Errorf("Incorrect buffer contents, got: '%s', wanted: '%s'", result, expected)
	}
}

func TestPrefixedLoggerPrintln(t *testing.T) {
	buf := bytes.NewBuffer([]byte(""))
	writer := bufio.NewWriter(buf)
	// Create new logger without date and time
	logger := gantry.NewPrefixedLogger("prefix", log.New(writer, "", 0))
	logger.Println("A line!")
	writer.Flush()
	result := buf.String()
	expected := fmt.Sprintf(gantry.PrefixedWriterFormat, "prefix", "A line!\n") + "\n"
	if result != expected {
		t.Errorf("Incorrect buffer contents, got: '%s', wanted: '%s'", result, expected)
	}
}

func TestPrefixedLoggerWrite(t *testing.T) {
	buf := bytes.NewBuffer([]byte(""))
	writer := bufio.NewWriter(buf)
	// Create new logger without date and time
	logger := gantry.NewPrefixedLogger("prefix", log.New(writer, "", 0))
	logger.Write([]byte("A\nB\n"))
	writer.Flush()
	result := buf.String()
	expected := fmt.Sprintf(gantry.PrefixedWriterFormat, "prefix", "A") + "\n" + fmt.Sprintf(gantry.PrefixedWriterFormat, "prefix", "B") + "\n"
	if result != expected {
		t.Errorf("Incorrect buffer contents, got: '%s', wanted: '%s'", result, expected)
		t.Errorf("Incorrect buffer contents, got: '%#v', wanted: '%#v'", result, expected)
	}
}
