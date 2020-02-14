package server

import (
	"bytes"
	"fmt"
	"os"
	"os/signal"
	"regexp"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"
)

func TestHelperProcess(t *testing.T) { fakeProcess() }

const assertFailedCode = 119

func fakeProcess() {
	if os.Getenv("_FAKEPROC_WANT_HELPER_PROCESS") != "1" {
		return
	}

	codes := os.Getenv("_FAKEPROC_EXITCODE")
	code, err := strconv.ParseInt(codes, 10, 8)
	if err != nil {
		if codes == "" {
			code = 0
		} else {
			panic(err)
		}
	}
	defer os.Exit(int(code))

	var expectSig bool
	var gotSig bool
	sigch := make(chan os.Signal, 1)
	if arg := os.Getenv("_FAKEPROC_EXPECT_SIGTERM"); arg != "" {
		expectSig = true
		signal.Notify(sigch, syscall.SIGTERM)
	}

	args := os.Args
	for len(args) > 0 {
		if args[0] == "--" {
			args = args[1:]
			break
		}
		args = args[1:]
	}

	argsStr := argsString(args)
	if expectArg := os.Getenv("_FAKEPROC_EXPECT_ARG"); expectArg != "" {
		if argsStr != expectArg {
			fmt.Fprintf(os.Stderr, "fakeprocess: assertion failed:\nexpected: '%s',\n     got: '%s'\n", expectArg, argsStr)
			os.Exit(assertFailedCode)
		}
	}

	if expectArg := os.Getenv("_FAKEPROC_EXPECT_ARG_SUFFIX"); expectArg != "" {
		if !strings.HasSuffix(argsStr, expectArg) {
			fmt.Fprintf(os.Stderr, "fakeprocess: assertion failed:\nexpected suffix: '%s',\n     got: '%s'\n", expectArg, argsStr)
			os.Exit(assertFailedCode)
		}
	}

	if expectArg := os.Getenv("_FAKEPROC_EXPECT_ARG_SUFFIX"); expectArg != "" {
		if match, err := regexp.MatchString(expectArg, argsStr); !match || err != nil {
			fmt.Fprintf(os.Stderr, "fakeprocess: assertion failed:\nexpected suffix: '%s',\n     got: '%s'\n(err: %v)\n", expectArg, argsStr, err)
			os.Exit(assertFailedCode)
		}
	}

	fmt.Fprint(os.Stderr, os.Getenv("_FAKEPROC_STDERR"))
	fmt.Fprint(os.Stdout, os.Getenv("_FAKEPROC_STDOUT"))

	if arg := os.Getenv("_FAKEPROC_SLEEP"); arg != "" {
		parsed, err := time.ParseDuration(arg)
		if err != nil {
			panic(err)
		}
		select {
		case <-sigch:
			gotSig = true
		case <-time.After(parsed):
		}
	}

	if expectSig && !gotSig {
		fmt.Fprintf(os.Stderr, "fakeprocess: assertion failed: expected SIGTERM/SIGINT\n")
		os.Exit(assertFailedCode)
	}
}

func argsString(args []string) string {
	b := &bytes.Buffer{}

	for i, arg := range args {
		if strings.Contains(arg, " ") {
			b.WriteString(`"`)
			b.WriteString(arg)
			b.WriteString(`"`)
		} else {
			b.WriteString(arg)
		}

		if i < len(args)-1 {
			b.WriteString(" ")
		}
	}

	return b.String()
}
