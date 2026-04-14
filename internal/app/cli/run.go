package cli

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/alecthomas/kong"
	"github.com/neatflowcv/worker/internal/app/flow"
	"github.com/neatflowcv/worker/internal/pkg/command"
	"github.com/neatflowcv/worker/internal/pkg/local"
	"github.com/neatflowcv/worker/internal/pkg/postgresql"
)

func Run() error {
	database, err := newDatabase()
	if err != nil {
		return fmt.Errorf("create postgres database: %w", err)
	}

	defer func() {
		_ = database.Close()
	}()

	projectRepository := postgresql.NewProjectRepository(database)
	backlogItemRepository := postgresql.NewBacklogItemRepository(database)

	projectWorkspacer, err := newWorkspacer()
	if err != nil {
		return fmt.Errorf("create workspacer: %w", err)
	}

	backlogActionRunner := command.NewBacklogActionRunner()

	return RunWithArgs(
		context.Background(),
		os.Args[1:],
		os.Stdout,
		flow.NewService(
			projectRepository,
			backlogItemRepository,
			projectWorkspacer,
			backlogActionRunner,
		),
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

func newDatabase() (*postgresql.Database, error) {
	database, err := postgresql.NewDatabase(databaseDSN())
	if err != nil {
		return nil, fmt.Errorf("open postgres database: %w", err)
	}

	err = postgresql.Migrate(database)
	if err != nil {
		_ = database.Close()

		return nil, fmt.Errorf("migrate postgres database: %w", err)
	}

	return database, nil
}

func databaseDSN() string {
	dsn := os.Getenv("WORKER_POSTGRES_DSN")
	if dsn != "" {
		return dsn
	}

	return "host=localhost user=worker password=password_change_me " +
		"dbname=worker port=5432 sslmode=disable TimeZone=Asia/Seoul"
}

func newWorkspacer() (*local.Workspacer, error) {
	dataHome, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("resolve user home directory: %w", err)
	}

	return local.NewWorkspacer(
		filepath.Join(dataHome, ".local", "share", "worker", "projects"),
	), nil
}
