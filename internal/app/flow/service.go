package flow

import (
	"context"

	"github.com/neatflowcv/worker/internal/pkg/domain"
	"github.com/oklog/ulid/v2"
)

type Service struct{}

func NewService() *Service {
	return &Service{}
}

func (s *Service) CreateProject(ctx context.Context, name, url string) (*domain.Project, error) {
	_ = ctx

	return domain.NewRepository(ulid.Make().String(), name, url), nil
}

func (s *Service) ListProjects(ctx context.Context) ([]*domain.Project, error) {
	_ = ctx

	return []*domain.Project{}, nil
}
