package memory_test

import (
	"testing"

	"github.com/neatflowcv/worker/internal/pkg/domain"
	"github.com/neatflowcv/worker/internal/pkg/memory"
	"github.com/stretchr/testify/require"
)

func TestProjectRepository_ListReturnsZeroProjects(t *testing.T) {
	t.Parallel()

	// Arrange
	repo := memory.NewProjectRepository()

	// Act
	projects, err := repo.List(t.Context())

	// Assert
	require.NoError(t, err)
	require.Empty(t, projects)
}

func TestProjectRepository_ListReturnsOneProject(t *testing.T) {
	t.Parallel()

	// Arrange
	repo := memory.NewProjectRepository()
	project := domain.NewRepository("project-1", "worker", "https://github.com/neatflowcv/worker.git")
	repo.Append(project)

	// Act
	projects, err := repo.List(t.Context())

	// Assert
	require.NoError(t, err)
	require.Len(t, projects, 1)
	require.Same(t, project, projects[0])
}

func TestProjectRepository_ListReturnsTwoProjects(t *testing.T) {
	t.Parallel()

	// Arrange
	repo := memory.NewProjectRepository()
	first := domain.NewRepository("project-1", "worker", "https://github.com/neatflowcv/worker.git")
	second := domain.NewRepository("project-2", "worker-docs", "https://github.com/neatflowcv/docs.git")

	repo.Append(first, second)

	// Act
	projects, err := repo.List(t.Context())

	// Assert
	require.NoError(t, err)
	require.Len(t, projects, 2)
	require.Same(t, first, projects[0])
	require.Same(t, second, projects[1])
}
