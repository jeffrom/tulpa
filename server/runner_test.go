package server

import (
	"os"
	"strings"
	"testing"
)

func TestRunner(t *testing.T) {
	mockCommand()
	defer resetCommand()

	cfg := newTestConfig()
	cfg.Wait = true
	runner := newRunner(cfg, []string{"cool"})
	if err := runner.run(); err != nil {
		t.Fatal(err)
	}
}

func TestRunnerFail(t *testing.T) {
	mockCommand()
	defer resetCommand()

	cfg := newTestConfig()
	cfg.Wait = true
	runner := newRunner(cfg, []string{"cool"})
	runner.env = []string{"_FAKEPROC_EXITCODE=1", "_FAKEPROC_STDERR=cool error"}
	err := runner.run()
	if err == nil {
		t.Fatal("expected error but got none")
	}

	if err.Error() != "cool error" {
		t.Fatal("expected cool error, got", err)
	}

	runner.env = []string{"_FAKEPROC_EXITCODE=1"}
	err = runner.run()
	if err == nil {
		t.Fatal("expected error but got none")
	}

	if !strings.Contains(err.Error(), "no output") {
		t.Fatal("expected no output from subprocess, got", err)
	}
}

func setEnv(environ []string) func() {
	var unsetEnv []string
	oldEnv := make(map[string]string)
	for _, env := range environ {
		parts := strings.SplitN(env, "=", 2)
		k, v := parts[0], parts[1]
		if _, ok := os.LookupEnv(k); !ok {
			unsetEnv = append(unsetEnv, k)
		} else {
			oldEnv[k] = os.Getenv(k)
		}
		if err := os.Setenv(k, v); err != nil {
			panic(err)
		}
	}
	return func() {
		for _, unset := range unsetEnv {
			os.Unsetenv(unset)
		}
		for k, v := range oldEnv {
			os.Setenv(k, v)
		}
	}
}
