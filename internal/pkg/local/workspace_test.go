package local_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/neatflowcv/worker/internal/pkg/domain"
	"github.com/neatflowcv/worker/internal/pkg/local"
	"github.com/stretchr/testify/require"
)

func TestWorkspace_CreateWorkspace(t *testing.T) {
	t.Parallel()

	// Arrange
	rootDir := t.TempDir()
	repositoryURL := createRepository(t)
	project := domain.NewProject("project-1", "worker", repositoryURL)
	workspace := local.NewWorkspace(rootDir)

	// Act
	err := workspace.CreateWorkspace(t.Context(), project)

	// Assert
	require.NoError(t, err)

	mainDir := filepath.Join(rootDir, project.ID(), "main")

	info, err := os.Stat(mainDir)
	require.NoError(t, err)
	require.True(t, info.IsDir())

	readmePath := filepath.Join(mainDir, "README.md")
	//nolint:gosec // Test reads a file from a path created within t.TempDir.
	content, err := os.ReadFile(readmePath)
	require.NoError(t, err)
	require.Equal(t, "worker\n", string(content))
}

func createRepository(t *testing.T) string {
	t.Helper()

	dir := t.TempDir()

	repository, err := git.PlainInit(dir, false)
	require.NoError(t, err)

	err = os.WriteFile(filepath.Join(dir, "README.md"), []byte("worker\n"), 0o600)
	require.NoError(t, err)

	worktree, err := repository.Worktree()
	require.NoError(t, err)

	_, err = worktree.Add("README.md")
	require.NoError(t, err)

	_, err = worktree.Commit("initial commit", &git.CommitOptions{
		All:               false,
		AllowEmptyCommits: false,
		Author: &object.Signature{
			Name:  "tester",
			Email: "tester@example.com",
			When:  time.Unix(0, 0),
		},
		Committer: nil,
		Parents:   nil,
		SignKey:   nil,
		Signer:    nil,
		Amend:     false,
	})
	require.NoError(t, err)

	return dir
}
