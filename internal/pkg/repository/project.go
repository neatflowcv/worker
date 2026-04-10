package repository

import (
	"context"
	"errors"

	"github.com/neatflowcv/worker/internal/pkg/domain"
)

var ErrProjectNotFound = errors.New("project not found")

type ProjectRepository interface {
	Create(ctx context.Context, project *domain.Project) error
	GetByName(ctx context.Context, name string) (*domain.Project, error)
	List(ctx context.Context) ([]*domain.Project, error)
}
