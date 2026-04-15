package plannercli

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/alecthomas/kong"
	"github.com/neatflowcv/worker/internal/app/planner"
	"github.com/neatflowcv/worker/internal/pkg/decider"
)

type Runner struct {
	rootDir string
	service *planner.Service
}

type app struct {
	runner *Runner
	Git    string `arg:"" help:"Git source reference." name:"git"`
	Title  string `arg:"" help:"Plan title."           name:"title"`
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

func (a *app) Run(stdout io.Writer) error {
	request := planner.CreatePlanRequest{
		RootDir: a.runner.rootDir,
		Title:   a.Title,
		Sources: []planner.Source{
			{
				Kind:      planner.SourceKindGit,
				Reference: a.Git,
			},
		},
	}

	response, err := a.runner.service.CreatePlan(request)
	if err != nil {
		return fmt.Errorf("create plan: %w", err)
	}

	_, err = fmt.Fprint(stdout, response.Markdown)
	if err != nil {
		return fmt.Errorf("write markdown: %w", err)
	}

	_, err = fmt.Fprint(stdout, formatItems(response.Items))
	if err != nil {
		return fmt.Errorf("write items: %w", err)
	}

	return nil
}

func formatItems(items []decider.Item) string {
	if len(items) == 0 {
		return ""
	}

	var builder strings.Builder

	builder.WriteString("\n\n## 결정사항\n")

	for idx, item := range items {
		fmt.Fprintf(&builder, "\n%d. %s\n", idx+1, item.Question)

		if len(item.ExpectedAnswers) == 0 {
			continue
		}

		builder.WriteString("   예상 답안:\n")

		for answerIdx, answer := range item.ExpectedAnswers {
			fmt.Fprintf(&builder, "   %d. %s\n", answerIdx+1, answer)
		}
	}

	return builder.String()
}

func newRootDir() (string, error) {
	dataHome, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolve user home directory: %w", err)
	}

	return filepath.Join(dataHome, ".local", "share", "worker", "plans"), nil
}
