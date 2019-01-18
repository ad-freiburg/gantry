package gantry // import "github.com/ad-freiburg/gantry"

import (
	"bytes"
	"fmt"
	"io"
)

const prefixedWriterFormat string = "\u001b[1m%s\u001b[0m %s\u001b[0m"

type PrefixedWriter struct {
	prefix string
	target io.Writer
	buf    *bytes.Buffer
}

func NewPrefixedWriter(prefix string, target io.Writer) *PrefixedWriter {
	return &PrefixedWriter{
		prefix: prefix,
		target: target,
		buf:    bytes.NewBuffer([]byte("")),
	}
}

func (p *PrefixedWriter) Write(b []byte) (int, error) {
	n, err := p.buf.Write(b)
	if err != nil {
		return n, err
	}
	err = p.Output()
	return n, err
}

func (p *PrefixedWriter) Output() error {
	for {
		line, err := p.buf.ReadString('\n')
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		fmt.Fprintf(p.target, prefixedWriterFormat, p.prefix, line)
	}
	return nil
}
