package memory

import (
	"context"
	"sync"

	"github.com/neatflowcv/worker/internal/pkg/domain"
	"github.com/neatflowcv/worker/internal/pkg/repository"
)

var _ repository.ProjectRepository = (*ProjectRepository)(nil)

type ProjectRepository struct {
	mu       sync.RWMutex
	projects []*domain.Project
}

func NewProjectRepository() *ProjectRepository {
	return &ProjectRepository{
		mu:       sync.RWMutex{},
		projects: []*domain.Project{},
	}
}

func (r *ProjectRepository) Create(ctx context.Context, project *domain.Project) error {
	_ = ctx

	r.Append(project)

	return nil
}

func (r *ProjectRepository) Append(projects ...*domain.Project) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.projects = append(r.projects, projects...)
}

func (r *ProjectRepository) List(ctx context.Context) ([]*domain.Project, error) {
	_ = ctx

	r.mu.RLock()
	defer r.mu.RUnlock()

	projects := make([]*domain.Project, len(r.projects))
	copy(projects, r.projects)

	return projects, nil
}
