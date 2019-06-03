package gantry // import "github.com/ad-freiburg/gantry"

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"strings"
	"sync"
)

// PrefixedWriterFormat provides formatting for to space separated strings
// resetting ANSI formatting after each string.
const PrefixedWriterFormat string = "%s\u001b[0m %s\u001b[0m"

// GenericStringFormat provides formatting for a single string enclodes in
// ANSI formatting tags.
const GenericStringFormat string = "\u001b[%sm%s\u001b[0m"

// AnsiStyleNormal stores int value for normal style
const AnsiStyleNormal int = 0

// AnsiStyleBold stores int value for bold style
const AnsiStyleBold int = 1

// AnsiStyleDim stores int value for dim style
const AnsiStyleDim int = 2

// AnsiStyleItalic stores int value for italic style
const AnsiStyleItalic int = 3

// AnsiStyleUnderline stores int value for underline style
const AnsiStyleUnderline int = 4

// AnsiStyleStrikethrough stores int value for strikethrough style
const AnsiStyleStrikethrough int = 9

// AnsiForegroundColorBlack stores int value for black foreground
const AnsiForegroundColorBlack int = 30

// AnsiForegroundColorRed stores int value for red foreground
const AnsiForegroundColorRed int = 31

// AnsiForegroundColorGreen stores int value for green foreground
const AnsiForegroundColorGreen int = 32

// AnsiForegroundColorYellow stores int value for yellow foreground
const AnsiForegroundColorYellow int = 33

// AnsiForegroundColorBlue stores int value for blue foreground
const AnsiForegroundColorBlue int = 34

// AnsiForegroundColorMagenta stores int value for magenta foreground
const AnsiForegroundColorMagenta int = 35

// AnsiForegroundColorCyan stores int value for cyan foreground
const AnsiForegroundColorCyan int = 36

// AnsiForegroundColorLightGray stores int value for light gray foreground
const AnsiForegroundColorLightGray int = 37

// AnsiForegroundColorDarkGray stores int value for dark gray foreground
const AnsiForegroundColorDarkGray int = 90

// AnsiForegroundColorLightRed stores int value for light red foreground
const AnsiForegroundColorLightRed int = 91

// AnsiForegroundColorLightGreen stores int value for light green foreground
const AnsiForegroundColorLightGreen int = 92

// AnsiForegroundColorLightYellow stores int value for light yellow foreground
const AnsiForegroundColorLightYellow int = 93

// AnsiForegroundColorLightBlue stores int value for light blue foreground
const AnsiForegroundColorLightBlue int = 94

// AnsiForegroundColorLightMagenta stores int value for light magenta foreground
const AnsiForegroundColorLightMagenta int = 95

// AnsiForegroundColorLightCyan stores int value for light cyan foreground
const AnsiForegroundColorLightCyan int = 96

// AnsiForegroundColorWhite stores int value for white foreground
const AnsiForegroundColorWhite int = 97

var (
	friendlyColors *ColorStore
)

func init() {
	friendlyColors = NewColorStore([]int{AnsiForegroundColorCyan, AnsiForegroundColorYellow, AnsiForegroundColorMagenta, AnsiForegroundColorBlue, AnsiForegroundColorLightCyan, AnsiForegroundColorLightYellow, AnsiForegroundColorLightMagenta, AnsiForegroundColorLightBlue})
}

type LogSettings struct {
}

// ApplyAnsiStyle applies ansi integer styles to text
func ApplyAnsiStyle(text string, style ...int) string {
	return fmt.Sprintf(GenericStringFormat, BuildAnsiStyle(style...), text)
}

// BuildAnsiStyle combines style integers with ;
func BuildAnsiStyle(parts ...int) string {
	return strings.Trim(strings.Replace(fmt.Sprint(parts), " ", ";", -1), "[]")
}

// GetNextFriendlyColor returns the next friendly color from the global
// friendlyColors store.
func GetNextFriendlyColor() int {
	return friendlyColors.NextColor()
}

// ColorStore provides a synced looping stylelist.
type ColorStore struct {
	index  int
	colors []int
	m      sync.Mutex
}

// NewColorStore creates a ColorStore for the provided list of styles.
func NewColorStore(colors []int) *ColorStore {
	return &ColorStore{
		index:  -1,
		colors: colors,
	}
}

// NextColor returns the next color. Loops if list is exhausted.
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
