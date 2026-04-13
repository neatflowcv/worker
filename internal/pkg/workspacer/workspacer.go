package workspacer

import (
	"context"

	"github.com/neatflowcv/worker/internal/pkg/domain"
)

//go:generate go run github.com/matryer/moq@v0.7.1 -pkg flow_test -skip-ensure -out ../../app/flow/workspacer_moq_generated_test.go . Workspacer
type Workspacer interface {
	PrepareWorkspace(ctx context.Context, project *domain.Project) (*domain.Workspace, error)
	CreateWorktree(
		ctx context.Context,
		project *domain.Project,
		workspace *domain.Workspace,
		worktree *domain.Worktree,
	) error
	CloseWorktree(ctx context.Context, project *domain.Project, worktree *domain.Worktree) error
}
