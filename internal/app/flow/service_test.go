package flow_test

import (
	"testing"

	"github.com/neatflowcv/worker/internal/app/flow"
	"github.com/neatflowcv/worker/internal/pkg/domain"
	"github.com/neatflowcv/worker/internal/pkg/memory"
	"github.com/oklog/ulid/v2"
	"github.com/stretchr/testify/require"
)

func TestService_CreateProject(t *testing.T) {
	t.Parallel()

	// Arrange
	service := flow.NewService(
		memory.NewProjectRepository(),
		memory.NewWorkspace(),
	)

	// Act
	project, err := service.CreateProject(t.Context(), "worker", "https://github.com/neatflowcv/worker.git")

	// Assert
	require.NoError(t, err)
	require.NotNil(t, project)
	require.Equal(t, "worker", project.Name())
	require.Equal(t, "https://github.com/neatflowcv/worker.git", project.RepositoryURL())

	_, err = ulid.Parse(project.ID())
	require.NoError(t, err)
}

func TestService_CreateProjectReturnsErrorWhenNameAlreadyExists(t *testing.T) {
	t.Parallel()

	// Arrange
	repository := memory.NewProjectRepository()
	service := flow.NewService(
		repository,
		memory.NewWorkspace(),
	)
	repository.Append(domain.NewProject("project-1", "worker", "https://github.com/neatflowcv/worker.git"))

	// Act
	project, err := service.CreateProject(t.Context(), "worker", "https://github.com/neatflowcv/worker-2.git")

	// Assert
	require.ErrorIs(t, err, flow.ErrProjectAlreadyExists)
	require.Nil(t, project)

	projects, err := service.ListProjects(t.Context())
	require.NoError(t, err)
	require.Len(t, projects, 1)
	require.Equal(t, "worker", projects[0].Name())
	require.Equal(t, "https://github.com/neatflowcv/worker.git", projects[0].RepositoryURL())
}

func TestService_ListProjects(t *testing.T) {
	t.Parallel()

	// Arrange
	service := flow.NewService(
		memory.NewProjectRepository(),
		memory.NewWorkspace(),
	)

	// Act
	projects, err := service.ListProjects(t.Context())

	// Assert
	require.NoError(t, err)
	require.Empty(t, projects)
}
