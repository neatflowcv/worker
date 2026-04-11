package repository

import (
	"context"
	"errors"

	"github.com/neatflowcv/worker/internal/pkg/domain"
)

var ErrBacklogItemNotFound = errors.New("backlog item not found")

//go:generate go run github.com/matryer/moq@v0.7.1 -pkg flow_test -skip-ensure -out ../../app/flow/backlog_item_repository_moq_generated_test.go . BacklogItemRepository
type BacklogItemRepository interface {
	CreateBacklogItem(ctx context.Context, item *domain.BacklogItem) error
	ListBacklogItems(
		ctx context.Context,
		projectID string,
		afterID string,
		limit int,
	) ([]*domain.BacklogItem, error)
}
