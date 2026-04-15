package planner_test

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/neatflowcv/worker/internal/app/planner"
	"github.com/neatflowcv/worker/internal/pkg/decider"
	"github.com/stretchr/testify/require"
)

var errUnexpectedDeciderCall = errors.New("unexpected decider call")

func TestServiceCreatePlan(t *testing.T) {
	t.Parallel()

	// Arrange
	rootDir := t.TempDir()
	repositoryURL := createGitRepository(t, "plan.md", "# worker\n")
	localDir := filepath.Join(rootDir, filepath.Base(repositoryURL))
	service := planner.NewService(newDeciderForCreatePlanTest(t, localDir))
	request := planner.CreatePlanRequest{
		RootDir: rootDir,
		Git:     repositoryURL,
		Title:   "Feedback Backlog Item 구현",
	}

	// Act
	plan, err := service.CreatePlan(request)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, plan)
	require.Equal(t, "Feedback Backlog Item 구현", plan.Title)
	require.Equal(t, localDir, plan.Git)
	require.Equal(
		t,
		[]decider.Item{
			{
				Question:        "무엇을 먼저 할까?",
				ExpectedAnswers: []string{"Git"},
			},
		},
		plan.Items,
	)
	require.Equal(
		t,
		"# Decision",
		plan.Markdown,
	)

	_, err = git.PlainOpen(localDir)
	require.NoError(t, err)
}

func TestServiceCreatePlanPullsExistingGitSource(t *testing.T) {
	t.Parallel()

	// Arrange
	service := planner.NewService(deciderFunc(func(_ decider.DecideRequest) (*decider.Decision, error) {
		return &decider.Decision{
			Markdown: "# Decision",
			Items:    nil,
		}, nil
	}))
	rootDir := t.TempDir()
	repositoryURL := createGitRepository(t, "README.md", "first\n")
	request := planner.CreatePlanRequest{
		RootDir: rootDir,
		Git:     repositoryURL,
		Title:   "Feedback Backlog Item 구현",
	}

	// Act
	_, err := service.CreatePlan(request)
	require.NoError(t, err)

	writeRepositoryFile(t, repositoryURL, "README.md", "second\n")

	plan, err := service.CreatePlan(request)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, plan)
	content, readErr := os.ReadFile(filepath.Join(plan.Git, "README.md"))
	require.NoError(t, readErr)
	require.Equal(t, "second\n", string(content))
}

func TestServiceCreatePlanReturnsErrorWhenTitleIsEmpty(t *testing.T) {
	t.Parallel()

	// Arrange
	service := planner.NewService(unusedDecider(t))
	request := planner.CreatePlanRequest{
		RootDir: t.TempDir(),
		Git:     createGitRepository(t, "plan.md", "# worker\n"),
		Title:   "  ",
	}

	// Act
	plan, err := service.CreatePlan(request)

	// Assert
	require.ErrorIs(t, err, planner.ErrPlanTitleRequired)
	require.Nil(t, plan)
}

func TestServiceCreatePlanReturnsErrorWhenGitIsEmpty(t *testing.T) {
	t.Parallel()

	// Arrange
	service := planner.NewService(unusedDecider(t))
	request := planner.CreatePlanRequest{
		RootDir: t.TempDir(),
		Git:     " ",
		Title:   "Feedback Backlog Item 구현",
	}

	// Act
	plan, err := service.CreatePlan(request)

	// Assert
	require.ErrorIs(t, err, planner.ErrPlanGitRequired)
	require.Nil(t, plan)
}

func TestServiceCreatePlanReturnsErrorWhenRootDirIsEmpty(t *testing.T) {
	t.Parallel()

	// Arrange
	service := planner.NewService(unusedDecider(t))
	request := planner.CreatePlanRequest{
		RootDir: " ",
		Git:     createGitRepository(t, "plan.md", "# worker\n"),
		Title:   "Feedback Backlog Item 구현",
	}

	// Act
	plan, err := service.CreatePlan(request)

	// Assert
	require.ErrorIs(t, err, planner.ErrPlanRootDirRequired)
	require.Nil(t, plan)
}

func createGitRepository(t *testing.T, name, content string) string {
	t.Helper()

	repositoryDir := filepath.Join(t.TempDir(), name)
	repository, err := git.PlainInit(repositoryDir, false)
	require.NoError(t, err)

	writeFileAndCommit(t, repository, repositoryDir, "README.md", content)

	return repositoryDir
}

func writeRepositoryFile(t *testing.T, repositoryDir, fileName, content string) {
	t.Helper()

	repository, err := git.PlainOpen(repositoryDir)
	require.NoError(t, err)

	writeFileAndCommit(t, repository, repositoryDir, fileName, content)
}

func writeFileAndCommit(
	t *testing.T,
	repository *git.Repository,
	repositoryDir, fileName, content string,
) {
	t.Helper()

	err := os.WriteFile(filepath.Join(repositoryDir, fileName), []byte(content), 0o600)
	require.NoError(t, err)

	worktree, err := repository.Worktree()
	require.NoError(t, err)

	_, err = worktree.Add(fileName)
	require.NoError(t, err)

	_, err = worktree.Commit("update "+fileName, &git.CommitOptions{
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
}

type deciderFunc func(request decider.DecideRequest) (*decider.Decision, error)

func (f deciderFunc) Decide(request decider.DecideRequest) (*decider.Decision, error) {
	return f(request)
}

func newDeciderForCreatePlanTest(t *testing.T, localDir string) deciderFunc {
	t.Helper()

	return deciderFunc(func(request decider.DecideRequest) (*decider.Decision, error) {
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
	})
}

func unusedDecider(t *testing.T) deciderFunc {
	t.Helper()

	return deciderFunc(func(_ decider.DecideRequest) (*decider.Decision, error) {
		t.Fatal("decider should not be called")

		return nil, errUnexpectedDeciderCall
	})
}
