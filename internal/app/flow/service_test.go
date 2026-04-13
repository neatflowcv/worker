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
	projectRepository.GetProjectByNameFunc = func(_ context.Context, _ string) (*domain.Project, error) {
		return nil, repository.ErrProjectNotFound
	}
	projectRepository.CreateProjectFunc = func(_ context.Context, project *domain.Project) error {
		createdProject = project

		return nil
	}
	workspace := newWorkspaceMock()
	service := flow.NewService(projectRepository, nil, workspace, nil)

	// Act
	project, err := service.CreateProject(
		t.Context(),
		"worker",
		"https://github.com/neatflowcv/worker.git",
		nil,
	)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, project)
	require.Equal(t, "worker", project.Name())
	require.Equal(t, "https://github.com/neatflowcv/worker.git", project.RepositoryURL())
	require.Same(t, project, createdProject)
	require.Len(t, projectRepository.GetProjectByNameCalls(), 1)
	require.Len(t, projectRepository.CreateProjectCalls(), 1)
	require.Len(t, workspace.PrepareWorkspaceCalls(), 1)

	_, err = ulid.Parse(project.ID())
	require.NoError(t, err)
}

func TestService_CreateProjectWithAuth(t *testing.T) {
	t.Parallel()

	// Arrange
	var createdProject *domain.Project

	projectRepository := newProjectRepositoryMock()
	projectRepository.GetProjectByNameFunc = func(_ context.Context, _ string) (*domain.Project, error) {
		return nil, repository.ErrProjectNotFound
	}
	projectRepository.CreateProjectFunc = func(_ context.Context, project *domain.Project) error {
		createdProject = project

		return nil
	}
	workspace := newWorkspaceMock()
	service := flow.NewService(projectRepository, nil, workspace, nil)
	auth := domain.NewAuth("user", "pass")

	// Act
	project, err := service.CreateProject(
		t.Context(),
		"worker",
		"https://github.com/neatflowcv/worker.git",
		auth,
	)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, project)
	require.NotNil(t, project.Auth())
	require.Equal(t, "user", project.Auth().Username())
	require.Equal(t, "pass", project.Auth().Password())
	require.Same(t, project, createdProject)
}

func TestService_CreateProjectReturnsErrorWhenNameAlreadyExists(t *testing.T) {
	t.Parallel()

	// Arrange
	existingProject := domain.NewProject("project-1", "worker", "https://github.com/neatflowcv/worker.git", nil)
	projectRepository := newProjectRepositoryMock()
	projectRepository.GetProjectByNameFunc = func(_ context.Context, _ string) (*domain.Project, error) {
		return existingProject, nil
	}
	workspace := newWorkspaceMock()
	service := flow.NewService(projectRepository, nil, workspace, nil)

	// Act
	project, err := service.CreateProject(
		t.Context(),
		"worker",
		"https://github.com/neatflowcv/worker-2.git",
		nil,
	)

	// Assert
	require.ErrorIs(t, err, flow.ErrProjectAlreadyExists)
	require.Nil(t, project)
	require.Len(t, projectRepository.GetProjectByNameCalls(), 1)
	require.Empty(t, projectRepository.CreateProjectCalls())
	require.Empty(t, workspace.PrepareWorkspaceCalls())
}

func TestService_ListProjects(t *testing.T) {
	t.Parallel()

	// Arrange
	expectedProjects := []*domain.Project{
		domain.NewProject("project-1", "worker", "https://github.com/neatflowcv/worker.git", nil),
	}
	projectRepository := newProjectRepositoryMock()
	projectRepository.ListProjectsFunc = func(_ context.Context) ([]*domain.Project, error) {
		return expectedProjects, nil
	}
	service := flow.NewService(projectRepository, nil, newWorkspaceMock(), nil)

	// Act
	projects, err := service.ListProjects(t.Context())

	// Assert
	require.NoError(t, err)
	require.Equal(t, expectedProjects, projects)
	require.Len(t, projectRepository.ListProjectsCalls(), 1)
}

func TestService_CreateBacklogItem(t *testing.T) {
	t.Parallel()

	// Arrange
	projectRepository := newProjectRepositoryMock()
	projectRepository.GetProjectByNameFunc = func(_ context.Context, _ string) (*domain.Project, error) {
		return domain.NewProject("project-1", "worker", "https://github.com/neatflowcv/worker.git", nil), nil
	}
	backlogItemRepository := newBacklogItemRepositoryMock()
	service := flow.NewService(
		projectRepository,
		backlogItemRepository,
		newWorkspaceMock(),
		nil,
	)

	// Act
	item, err := service.CreateBacklogItem(t.Context(), "worker", "Add backlog create", "cli implementation")

	// Assert
	require.NoError(t, err)
	require.NotNil(t, item)
	require.Equal(t, "project-1", item.ProjectID())
	require.Equal(t, "Add backlog create", item.Title())
	require.Equal(t, "cli implementation", item.Description())
	require.Equal(t, domain.BacklogItemStatusOpen, item.Status())
	require.Len(t, backlogItemRepository.CreateBacklogItemCalls(), 1)
	require.Same(t, item, backlogItemRepository.CreateBacklogItemCalls()[0].Item)

	_, err = ulid.Parse(item.ID())
	require.NoError(t, err)
}

func TestService_CreateBacklogItemReturnsErrorWhenProjectDoesNotExist(t *testing.T) {
	t.Parallel()

	// Arrange
	projectRepository := newProjectRepositoryMock()
	projectRepository.GetProjectByNameFunc = func(_ context.Context, _ string) (*domain.Project, error) {
		return nil, repository.ErrProjectNotFound
	}
	backlogItemRepository := newBacklogItemRepositoryMock()
	service := flow.NewService(projectRepository, backlogItemRepository, newWorkspaceMock(), nil)

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
	projectRepository.GetProjectByNameFunc = func(_ context.Context, _ string) (*domain.Project, error) {
		return domain.NewProject("project-1", "worker", "https://github.com/neatflowcv/worker.git", nil), nil
	}
	backlogItemRepository := newBacklogItemRepositoryMock()
	backlogItemRepository.CreateBacklogItemFunc = func(_ context.Context, _ *domain.BacklogItem) error {
		return errBacklogItemRepository
	}
	service := flow.NewService(projectRepository, backlogItemRepository, newWorkspaceMock(), nil)

	// Act
	item, err := service.CreateBacklogItem(t.Context(), "worker", "Add backlog create", "cli implementation")

	// Assert
	require.ErrorIs(t, err, errBacklogItemRepository)
	require.Nil(t, item)
}

func TestService_GetBacklogItem(t *testing.T) {
	t.Parallel()

	// Arrange
	expectedItem := mustNewBacklogItem(t, "backlog-1", "project-1", "First", "desc", "000000000001")
	projectRepository := newProjectRepositoryMock()
	projectRepository.GetProjectByNameFunc = func(_ context.Context, _ string) (*domain.Project, error) {
		return domain.NewProject("project-1", "worker", "https://github.com/neatflowcv/worker.git", nil), nil
	}
	backlogItemRepository := newBacklogItemRepositoryMock()
	backlogItemRepository.GetBacklogItemFunc = func(_ context.Context, _ string) (*domain.BacklogItem, error) {
		return expectedItem, nil
	}
	service := flow.NewService(projectRepository, backlogItemRepository, newWorkspaceMock(), nil)

	// Act
	item, err := service.GetBacklogItem(t.Context(), "worker", "backlog-1")

	// Assert
	require.NoError(t, err)
	require.Same(t, expectedItem, item)
	require.Len(t, backlogItemRepository.GetBacklogItemCalls(), 1)
	require.Equal(t, "backlog-1", backlogItemRepository.GetBacklogItemCalls()[0].ID)
}

func TestService_GetBacklogItemReturnsErrorWhenProjectDoesNotExist(t *testing.T) {
	t.Parallel()

	// Arrange
	projectRepository := newProjectRepositoryMock()
	projectRepository.GetProjectByNameFunc = func(_ context.Context, _ string) (*domain.Project, error) {
		return nil, repository.ErrProjectNotFound
	}
	backlogItemRepository := newBacklogItemRepositoryMock()
	service := flow.NewService(projectRepository, backlogItemRepository, newWorkspaceMock(), nil)

	// Act
	item, err := service.GetBacklogItem(t.Context(), "worker", "backlog-1")

	// Assert
	require.ErrorIs(t, err, repository.ErrProjectNotFound)
	require.Nil(t, item)
	require.Empty(t, backlogItemRepository.GetBacklogItemCalls())
}

func TestService_GetBacklogItemReturnsErrorWhenBacklogItemBelongsToAnotherProject(t *testing.T) {
	t.Parallel()

	// Arrange
	projectRepository := newProjectRepositoryMock()
	projectRepository.GetProjectByNameFunc = func(_ context.Context, _ string) (*domain.Project, error) {
		return domain.NewProject("project-1", "worker", "https://github.com/neatflowcv/worker.git", nil), nil
	}
	backlogItemRepository := newBacklogItemRepositoryMock()
	backlogItemRepository.GetBacklogItemFunc = func(_ context.Context, _ string) (*domain.BacklogItem, error) {
		return mustNewBacklogItem(t, "backlog-1", "project-2", "First", "desc", "000000000001"), nil
	}
	service := flow.NewService(projectRepository, backlogItemRepository, newWorkspaceMock(), nil)

	// Act
	item, err := service.GetBacklogItem(t.Context(), "worker", "backlog-1")

	// Assert
	require.ErrorIs(t, err, repository.ErrBacklogItemNotFound)
	require.Nil(t, item)
}

func TestService_GetBacklogItemReturnsErrorWhenRepositoryFails(t *testing.T) {
	t.Parallel()

	// Arrange
	projectRepository := newProjectRepositoryMock()
	projectRepository.GetProjectByNameFunc = func(_ context.Context, _ string) (*domain.Project, error) {
		return domain.NewProject("project-1", "worker", "https://github.com/neatflowcv/worker.git", nil), nil
	}
	backlogItemRepository := newBacklogItemRepositoryMock()
	backlogItemRepository.GetBacklogItemFunc = func(_ context.Context, _ string) (*domain.BacklogItem, error) {
		return nil, errBacklogItemRepository
	}
	service := flow.NewService(projectRepository, backlogItemRepository, newWorkspaceMock(), nil)

	// Act
	item, err := service.GetBacklogItem(t.Context(), "worker", "backlog-1")

	// Assert
	require.ErrorIs(t, err, errBacklogItemRepository)
	require.Nil(t, item)
}

func TestService_ListBacklogItems(t *testing.T) {
	t.Parallel()

	// Arrange
	expectedItems := []*domain.BacklogItem{
		mustNewBacklogItem(t, "backlog-2", "project-1", "Second", "", "000000000002"),
		mustNewBacklogItem(t, "backlog-3", "project-1", "Third", "", "000000000003"),
	}
	projectRepository := newProjectRepositoryMock()
	projectRepository.GetProjectByNameFunc = func(_ context.Context, _ string) (*domain.Project, error) {
		return domain.NewProject("project-1", "worker", "https://github.com/neatflowcv/worker.git", nil), nil
	}
	backlogItemRepository := newBacklogItemRepositoryMock()
	backlogItemRepository.ListBacklogItemsFunc = func(
		_ context.Context,
		projectID, afterID string,
		limit int,
	) ([]*domain.BacklogItem, error) {
		return expectedItems, nil
	}
	service := flow.NewService(projectRepository, backlogItemRepository, newWorkspaceMock(), nil)

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
	projectRepository.GetProjectByNameFunc = func(_ context.Context, _ string) (*domain.Project, error) {
		return domain.NewProject("project-1", "worker", "https://github.com/neatflowcv/worker.git", nil), nil
	}
	backlogItemRepository := newBacklogItemRepositoryMock()
	service := flow.NewService(projectRepository, backlogItemRepository, newWorkspaceMock(), nil)

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
	projectRepository.GetProjectByNameFunc = func(_ context.Context, _ string) (*domain.Project, error) {
		return nil, repository.ErrProjectNotFound
	}
	backlogItemRepository := newBacklogItemRepositoryMock()
	service := flow.NewService(projectRepository, backlogItemRepository, newWorkspaceMock(), nil)

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
	projectRepository.GetProjectByNameFunc = func(_ context.Context, _ string) (*domain.Project, error) {
		return domain.NewProject("project-1", "worker", "https://github.com/neatflowcv/worker.git", nil), nil
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
	service := flow.NewService(projectRepository, backlogItemRepository, newWorkspaceMock(), nil)

	// Act
	items, err := service.ListBacklogItems(t.Context(), "worker", "", 20)

	// Assert
	require.ErrorIs(t, err, errBacklogItemRepository)
	require.Nil(t, items)
}

func TestService_UpdateBacklogItem(t *testing.T) {
	t.Parallel()

	const updatedDescription = "new"

	// Arrange
	projectRepository := newProjectRepositoryMock()
	projectRepository.GetProjectByNameFunc = func(_ context.Context, _ string) (*domain.Project, error) {
		return domain.NewProject("project-1", "worker", "https://github.com/neatflowcv/worker.git", nil), nil
	}
	existingItem := mustNewBacklogItem(t, "backlog-1", "project-1", "Before", "old", "000000000001")
	backlogItemRepository := newBacklogItemRepositoryMock()
	backlogItemRepository.GetBacklogItemFunc = func(_ context.Context, _ string) (*domain.BacklogItem, error) {
		return existingItem, nil
	}

	var updatedItem *domain.BacklogItem

	backlogItemRepository.UpdateBacklogItemFunc = func(_ context.Context, item *domain.BacklogItem) error {
		updatedItem = item

		return nil
	}
	service := flow.NewService(projectRepository, backlogItemRepository, newWorkspaceMock(), nil)
	title := "After"
	description := updatedDescription

	// Act
	item, err := service.UpdateBacklogItem(t.Context(), "worker", "backlog-1", &title, &description)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, item)
	require.NotSame(t, existingItem, item)
	require.Equal(t, "After", item.Title())
	require.Equal(t, "new", item.Description())
	require.Same(t, item, updatedItem)
	require.Len(t, backlogItemRepository.GetBacklogItemCalls(), 1)
	require.Len(t, backlogItemRepository.UpdateBacklogItemCalls(), 1)
}

func TestService_UpdateBacklogItemReturnsErrorWhenProjectDoesNotExist(t *testing.T) {
	t.Parallel()

	// Arrange
	const updatedDescription = "new"

	projectRepository := newProjectRepositoryMock()
	projectRepository.GetProjectByNameFunc = func(_ context.Context, _ string) (*domain.Project, error) {
		return nil, repository.ErrProjectNotFound
	}
	backlogItemRepository := newBacklogItemRepositoryMock()
	service := flow.NewService(projectRepository, backlogItemRepository, newWorkspaceMock(), nil)
	description := updatedDescription

	// Act
	item, err := service.UpdateBacklogItem(
		t.Context(),
		"worker",
		"backlog-1",
		nil,
		&description,
	)

	// Assert
	require.ErrorIs(t, err, repository.ErrProjectNotFound)
	require.Nil(t, item)
	require.Empty(t, backlogItemRepository.GetBacklogItemCalls())
	require.Empty(t, backlogItemRepository.UpdateBacklogItemCalls())
}

func TestService_UpdateBacklogItemReturnsErrorWhenBacklogItemBelongsToAnotherProject(t *testing.T) {
	t.Parallel()

	// Arrange
	const updatedDescription = "new"

	projectRepository := newProjectRepositoryMock()
	projectRepository.GetProjectByNameFunc = func(_ context.Context, _ string) (*domain.Project, error) {
		return domain.NewProject("project-1", "worker", "https://github.com/neatflowcv/worker.git", nil), nil
	}
	backlogItemRepository := newBacklogItemRepositoryMock()
	backlogItemRepository.GetBacklogItemFunc = func(_ context.Context, _ string) (*domain.BacklogItem, error) {
		return mustNewBacklogItem(t, "backlog-1", "project-2", "Before", "old", "000000000001"), nil
	}
	service := flow.NewService(projectRepository, backlogItemRepository, newWorkspaceMock(), nil)
	description := updatedDescription

	// Act
	item, err := service.UpdateBacklogItem(
		t.Context(),
		"worker",
		"backlog-1",
		nil,
		&description,
	)

	// Assert
	require.ErrorIs(t, err, repository.ErrBacklogItemNotFound)
	require.Nil(t, item)
	require.Empty(t, backlogItemRepository.UpdateBacklogItemCalls())
}

func TestService_UpdateBacklogItemReturnsErrorWhenTitleIsEmpty(t *testing.T) {
	t.Parallel()

	// Arrange
	projectRepository := newProjectRepositoryMock()
	projectRepository.GetProjectByNameFunc = func(_ context.Context, _ string) (*domain.Project, error) {
		return domain.NewProject("project-1", "worker", "https://github.com/neatflowcv/worker.git", nil), nil
	}
	backlogItemRepository := newBacklogItemRepositoryMock()
	backlogItemRepository.GetBacklogItemFunc = func(_ context.Context, _ string) (*domain.BacklogItem, error) {
		return mustNewBacklogItem(t, "backlog-1", "project-1", "Before", "old", "000000000001"), nil
	}
	service := flow.NewService(projectRepository, backlogItemRepository, newWorkspaceMock(), nil)
	title := ""

	// Act
	item, err := service.UpdateBacklogItem(t.Context(), "worker", "backlog-1", &title, nil)

	// Assert
	require.ErrorIs(t, err, domain.ErrBacklogItemTitleRequired)
	require.Nil(t, item)
	require.Empty(t, backlogItemRepository.UpdateBacklogItemCalls())
}

func TestService_UpdateBacklogItemReturnsErrorWhenRepositoryUpdateFails(t *testing.T) {
	t.Parallel()

	// Arrange
	const updatedDescription = "new"

	projectRepository := newProjectRepositoryMock()
	projectRepository.GetProjectByNameFunc = func(_ context.Context, _ string) (*domain.Project, error) {
		return domain.NewProject("project-1", "worker", "https://github.com/neatflowcv/worker.git", nil), nil
	}
	backlogItemRepository := newBacklogItemRepositoryMock()
	backlogItemRepository.GetBacklogItemFunc = func(_ context.Context, _ string) (*domain.BacklogItem, error) {
		return mustNewBacklogItem(t, "backlog-1", "project-1", "Before", "old", "000000000001"), nil
	}
	backlogItemRepository.UpdateBacklogItemFunc = func(_ context.Context, _ *domain.BacklogItem) error {
		return errBacklogItemRepository
	}
	service := flow.NewService(projectRepository, backlogItemRepository, newWorkspaceMock(), nil)
	description := updatedDescription

	// Act
	item, err := service.UpdateBacklogItem(
		t.Context(),
		"worker",
		"backlog-1",
		nil,
		&description,
	)

	// Assert
	require.ErrorIs(t, err, errBacklogItemRepository)
	require.Nil(t, item)
}

func TestService_RefineBacklogItem(t *testing.T) {
	t.Parallel()

	// Arrange
	project := domain.NewProject("project-1", "worker", "https://github.com/neatflowcv/worker.git", nil)
	backlogItem := mustNewBacklogItem(
		t,
		"backlog-1",
		"project-1",
		"First",
		"original",
		"000000000001",
	)
	projectRepository := newProjectRepositoryMock()
	projectRepository.GetProjectByNameFunc = func(_ context.Context, _ string) (*domain.Project, error) {
		return project, nil
	}
	backlogItemRepository := newBacklogItemRepositoryMock()
	backlogItemRepository.GetBacklogItemFunc = func(_ context.Context, _ string) (*domain.BacklogItem, error) {
		return backlogItem, nil
	}
	workspace := newWorkspaceMock()
	workspace.PrepareWorkspaceFunc = func(_ context.Context, gotProject *domain.Project) (*domain.Workspace, error) {
		require.Same(t, project, gotProject)

		return domain.NewWorkspace("/tmp/project-1", "/tmp/project-1/main", nil), nil
	}
	executor := newBacklogActionRunnerMock()
	executor.RefineBacklogItemFunc = func(
		_ context.Context,
		projectDir string,
		item *domain.BacklogItem,
	) (*domain.BacklogItem, error) {
		require.Equal(t, "/tmp/project-1/main", projectDir)
		require.Same(t, backlogItem, item)

		return item.SetDescription("refined"), nil
	}
	service := flow.NewService(projectRepository, backlogItemRepository, workspace, executor)

	// Act
	item, err := service.RefineBacklogItem(t.Context(), "worker", "backlog-1")

	// Assert
	require.NoError(t, err)
	require.NotNil(t, item)
	require.NotSame(t, backlogItem, item)
	require.Equal(t, backlogItem.ID(), item.ID())
	require.Equal(t, backlogItem.ProjectID(), item.ProjectID())
	require.Equal(t, backlogItem.Title(), item.Title())
	require.Equal(t, "refined", item.Description())
	require.Equal(t, backlogItem.Status(), item.Status())
	require.Equal(t, backlogItem.OrderKey(), item.OrderKey())
	require.Len(t, workspace.PrepareWorkspaceCalls(), 1)
	require.Len(t, executor.RefineBacklogItemCalls(), 1)
	require.Empty(t, backlogItemRepository.CreateBacklogItemCalls())
}

func TestService_RefineBacklogItemReturnsErrorWhenPrepareWorkspaceFails(t *testing.T) {
	t.Parallel()

	// Arrange
	project := domain.NewProject("project-1", "worker", "https://github.com/neatflowcv/worker.git", nil)
	projectRepository := newProjectRepositoryMock()
	projectRepository.GetProjectByNameFunc = func(_ context.Context, _ string) (*domain.Project, error) {
		return project, nil
	}
	backlogItemRepository := newBacklogItemRepositoryMock()
	backlogItemRepository.GetBacklogItemFunc = func(_ context.Context, _ string) (*domain.BacklogItem, error) {
		return mustNewBacklogItem(t, "backlog-1", "project-1", "First", "desc", "000000000001"), nil
	}
	workspace := newWorkspaceMock()
	workspace.PrepareWorkspaceFunc = func(_ context.Context, gotProject *domain.Project) (*domain.Workspace, error) {
		require.Same(t, project, gotProject)

		return nil, errWorkspace
	}
	executor := newBacklogActionRunnerMock()
	service := flow.NewService(projectRepository, backlogItemRepository, workspace, executor)

	// Act
	item, err := service.RefineBacklogItem(t.Context(), "worker", "backlog-1")

	// Assert
	require.ErrorIs(t, err, errWorkspace)
	require.Nil(t, item)
	require.Len(t, workspace.PrepareWorkspaceCalls(), 1)
	require.Empty(t, executor.RefineBacklogItemCalls())
}

func TestService_RefineBacklogItemReturnsErrorWhenExecutorFails(t *testing.T) {
	t.Parallel()

	// Arrange
	project := domain.NewProject("project-1", "worker", "https://github.com/neatflowcv/worker.git", nil)
	projectRepository := newProjectRepositoryMock()
	projectRepository.GetProjectByNameFunc = func(_ context.Context, _ string) (*domain.Project, error) {
		return project, nil
	}
	backlogItemRepository := newBacklogItemRepositoryMock()
	backlogItemRepository.GetBacklogItemFunc = func(_ context.Context, _ string) (*domain.BacklogItem, error) {
		return mustNewBacklogItem(t, "backlog-1", "project-1", "First", "desc", "000000000001"), nil
	}
	workspace := newWorkspaceMock()
	workspace.PrepareWorkspaceFunc = func(_ context.Context, gotProject *domain.Project) (*domain.Workspace, error) {
		require.Same(t, project, gotProject)

		return domain.NewWorkspace("/tmp/project-1", "/tmp/project-1/main", nil), nil
	}
	executor := newBacklogActionRunnerMock()
	executor.RefineBacklogItemFunc = func(
		_ context.Context,
		_ string,
		_ *domain.BacklogItem,
	) (*domain.BacklogItem, error) {
		return nil, errBacklogRefineExecutor
	}
	service := flow.NewService(projectRepository, backlogItemRepository, workspace, executor)

	// Act
	item, err := service.RefineBacklogItem(t.Context(), "worker", "backlog-1")

	// Assert
	require.ErrorIs(t, err, errBacklogRefineExecutor)
	require.Nil(t, item)
	require.Len(t, workspace.PrepareWorkspaceCalls(), 1)
}

func TestService_RefineBacklogItemReturnsErrorWhenBacklogItemBelongsToAnotherProject(t *testing.T) {
	t.Parallel()

	// Arrange
	projectRepository := newProjectRepositoryMock()
	projectRepository.GetProjectByNameFunc = func(_ context.Context, _ string) (*domain.Project, error) {
		return domain.NewProject("project-1", "worker", "https://github.com/neatflowcv/worker.git", nil), nil
	}
	backlogItemRepository := newBacklogItemRepositoryMock()
	backlogItemRepository.GetBacklogItemFunc = func(_ context.Context, _ string) (*domain.BacklogItem, error) {
		return mustNewBacklogItem(t, "backlog-1", "project-2", "First", "desc", "000000000001"), nil
	}
	workspace := newWorkspaceMock()
	executor := newBacklogActionRunnerMock()
	service := flow.NewService(projectRepository, backlogItemRepository, workspace, executor)

	// Act
	item, err := service.RefineBacklogItem(t.Context(), "worker", "backlog-1")

	// Assert
	require.ErrorIs(t, err, repository.ErrBacklogItemNotFound)
	require.Nil(t, item)
	require.Empty(t, executor.RefineBacklogItemCalls())
}

func TestService_RefineBacklogItemReturnsErrorWhenProjectDoesNotExist(t *testing.T) {
	t.Parallel()

	// Arrange
	projectRepository := newProjectRepositoryMock()
	projectRepository.GetProjectByNameFunc = func(_ context.Context, _ string) (*domain.Project, error) {
		return nil, repository.ErrProjectNotFound
	}
	backlogItemRepository := newBacklogItemRepositoryMock()
	workspace := newWorkspaceMock()
	executor := newBacklogActionRunnerMock()
	service := flow.NewService(projectRepository, backlogItemRepository, workspace, executor)

	// Act
	item, err := service.RefineBacklogItem(t.Context(), "worker", "backlog-1")

	// Assert
	require.ErrorIs(t, err, repository.ErrProjectNotFound)
	require.Nil(t, item)
	require.Empty(t, backlogItemRepository.GetBacklogItemCalls())
	require.Empty(t, executor.RefineBacklogItemCalls())
}

func TestService_RefineBacklogItemReturnsErrorWhenRepositoryFails(t *testing.T) {
	t.Parallel()

	// Arrange
	projectRepository := newProjectRepositoryMock()
	projectRepository.GetProjectByNameFunc = func(_ context.Context, _ string) (*domain.Project, error) {
		return domain.NewProject("project-1", "worker", "https://github.com/neatflowcv/worker.git", nil), nil
	}
	backlogItemRepository := newBacklogItemRepositoryMock()
	backlogItemRepository.GetBacklogItemFunc = func(_ context.Context, _ string) (*domain.BacklogItem, error) {
		return nil, errBacklogItemRepository
	}
	workspace := newWorkspaceMock()
	executor := newBacklogActionRunnerMock()
	service := flow.NewService(projectRepository, backlogItemRepository, workspace, executor)

	// Act
	item, err := service.RefineBacklogItem(t.Context(), "worker", "backlog-1")

	// Assert
	require.ErrorIs(t, err, errBacklogItemRepository)
	require.Nil(t, item)
	require.Empty(t, executor.RefineBacklogItemCalls())
}

func newProjectRepositoryMock() *ProjectRepositoryMock {
	var mock ProjectRepositoryMock

	mock.CreateProjectFunc = func(_ context.Context, _ *domain.Project) error {
		return nil
	}
	mock.GetProjectByNameFunc = func(_ context.Context, _ string) (*domain.Project, error) {
		return nil, repository.ErrProjectNotFound
	}
	mock.ListProjectsFunc = func(_ context.Context) ([]*domain.Project, error) {
		return nil, nil
	}

	return &mock
}

func newBacklogItemRepositoryMock() *BacklogItemRepositoryMock {
	var mock BacklogItemRepositoryMock

	mock.CreateBacklogItemFunc = func(_ context.Context, _ *domain.BacklogItem) error {
		return nil
	}
	mock.GetBacklogItemFunc = func(_ context.Context, _ string) (*domain.BacklogItem, error) {
		return nil, repository.ErrBacklogItemNotFound
	}
	mock.ListBacklogItemsFunc = func(
		_ context.Context,
		_ string,
		_ string,
		_ int,
	) ([]*domain.BacklogItem, error) {
		return []*domain.BacklogItem{}, nil
	}
	mock.UpdateBacklogItemFunc = func(_ context.Context, _ *domain.BacklogItem) error {
		return nil
	}

	return &mock
}

func newWorkspaceMock() *WorkspaceMock {
	var mock WorkspaceMock

	mock.PrepareWorkspaceFunc = func(_ context.Context, project *domain.Project) (*domain.Workspace, error) {
		return domain.NewWorkspace("/tmp/"+project.ID(), "/tmp/"+project.ID()+"/main", nil), nil
	}

	return &mock
}

var errBacklogItemRepository = errors.New("backlog item repository error")
var errBacklogRefineExecutor = errors.New("backlog refine executor error")
var errWorkspace = errors.New("workspace error")

func mustNewBacklogItem(
	t *testing.T,
	id string,
	projectID string,
	title string,
	description string,
	orderKey string,
) *domain.BacklogItem {
	t.Helper()

	item, err := domain.NewBacklogItem(
		id,
		projectID,
		title,
		description,
		domain.BacklogItemStatusOpen,
		orderKey,
	)
	require.NoError(t, err)

	return item
}

func newBacklogActionRunnerMock() *BacklogActionRunnerMock {
	var mock BacklogActionRunnerMock

	mock.RefineBacklogItemFunc = func(
		_ context.Context,
		_ string,
		item *domain.BacklogItem,
	) (*domain.BacklogItem, error) {
		return item, nil
	}

	return &mock
}
