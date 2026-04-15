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
	"unicode/utf8"

	"github.com/alecthomas/kong"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/neatflowcv/worker/internal/app/planner"
	"github.com/neatflowcv/worker/internal/pkg/decider"
)

const outputFileMode = 0o600

const extraQuestionOptionCount = 2

var errExpectedAnswersRequired = errors.New("expected answers are required")
var errUnexpectedFinalModelType = errors.New("unexpected final model type")

const (
	ignoreQuestionChoice = "질문 무시"
	customAnswerChoice   = "직접 입력"
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

type phase string

const (
	phaseLoading phase = "loading"
	phaseSelect  phase = "select"
	phaseInput   phase = "input"
	phaseDone    phase = "done"
	phaseError   phase = "error"
)

type planCreatedMsg struct {
	response *planner.CreatePlanResponse
	err      error
}

type planRefinedMsg struct {
	response *planner.CreatePlanResponse
	err      error
}

type model struct {
	rootDir string
	service *planner.Service
	output  string
	git     string
	title   string

	phase       phase
	response    *planner.CreatePlanResponse
	selected    int
	customInput string
	err         error
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

	if !isInteractiveTerminal(a.runner.stdin, stdout) {
		return a.runTextMode(stdout)
	}

	program := tea.NewProgram(
		newModel(a.runner.service, a.runner.rootDir, a.Output, a.Git, a.Title),
		tea.WithInput(a.runner.stdin),
		tea.WithOutput(stdout),
		tea.WithoutSignalHandler(),
	)

	result, err := program.Run()
	if err != nil {
		return fmt.Errorf("run tui: %w", err)
	}

	finalModel, ok := result.(model)
	if !ok {
		return errUnexpectedFinalModelType
	}

	if finalModel.err != nil {
		return finalModel.err
	}

	return nil
}

func (a *app) runTextMode(stdout io.Writer) error {
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

func newModel(
	service *planner.Service,
	rootDir string,
	output string,
	git string,
	title string,
) model {
	return model{
		rootDir:     rootDir,
		service:     service,
		output:      output,
		git:         git,
		title:       title,
		phase:       phaseLoading,
		response:    nil,
		selected:    0,
		customInput: "",
		err:         nil,
	}
}

func (m model) Init() tea.Cmd {
	return m.createPlanCmd()
}

func (m model) Update(message tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := message.(type) {
	case planCreatedMsg:
		if msg.err != nil {
			m.phase = phaseError
			m.err = msg.err

			return m, tea.Quit
		}

		m.response = msg.response
		m.selected = 0
		m.customInput = ""

		if len(msg.response.Items) == 0 {
			m.phase = phaseDone

			return m, tea.Quit
		}

		m.phase = phaseSelect

		return m, nil

	case planRefinedMsg:
		if msg.err != nil {
			m.phase = phaseError
			m.err = msg.err

			return m, tea.Quit
		}

		m.response = msg.response
		m.selected = 0
		m.customInput = ""

		if len(msg.response.Items) == 0 {
			m.phase = phaseDone

			return m, tea.Quit
		}

		m.phase = phaseSelect

		return m, nil

	case tea.KeyMsg:
		return m.updateKey(msg)
	}

	return m, nil
}

func (m model) View() string {
	switch m.phase {
	case phaseLoading:
		return m.loadingView()
	case phaseSelect:
		return m.selectView()
	case phaseInput:
		return m.inputView()
	case phaseDone:
		return fmt.Sprintf("플랜을 저장했어: %s\n", m.output)
	case phaseError:
		return fmt.Sprintf("오류: %v\n", m.err)
	}

	return ""
}

func (m model) updateKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.phase {
	case phaseLoading:
		return m, nil
	case phaseSelect:
		return m.updateSelect(msg)
	case phaseInput:
		return m.updateInput(msg)
	case phaseDone:
		return m, nil
	case phaseError:
		return m, nil
	}

	return m, nil
}

func (m model) updateSelect(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	options := m.currentOptions()
	if len(options) == 0 {
		return m, nil
	}

	switch msg.String() {
	case "up", "k":
		if m.selected > 0 {
			m.selected--
		}

		return m, nil

	case "down", "j":
		if m.selected < len(options)-1 {
			m.selected++
		}

		return m, nil

	case "enter":
		selection := options[m.selected]
		if selection == ignoreQuestionChoice || selection == customAnswerChoice {
			m.phase = phaseInput
			m.customInput = ""

			return m, nil
		}

		answer, err := m.buildSelectionAnswer(strconv.Itoa(m.selected + 1))
		if err != nil {
			m.phase = phaseError
			m.err = err

			return m, tea.Quit
		}

		m.phase = phaseLoading

		return m, m.refinePlanCmd(answer)
	}

	return m, nil
}

func (m model) updateInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.phase = phaseSelect
		m.customInput = ""

		return m, nil

	case "enter":
		answer, err := m.buildInputAnswer()
		if err != nil {
			return m, nil
		}

		m.phase = phaseLoading

		return m, m.refinePlanCmd(answer)

	case "backspace":
		m.customInput = trimLastRune(m.customInput)

		return m, nil

	default:
		if msg.Type == tea.KeyRunes || msg.Type == tea.KeySpace {
			m.customInput += msg.String()
		}

		return m, nil
	}
}

func (m model) loadingView() string {
	if m.response == nil {
		return "플랜 생성 중...\n"
	}

	return "답변을 반영해 플랜 갱신 중...\n"
}

func (m model) selectView() string {
	item := m.currentItem()
	if item == nil {
		return ""
	}

	lines := []string{
		strings.TrimSpace(item.Question),
		"",
	}

	for index, option := range m.currentOptions() {
		prefix := " "
		if index == m.selected {
			prefix = ">"
		}

		lines = append(lines, fmt.Sprintf("%s %d. %s", prefix, index+1, option))
	}

	lines = append(lines, "")
	lines = append(lines, "위아래로 이동하고 Enter로 선택해.")

	return strings.Join(lines, "\n") + "\n"
}

func (m model) inputView() string {
	item := m.currentItem()
	if item == nil {
		return ""
	}

	label := "답변"
	if m.selected == len(item.ExpectedAnswers) {
		label = "질문 없이 전달할 메모"
	}

	lines := []string{
		strings.TrimSpace(item.Question),
		"",
		fmt.Sprintf("%s: %s", label, m.customInput),
		"",
		"Enter로 제출하고 Esc로 돌아가.",
	}

	return strings.Join(lines, "\n") + "\n"
}

func (m model) currentItem() *decider.Item {
	if m.response == nil || len(m.response.Items) == 0 {
		return nil
	}

	return &m.response.Items[0]
}

func (m model) currentOptions() []string {
	item := m.currentItem()
	if item == nil {
		return nil
	}

	options := make([]string, 0, len(item.ExpectedAnswers)+extraQuestionOptionCount)
	options = append(options, item.ExpectedAnswers...)
	options = append(options, ignoreQuestionChoice, customAnswerChoice)

	return options
}

func (m model) createPlanCmd() tea.Cmd {
	request := planner.CreatePlanRequest{
		RootDir: m.rootDir,
		Git:     m.git,
		Title:   m.title,
	}

	return func() tea.Msg {
		response, err := m.service.CreatePlan(request)
		if err != nil {
			return planCreatedMsg{
				response: nil,
				err:      fmt.Errorf("create plan: %w", err),
			}
		}

		err = writeMarkdownFile(m.output, response.Markdown)
		if err != nil {
			return planCreatedMsg{
				response: nil,
				err:      err,
			}
		}

		return planCreatedMsg{
			response: response,
			err:      nil,
		}
	}
}

func (m model) refinePlanCmd(answer planner.QuestionAnswer) tea.Cmd {
	return func() tea.Msg {
		refinedResponse, err := m.service.RefinePlan(planner.RefinePlanRequest{
			RootDir:  m.rootDir,
			Git:      m.git,
			Markdown: m.response.Markdown,
			Answers:  []planner.QuestionAnswer{answer},
		})
		if err != nil {
			return planRefinedMsg{
				response: nil,
				err:      fmt.Errorf("refine plan: %w", err),
			}
		}

		response := &planner.CreatePlanResponse{
			Title:    m.response.Title,
			Git:      refinedResponse.Git,
			Items:    refinedResponse.Items,
			Markdown: refinedResponse.Markdown,
		}

		err = writeMarkdownFile(m.output, response.Markdown)
		if err != nil {
			return planRefinedMsg{
				response: nil,
				err:      err,
			}
		}

		return planRefinedMsg{
			response: response,
			err:      nil,
		}
	}
}

func (m model) buildSelectionAnswer(selection string) (planner.QuestionAnswer, error) {
	item := m.currentItem()
	if item == nil {
		return planner.QuestionAnswer{}, errExpectedAnswersRequired
	}

	selectedIndex, err := strconv.Atoi(selection)
	if err != nil {
		return planner.QuestionAnswer{}, fmt.Errorf("parse selected index: %w", err)
	}

	if selectedIndex < 1 || selectedIndex > len(item.ExpectedAnswers) {
		return planner.QuestionAnswer{}, errExpectedAnswersRequired
	}

	return planner.QuestionAnswer{
		Question: strings.TrimSpace(item.Question),
		Answer:   strings.TrimSpace(item.ExpectedAnswers[selectedIndex-1]),
	}, nil
}

func (m model) buildInputAnswer() (planner.QuestionAnswer, error) {
	item := m.currentItem()
	if item == nil {
		return planner.QuestionAnswer{}, errExpectedAnswersRequired
	}

	answer := strings.TrimSpace(m.customInput)
	if answer == "" {
		return planner.QuestionAnswer{}, planner.ErrPlanAnswerRequired
	}

	if m.selected == len(item.ExpectedAnswers) {
		return planner.QuestionAnswer{
			Question: "",
			Answer:   answer,
		}, nil
	}

	return planner.QuestionAnswer{
		Question: strings.TrimSpace(item.Question),
		Answer:   answer,
	}, nil
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

	_, err = fmt.Fprintf(stdout, "%d. %s\n", len(item.ExpectedAnswers)+1, ignoreQuestionChoice)
	if err != nil {
		return fmt.Errorf("write ignore question option: %w", err)
	}

	_, err = fmt.Fprintf(stdout, "%d. %s\n", len(item.ExpectedAnswers)+extraQuestionOptionCount, customAnswerChoice)
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
	selectedIndex, err := strconv.Atoi(selection)
	if err != nil {
		return planner.QuestionAnswer{}, fmt.Errorf("parse selected index: %w", err)
	}

	if selectedIndex == len(item.ExpectedAnswers)+1 {
		answer, readErr := readCustomAnswer(reader, stdout)
		if readErr != nil {
			return planner.QuestionAnswer{}, readErr
		}

		return planner.QuestionAnswer{
			Question: "",
			Answer:   answer,
		}, nil
	}

	if selectedIndex == len(item.ExpectedAnswers)+2 {
		answer, readErr := readCustomAnswer(reader, stdout)
		if readErr != nil {
			return planner.QuestionAnswer{}, readErr
		}

		return planner.QuestionAnswer{
			Question: strings.TrimSpace(item.Question),
			Answer:   answer,
		}, nil
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

		selectedIndex, err := strconv.Atoi(selection)
		if err == nil && selectedIndex >= 1 && selectedIndex <= answerCount+2 {
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

func isInteractiveTerminal(stdin io.Reader, stdout io.Writer) bool {
	inputFile, inputOK := stdin.(*os.File)

	outputFile, outputOK := stdout.(*os.File)
	if !inputOK || !outputOK {
		return false
	}

	inputInfo, err := inputFile.Stat()
	if err != nil {
		return false
	}

	outputInfo, err := outputFile.Stat()
	if err != nil {
		return false
	}

	return inputInfo.Mode()&os.ModeCharDevice != 0 &&
		outputInfo.Mode()&os.ModeCharDevice != 0
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

func trimLastRune(value string) string {
	if value == "" {
		return ""
	}

	_, size := utf8.DecodeLastRuneInString(value)
	if size <= 0 {
		return ""
	}

	return value[:len(value)-size]
}
