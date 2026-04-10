package workspace

import (
	"context"

	"github.com/neatflowcv/worker/internal/pkg/domain"
)

type Workspace interface {
	CreateWorkspace(ctx context.Context, project *domain.Project) error
}
