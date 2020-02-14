package server

import (
	"bytes"
	"context"
	"errors"
	"io"
	"os"
	"os/exec"
	"strings"
	"sync"
	"syscall"
)

type runner struct {
	cfg    *Config
	args   []string
	errors chan error
	cmd    *exec.Cmd
	pid    int
	stderr *bytes.Buffer
	env    []string // for testing
	mu     sync.Mutex
	stop   chan struct{}
}

func newRunner(cfg *Config, args []string) *runner {
	return &runner{
		cfg:    cfg,
		args:   args,
		errors: make(chan error),
		stop:   make(chan struct{}),
	}
}

func (r *runner) run() error {
	r.kill()

	if err := r.execute(); err != nil {
		return err
	}

	if r.cfg.Wait {
		return r.wait()
	} else {
		go func() {
			ignoreError(r.wait())
		}()
	}

	return nil
}

func (r *runner) execute() error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.cmd != nil && r.cmd.ProcessState != nil && r.cmd.ProcessState.Exited() {
		return nil
	}

	r.stderr = &bytes.Buffer{}
	mw := io.MultiWriter(r.stderr, os.Stderr)

	r.cmd = execCommand(context.TODO(), "/bin/sh", "-c", strings.Join(r.args, " "))
	r.cmd.Env = append(r.cmd.Env, r.env...)
	r.cmd.Stdout = os.Stdout
	r.cmd.Stderr = mw

	// Setup a process group so when this process gets stopped, so do any child
	// process that it may spawn.
	r.cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	if err := r.cmd.Start(); err != nil {
		return errors.New(r.stderr.String())
	}
	r.pid = r.cmd.Process.Pid
	return nil
}

// Wait for the command to finish. If the process exits with an error, only log
// it if it exit status is postive, as status code -1 is returned when the
// process was killed by runner#kill.
func (r *runner) wait() error {
	select {
	case <-r.stop:
		return nil
	default:
	}

	r.mu.Lock()
	err := r.cmd.Wait()
	r.mu.Unlock()

	if err != nil {
		if exiterr, ok := err.(*exec.ExitError); ok {
			ws := exiterr.Sys().(syscall.WaitStatus)
			if ws.ExitStatus() > 0 {
				errStr := r.stderr.String()
				if errStr == "" {
					errStr = "non-zero exit (but no output) from subprocess"
				}
				err = errors.New(errStr)

				if r.cfg.Wait {
					return err
				} else {
					r.errors <- err
				}
			}
		}
	}

	return nil
}

// Kill the existing process & process group
func (r *runner) kill() {
	if r.pid > 0 {
		if pgid, err := syscall.Getpgid(r.pid); err == nil {
			ignoreError(syscall.Kill(-pgid, syscall.SIGKILL))
		}

		ignoreError(syscall.Kill(-r.pid, syscall.SIGKILL))

		r.pid = -1
		r.mu.Lock()
		r.cmd = nil
		r.mu.Unlock()
	}
}
