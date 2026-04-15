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
	var stdout bytes.Buffer

	rootDir := filepath.Join(t.TempDir(), "plans")
	repositoryURL := createGitRepository(t, "plan", "# worker\n")
	localDir := filepath.Join(rootDir, filepath.Base(repositoryURL))
	runner := plannercli.NewRunner(
		planner.NewService(deciderFunc(func(request decider.DecideRequest) (*decider.Decision, error) {
			require.Equal(t, "Feedback Backlog Item 구현", request.Title)
			require.Equal(t, []string{localDir}, request.Directories)

			return &decider.Decision{
				Markdown: "# Decision",
				Items: []decider.Item{
					{
						Question:        "무엇을 먼저 할까?",
						ExpectedAnswers: []string{"Git"},
					},
				},
			}, nil
		})),
		rootDir,
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
	require.Equal(
		t,
		"# Decision\n\n## 결정사항\n\n1. 무엇을 먼저 할까?\n   예상 답안:\n   1. Git\n",
		stdout.String(),
	)
}

func TestRunnerRunReturnsErrorWhenTitleIsMissing(t *testing.T) {
	t.Parallel()

	// Arrange
	var stdout bytes.Buffer

	runner := plannercli.NewRunner(
		planner.NewService(unusedDecider(t)),
		filepath.Join(t.TempDir(), "plans"),
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

func createGitRepository(t *testing.T, name, content string) string {
	t.Helper()

	repositoryDir := filepath.Join(t.TempDir(), name)
	repository, err := git.PlainInit(repositoryDir, false)
	require.NoError(t, err)

	err = os.WriteFile(filepath.Join(repositoryDir, "README.md"), []byte(content), 0o600)
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

type deciderFunc func(request decider.DecideRequest) (*decider.Decision, error)

func (f deciderFunc) Decide(request decider.DecideRequest) (*decider.Decision, error) {
	return f(request)
}

func unusedDecider(t *testing.T) deciderFunc {
	t.Helper()

	return deciderFunc(func(_ decider.DecideRequest) (*decider.Decision, error) {
		t.Fatal("decider should not be called")

		return nil, errUnexpectedDeciderCall
	})
}
