package badger

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	badgerdb "github.com/dgraph-io/badger/v4"
	"github.com/neatflowcv/worker/internal/pkg/domain"
	"github.com/neatflowcv/worker/internal/pkg/orderkey"
	"github.com/neatflowcv/worker/internal/pkg/repository"
)

const backlogItemKeyPrefix = "backlog_item/"
const backlogItemProjectKeyPrefix = "backlog_item_project/"

var _ repository.BacklogItemRepository = (*BacklogItemRepository)(nil)

type BacklogItemRepository struct {
	db *badgerdb.DB
}

func NewBacklogItemRepository(database *Database) *BacklogItemRepository {
	return &BacklogItemRepository{
		db: database.db,
	}
}

type backlogItemRecord struct {
	ID          string                   `json:"id"`
	ProjectID   string                   `json:"projectId"`
	Title       string                   `json:"title"`
	Description string                   `json:"description"`
	Status      domain.BacklogItemStatus `json:"status"`
	OrderKey    string                   `json:"orderKey"`
}

func newBacklogItemRecord(item *domain.BacklogItem) backlogItemRecord {
	return backlogItemRecord{
		ID:          item.ID(),
		ProjectID:   item.ProjectID(),
		Title:       item.Title(),
		Description: item.Description(),
		Status:      item.Status(),
		OrderKey:    item.OrderKey(),
	}
}

func (r *BacklogItemRepository) CreateBacklogItem(
	ctx context.Context,
	item *domain.BacklogItem,
) error {
	record := newBacklogItemRecord(item)

	err := r.db.Update(func(txn *badgerdb.Txn) error {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		if record.OrderKey == "" {
			nextOrderKey, err := nextTopBacklogItemOrderKey(txn, item.ProjectID())
			if err != nil {
				return err
			}

			record.OrderKey = nextOrderKey
		}

		value, err := json.Marshal(record)
		if err != nil {
			return fmt.Errorf("marshal backlog item record: %w", err)
		}

		err = txn.Set(backlogItemKey(record.ID), value)
		if err != nil {
			return fmt.Errorf("persist backlog item record: %w", err)
		}

		err = txn.Set(backlogItemProjectKey(record.ProjectID, record.OrderKey, record.ID), []byte(record.ID))
		if err != nil {
			return fmt.Errorf("persist backlog item project index: %w", err)
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("persist backlog item: %w", err)
	}

	return nil
}

func (r *BacklogItemRepository) UpdateBacklogItem(
	ctx context.Context,
	item *domain.BacklogItem,
) error {
	record := newBacklogItemRecord(item)

	err := r.db.Update(func(txn *badgerdb.Txn) error {
		err := checkContext(ctx)
		if err != nil {
			return err
		}

		_, err = readBacklogItem(txn, item.ID())
		if err != nil {
			return err
		}

		value, err := json.Marshal(record)
		if err != nil {
			return fmt.Errorf("marshal backlog item record: %w", err)
		}

		err = txn.Set(backlogItemKey(record.ID), value)
		if err != nil {
			return fmt.Errorf("persist backlog item record: %w", err)
		}

		return nil
	})
	if err != nil {
		if errors.Is(err, repository.ErrBacklogItemNotFound) {
			return repository.ErrBacklogItemNotFound
		}

		return fmt.Errorf("persist updated backlog item: %w", err)
	}

	return nil
}

func (r *BacklogItemRepository) GetBacklogItem(
	ctx context.Context,
	id string,
) (*domain.BacklogItem, error) {
	var backlogItem *domain.BacklogItem

	err := r.db.View(func(txn *badgerdb.Txn) error {
		err := checkContext(ctx)
		if err != nil {
			return err
		}

		backlogItem, err = readBacklogItem(txn, id)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		if errors.Is(err, repository.ErrBacklogItemNotFound) {
			return nil, repository.ErrBacklogItemNotFound
		}

		return nil, fmt.Errorf("load backlog item: %w", err)
	}

	return backlogItem, nil
}

func (r *BacklogItemRepository) ListBacklogItems(
	ctx context.Context,
	projectID string,
	afterID string,
	limit int,
) ([]*domain.BacklogItem, error) {
	items := make([]*domain.BacklogItem, 0)

	err := r.db.View(func(txn *badgerdb.Txn) error {
		return r.listBacklogItems(ctx, txn, projectID, afterID, limit, &items)
	})
	if err != nil {
		if errors.Is(err, repository.ErrBacklogItemNotFound) {
			return nil, repository.ErrBacklogItemNotFound
		}

		return nil, fmt.Errorf("load backlog items: %w", err)
	}

	return items, nil
}

func (r *BacklogItemRepository) listBacklogItems(
	ctx context.Context,
	txn *badgerdb.Txn,
	projectID string,
	afterID string,
	limit int,
	items *[]*domain.BacklogItem,
) error {
	err := checkContext(ctx)
	if err != nil {
		return err
	}

	iteratorOptions := badgerdb.DefaultIteratorOptions
	iteratorOptions.PrefetchValues = false

	iterator := txn.NewIterator(iteratorOptions)
	defer iterator.Close()

	prefix := backlogItemProjectPrefix(projectID)

	startKey, shouldSkipStartKey, err := resolveBacklogItemStartKey(txn, projectID, afterID)
	if err != nil {
		return err
	}

	for iterator.Seek(startKey); iterator.ValidForPrefix(prefix); iterator.Next() {
		err = checkContext(ctx)
		if err != nil {
			return err
		}

		item := iterator.Item()
		if shouldSkipStartKey && string(item.Key()) == string(startKey) {
			continue
		}

		backlogItem, err := readIndexedBacklogItem(txn, item)
		if err != nil {
			return err
		}

		*items = append(*items, backlogItem)
		if limit > 0 && len(*items) >= limit {
			break
		}
	}

	return nil
}

func backlogItemKey(id string) []byte {
	return []byte(backlogItemKeyPrefix + id)
}

func backlogItemProjectPrefix(projectID string) []byte {
	return []byte(backlogItemProjectKeyPrefix + projectID + "/")
}

func backlogItemProjectKey(projectID, orderKeyValue, id string) []byte {
	return []byte(backlogItemProjectKeyPrefix + projectID + "/" + orderKeyValue + "/" + id)
}

func nextTopBacklogItemOrderKey(txn *badgerdb.Txn, projectID string) (string, error) {
	iteratorOptions := badgerdb.DefaultIteratorOptions
	iteratorOptions.PrefetchValues = false

	iterator := txn.NewIterator(iteratorOptions)
	defer iterator.Close()

	prefix := backlogItemProjectPrefix(projectID)
	iterator.Seek(prefix)

	if !iterator.ValidForPrefix(prefix) {
		return orderkey.First(), nil
	}

	item := iterator.Item()

	backlogItem, err := readIndexedBacklogItem(txn, item)
	if err != nil {
		return "", err
	}

	return orderkey.Before(backlogItem.OrderKey()), nil
}

func readBacklogItem(txn *badgerdb.Txn, id string) (*domain.BacklogItem, error) {
	item, err := txn.Get(backlogItemKey(id))
	if err != nil {
		if errors.Is(err, badgerdb.ErrKeyNotFound) {
			return nil, repository.ErrBacklogItemNotFound
		}

		return nil, fmt.Errorf("read backlog item record: %w", err)
	}

	var backlogItem *domain.BacklogItem

	err = item.Value(func(value []byte) error {
		var record backlogItemRecord

		err := json.Unmarshal(value, &record)
		if err != nil {
			return fmt.Errorf("unmarshal backlog item record: %w", err)
		}

		backlogItem, err = domain.NewBacklogItem(
			record.ID,
			record.ProjectID,
			record.Title,
			record.Description,
			record.OrderKey,
		)
		if err != nil {
			return fmt.Errorf("new backlog item: %w", err)
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("read backlog item record value: %w", err)
	}

	return backlogItem, nil
}

func readIndexedBacklogItem(txn *badgerdb.Txn, item *badgerdb.Item) (*domain.BacklogItem, error) {
	var backlogItemID []byte

	err := item.Value(func(value []byte) error {
		backlogItemID = append(backlogItemID, value...)

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("read backlog item project index value: %w", err)
	}

	return readBacklogItem(txn, string(backlogItemID))
}

func resolveBacklogItemStartKey(
	txn *badgerdb.Txn,
	projectID string,
	afterID string,
) ([]byte, bool, error) {
	prefix := backlogItemProjectPrefix(projectID)
	if afterID == "" {
		return prefix, false, nil
	}

	afterItem, err := readBacklogItem(txn, afterID)
	if err != nil {
		return nil, false, err
	}

	if afterItem.ProjectID() != projectID {
		return nil, false, repository.ErrBacklogItemNotFound
	}

	return backlogItemProjectKey(projectID, afterItem.OrderKey(), afterItem.ID()), true, nil
}

func checkContext(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return fmt.Errorf("context done: %w", ctx.Err())
	default:
		return nil
	}
}
