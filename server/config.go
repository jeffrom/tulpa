package server

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/fatih/color"
)

var logPrefixColor = color.New(color.FgMagenta, color.Bold)

func logPrefix() string {
	return logPrefixColor.Sprint("¤")
}

type Config struct {
	AppPort    int
	ProxyPort  int
	IgnoreDirs []string
	Timeout    time.Duration
	Debounce   time.Duration
	// DebouncePoll is the interval between watches while the request debouncer
	// is saturated. It will be disabled if <= 0.
	DebouncePoll time.Duration
	Wait         bool
	Verbose      bool
	stdout       io.Writer
	stderr       io.Writer
}

func (c *Config) Initialize() {
	if c.stdout == nil {
		c.stdout = os.Stdout
	}
	if c.stderr == nil {
		c.stderr = os.Stderr
	}
}

func (c *Config) Print(args ...interface{}) {
	fmt.Fprintln(c.stdout, append([]interface{}{logPrefix()}, args...)...)
}

func (c *Config) Printf(msg string, args ...interface{}) {
	finalMsg := fmt.Sprintf("%s %s\n", logPrefix(), msg)
	fmt.Fprintf(c.stdout, finalMsg, args...)
}

func (c *Config) Debug(args ...interface{}) {
	if !c.Verbose {
		return
	}
	fmt.Fprintln(c.stdout, append([]interface{}{logPrefix()}, args...)...)
}

func (c *Config) Debugf(msg string, args ...interface{}) {
	if !c.Verbose {
		return
	}
	finalMsg := fmt.Sprintf("%s %s\n", logPrefix(), msg)
	fmt.Fprintf(c.stdout, finalMsg, args...)
}
