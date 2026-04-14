package codex

import (
	"context"
	"fmt"
	"io"
	"os/exec"
)

const codexExecBaseArgCount = 5

type Runner struct{}

func (r *Runner) Run(ctx context.Context, dir string, args ...string) ([]byte, error) {
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
	command.Stderr = io.Discard

	output, err := command.Output()
	if err != nil {
		return nil, fmt.Errorf("run command: %w", err)
	}

	return output, nil
}
