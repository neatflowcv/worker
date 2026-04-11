package workspace

import (
	"context"

	"github.com/neatflowcv/worker/internal/pkg/domain"
)

//go:generate go run github.com/matryer/moq@v0.7.1 -pkg flow_test -skip-ensure -out ../../app/flow/workspace_moq_generated_test.go . Workspace
type Workspace interface {
	CreateWorkspace(ctx context.Context, project *domain.Project) error
}
