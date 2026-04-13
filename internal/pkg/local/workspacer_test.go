package local_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/neatflowcv/worker/internal/pkg/domain"
	"github.com/neatflowcv/worker/internal/pkg/local"
	"github.com/stretchr/testify/require"
)

func TestWorkspacer_PrepareWorkspace(t *testing.T) {
	t.Parallel()

	// Arrange
	rootDir := t.TempDir()
	repositoryURL := createRepository(t)
	project := domain.NewProject("project-1", "worker", repositoryURL, nil)
	workspacer := local.NewWorkspacer(rootDir)

	// Act
	workspace, err := workspacer.PrepareWorkspace(t.Context(), project)

	// Assert
	require.NoError(t, err)
	require.Equal(t, filepath.Join(rootDir, project.ID()), workspace.Root())
	require.Equal(t, filepath.Join(rootDir, project.ID(), "main"), workspace.Main())

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

func TestWorkspacer_PrepareWorkspacePullsWhenRepositoryExists(t *testing.T) {
	t.Parallel()

	// Arrange
	rootDir := t.TempDir()
	repositoryURL := createRepository(t)
	project := domain.NewProject("project-1", "worker", repositoryURL, nil)
	workspacer := local.NewWorkspacer(rootDir)

	_, err := workspacer.PrepareWorkspace(t.Context(), project)
	require.NoError(t, err)

	appendCommitToRepository(t, repositoryURL, "README.md", "updated\n", "update readme")

	// Act
	workspace, err := workspacer.PrepareWorkspace(t.Context(), project)

	// Assert
	require.NoError(t, err)
	require.Equal(t, filepath.Join(rootDir, project.ID(), "main"), workspace.Main())

	readmePath := filepath.Join(rootDir, project.ID(), "main", "README.md")
	//nolint:gosec // Test reads a file from a path created within t.TempDir.
	content, err := os.ReadFile(readmePath)
	require.NoError(t, err)
	require.Equal(t, "worker\nupdated\n", string(content))
}

func TestWorkspacer_PrepareWorkspaceReturnsWorkspace(t *testing.T) {
	t.Parallel()

	// Arrange
	rootDir := t.TempDir()
	repositoryURL := createRepository(t)
	project := domain.NewProject("project-1", "worker", repositoryURL, nil)
	workspacer := local.NewWorkspacer(rootDir)
	workspace, err := workspacer.PrepareWorkspace(t.Context(), project)

	// Assert
	require.NoError(t, err)
	require.Equal(t, filepath.Join(rootDir, "project-1"), workspace.Root())
	require.Equal(t, filepath.Join(rootDir, "project-1", "main"), workspace.Main())
}

func TestWorkspacer_CreateWorktree(t *testing.T) {
	t.Parallel()

	// Arrange
	rootDir := t.TempDir()
	repositoryURL := createRepository(t)
	project := domain.NewProject("project-1", "worker", repositoryURL, nil)
	item, err := domain.NewBacklogItem(
		"backlog-1",
		project.ID(),
		"First",
		"description",
		domain.BacklogItemStatusOpen,
		"000000000001",
	)
	require.NoError(t, err)

	workspacer := local.NewWorkspacer(rootDir)
	workspace, err := workspacer.PrepareWorkspace(t.Context(), project)
	require.NoError(t, err)

	// Act
	worktree, err := workspacer.CreateWorktree(t.Context(), project, workspace, item)

	// Assert
	require.NoError(t, err)
	require.Equal(t, item.ID(), worktree.Branch())
	require.Equal(t, filepath.Join(rootDir, project.ID(), item.ID()), worktree.Dir())

	info, err := os.Stat(worktree.Dir())
	require.NoError(t, err)
	require.True(t, info.IsDir())

	readmePath := filepath.Join(worktree.Dir(), "README.md")
	//nolint:gosec // Test reads a file from a path created within t.TempDir.
	content, err := os.ReadFile(readmePath)
	require.NoError(t, err)
	require.Equal(t, "worker\n", string(content))

	assertGitBranchExists(t, workspace.Main(), item.ID())
}

func TestWorkspacer_CloseWorktree(t *testing.T) {
	t.Parallel()

	// Arrange
	rootDir := t.TempDir()
	repositoryURL := createRepository(t)
	project := domain.NewProject("project-1", "worker", repositoryURL, nil)
	item, err := domain.NewBacklogItem(
		"backlog-1",
		project.ID(),
		"First",
		"description",
		domain.BacklogItemStatusOpen,
		"000000000001",
	)
	require.NoError(t, err)

	workspacer := local.NewWorkspacer(rootDir)
	workspace, err := workspacer.PrepareWorkspace(t.Context(), project)
	require.NoError(t, err)

	worktree, err := workspacer.CreateWorktree(t.Context(), project, workspace, item)
	require.NoError(t, err)

	// Act
	err = workspacer.CloseWorktree(t.Context(), project, worktree)

	// Assert
	require.NoError(t, err)
	info, err := os.Stat(worktree.Dir())
	require.NoError(t, err)
	require.True(t, info.IsDir())
	assertGitBranchExists(t, workspace.Main(), item.ID())
	assertGitBranchExists(t, repositoryURL, item.ID())
}

func assertGitBranchExists(t *testing.T, repositoryDir string, branch string) {
	t.Helper()

	repository, err := git.PlainOpen(repositoryDir)
	require.NoError(t, err)

	reference, err := repository.Reference(plumbing.NewBranchReferenceName(branch), true)
	require.NoError(t, err)
	require.Equal(t, plumbing.NewBranchReferenceName(branch), reference.Name())
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

func appendCommitToRepository(t *testing.T, repositoryDir string, path string, content string, message string) {
	t.Helper()

	//nolint:gosec // Test writes a file within t.TempDir.
	file, err := os.OpenFile(filepath.Join(repositoryDir, path), os.O_APPEND|os.O_WRONLY, 0)
	require.NoError(t, err)

	_, err = file.WriteString(content)
	require.NoError(t, err)

	err = file.Close()
	require.NoError(t, err)

	repository, err := git.PlainOpen(repositoryDir)
	require.NoError(t, err)

	worktree, err := repository.Worktree()
	require.NoError(t, err)

	_, err = worktree.Add(path)
	require.NoError(t, err)

	_, err = worktree.Commit(message, &git.CommitOptions{
		All:               false,
		AllowEmptyCommits: false,
		Author: &object.Signature{
			Name:  "tester",
			Email: "tester@example.com",
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
