package command_test

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/neatflowcv/worker/internal/pkg/command"
	"github.com/neatflowcv/worker/internal/pkg/domain"
	"github.com/stretchr/testify/require"
)

var errRunner = errors.New("runner error")

func TestNewBacklogActionRunnerWithRunnerStartBacklogItem(t *testing.T) {
	t.Parallel()

	// Arrange
	projectDir := t.TempDir()
	item := mustNewBacklogItem(
		t,
		"Start title",
		"implement the start flow",
	)

	var (
		gotDir    string
		gotPrompt string
	)

	actionRunner := command.NewBacklogActionRunnerWithRunner(
		runnerFunc(func(_ context.Context, dir string, args ...string) error {
			gotDir = dir

			require.Len(t, args, 1)
			gotPrompt = args[0]
			//nolint:gosec // The test reads a file created under the temp project directory.
			content, err := os.ReadFile(filepath.Join(dir, "backlog-1.md"))
			require.NoError(t, err)
			require.Equal(t, "implement the start flow", string(content))

			return nil
		}),
	)

	// Act
	err := actionRunner.StartBacklogItem(t.Context(), projectDir, item)

	// Assert
	require.NoError(t, err)
	require.Equal(t, projectDir, gotDir)
	require.Equal(
		t,
		`"backlog-1.md" 파일에 명시된 작업을 수행해.`,
		gotPrompt,
	)

	_, err = os.Stat(filepath.Join(projectDir, "backlog-1.md"))
	require.ErrorIs(t, err, os.ErrNotExist)
}

func TestNewBacklogActionRunnerWithRunnerStartBacklogItemReturnsErrorWhenRunnerFails(t *testing.T) {
	t.Parallel()

	// Arrange
	projectDir := t.TempDir()
	item := mustNewBacklogItem(
		t,
		"Start title",
		"implement the start flow",
	)
	actionRunner := command.NewBacklogActionRunnerWithRunner(
		runnerFunc(func(_ context.Context, _ string, _ ...string) error {
			return errRunner
		}),
	)

	// Act
	err := actionRunner.StartBacklogItem(t.Context(), projectDir, item)

	// Assert
	require.ErrorIs(t, err, errRunner)
}

func TestNewBacklogActionRunnerWithRunnerRefineBacklogItem(t *testing.T) {
	t.Parallel()

	// Arrange
	projectDir := t.TempDir()
	item := mustNewBacklogItem(
		t,
		"Refine title",
		"original body",
	)

	var (
		gotDir    string
		gotPrompt string
	)

	actionRunner := command.NewBacklogActionRunnerWithRunner(
		runnerFunc(func(_ context.Context, dir string, args ...string) error {
			gotDir = dir

			require.Len(t, args, 1)
			gotPrompt = args[0]

			path := filepath.Join(dir, "backlog-1.md")
			err := os.WriteFile(path, []byte("refined body"), 0o600)
			require.NoError(t, err)

			return nil
		}),
	)

	// Act
	refinedItem, err := actionRunner.RefineBacklogItem(t.Context(), projectDir, item)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, refinedItem)
	require.NotSame(t, item, refinedItem)
	require.Equal(t, "refined body", refinedItem.Description())
	require.Equal(t, item.ID(), refinedItem.ID())
	require.Equal(t, projectDir, gotDir)
	require.Equal(
		t,
		`"backlog-1.md" 파일의 계획을 더 명확하고 간결하게 정리해. `+
			`만약, 추가적으로 결정해야 할 사항이 있다면, 추가해.`,
		gotPrompt,
	)

	_, err = os.Stat(filepath.Join(projectDir, "backlog-1.md"))
	require.ErrorIs(t, err, os.ErrNotExist)
}

func TestNewBacklogActionRunnerWithRunnerReturnsErrorWhenRunnerFails(t *testing.T) {
	t.Parallel()

	// Arrange
	projectDir := t.TempDir()
	item := mustNewBacklogItem(
		t,
		"Refine title",
		"original body",
	)
	actionRunner := command.NewBacklogActionRunnerWithRunner(
		runnerFunc(func(_ context.Context, _ string, _ ...string) error {
			return errRunner
		}),
	)

	// Act
	refinedItem, err := actionRunner.RefineBacklogItem(t.Context(), projectDir, item)

	// Assert
	require.ErrorIs(t, err, errRunner)
	require.Nil(t, refinedItem)
}

func TestNewBacklogActionRunnerWithRunnerRecommendWorktree(t *testing.T) {
	t.Parallel()

	// Arrange
	projectDir := t.TempDir()
	item := mustNewBacklogItem(
		t,
		"Recommend title",
		"plan body",
	)

	var (
		gotDir  string
		gotArgs []string
	)

	actionRunner := command.NewBacklogActionRunnerWithRunner(
		runnerFunc(func(_ context.Context, dir string, args ...string) error {
			gotDir = dir

			gotArgs = append([]string(nil), args...)

			//nolint:gosec // The test reads a file created under the temp project directory.
			content, err := os.ReadFile(filepath.Join(dir, "backlog-1.md"))
			require.NoError(t, err)
			require.Equal(t, "plan body", string(content))

			assertRecommendWorktreeArgs(t, args)

			err = os.WriteFile(filepath.Join(dir, "backlog-1.json"), []byte(`{
				"branch_name":"feat/project-backlog-start",
				"directory_name":"project-backlog-start"
			}`), 0o600)
			require.NoError(t, err)

			return nil
		}),
	)

	// Act
	worktree, err := actionRunner.RecommendWorktree(t.Context(), projectDir, item)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, worktree)
	require.Equal(t, "feat/project-backlog-start", worktree.Branch())
	require.Equal(t, "project-backlog-start", worktree.Dir())
	require.Equal(t, projectDir, gotDir)
	require.Len(t, gotArgs, 5)

	_, err = os.Stat(filepath.Join(projectDir, "backlog-1.md"))
	require.ErrorIs(t, err, os.ErrNotExist)

	_, err = os.Stat(filepath.Join(projectDir, "backlog-1.json"))
	require.ErrorIs(t, err, os.ErrNotExist)

	_, err = os.Stat(gotArgs[1])
	require.ErrorIs(t, err, os.ErrNotExist)
}

func assertRecommendWorktreeArgs(t *testing.T, args []string) {
	t.Helper()

	require.Len(t, args, 5)
	require.Equal(t, "--output-schema", args[0])

	schemaPath := args[1]
	//nolint:gosec // The test reads the temp schema file created for this runner invocation.
	schemaContent, err := os.ReadFile(schemaPath)
	require.NoError(t, err)
	require.JSONEq(t, `{
		"type":"object",
		"properties":{
			"branch_name":{"type":"string"},
			"directory_name":{"type":"string"}
		},
		"required":["branch_name","directory_name"],
		"additionalProperties":false
	}`, string(schemaContent))
	require.Equal(
		t,
		`Based on the contents of the "backlog-1.md" file, suggest a branch name and directory name for a git worktree.`,
		args[2],
	)
	require.Equal(t, "-o", args[3])
	require.Equal(t, "backlog-1.json", args[4])
}

func TestNewBacklogActionRunnerWithRunnerRecommendWorktreeReturnsErrorWhenRunnerFails(t *testing.T) {
	t.Parallel()

	// Arrange
	projectDir := t.TempDir()
	item := mustNewBacklogItem(
		t,
		"Recommend title",
		"plan body",
	)

	var schemaPath string

	actionRunner := command.NewBacklogActionRunnerWithRunner(
		runnerFunc(func(_ context.Context, _ string, args ...string) error {
			schemaPath = args[1]

			return errRunner
		}),
	)

	// Act
	worktree, err := actionRunner.RecommendWorktree(t.Context(), projectDir, item)

	// Assert
	require.ErrorIs(t, err, errRunner)
	require.Nil(t, worktree)

	_, statErr := os.Stat(filepath.Join(projectDir, "backlog-1.md"))
	require.ErrorIs(t, statErr, os.ErrNotExist)

	_, statErr = os.Stat(filepath.Join(projectDir, "backlog-1.json"))
	require.ErrorIs(t, statErr, os.ErrNotExist)

	_, statErr = os.Stat(schemaPath)
	require.ErrorIs(t, statErr, os.ErrNotExist)
}

func TestNewBacklogActionRunnerWithRunnerRecommendWorktreeReturnsErrorWhenOutputIsInvalid(t *testing.T) {
	t.Parallel()

	// Arrange
	projectDir := t.TempDir()
	item := mustNewBacklogItem(
		t,
		"Recommend title",
		"plan body",
	)
	actionRunner := command.NewBacklogActionRunnerWithRunner(
		runnerFunc(func(_ context.Context, dir string, _ ...string) error {
			err := os.WriteFile(filepath.Join(dir, "backlog-1.json"), []byte(`{`), 0o600)
			require.NoError(t, err)

			return nil
		}),
	)

	// Act
	worktree, err := actionRunner.RecommendWorktree(t.Context(), projectDir, item)

	// Assert
	require.Error(t, err)
	require.ErrorContains(t, err, "unmarshal recommended worktree file")
	require.Nil(t, worktree)
}

func TestNewBacklogActionRunnerWithRunnerRecommendWorktreeReturnsErrorWhenOutputFieldsAreEmpty(t *testing.T) {
	t.Parallel()

	// Arrange
	projectDir := t.TempDir()
	item := mustNewBacklogItem(
		t,
		"Recommend title",
		"plan body",
	)
	actionRunner := command.NewBacklogActionRunnerWithRunner(
		runnerFunc(func(_ context.Context, dir string, _ ...string) error {
			err := os.WriteFile(filepath.Join(dir, "backlog-1.json"), []byte(`{
				"branch_name":"",
				"directory_name":""
			}`), 0o600)
			require.NoError(t, err)

			return nil
		}),
	)

	// Act
	worktree, err := actionRunner.RecommendWorktree(t.Context(), projectDir, item)

	// Assert
	require.Error(t, err)
	require.ErrorContains(t, err, "recommended worktree branch name is empty")
	require.Nil(t, worktree)
}

func TestNewBacklogActionRunnerWithRunnerBuildsDraftPromptWhenDescriptionIsEmpty(t *testing.T) {
	t.Parallel()

	// Arrange
	projectDir := t.TempDir()
	item := mustNewBacklogItem(t, "Draft title", "")

	var gotPrompt string

	actionRunner := command.NewBacklogActionRunnerWithRunner(
		runnerFunc(func(_ context.Context, dir string, args ...string) error {
			require.Len(t, args, 1)
			gotPrompt = args[0]

			path := filepath.Join(dir, "backlog-1.md")
			err := os.WriteFile(path, []byte("# draft"), 0o600)
			require.NoError(t, err)

			return nil
		}),
	)

	// Act
	refinedItem, err := actionRunner.RefineBacklogItem(t.Context(), projectDir, item)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, refinedItem)
	require.Equal(t, "# draft", refinedItem.Description())
	require.Equal(
		t,
		`"backlog-1.md" 파일에 "Draft title" 에 대한 목표와 범위, `+
			`그리고 커밋 계획, 결정해야 할 사항 등을 포함한 계획 초안을 작성해. `+
			`특히, 결정해야 할 사항은 답변하기 쉬운 형식으로 만들어.`,
		gotPrompt,
	)
}

func TestNewBacklogActionRunnerWithRunnerBuildsStartPromptWhenDescriptionIsEmpty(t *testing.T) {
	t.Parallel()

	// Arrange
	projectDir := t.TempDir()
	item := mustNewBacklogItem(t, "Draft title", "")

	var gotPrompt string

	actionRunner := command.NewBacklogActionRunnerWithRunner(
		runnerFunc(func(_ context.Context, _ string, args ...string) error {
			require.Len(t, args, 1)
			gotPrompt = args[0]
			//nolint:gosec // The test reads a file created under the temp project directory.
			content, err := os.ReadFile(filepath.Join(projectDir, "backlog-1.md"))
			require.NoError(t, err)
			require.Empty(t, string(content))

			return nil
		}),
	)

	// Act
	err := actionRunner.StartBacklogItem(t.Context(), projectDir, item)

	// Assert
	require.NoError(t, err)
	require.Equal(
		t,
		`"Draft title" 작업을 수행해.`,
		gotPrompt,
	)

	_, err = os.Stat(filepath.Join(projectDir, "backlog-1.md"))
	require.ErrorIs(t, err, os.ErrNotExist)
}

type runnerFunc func(ctx context.Context, dir string, args ...string) error

func (f runnerFunc) Run(ctx context.Context, dir string, args ...string) error {
	return f(ctx, dir, args...)
}

func mustNewBacklogItem(
	t *testing.T,
	title string,
	description string,
) *domain.BacklogItem {
	t.Helper()

	item, err := domain.NewBacklogItem(
		"backlog-1",
		"project-1",
		title,
		description,
		domain.BacklogItemStatusOpen,
		"001",
	)
	require.NoError(t, err)

	return item
}
