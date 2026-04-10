package flow

import (
	"context"
	"fmt"

	"github.com/neatflowcv/worker/internal/pkg/domain"
	"github.com/neatflowcv/worker/internal/pkg/repository"
	"github.com/oklog/ulid/v2"
)

type Service struct {
	projectRepository repository.ProjectRepository
}

func NewService(projectRepository repository.ProjectRepository) *Service {
	return &Service{
		projectRepository: projectRepository,
	}
}

func (s *Service) CreateProject(ctx context.Context, name, url string) (*domain.Project, error) {
	project := domain.NewRepository(ulid.Make().String(), name, url)

	err := s.projectRepository.Create(ctx, project)
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
