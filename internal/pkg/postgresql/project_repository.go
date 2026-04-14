package postgresql

import (
	"context"
	"errors"
	"fmt"

	"github.com/neatflowcv/worker/internal/pkg/domain"
	"github.com/neatflowcv/worker/internal/pkg/repository"
	"gorm.io/gorm"
)

var _ repository.ProjectRepository = (*ProjectRepository)(nil)

type ProjectRepository struct {
	db *gorm.DB
}

func NewProjectRepository(database *Database) *ProjectRepository {
	return &ProjectRepository{
		db: database.db,
	}
}

func (r *ProjectRepository) CreateProject(ctx context.Context, project *domain.Project) error {
	model := newProjectModel(project)

	err := gorm.G[ProjectModel](r.db).Create(ctx, &model)
	if err != nil {
		return fmt.Errorf("create project: %w", err)
	}

	return nil
}

func (r *ProjectRepository) GetProjectByName(
	ctx context.Context,
	name string,
) (*domain.Project, error) {
	model, err := gorm.G[ProjectModel](r.db).
		Where("name = ?", name).
		First(ctx)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, repository.ErrProjectNotFound
		}

		return nil, fmt.Errorf("query project by name: %w", err)
	}

	return model.toDomain(), nil
}

func (r *ProjectRepository) ListProjects(ctx context.Context) ([]*domain.Project, error) {
	models, err := gorm.G[ProjectModel](r.db).
		Order("created_at asc, id asc").
		Find(ctx)
	if err != nil {
		return nil, fmt.Errorf("list projects: %w", err)
	}

	projects := make([]*domain.Project, 0, len(models))
	for _, model := range models {
		projects = append(projects, model.toDomain())
	}

	return projects, nil
}
