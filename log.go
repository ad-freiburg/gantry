package gantry // import "github.com/ad-freiburg/gantry"

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"strings"
	"sync"
)

const PrefixedWriterFormat string = "%s\u001b[0m %s\u001b[0m"
const GenericStringFormat string = "\u001b[%sm%s\u001b[0m"
const STYLE_NORMAL int = 0
const STYLE_BOLD int = 1
const STYLE_DIM int = 2
const STYLE_ITALIC int = 3
const STYLE_UNDERLINE int = 4
const STYLE_STRIKETHROUGH int = 9

const FG_COLOR_BLACK int = 30
const FG_COLOR_RED int = 31
const FG_COLOR_GREEN int = 32
const FG_COLOR_YELLOW int = 33
const FG_COLOR_BLUE int = 34
const FG_COLOR_MAGENTA int = 35
const FG_COLOR_CYAN int = 36
const FG_COLOR_LIGHT_GRAY int = 37
const FG_COLOR_DARK_GRAY int = 90
const FG_COLOR_LIGHT_RED int = 91
const FG_COLOR_LIGHT_GREEN int = 92
const FG_COLOR_LIGHT_YELLOW int = 93
const FG_COLOR_LIGHT_BLUE int = 94
const FG_COLOR_LIGHT_MAGENTA int = 95
const FG_COLOR_LIGHT_CYAN int = 96
const FG_COLOR_WHITE int = 97

var (
	friendlyColors *ColorStore
)

func init() {
	friendlyColors = NewColorStore([]int{FG_COLOR_CYAN, FG_COLOR_YELLOW, FG_COLOR_MAGENTA, FG_COLOR_BLUE, FG_COLOR_LIGHT_CYAN, FG_COLOR_LIGHT_YELLOW, FG_COLOR_LIGHT_MAGENTA, FG_COLOR_LIGHT_BLUE})
}

func ApplyStyle(text string, style ...int) string {
	return fmt.Sprintf(GenericStringFormat, BuildPrefixStyle(style...), text)
}

func BuildPrefixStyle(parts ...int) string {
	return strings.Trim(strings.Replace(fmt.Sprint(parts), " ", ";", -1), "[]")
}

func GetNextFriendlyColor() int {
	return friendlyColors.NextColor()
}

type ColorStore struct {
	index  int
	colors []int
	m      sync.Mutex
}

func NewColorStore(colors []int) *ColorStore {
	return &ColorStore{
		index:  -1,
		colors: colors,
	}
}

func (c *ColorStore) NextColor() int {
	defer c.m.Unlock()
	c.m.Lock()
	c.index++
	if c.index >= len(c.colors) {
		c.index = 0
	}
	return c.colors[c.index]
}

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
			fmt.Fprintf(p.target, PrefixedWriterFormat, p.prefix, line)
			break
		}
		if err != nil {
			return err
		}
		fmt.Fprintf(p.target, PrefixedWriterFormat, p.prefix, line)
	}
	return nil
}

type PrefixedLogger struct {
	prefix string
	logger *log.Logger
}

func NewPrefixedLogger(prefix string, logger *log.Logger) *PrefixedLogger {
	return &PrefixedLogger{
		prefix: prefix,
		logger: logger,
	}
}

func (p *PrefixedLogger) Printf(format string, v ...interface{}) {
	p.logger.Output(2, fmt.Sprintf(PrefixedWriterFormat, p.prefix, fmt.Sprintf(format, v...)))
}

func (p *PrefixedLogger) Println(v ...interface{}) {
	p.logger.Output(2, fmt.Sprintf(PrefixedWriterFormat, p.prefix, fmt.Sprintln(v...)))
}

func (p *PrefixedLogger) Write(b []byte) (int, error) {
	n := len(b)
	if n > 0 && b[n-1] == '\n' {
		b = b[:n-1]
	}
	for _, s := range strings.Split(string(b), "\n") {
		p.logger.Output(2, fmt.Sprintf(PrefixedWriterFormat, p.prefix, s))
	}
	return n, nil
}
