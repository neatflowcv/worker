package flow

import (
	"context"
	"errors"
	"fmt"

	"github.com/neatflowcv/worker/internal/pkg/domain"
	"github.com/neatflowcv/worker/internal/pkg/repository"
	"github.com/neatflowcv/worker/internal/pkg/runner"
	workspacepkg "github.com/neatflowcv/worker/internal/pkg/workspace"
	"github.com/oklog/ulid/v2"
)

var ErrProjectAlreadyExists = errors.New("project already exists")

const defaultBacklogItemLimit = 20

type Service struct {
	projectRepository     repository.ProjectRepository
	backlogItemRepository repository.BacklogItemRepository
	workspace             workspacepkg.Workspace
	backlogActionRunner   runner.BacklogActionRunner
}

func NewService(
	projectRepository repository.ProjectRepository,
	backlogItemRepository repository.BacklogItemRepository,
	workspace workspacepkg.Workspace,
	backlogActionRunner runner.BacklogActionRunner,
) *Service {
	return &Service{
		projectRepository:     projectRepository,
		backlogItemRepository: backlogItemRepository,
		workspace:             workspace,
		backlogActionRunner:   backlogActionRunner,
	}
}

func (s *Service) CreateProject(ctx context.Context, name, url string) (*domain.Project, error) {
	_, err := s.projectRepository.GetProjectByName(ctx, name)
	if err == nil {
		return nil, ErrProjectAlreadyExists
	}

	if !errors.Is(err, repository.ErrProjectNotFound) {
		return nil, fmt.Errorf("get project by name: %w", err)
	}

	project := domain.NewProject(ulid.Make().String(), name, url)

	err = s.workspace.PrepareWorkspace(ctx, project)
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

	err = s.workspace.PrepareWorkspace(ctx, project)
	if err != nil {
		return nil, fmt.Errorf("prepare workspace: %w", err)
	}

	projectDir, err := s.workspace.ProjectDir(ctx, project)
	if err != nil {
		return nil, fmt.Errorf("get project directory: %w", err)
	}

	refinedItem, err := s.backlogActionRunner.RefineBacklogItem(ctx, projectDir, item)
	if err != nil {
		return nil, fmt.Errorf("refine backlog item: %w", err)
	}

	return refinedItem, nil
}
