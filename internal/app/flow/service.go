package flow

import (
	"context"
	"errors"
	"fmt"

	"github.com/neatflowcv/worker/internal/pkg/domain"
	"github.com/neatflowcv/worker/internal/pkg/repository"
	"github.com/neatflowcv/worker/internal/pkg/runner"
	workspacerpkg "github.com/neatflowcv/worker/internal/pkg/workspacer"
	"github.com/oklog/ulid/v2"
)

var ErrProjectAlreadyExists = errors.New("project already exists")

const defaultBacklogItemLimit = 20

type Service struct {
	projectRepository     repository.ProjectRepository
	backlogItemRepository repository.BacklogItemRepository
	workspacer            workspacerpkg.Workspacer
	backlogActionRunner   runner.BacklogActionRunner
}

func NewService(
	projectRepository repository.ProjectRepository,
	backlogItemRepository repository.BacklogItemRepository,
	workspacer workspacerpkg.Workspacer,
	backlogActionRunner runner.BacklogActionRunner,
) *Service {
	return &Service{
		projectRepository:     projectRepository,
		backlogItemRepository: backlogItemRepository,
		workspacer:            workspacer,
		backlogActionRunner:   backlogActionRunner,
	}
}

func (s *Service) CreateProject(
	ctx context.Context,
	name string,
	url string,
	auth *domain.Auth,
) (*domain.Project, error) {
	_, err := s.projectRepository.GetProjectByName(ctx, name)
	if err == nil {
		return nil, ErrProjectAlreadyExists
	}

	if !errors.Is(err, repository.ErrProjectNotFound) {
		return nil, fmt.Errorf("get project by name: %w", err)
	}

	project := domain.NewProject(ulid.Make().String(), name, url, auth)

	_, err = s.workspacer.PrepareWorkspace(ctx, project)
	if err != nil {
		return nil, fmt.Errorf("prepare workspace: %w", err)
	}

	err = s.projectRepository.CreateProject(ctx, project)
	if err != nil {
		return nil, fmt.Errorf("create project: %w", err)
	}

	return project, nil
}

func (s *Service) ListProjects(ctx context.Context) ([]*domain.Project, error) {
	projects, err := s.projectRepository.ListProjects(ctx)
	if err != nil {
		return nil, fmt.Errorf("list projects: %w", err)
	}

	return projects, nil
}

func (s *Service) CreateBacklogItem(
	ctx context.Context,
	projectName string,
	title string,
	description string,
) (*domain.BacklogItem, error) {
	project, err := s.projectRepository.GetProjectByName(ctx, projectName)
	if err != nil {
		return nil, fmt.Errorf("get project by name: %w", err)
	}

	item, err := domain.NewBacklogItem(
		ulid.Make().String(),
		project.ID(),
		title,
		description,
		domain.BacklogItemStatusOpen,
		"",
	)
	if err != nil {
		return nil, fmt.Errorf("new backlog item: %w", err)
	}

	err = s.backlogItemRepository.CreateBacklogItem(ctx, item)
	if err != nil {
		return nil, fmt.Errorf("create backlog item: %w", err)
	}

	return item, nil
}

func (s *Service) GetBacklogItem(
	ctx context.Context,
	projectName string,
	backlogItemID string,
) (*domain.BacklogItem, error) {
	project, err := s.projectRepository.GetProjectByName(ctx, projectName)
	if err != nil {
		return nil, fmt.Errorf("get project by name: %w", err)
	}

	item, err := s.backlogItemRepository.GetBacklogItem(ctx, backlogItemID)
	if err != nil {
		return nil, fmt.Errorf("get backlog item: %w", err)
	}

	if item.ProjectID() != project.ID() {
		return nil, repository.ErrBacklogItemNotFound
	}

	return item, nil
}

func (s *Service) ListBacklogItems(
	ctx context.Context,
	projectName string,
	afterID string,
	limit int,
) ([]*domain.BacklogItem, error) {
	project, err := s.projectRepository.GetProjectByName(ctx, projectName)
	if err != nil {
		return nil, fmt.Errorf("get project by name: %w", err)
	}

	if limit <= 0 {
		limit = defaultBacklogItemLimit
	}

	items, err := s.backlogItemRepository.ListBacklogItems(
		ctx,
		project.ID(),
		afterID,
		limit,
	)
	if err != nil {
		return nil, fmt.Errorf("list backlog items: %w", err)
	}

	return items, nil
}

func (s *Service) UpdateBacklogItem(
	ctx context.Context,
	projectName string,
	backlogItemID string,
	title *string,
	description *string,
) (*domain.BacklogItem, error) {
	project, err := s.projectRepository.GetProjectByName(ctx, projectName)
	if err != nil {
		return nil, fmt.Errorf("get project by name: %w", err)
	}

	item, err := s.backlogItemRepository.GetBacklogItem(ctx, backlogItemID)
	if err != nil {
		return nil, fmt.Errorf("get backlog item: %w", err)
	}

	if item.ProjectID() != project.ID() {
		return nil, repository.ErrBacklogItemNotFound
	}

	updatedItem := item
	if title != nil {
		updatedItem, err = updatedItem.SetTitle(*title)
		if err != nil {
			return nil, fmt.Errorf("set backlog item title: %w", err)
		}
	}

	if description != nil {
		updatedItem = updatedItem.SetDescription(*description)
	}

	err = s.backlogItemRepository.UpdateBacklogItem(ctx, updatedItem)
	if err != nil {
		return nil, fmt.Errorf("update backlog item repository: %w", err)
	}

	return updatedItem, nil
}

//nolint:cyclop,funlen // Backlog start flow is kept inline by request.
func (s *Service) StartBacklogItem(
	ctx context.Context,
	projectName string,
	backlogItemID string,
) (*domain.BacklogItem, error) {
	project, err := s.projectRepository.GetProjectByName(ctx, projectName)
	if err != nil {
		return nil, fmt.Errorf("get project by name: %w", err)
	}

	item, err := s.backlogItemRepository.GetBacklogItem(ctx, backlogItemID)
	if err != nil {
		return nil, fmt.Errorf("get backlog item: %w", err)
	}

	if item.ProjectID() != project.ID() {
		return nil, repository.ErrBacklogItemNotFound
	}

	startedItem, err := item.Start()
	if err != nil {
		return nil, fmt.Errorf("start backlog item: %w", err)
	}

	workspace, err := s.workspacer.PrepareWorkspace(ctx, project)
	if err != nil {
		return nil, fmt.Errorf("prepare workspace: %w", err)
	}

	err = s.backlogItemRepository.UpdateBacklogItem(ctx, startedItem)
	if err != nil {
		return nil, fmt.Errorf("update backlog item repository: %w", err)
	}

	tree, err := s.backlogActionRunner.RecommendWorktree(
		ctx,
		workspace.Main(),
		startedItem,
	)
	if err != nil {
		return nil, fmt.Errorf("recommend worktree: %w", err)
	}

	err = s.workspacer.CreateWorktree(ctx, project, workspace, tree)
	if err != nil {
		return nil, fmt.Errorf("create worktree: %w", err)
	}

	err = s.backlogActionRunner.StartBacklogItem(
		ctx,
		workspace.Root()+"/"+tree.Dir(),
		startedItem,
	)
	if err != nil {
		return nil, fmt.Errorf("start backlog item: %w", err)
	}

	err = s.workspacer.CloseWorktree(ctx, project, tree)
	if err != nil {
		return nil, fmt.Errorf("close worktree: %w", err)
	}

	blockedItem, err := startedItem.Blocked()
	if err != nil {
		return nil, fmt.Errorf("block backlog item: %w", err)
	}

	err = s.backlogItemRepository.UpdateBacklogItem(ctx, blockedItem)
	if err != nil {
		return nil, fmt.Errorf("update backlog item repository: %w", err)
	}

	return blockedItem, nil
}

func (s *Service) RefineBacklogItem(
	ctx context.Context,
	projectName string,
	backlogItemID string,
) (*domain.BacklogItem, error) {
	project, err := s.projectRepository.GetProjectByName(ctx, projectName)
	if err != nil {
		return nil, fmt.Errorf("get project by name: %w", err)
	}

	item, err := s.backlogItemRepository.GetBacklogItem(ctx, backlogItemID)
	if err != nil {
		return nil, fmt.Errorf("get backlog item: %w", err)
	}

	if item.ProjectID() != project.ID() {
		return nil, repository.ErrBacklogItemNotFound
	}

	workspace, err := s.workspacer.PrepareWorkspace(ctx, project)
	if err != nil {
		return nil, fmt.Errorf("prepare workspace: %w", err)
	}

	refinedItem, err := s.backlogActionRunner.RefineBacklogItem(ctx, workspace.Main(), item)
	if err != nil {
		return nil, fmt.Errorf("refine backlog item: %w", err)
	}

	return refinedItem, nil
}
