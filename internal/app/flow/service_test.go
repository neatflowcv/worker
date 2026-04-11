package flow_test

import (
	"context"
	"errors"
	"testing"

	"github.com/neatflowcv/worker/internal/app/flow"
	"github.com/neatflowcv/worker/internal/pkg/domain"
	"github.com/neatflowcv/worker/internal/pkg/repository"
	"github.com/oklog/ulid/v2"
	"github.com/stretchr/testify/require"
)

func TestService_CreateProject(t *testing.T) {
	t.Parallel()

	// Arrange
	var createdProject *domain.Project

	projectRepository := newProjectRepositoryMock()
	projectRepository.GetByNameFunc = func(_ context.Context, _ string) (*domain.Project, error) {
		return nil, repository.ErrProjectNotFound
	}
	projectRepository.CreateFunc = func(_ context.Context, project *domain.Project) error {
		createdProject = project

		return nil
	}
	workspace := newWorkspaceMock()
	service := flow.NewService(projectRepository, nil, workspace)

	// Act
	project, err := service.CreateProject(t.Context(), "worker", "https://github.com/neatflowcv/worker.git")

	// Assert
	require.NoError(t, err)
	require.NotNil(t, project)
	require.Equal(t, "worker", project.Name())
	require.Equal(t, "https://github.com/neatflowcv/worker.git", project.RepositoryURL())
	require.Same(t, project, createdProject)
	require.Len(t, projectRepository.GetByNameCalls(), 1)
	require.Len(t, projectRepository.CreateCalls(), 1)
	require.Len(t, workspace.CreateWorkspaceCalls(), 1)

	_, err = ulid.Parse(project.ID())
	require.NoError(t, err)
}

func TestService_CreateProjectReturnsErrorWhenNameAlreadyExists(t *testing.T) {
	t.Parallel()

	// Arrange
	existingProject := domain.NewProject("project-1", "worker", "https://github.com/neatflowcv/worker.git")
	projectRepository := newProjectRepositoryMock()
	projectRepository.GetByNameFunc = func(_ context.Context, _ string) (*domain.Project, error) {
		return existingProject, nil
	}
	workspace := newWorkspaceMock()
	service := flow.NewService(projectRepository, nil, workspace)

	// Act
	project, err := service.CreateProject(t.Context(), "worker", "https://github.com/neatflowcv/worker-2.git")

	// Assert
	require.ErrorIs(t, err, flow.ErrProjectAlreadyExists)
	require.Nil(t, project)
	require.Len(t, projectRepository.GetByNameCalls(), 1)
	require.Empty(t, projectRepository.CreateCalls())
	require.Empty(t, workspace.CreateWorkspaceCalls())
}

func TestService_ListProjects(t *testing.T) {
	t.Parallel()

	// Arrange
	expectedProjects := []*domain.Project{
		domain.NewProject("project-1", "worker", "https://github.com/neatflowcv/worker.git"),
	}
	projectRepository := newProjectRepositoryMock()
	projectRepository.ListFunc = func(_ context.Context) ([]*domain.Project, error) {
		return expectedProjects, nil
	}
	service := flow.NewService(projectRepository, nil, newWorkspaceMock())

	// Act
	projects, err := service.ListProjects(t.Context())

	// Assert
	require.NoError(t, err)
	require.Equal(t, expectedProjects, projects)
	require.Len(t, projectRepository.ListCalls(), 1)
}

func TestService_CreateBacklogItem(t *testing.T) {
	t.Parallel()

	// Arrange
	projectRepository := newProjectRepositoryMock()
	projectRepository.GetByNameFunc = func(_ context.Context, _ string) (*domain.Project, error) {
		return domain.NewProject("project-1", "worker", "https://github.com/neatflowcv/worker.git"), nil
	}
	backlogItemRepository := newBacklogItemRepositoryMock()
	service := flow.NewService(
		projectRepository,
		backlogItemRepository,
		newWorkspaceMock(),
	)

	// Act
	item, err := service.CreateBacklogItem(t.Context(), "worker", "Add backlog create", "cli implementation")

	// Assert
	require.NoError(t, err)
	require.NotNil(t, item)
	require.Equal(t, "project-1", item.ProjectID())
	require.Equal(t, "Add backlog create", item.Title())
	require.Equal(t, "cli implementation", item.Description())
	require.Len(t, backlogItemRepository.CreateBacklogItemCalls(), 1)
	require.Same(t, item, backlogItemRepository.CreateBacklogItemCalls()[0].Item)

	_, err = ulid.Parse(item.ID())
	require.NoError(t, err)
}

func TestService_CreateBacklogItemReturnsErrorWhenProjectDoesNotExist(t *testing.T) {
	t.Parallel()

	// Arrange
	projectRepository := newProjectRepositoryMock()
	projectRepository.GetByNameFunc = func(_ context.Context, _ string) (*domain.Project, error) {
		return nil, repository.ErrProjectNotFound
	}
	backlogItemRepository := newBacklogItemRepositoryMock()
	service := flow.NewService(projectRepository, backlogItemRepository, newWorkspaceMock())

	// Act
	item, err := service.CreateBacklogItem(t.Context(), "worker", "Add backlog create", "cli implementation")

	// Assert
	require.ErrorIs(t, err, repository.ErrProjectNotFound)
	require.Nil(t, item)
	require.Empty(t, backlogItemRepository.CreateBacklogItemCalls())
}

func TestService_CreateBacklogItemReturnsErrorWhenRepositoryFails(t *testing.T) {
	t.Parallel()

	// Arrange
	projectRepository := newProjectRepositoryMock()
	projectRepository.GetByNameFunc = func(_ context.Context, _ string) (*domain.Project, error) {
		return domain.NewProject("project-1", "worker", "https://github.com/neatflowcv/worker.git"), nil
	}
	backlogItemRepository := newBacklogItemRepositoryMock()
	backlogItemRepository.CreateBacklogItemFunc = func(_ context.Context, _ *domain.BacklogItem) error {
		return errBacklogItemRepository
	}
	service := flow.NewService(projectRepository, backlogItemRepository, newWorkspaceMock())

	// Act
	item, err := service.CreateBacklogItem(t.Context(), "worker", "Add backlog create", "cli implementation")

	// Assert
	require.ErrorIs(t, err, errBacklogItemRepository)
	require.Nil(t, item)
}

func TestService_ListBacklogItems(t *testing.T) {
	t.Parallel()

	// Arrange
	expectedItems := []*domain.BacklogItem{
		domain.NewBacklogItem("backlog-2", "project-1", "Second", "", "000000000002"),
		domain.NewBacklogItem("backlog-3", "project-1", "Third", "", "000000000003"),
	}
	projectRepository := newProjectRepositoryMock()
	projectRepository.GetByNameFunc = func(_ context.Context, _ string) (*domain.Project, error) {
		return domain.NewProject("project-1", "worker", "https://github.com/neatflowcv/worker.git"), nil
	}
	backlogItemRepository := newBacklogItemRepositoryMock()
	backlogItemRepository.ListBacklogItemsFunc = func(
		_ context.Context,
		projectID, afterID string,
		limit int,
	) ([]*domain.BacklogItem, error) {
		return expectedItems, nil
	}
	service := flow.NewService(projectRepository, backlogItemRepository, newWorkspaceMock())

	// Act
	items, err := service.ListBacklogItems(t.Context(), "worker", "backlog-1", 5)

	// Assert
	require.NoError(t, err)
	require.Equal(t, expectedItems, items)
	require.Len(t, backlogItemRepository.ListBacklogItemsCalls(), 1)
	require.Equal(t, "project-1", backlogItemRepository.ListBacklogItemsCalls()[0].ProjectID)
	require.Equal(t, "backlog-1", backlogItemRepository.ListBacklogItemsCalls()[0].AfterID)
	require.Equal(t, 5, backlogItemRepository.ListBacklogItemsCalls()[0].Limit)
}

func TestService_ListBacklogItemsAppliesDefaultLimit(t *testing.T) {
	t.Parallel()

	// Arrange
	projectRepository := newProjectRepositoryMock()
	projectRepository.GetByNameFunc = func(_ context.Context, _ string) (*domain.Project, error) {
		return domain.NewProject("project-1", "worker", "https://github.com/neatflowcv/worker.git"), nil
	}
	backlogItemRepository := newBacklogItemRepositoryMock()
	service := flow.NewService(projectRepository, backlogItemRepository, newWorkspaceMock())

	// Act
	items, err := service.ListBacklogItems(t.Context(), "worker", "", 0)

	// Assert
	require.NoError(t, err)
	require.Empty(t, items)
	require.Len(t, backlogItemRepository.ListBacklogItemsCalls(), 1)
	require.Equal(t, 20, backlogItemRepository.ListBacklogItemsCalls()[0].Limit)
}

func TestService_ListBacklogItemsReturnsErrorWhenProjectDoesNotExist(t *testing.T) {
	t.Parallel()

	// Arrange
	projectRepository := newProjectRepositoryMock()
	projectRepository.GetByNameFunc = func(_ context.Context, _ string) (*domain.Project, error) {
		return nil, repository.ErrProjectNotFound
	}
	backlogItemRepository := newBacklogItemRepositoryMock()
	service := flow.NewService(projectRepository, backlogItemRepository, newWorkspaceMock())

	// Act
	items, err := service.ListBacklogItems(t.Context(), "worker", "", 20)

	// Assert
	require.ErrorIs(t, err, repository.ErrProjectNotFound)
	require.Nil(t, items)
	require.Empty(t, backlogItemRepository.ListBacklogItemsCalls())
}

func TestService_ListBacklogItemsReturnsErrorWhenRepositoryFails(t *testing.T) {
	t.Parallel()

	// Arrange
	projectRepository := newProjectRepositoryMock()
	projectRepository.GetByNameFunc = func(_ context.Context, _ string) (*domain.Project, error) {
		return domain.NewProject("project-1", "worker", "https://github.com/neatflowcv/worker.git"), nil
	}
	backlogItemRepository := newBacklogItemRepositoryMock()
	backlogItemRepository.ListBacklogItemsFunc = func(
		_ context.Context,
		_ string,
		_ string,
		_ int,
	) ([]*domain.BacklogItem, error) {
		return nil, errBacklogItemRepository
	}
	service := flow.NewService(projectRepository, backlogItemRepository, newWorkspaceMock())

	// Act
	items, err := service.ListBacklogItems(t.Context(), "worker", "", 20)

	// Assert
	require.ErrorIs(t, err, errBacklogItemRepository)
	require.Nil(t, items)
}

func newProjectRepositoryMock() *ProjectRepositoryMock {
	var mock ProjectRepositoryMock

	mock.CreateFunc = func(_ context.Context, _ *domain.Project) error {
		return nil
	}
	mock.GetByNameFunc = func(_ context.Context, _ string) (*domain.Project, error) {
		return nil, repository.ErrProjectNotFound
	}
	mock.ListFunc = func(_ context.Context) ([]*domain.Project, error) {
		return nil, nil
	}

	return &mock
}

func newBacklogItemRepositoryMock() *BacklogItemRepositoryMock {
	var mock BacklogItemRepositoryMock

	mock.CreateBacklogItemFunc = func(_ context.Context, _ *domain.BacklogItem) error {
		return nil
	}
	mock.ListBacklogItemsFunc = func(
		_ context.Context,
		_ string,
		_ string,
		_ int,
	) ([]*domain.BacklogItem, error) {
		return []*domain.BacklogItem{}, nil
	}

	return &mock
}

func newWorkspaceMock() *WorkspaceMock {
	var mock WorkspaceMock

	mock.CreateWorkspaceFunc = func(_ context.Context, _ *domain.Project) error {
		return nil
	}

	return &mock
}

var errBacklogItemRepository = errors.New("backlog item repository error")
