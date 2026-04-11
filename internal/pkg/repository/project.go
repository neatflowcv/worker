package repository

import (
	"context"
	"errors"

	"github.com/neatflowcv/worker/internal/pkg/domain"
)

var ErrProjectNotFound = errors.New("project not found")

//go:generate go run github.com/matryer/moq@v0.7.1 -pkg flow_test -skip-ensure -out ../../app/flow/project_repository_moq_generated_test.go . ProjectRepository
type ProjectRepository interface {
	CreateProject(ctx context.Context, project *domain.Project) error
	GetProjectByName(ctx context.Context, name string) (*domain.Project, error)
	ListProjects(ctx context.Context) ([]*domain.Project, error)
}
