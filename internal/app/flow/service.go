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

type Service struct {
	projectRepository repository.ProjectRepository
	workspace         workspacepkg.Workspace
}

func NewService(
	projectRepository repository.ProjectRepository,
	workspace workspacepkg.Workspace,
) *Service {
	return &Service{
		projectRepository: projectRepository,
		workspace:         workspace,
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
