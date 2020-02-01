package server

import (
	"context"
	"os"
	"os/exec"
)

var execCommand = exec.CommandContext

func fakeCommand(ctx context.Context, name string, arg ...string) *exec.Cmd {
	cs := []string{"-test.run=TestHelperProcess", "--", name}
	cs = append(cs, arg...)
	cmd := exec.Command(os.Args[0], cs...)

	return cmd
}

func mockCommand()  { execCommand = fakeCommand }
func resetCommand() { execCommand = exec.CommandContext }
