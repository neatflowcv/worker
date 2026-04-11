package badger_test

import (
	"testing"

	"github.com/neatflowcv/worker/internal/pkg/badger"
	"github.com/neatflowcv/worker/internal/pkg/domain"
	"github.com/neatflowcv/worker/internal/pkg/repository"
	"github.com/stretchr/testify/require"
)

func TestProjectRepository_ListReturnsZeroProjects(t *testing.T) {
	t.Parallel()

	// Arrange
	repo := newProjectRepository(t)

	// Act
	projects, err := repo.List(t.Context())

	// Assert
	require.NoError(t, err)
	require.Empty(t, projects)
}

func TestProjectRepository_CreatePersistsProject(t *testing.T) {
	t.Parallel()

	// Arrange
	dir := t.TempDir()
	repo, database := newProjectRepositoryAt(t, dir)
	project := domain.NewProject("project-1", "worker", "https://github.com/neatflowcv/worker.git")

	// Act
	err := repo.Create(t.Context(), project)
	require.NoError(t, err)
	require.NoError(t, database.Close())

	reopened, _ := newProjectRepositoryAt(t, dir)
	projects, err := reopened.List(t.Context())
	require.NoError(t, err)

	projectByName, err := reopened.GetByName(t.Context(), "worker")

	// Assert
	require.NoError(t, err)
	require.Len(t, projects, 1)
	require.Equal(t, project.ID(), projects[0].ID())
	require.Equal(t, project.Name(), projects[0].Name())
	require.Equal(t, project.RepositoryURL(), projects[0].RepositoryURL())
	require.NotNil(t, projectByName)
	require.Equal(t, project.ID(), projectByName.ID())
	require.Equal(t, project.Name(), projectByName.Name())
	require.Equal(t, project.RepositoryURL(), projectByName.RepositoryURL())
}

func TestProjectRepository_ListReturnsTwoProjects(t *testing.T) {
	t.Parallel()

	// Arrange
	repo := newProjectRepository(t)
	first := domain.NewProject("project-1", "worker", "https://github.com/neatflowcv/worker.git")
	second := domain.NewProject("project-2", "worker-docs", "https://github.com/neatflowcv/docs.git")

	require.NoError(t, repo.Create(t.Context(), first))
	require.NoError(t, repo.Create(t.Context(), second))

	// Act
	projects, err := repo.List(t.Context())

	// Assert
	require.NoError(t, err)
	require.Len(t, projects, 2)
	require.Equal(t, first.ID(), projects[0].ID())
	require.Equal(t, first.Name(), projects[0].Name())
	require.Equal(t, first.RepositoryURL(), projects[0].RepositoryURL())
	require.Equal(t, second.ID(), projects[1].ID())
	require.Equal(t, second.Name(), projects[1].Name())
	require.Equal(t, second.RepositoryURL(), projects[1].RepositoryURL())
}

func TestProjectRepository_GetByNameReturnsProject(t *testing.T) {
	t.Parallel()

	// Arrange
	repo := newProjectRepository(t)
	first := domain.NewProject("project-1", "worker", "https://github.com/neatflowcv/worker.git")
	second := domain.NewProject("project-2", "worker-docs", "https://github.com/neatflowcv/docs.git")

	require.NoError(t, repo.Create(t.Context(), first))
	require.NoError(t, repo.Create(t.Context(), second))

	// Act
	project, err := repo.GetByName(t.Context(), "worker-docs")

	// Assert
	require.NoError(t, err)
	require.NotNil(t, project)
	require.Equal(t, second.ID(), project.ID())
	require.Equal(t, second.Name(), project.Name())
	require.Equal(t, second.RepositoryURL(), project.RepositoryURL())
}

func TestProjectRepository_GetByNameReturnsNilWhenMissing(t *testing.T) {
	t.Parallel()

	// Arrange
	repo := newProjectRepository(t)
	project := domain.NewProject("project-1", "worker", "https://github.com/neatflowcv/worker.git")

	require.NoError(t, repo.Create(t.Context(), project))

	// Act
	actual, err := repo.GetByName(t.Context(), "missing")

	// Assert
	require.ErrorIs(t, err, repository.ErrProjectNotFound)
	require.Nil(t, actual)
}

func newProjectRepository(t *testing.T) *badger.ProjectRepository {
	t.Helper()

	repo, _ := newProjectRepositoryAt(t, t.TempDir())

	return repo
}

func newProjectRepositoryAt(t *testing.T, dir string) (*badger.ProjectRepository, *badger.Database) {
	t.Helper()

	database, err := badger.NewDatabase(dir)
	require.NoError(t, err)

	t.Cleanup(func() {
		require.NoError(t, database.Close())
	})

	return badger.NewProjectRepository(database), database
}
