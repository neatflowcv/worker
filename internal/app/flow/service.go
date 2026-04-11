package flow

import (
	"context"
	"errors"
	"fmt"

	"github.com/neatflowcv/worker/internal/pkg/domain"
	"github.com/neatflowcv/worker/internal/pkg/repository"
	workspacepkg "github.com/neatflowcv/worker/internal/pkg/workspace"
	"github.com/oklog/ulid/v2"
)

var ErrProjectAlreadyExists = errors.New("project already exists")

const defaultBacklogItemLimit = 20

type Service struct {
	projectRepository     repository.ProjectRepository
	backlogItemRepository repository.BacklogItemRepository
	workspace             workspacepkg.Workspace
}

func NewService(
	projectRepository repository.ProjectRepository,
	backlogItemRepository repository.BacklogItemRepository,
	workspace workspacepkg.Workspace,
) *Service {
	return &Service{
		projectRepository:     projectRepository,
		backlogItemRepository: backlogItemRepository,
		workspace:             workspace,
	}
}

func (s *Service) CreateProject(ctx context.Context, name, url string) (*domain.Project, error) {
	_, err := s.projectRepository.GetByName(ctx, name)
	if err == nil {
		return nil, ErrProjectAlreadyExists
	}

	if !errors.Is(err, repository.ErrProjectNotFound) {
		return nil, fmt.Errorf("get project by name: %w", err)
	}

	project := domain.NewProject(ulid.Make().String(), name, url)

	err = s.workspace.CreateWorkspace(ctx, project)
	if err != nil {
		return nil, fmt.Errorf("create workspace: %w", err)
	}

	err = s.projectRepository.Create(ctx, project)
	if err != nil {
		return nil, fmt.Errorf("create project: %w", err)
	}

	return project, nil
}

func (s *Service) ListProjects(ctx context.Context) ([]*domain.Project, error) {
	projects, err := s.projectRepository.List(ctx)
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
	project, err := s.projectRepository.GetByName(ctx, projectName)
	if err != nil {
		return nil, fmt.Errorf("get project by name: %w", err)
	}

	item := domain.NewBacklogItem(
		ulid.Make().String(),
		project.ID(),
		title,
		description,
		"",
	)

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
	project, err := s.projectRepository.GetByName(ctx, projectName)
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
	project, err := s.projectRepository.GetByName(ctx, projectName)
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
