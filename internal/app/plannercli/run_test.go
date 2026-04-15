package plannercli_test

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/neatflowcv/worker/internal/app/planner"
	"github.com/neatflowcv/worker/internal/app/plannercli"
	"github.com/neatflowcv/worker/internal/pkg/decider"
	"github.com/stretchr/testify/require"
)

var errUnexpectedDeciderCall = errors.New("unexpected decider call")

func TestRunnerRun(t *testing.T) {
	t.Parallel()

	// Arrange
	var (
		stdin  bytes.Buffer
		stdout bytes.Buffer
	)

	_, err := stdin.WriteString("1\n")
	require.NoError(t, err)

	workingDir := t.TempDir()
	outputPath := filepath.Join(workingDir, "custom-plan.md")

	rootDir := filepath.Join(t.TempDir(), "plans")
	repositoryURL := createGitRepository(t)
	localDir := filepath.Join(rootDir, filepath.Base(repositoryURL))
	runner := newInteractiveRunner(t, rootDir, localDir, &stdin)

	// Act
	err = runner.Run(
		[]string{
			"--output",
			outputPath,
			repositoryURL,
			"Feedback Backlog Item 구현",
		},
		&stdout,
	)

	// Assert
	require.NoError(t, err)
	//nolint:gosec // Test controls the temporary output path.
	content, err := os.ReadFile(outputPath)
	require.NoError(t, err)
	require.Equal(t, "# Refined Decision", string(content))
	require.Contains(t, stdout.String(), "무엇을 먼저 할까?")
	require.Contains(t, stdout.String(), "1. Git")
	require.Contains(t, stdout.String(), "i. 질문 무시")
	require.Contains(t, stdout.String(), "c. 직접 입력")
}

func TestRunnerRunRepeatsRefineUntilItemsAreEmpty(t *testing.T) {
	t.Parallel()

	stdout, outputPath, repositoryURL, runner, feedbackCallCount := newRepeatRefineTestContext(t)

	// Act
	err := runner.Run(
		[]string{
			"--output",
			outputPath,
			repositoryURL,
			"Feedback Backlog Item 구현",
		},
		stdout,
	)

	// Assert
	require.NoError(t, err)
	require.Equal(t, 2, *feedbackCallCount)
	//nolint:gosec // Test controls the temporary output path.
	content, err := os.ReadFile(outputPath)
	require.NoError(t, err)
	require.Equal(t, "# Final Decision", string(content))
	require.Contains(t, stdout.String(), "무엇을 먼저 할까?")
	require.Contains(t, stdout.String(), "무엇을 다음에 할까?")
}

func TestRunnerRunAcceptsCustomAnswer(t *testing.T) {
	t.Parallel()

	content, stdout := runInteractiveAnswerTest(
		t,
		"c\n직접 정한 답변\n",
		decider.QuestionAnswer{
			Question: "무엇을 먼저 할까?",
			Answer:   "직접 정한 답변",
		},
		"# Custom Decision",
	)

	require.Equal(t, "# Custom Decision", content)
	require.Contains(t, stdout, "답변: ")
}

func TestRunnerRunAllowsIgnoringQuestion(t *testing.T) {
	t.Parallel()

	content, stdout := runInteractiveAnswerTest(
		t,
		"i\n질문 없이 전달할 답변\n",
		decider.QuestionAnswer{
			Question: "",
			Answer:   "질문 없이 전달할 답변",
		},
		"# Ignored Question Decision",
	)

	require.Equal(t, "# Ignored Question Decision", content)
	require.Contains(t, stdout, "답변: ")
}

func TestRunnerRunReturnsErrorWhenTitleIsMissing(t *testing.T) {
	t.Parallel()

	// Arrange
	var stdout bytes.Buffer

	runner := plannercli.NewRunner(
		planner.NewService(unusedDecider(t)),
		filepath.Join(t.TempDir(), "plans"),
		bytes.NewBuffer(nil),
	)

	// Act
	err := runner.Run(
		[]string{
			"Feedback Backlog Item 구현",
		},
		&stdout,
	)

	// Assert
	require.ErrorContains(t, err, "expected \"<title>\"")
	require.Empty(t, stdout.String())
}

//nolint:paralleltest // Changes working directory to verify default relative output path.
func TestRunnerRunWritesPlanToDefaultOutputFile(t *testing.T) {
	// Arrange
	var stdout bytes.Buffer

	workingDir := t.TempDir()
	t.Chdir(workingDir)

	rootDir := filepath.Join(t.TempDir(), "plans")
	repositoryURL := createGitRepository(t)
	runner := plannercli.NewRunner(
		planner.NewService(deciderStub{
			decideFunc: func(_ decider.DecideRequest) (*decider.Decision, error) {
				return &decider.Decision{
					Markdown: "# Default",
					Items:    nil,
				}, nil
			},
			refinePlanFunc: func(_ decider.RefinePlanRequest) (*decider.Decision, error) {
				t.Fatal("refine plan should not be called")

				return nil, errUnexpectedDeciderCall
			},
		}),
		rootDir,
		bytes.NewBuffer(nil),
	)

	// Act
	err := runner.Run(
		[]string{
			repositoryURL,
			"Feedback Backlog Item 구현",
		},
		&stdout,
	)

	// Assert
	require.NoError(t, err)
	//nolint:gosec // Test controls the temporary output path.
	content, err := os.ReadFile(filepath.Join(workingDir, "plan.md"))
	require.NoError(t, err)
	require.Equal(t, "# Default", string(content))
	require.Empty(t, stdout.String())
}

func TestRunnerRunReturnsErrorWhenOutputFileAlreadyExists(t *testing.T) {
	t.Parallel()

	// Arrange
	var stdout bytes.Buffer

	outputPath := filepath.Join(t.TempDir(), "plan.md")
	err := os.WriteFile(outputPath, []byte("# existing"), 0o600)
	require.NoError(t, err)

	rootDir := filepath.Join(t.TempDir(), "plans")
	repositoryURL := createGitRepository(t)
	runner := plannercli.NewRunner(
		planner.NewService(unusedDecider(t)),
		rootDir,
		bytes.NewBuffer(nil),
	)

	// Act
	err = runner.Run(
		[]string{
			"--output",
			outputPath,
			repositoryURL,
			"Feedback Backlog Item 구현",
		},
		&stdout,
	)

	// Assert
	require.ErrorContains(t, err, "output file already exists")
	require.ErrorIs(t, err, os.ErrExist)
	require.Empty(t, stdout.String())
}

func createGitRepository(t *testing.T) string {
	t.Helper()

	repositoryDir := filepath.Join(t.TempDir(), "plan")
	repository, err := git.PlainInit(repositoryDir, false)
	require.NoError(t, err)

	err = os.WriteFile(filepath.Join(repositoryDir, "README.md"), []byte("# worker\n"), 0o600)
	require.NoError(t, err)

	worktree, err := repository.Worktree()
	require.NoError(t, err)

	_, err = worktree.Add("README.md")
	require.NoError(t, err)

	_, err = worktree.Commit("init", &git.CommitOptions{
		All:               false,
		AllowEmptyCommits: false,
		Author: &object.Signature{
			Name:  "worker",
			Email: "worker@example.com",
			When:  time.Unix(1, 0),
		},
		Committer: nil,
		Parents:   nil,
		SignKey:   nil,
		Signer:    nil,
		Amend:     false,
	})
	require.NoError(t, err)

	return repositoryDir
}

type deciderStub struct {
	decideFunc     func(request decider.DecideRequest) (*decider.Decision, error)
	refinePlanFunc func(request decider.RefinePlanRequest) (*decider.Decision, error)
}

func (s deciderStub) Decide(request decider.DecideRequest) (*decider.Decision, error) {
	return s.decideFunc(request)
}

func (s deciderStub) RefinePlan(request decider.RefinePlanRequest) (*decider.Decision, error) {
	return s.refinePlanFunc(request)
}

func unusedDecider(t *testing.T) deciderStub {
	t.Helper()

	return deciderStub{
		decideFunc: func(_ decider.DecideRequest) (*decider.Decision, error) {
			t.Fatal("decide should not be called")

			return nil, errUnexpectedDeciderCall
		},
		refinePlanFunc: func(_ decider.RefinePlanRequest) (*decider.Decision, error) {
			t.Fatal("refine plan should not be called")

			return nil, errUnexpectedDeciderCall
		},
	}
}

func newInteractiveRunner(
	t *testing.T,
	rootDir string,
	localDir string,
	stdin *bytes.Buffer,
) *plannercli.Runner {
	t.Helper()

	return plannercli.NewRunner(
		planner.NewService(deciderStub{
			decideFunc: func(request decider.DecideRequest) (*decider.Decision, error) {
				require.Equal(t, "Feedback Backlog Item 구현", request.Title)
				require.Equal(t, localDir, request.Directory)

				return &decider.Decision{
					Markdown: "# Decision",
					Items: []decider.Item{
						{
							Question:        "무엇을 먼저 할까?",
							ExpectedAnswers: []string{"Git"},
						},
					},
				}, nil
			},
			refinePlanFunc: func(request decider.RefinePlanRequest) (*decider.Decision, error) {
				require.Equal(t, localDir, request.Directory)
				require.Equal(t, "# Decision", request.Markdown)
				require.Equal(
					t,
					[]decider.QuestionAnswer{
						{
							Question: "무엇을 먼저 할까?",
							Answer:   "Git",
						},
					},
					request.Answers,
				)

				return &decider.Decision{
					Markdown: "# Refined Decision",
					Items:    nil,
				}, nil
			},
		}),
		rootDir,
		stdin,
	)
}

func newRepeatRefineTestContext(
	t *testing.T,
) (*bytes.Buffer, string, string, *plannercli.Runner, *int) {
	t.Helper()

	var (
		stdin  bytes.Buffer
		stdout bytes.Buffer
	)

	_, err := stdin.WriteString("1\n1\n")
	require.NoError(t, err)

	workingDir := t.TempDir()
	outputPath := filepath.Join(workingDir, "custom-plan.md")
	rootDir := filepath.Join(t.TempDir(), "plans")
	repositoryURL := createGitRepository(t)
	localDir := filepath.Join(rootDir, filepath.Base(repositoryURL))
	refineCallCount := 0

	runner := plannercli.NewRunner(
		planner.NewService(newRepeatRefineDecider(t, localDir, &refineCallCount)),
		rootDir,
		&stdin,
	)

	return &stdout, outputPath, repositoryURL, runner, &refineCallCount
}

func newRepeatRefineDecider(
	t *testing.T,
	localDir string,
	refineCallCount *int,
) deciderStub {
	t.Helper()

	return deciderStub{
		decideFunc: func(request decider.DecideRequest) (*decider.Decision, error) {
			require.Equal(t, localDir, request.Directory)

			return &decider.Decision{
				Markdown: "# First Decision",
				Items: []decider.Item{
					{
						Question:        "무엇을 먼저 할까?",
						ExpectedAnswers: []string{"Git"},
					},
				},
			}, nil
		},
		refinePlanFunc: func(request decider.RefinePlanRequest) (*decider.Decision, error) {
			return buildRepeatRefineDecision(t, request, refineCallCount)
		},
	}
}

func buildRepeatRefineDecision(
	t *testing.T,
	request decider.RefinePlanRequest,
	refineCallCount *int,
) (*decider.Decision, error) {
	t.Helper()

	(*refineCallCount)++
	if *refineCallCount == 1 {
		require.Equal(t, "# First Decision", request.Markdown)
		require.Equal(
			t,
			[]decider.QuestionAnswer{
				{
					Question: "무엇을 먼저 할까?",
					Answer:   "Git",
				},
			},
			request.Answers,
		)

		return &decider.Decision{
			Markdown: "# Second Decision",
			Items: []decider.Item{
				{
					Question:        "무엇을 다음에 할까?",
					ExpectedAnswers: []string{"Test"},
				},
			},
		}, nil
	}

	require.Equal(t, 2, *refineCallCount)
	require.Equal(t, "# Second Decision", request.Markdown)
	require.Equal(
		t,
		[]decider.QuestionAnswer{
			{
				Question: "무엇을 다음에 할까?",
				Answer:   "Test",
			},
		},
		request.Answers,
	)

	return &decider.Decision{
		Markdown: "# Final Decision",
		Items:    nil,
	}, nil
}

func runInteractiveAnswerTest(
	t *testing.T,
	input string,
	expectedAnswer decider.QuestionAnswer,
	finalMarkdown string,
) (string, string) {
	t.Helper()

	stdout, outputPath, repositoryURL, runner := newInteractiveAnswerTestContext(
		t,
		input,
		expectedAnswer,
		finalMarkdown,
	)

	err := runner.Run(
		[]string{
			"--output",
			outputPath,
			repositoryURL,
			"Feedback Backlog Item 구현",
		},
		stdout,
	)
	require.NoError(t, err)

	//nolint:gosec // Test controls the temporary output path.
	content, err := os.ReadFile(outputPath)
	require.NoError(t, err)

	return string(content), stdout.String()
}

func newInteractiveAnswerTestContext(
	t *testing.T,
	input string,
	expectedAnswer decider.QuestionAnswer,
	finalMarkdown string,
) (*bytes.Buffer, string, string, *plannercli.Runner) {
	t.Helper()

	var (
		stdin  bytes.Buffer
		stdout bytes.Buffer
	)

	_, err := stdin.WriteString(input)
	require.NoError(t, err)

	workingDir := t.TempDir()
	outputPath := filepath.Join(workingDir, "custom-plan.md")
	rootDir := filepath.Join(t.TempDir(), "plans")
	repositoryURL := createGitRepository(t)
	localDir := filepath.Join(rootDir, filepath.Base(repositoryURL))

	return &stdout, outputPath, repositoryURL, plannercli.NewRunner(
		planner.NewService(deciderStub{
			decideFunc: func(request decider.DecideRequest) (*decider.Decision, error) {
				require.Equal(t, localDir, request.Directory)

				return &decider.Decision{
					Markdown: "# Decision",
					Items: []decider.Item{
						{
							Question:        "무엇을 먼저 할까?",
							ExpectedAnswers: []string{"Git"},
						},
					},
				}, nil
			},
			refinePlanFunc: func(request decider.RefinePlanRequest) (*decider.Decision, error) {
				require.Equal(t, []decider.QuestionAnswer{expectedAnswer}, request.Answers)

				return &decider.Decision{
					Markdown: finalMarkdown,
					Items:    nil,
				}, nil
			},
		}),
		rootDir,
		&stdin,
	)
}
