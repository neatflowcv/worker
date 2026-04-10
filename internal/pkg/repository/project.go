package repository

import (
	"context"

	"github.com/neatflowcv/worker/internal/pkg/domain"
)

type ProjectRepository interface {
	Create(ctx context.Context, project *domain.Project) error
	List(ctx context.Context) ([]*domain.Project, error)
}
