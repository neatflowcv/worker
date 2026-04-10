package cli

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/alecthomas/kong"
	"github.com/neatflowcv/worker/internal/app/flow"
	projectbadger "github.com/neatflowcv/worker/internal/pkg/badger"
	"github.com/neatflowcv/worker/internal/pkg/local"
)

func Run() error {
	projectRepository, err := newProjectRepository()
	if err != nil {
		return fmt.Errorf("create project repository: %w", err)
	}

	defer func() {
		_ = projectRepository.Close()
	}()

	projectWorkspace, err := newWorkspace()
	if err != nil {
		return fmt.Errorf("create workspace: %w", err)
	}

	return RunWithArgs(
		context.Background(),
		os.Args[1:],
		os.Stdout,
		flow.NewService(projectRepository, projectWorkspace),
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

func newProjectRepository() (*projectbadger.ProjectRepository, error) {
	dataHome, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("resolve user home directory: %w", err)
	}

	projectRepository, err := projectbadger.NewProjectRepository(
		filepath.Join(dataHome, ".local", "share", "worker", "projects.badger"),
	)
	if err != nil {
		return nil, fmt.Errorf("open project repository: %w", err)
	}

	return projectRepository, nil
}

func newWorkspace() (*local.Workspace, error) {
	dataHome, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("resolve user home directory: %w", err)
	}

	return local.NewWorkspace(
		filepath.Join(dataHome, ".local", "share", "worker", "projects"),
	), nil
}
