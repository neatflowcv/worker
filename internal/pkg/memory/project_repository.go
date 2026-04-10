package memory

import (
	"context"
	"slices"
	"sync"

	"github.com/neatflowcv/worker/internal/pkg/domain"
	"github.com/neatflowcv/worker/internal/pkg/repository"
)

var _ repository.ProjectRepository = (*ProjectRepository)(nil)

type ProjectRepository struct {
	mu       sync.RWMutex
	projects map[string]*domain.Project
}

func NewProjectRepository() *ProjectRepository {
	return &ProjectRepository{
		mu:       sync.RWMutex{},
		projects: make(map[string]*domain.Project),
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

	for _, project := range projects {
		r.projects[project.Name()] = project
	}
}

func (r *ProjectRepository) GetByName(ctx context.Context, name string) (*domain.Project, error) {
	_ = ctx

	r.mu.RLock()
	defer r.mu.RUnlock()

	project, ok := r.projects[name]
	if ok {
		return project, nil
	}

	return nil, repository.ErrProjectNotFound
}

func (r *ProjectRepository) List(ctx context.Context) ([]*domain.Project, error) {
	_ = ctx

	r.mu.RLock()
	defer r.mu.RUnlock()

	projects := make([]*domain.Project, 0, len(r.projects))
	for _, project := range r.projects {
		projects = append(projects, project)
	}

	slices.SortFunc(projects, func(a, b *domain.Project) int {
		if a.ID() < b.ID() {
			return -1
		}

		if a.ID() > b.ID() {
			return 1
		}

		return 0
	})

	return projects, nil
}
