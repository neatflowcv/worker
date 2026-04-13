package command

import (
	"context"
	"fmt"
	"os"
	"os/exec"
)

const codexExecBaseArgCount = 5

type Runner interface {
	Run(ctx context.Context, dir string, args ...string) error
}

type ExecRunner struct{}

func (r *ExecRunner) Run(ctx context.Context, dir string, args ...string) error {
	commandArgs := make([]string, 0, codexExecBaseArgCount+len(args))
	commandArgs = append(commandArgs,
		"--bun",
		"codex",
		"exec",
		"--sandbox",
		"workspace-write",
	)
	commandArgs = append(commandArgs, args...)

	//nolint:gosec // The executable is fixed and only codex CLI arguments vary.
	command := exec.CommandContext(ctx, "bunx", commandArgs...)
	command.Dir = dir
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr

	err := command.Run()
	if err != nil {
		return fmt.Errorf("run command: %w", err)
	}

	return nil
}
