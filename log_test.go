package gantry_test

import (
	"bufio"
	"bytes"
	"fmt"
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
}
