package server

import "testing"

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
