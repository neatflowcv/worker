package plannercli

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/alecthomas/kong"
	"github.com/neatflowcv/worker/internal/app/planner"
	"github.com/neatflowcv/worker/internal/pkg/decider"
)

const outputFileMode = 0o600

var errExpectedAnswersRequired = errors.New("expected answers are required")

const (
	ignoreQuestionChoice = "i"
	customAnswerChoice   = "c"
)

type Runner struct {
	rootDir string
	service *planner.Service
	stdin   io.Reader
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

	return NewRunner(service, rootDir, os.Stdin).Run(os.Args[1:], os.Stdout)
}

func NewRunner(service *planner.Service, rootDir string, stdin io.Reader) *Runner {
	return &Runner{
		rootDir: rootDir,
		service: service,
		stdin:   stdin,
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

func (a *app) Run(stdout io.Writer) error {
	_ = stdout

	_, err := os.Stat(a.Output)
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
	reader := bufio.NewReader(a.runner.stdin)

	response, err := a.runner.service.CreatePlan(request)
	if err != nil {
		return fmt.Errorf("create plan: %w", err)
	}

	err = writeMarkdownFile(a.Output, response.Markdown)
	if err != nil {
		return err
	}

	for len(response.Items) > 0 {
		response, err = a.refinePlan(reader, stdout, response)
		if err != nil {
			return err
		}

		err = writeMarkdownFile(a.Output, response.Markdown)
		if err != nil {
			return err
		}
	}

	return nil
}

func (a *app) refinePlan(
	reader *bufio.Reader,
	stdout io.Writer,
	response *planner.CreatePlanResponse,
) (*planner.CreatePlanResponse, error) {
	answer, err := askQuestion(reader, stdout, response.Items[0])
	if err != nil {
		return nil, err
	}

	refinedResponse, err := a.runner.service.RefinePlan(planner.RefinePlanRequest{
		RootDir:  a.runner.rootDir,
		Git:      a.Git,
		Markdown: response.Markdown,
		Answers:  []planner.QuestionAnswer{answer},
	})
	if err != nil {
		return nil, fmt.Errorf("refine plan: %w", err)
	}

	return &planner.CreatePlanResponse{
		Title:    response.Title,
		Git:      refinedResponse.Git,
		Items:    refinedResponse.Items,
		Markdown: refinedResponse.Markdown,
	}, nil
}

func askQuestion(
	reader *bufio.Reader,
	stdout io.Writer,
	item decider.Item,
) (planner.QuestionAnswer, error) {
	if len(item.ExpectedAnswers) == 0 {
		return planner.QuestionAnswer{}, errExpectedAnswersRequired
	}

	err := writeQuestion(stdout, item)
	if err != nil {
		return planner.QuestionAnswer{}, err
	}

	selection, err := readSelection(reader, stdout, len(item.ExpectedAnswers))
	if err != nil {
		return planner.QuestionAnswer{}, err
	}

	return buildQuestionAnswer(reader, stdout, item, selection)
}

func writeQuestion(stdout io.Writer, item decider.Item) error {
	_, err := fmt.Fprintf(stdout, "%s\n", strings.TrimSpace(item.Question))
	if err != nil {
		return fmt.Errorf("write question: %w", err)
	}

	for index, expectedAnswer := range item.ExpectedAnswers {
		_, err = fmt.Fprintf(stdout, "%d. %s\n", index+1, strings.TrimSpace(expectedAnswer))
		if err != nil {
			return fmt.Errorf("write expected answer: %w", err)
		}
	}

	_, err = fmt.Fprintf(stdout, "%s. 질문 무시\n", ignoreQuestionChoice)
	if err != nil {
		return fmt.Errorf("write ignore question option: %w", err)
	}

	_, err = fmt.Fprintf(stdout, "%s. 직접 입력\n", customAnswerChoice)
	if err != nil {
		return fmt.Errorf("write custom answer option: %w", err)
	}

	return nil
}

func buildQuestionAnswer(
	reader *bufio.Reader,
	stdout io.Writer,
	item decider.Item,
	selection string,
) (planner.QuestionAnswer, error) {
	if selection == ignoreQuestionChoice {
		answer, err := readCustomAnswer(reader, stdout)
		if err != nil {
			return planner.QuestionAnswer{}, err
		}

		return planner.QuestionAnswer{
			Question: "",
			Answer:   answer,
		}, nil
	}

	if selection == customAnswerChoice {
		answer, err := readCustomAnswer(reader, stdout)
		if err != nil {
			return planner.QuestionAnswer{}, err
		}

		return planner.QuestionAnswer{
			Question: strings.TrimSpace(item.Question),
			Answer:   answer,
		}, nil
	}

	selectedIndex, err := strconv.Atoi(selection)
	if err != nil {
		return planner.QuestionAnswer{}, fmt.Errorf("parse selected index: %w", err)
	}

	return planner.QuestionAnswer{
		Question: strings.TrimSpace(item.Question),
		Answer:   strings.TrimSpace(item.ExpectedAnswers[selectedIndex-1]),
	}, nil
}

func readSelection(
	reader *bufio.Reader,
	stdout io.Writer,
	answerCount int,
) (string, error) {
	for {
		_, err := fmt.Fprint(stdout, "선택: ")
		if err != nil {
			return "", fmt.Errorf("write selection prompt: %w", err)
		}

		line, err := reader.ReadString('\n')
		if err != nil {
			return "", fmt.Errorf("read selection: %w", err)
		}

		selection := strings.TrimSpace(line)
		if selection == ignoreQuestionChoice || selection == customAnswerChoice {
			return selection, nil
		}

		selectedIndex, err := strconv.Atoi(selection)
		if err == nil && selectedIndex >= 1 && selectedIndex <= answerCount {
			return selection, nil
		}

		_, err = fmt.Fprintln(stdout, "올바른 번호를 입력해.")
		if err != nil {
			return "", fmt.Errorf("write invalid selection message: %w", err)
		}
	}
}

func readCustomAnswer(reader *bufio.Reader, stdout io.Writer) (string, error) {
	for {
		_, err := fmt.Fprint(stdout, "답변: ")
		if err != nil {
			return "", fmt.Errorf("write custom answer prompt: %w", err)
		}

		line, err := reader.ReadString('\n')
		if err != nil {
			return "", fmt.Errorf("read custom answer: %w", err)
		}

		answer := strings.TrimSpace(line)
		if answer != "" {
			return answer, nil
		}

		_, err = fmt.Fprintln(stdout, "비어 있지 않은 답변을 입력해.")
		if err != nil {
			return "", fmt.Errorf("write invalid custom answer message: %w", err)
		}
	}
}

func newRootDir() (string, error) {
	dataHome, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolve user home directory: %w", err)
	}

	return filepath.Join(dataHome, ".local", "share", "worker", "plans"), nil
}

func writeMarkdownFile(outputPath, markdown string) error {
	err := os.WriteFile(outputPath, []byte(markdown), outputFileMode)
	if err != nil {
		return fmt.Errorf("write markdown: %w", err)
	}

	return nil
}
