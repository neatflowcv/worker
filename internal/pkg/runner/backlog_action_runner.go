package runner

import (
	"context"

	"github.com/neatflowcv/worker/internal/pkg/domain"
)

//go:generate go run github.com/matryer/moq@v0.7.1 -pkg flow_test -skip-ensure -out ../../app/flow/backlog_action_runner_moq_generated_test.go . BacklogActionRunner
type BacklogActionRunner interface {
	RefineBacklogItem(
		ctx context.Context,
		projectDir string,
		item *domain.BacklogItem,
	) (*domain.BacklogItem, error)
}
