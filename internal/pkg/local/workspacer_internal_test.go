package local

import (
	"testing"

	githttp "github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/neatflowcv/worker/internal/pkg/domain"
	"github.com/stretchr/testify/require"
)

func TestNewPushOptionsUsesProjectAuth(t *testing.T) {
	t.Parallel()

	// Arrange
	project := domain.NewProject(
		"project-1",
		"worker",
		"https://github.com/neatflowcv/worker.git",
		domain.NewAuth("user", "pass"),
	)
	worktree := domain.NewWorktree("backlog-1", "/tmp/project-1/backlog-1")

	// Act
	options := newPushOptions(project, worktree)

	// Assert
	auth, ok := options.Auth.(*githttp.BasicAuth)
	require.True(t, ok)
	require.Equal(t, "user", auth.Username)
	require.Equal(t, "pass", auth.Password)
}

func TestNewPushOptionsReturnsNilAuthWhenProjectHasNoAuth(t *testing.T) {
	t.Parallel()

	// Arrange
	project := domain.NewProject("project-1", "worker", "https://github.com/neatflowcv/worker.git", nil)
	worktree := domain.NewWorktree("backlog-1", "/tmp/project-1/backlog-1")

	// Act
	options := newPushOptions(project, worktree)

	// Assert
	require.Nil(t, options.Auth)
}
