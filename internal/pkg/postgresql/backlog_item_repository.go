package postgresql

import (
	"context"
	"errors"
	"fmt"

	"github.com/neatflowcv/worker/internal/pkg/domain"
	"github.com/neatflowcv/worker/internal/pkg/orderkey"
	"github.com/neatflowcv/worker/internal/pkg/repository"
	"gorm.io/gorm"
)

var _ repository.BacklogItemRepository = (*BacklogItemRepository)(nil)

type BacklogItemRepository struct {
	db *gorm.DB
}

func NewBacklogItemRepository(database *Database) *BacklogItemRepository {
	return &BacklogItemRepository{
		db: database.db,
	}
}

func (r *BacklogItemRepository) CreateBacklogItem(
	ctx context.Context,
	item *domain.BacklogItem,
) error {
	model := newBacklogItemModel(item)

	if model.OrderKey == "" {
		nextOrderKey, err := r.nextTopBacklogItemOrderKey(ctx, item.ProjectID())
		if err != nil {
			return err
		}

		model.OrderKey = nextOrderKey
	}

	err := gorm.G[BacklogItemModel](r.db).Create(ctx, &model)
	if err != nil {
		return fmt.Errorf("create backlog item: %w", err)
	}

	return nil
}

func (r *BacklogItemRepository) GetBacklogItem(
	ctx context.Context,
	id string,
) (*domain.BacklogItem, error) {
	model, err := gorm.G[BacklogItemModel](r.db).
		Where("id = ?", id).
		First(ctx)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, repository.ErrBacklogItemNotFound
		}

		return nil, fmt.Errorf("load backlog item: %w", err)
	}

	item, err := model.toDomain()
	if err != nil {
		return nil, fmt.Errorf("map backlog item: %w", err)
	}

	return item, nil
}

func (r *BacklogItemRepository) ListBacklogItems(
	ctx context.Context,
	projectID string,
	afterID string,
	limit int,
) ([]*domain.BacklogItem, error) {
	query := gorm.G[BacklogItemModel](r.db).
		Where("project_id = ?", projectID).
		Order("order_key asc, id asc")

	if afterID != "" {
		afterItem, err := r.GetBacklogItem(ctx, afterID)
		if err != nil {
			if errors.Is(err, repository.ErrBacklogItemNotFound) {
				return nil, repository.ErrBacklogItemNotFound
			}

			return nil, fmt.Errorf("resolve backlog item cursor: %w", err)
		}

		if afterItem.ProjectID() != projectID {
			return nil, repository.ErrBacklogItemNotFound
		}

		query = query.Where(
			"(order_key > ?) OR (order_key = ? AND id > ?)",
			afterItem.OrderKey(),
			afterItem.OrderKey(),
			afterItem.ID(),
		)
	}

	if limit > 0 {
		query = query.Limit(limit)
	}

	models, err := query.Find(ctx)
	if err != nil {
		return nil, fmt.Errorf("list backlog items: %w", err)
	}

	items := make([]*domain.BacklogItem, 0, len(models))
	for _, model := range models {
		item, itemErr := model.toDomain()
		if itemErr != nil {
			return nil, fmt.Errorf("map backlog item: %w", itemErr)
		}

		items = append(items, item)
	}

	return items, nil
}

func (r *BacklogItemRepository) UpdateBacklogItem(
	ctx context.Context,
	item *domain.BacklogItem,
) error {
	model := newBacklogItemModel(item)

	rows, err := gorm.G[BacklogItemModel](r.db).
		Where("id = ?", item.ID()).
		Select("project_id", "title", "description", "status", "order_key").
		Updates(ctx, model)
	if err != nil {
		return fmt.Errorf("update backlog item: %w", err)
	}

	if rows == 0 {
		return repository.ErrBacklogItemNotFound
	}

	return nil
}

func (r *BacklogItemRepository) nextTopBacklogItemOrderKey(
	ctx context.Context,
	projectID string,
) (string, error) {
	model, err := gorm.G[BacklogItemModel](r.db).
		Where("project_id = ?", projectID).
		Order("order_key asc, id asc").
		First(ctx)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return orderkey.First(), nil
		}

		return "", fmt.Errorf("query first backlog item order key: %w", err)
	}

	return orderkey.Before(model.OrderKey), nil
}
