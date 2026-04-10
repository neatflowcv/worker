package memory

import (
	"context"

	"github.com/neatflowcv/worker/internal/pkg/domain"
	workspacepkg "github.com/neatflowcv/worker/internal/pkg/workspace"
)

var _ workspacepkg.Workspace = (*Workspace)(nil)

type Workspace struct{}

func NewWorkspace() *Workspace {
	return &Workspace{}
}

func (w *Workspace) CreateWorkspace(ctx context.Context, project *domain.Project) error {
	_ = ctx
	_ = project

	return nil
}
