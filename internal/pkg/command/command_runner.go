package command

import (
	"context"
	"fmt"
	"os"
	"os/exec"
)

type Runner interface {
	Run(ctx context.Context, dir string, prompt string) error
}

type ExecRunner struct{}

func (r *ExecRunner) Run(ctx context.Context, dir string, prompt string) error {
	//nolint:gosec // The command is fixed; only the prompt argument is variable.
	command := exec.CommandContext(
		ctx,
		"bunx",
		"--bun",
		"codex",
		"exec",
		"--sandbox",
		"workspace-write",
		prompt,
	)
	command.Dir = dir
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr

	err := command.Run()
	if err != nil {
		return fmt.Errorf("run command: %w", err)
	}

	return nil
}
