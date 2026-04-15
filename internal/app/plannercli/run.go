package plannercli

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/alecthomas/kong"
	"github.com/neatflowcv/worker/internal/app/planner"
)

const outputFileMode = 0o600

type Runner struct {
	rootDir string
	service *planner.Service
}

type app struct {
	runner *Runner
	Output string `default:"plan.md" help:"Output file path."     name:"output"`
	Git    string `arg:""            help:"Git source reference." name:"git"`
	Title  string `arg:""            help:"Plan title."           name:"title"`
}

func Run() error {
	rootDir, err := newRootDir()
	if err != nil {
		return fmt.Errorf("resolve planner root directory: %w", err)
	}

	service, err := planner.NewCodexService()
	if err != nil {
		return fmt.Errorf("create planner service: %w", err)
	}

	return NewRunner(service, rootDir).Run(os.Args[1:], os.Stdout)
}

func NewRunner(service *planner.Service, rootDir string) *Runner {
	return &Runner{
		rootDir: rootDir,
		service: service,
	}
}

func (r *Runner) Run(args []string, stdout io.Writer) error {
	parser, err := kong.New(
		&app{
			runner: r,
			Output: "plan.md",
			Git:    "",
			Title:  "",
		},
		kong.Name("planner"),
		kong.BindTo(stdout, (*io.Writer)(nil)),
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

func (a *app) Run(stdout io.Writer) (err error) {
	_ = stdout

	_, err = os.Stat(a.Output)
	if err == nil {
		return fmt.Errorf("output file already exists: %w", os.ErrExist)
	}

	if !os.IsNotExist(err) {
		return fmt.Errorf("stat output file: %w", err)
	}

	request := planner.CreatePlanRequest{
		RootDir: a.runner.rootDir,
		Git:     a.Git,
		Title:   a.Title,
	}

	response, err := a.runner.service.CreatePlan(request)
	if err != nil {
		return fmt.Errorf("create plan: %w", err)
	}

	file, err := os.OpenFile(a.Output, os.O_WRONLY|os.O_CREATE|os.O_EXCL, outputFileMode)
	if err != nil {
		return fmt.Errorf("open output file: %w", err)
	}

	defer func() {
		closeErr := file.Close()
		if closeErr != nil {
			err = fmt.Errorf("close output file: %w", closeErr)
		}
	}()

	_, err = file.WriteString(response.Markdown)
	if err != nil {
		return fmt.Errorf("write markdown: %w", err)
	}

	return nil
}

func newRootDir() (string, error) {
	dataHome, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolve user home directory: %w", err)
	}

	return filepath.Join(dataHome, ".local", "share", "worker", "plans"), nil
}
