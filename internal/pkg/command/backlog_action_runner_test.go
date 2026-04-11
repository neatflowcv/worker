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

func TestNewBacklogActionRunnerWithRunnerRefineBacklogItem(t *testing.T) {
	t.Parallel()

	// Arrange
	projectDir := t.TempDir()
	item := mustNewBacklogItem(
		t,
		"backlog-1",
		"project-1",
		"Refine title",
		"original body",
		"001",
	)

	var (
		gotDir    string
		gotPrompt string
	)

	actionRunner := command.NewBacklogActionRunnerWithRunner(
		runnerFunc(func(_ context.Context, dir string, prompt string) error {
			gotDir = dir
			gotPrompt = prompt

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
		`Read and refine the backlog item markdown file "backlog-1.md". `+
			`Improve clarity and structure while preserving intent. `+
			`Modify the file in place. Do not rename the file.`,
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
		"backlog-1",
		"project-1",
		"Refine title",
		"original body",
		"001",
	)
	actionRunner := command.NewBacklogActionRunnerWithRunner(
		runnerFunc(func(_ context.Context, _ string, _ string) error {
			return errRunner
		}),
	)

	// Act
	refinedItem, err := actionRunner.RefineBacklogItem(t.Context(), projectDir, item)

	// Assert
	require.ErrorIs(t, err, errRunner)
	require.Nil(t, refinedItem)
}

func TestNewBacklogActionRunnerWithRunnerBuildsDraftPromptWhenDescriptionIsEmpty(t *testing.T) {
	t.Parallel()

	// Arrange
	projectDir := t.TempDir()
	item := mustNewBacklogItem(t, "backlog-1", "project-1", "Draft title", "", "001")

	var gotPrompt string

	actionRunner := command.NewBacklogActionRunnerWithRunner(
		runnerFunc(func(_ context.Context, dir string, prompt string) error {
			gotPrompt = prompt

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
		`Read the backlog item markdown file "backlog-1.md". `+
			`The file is empty, so write an initial draft for the backlog item based on its title "Draft title". `+
			`Create a concise, actionable markdown draft with clear structure. `+
			`Modify the file in place. Do not rename the file.`,
		gotPrompt,
	)
}

type runnerFunc func(ctx context.Context, dir string, prompt string) error

func (f runnerFunc) Run(ctx context.Context, dir string, prompt string) error {
	return f(ctx, dir, prompt)
}

func mustNewBacklogItem(
	t *testing.T,
	id string,
	projectID string,
	title string,
	description string,
	orderKey string,
) *domain.BacklogItem {
	t.Helper()

	item, err := domain.NewBacklogItem(id, projectID, title, description, orderKey)
	require.NoError(t, err)

	return item
}
