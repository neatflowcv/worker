package cli

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/alecthomas/kong"
	"github.com/neatflowcv/worker/internal/app/flow"
	"github.com/neatflowcv/worker/internal/pkg/memory"
)

func Run() error {
	return RunWithArgs(
		context.Background(),
		os.Args[1:],
		os.Stdout,
		flow.NewService(memory.NewProjectRepository()),
	)
}

func RunWithArgs(ctx context.Context, args []string, stdout io.Writer, service *flow.Service) error {
	parser, err := kong.New(
		newApp(),
		kong.Name("worker"),
		kong.BindTo(ctx, (*context.Context)(nil)),
		kong.BindTo(stdout, (*io.Writer)(nil)),
		kong.Bind(service),
	)
	if err != nil {
		return fmt.Errorf("create CLI parser: %w", err)
	}

	kctx, err := parser.Parse(args)
	if err != nil {
		return fmt.Errorf("parse CLI arguments: %w", err)
	}

	err = kctx.Run()
	if err != nil {
		return fmt.Errorf("run CLI command: %w", err)
	}

	return nil
}
