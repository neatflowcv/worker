package badger

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	badgerdb "github.com/dgraph-io/badger/v4"
	"github.com/neatflowcv/worker/internal/pkg/domain"
	"github.com/neatflowcv/worker/internal/pkg/repository"
)

const projectKeyPrefix = "project/"
const databaseDirectoryPermission = 0o750

var _ repository.ProjectRepository = (*ProjectRepository)(nil)

type ProjectRepository struct {
	db        *badgerdb.DB
	closeOnce sync.Once
	closeErr  error
}

type projectRecord struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	RepositoryURL string `json:"repositoryUrl"`
}

func NewProjectRepository(dir string) (*ProjectRepository, error) {
	cleanDir := filepath.Clean(dir)

	err := os.MkdirAll(cleanDir, databaseDirectoryPermission)
	if err != nil {
		return nil, fmt.Errorf("create badger database directory: %w", err)
	}

	options := badgerdb.DefaultOptions(cleanDir)
	options.Logger = nil

	database, err := badgerdb.Open(options)
	if err != nil {
		return nil, fmt.Errorf("open badger database: %w", err)
	}

	return &ProjectRepository{
		db:        database,
		closeOnce: sync.Once{},
		closeErr:  nil,
	}, nil
}

func (r *ProjectRepository) Close() error {
	r.closeOnce.Do(func() {
		r.closeErr = r.db.Close()
		if r.closeErr != nil {
			r.closeErr = fmt.Errorf("close badger database: %w", r.closeErr)
		}
	})

	return r.closeErr
}

func (r *ProjectRepository) Create(ctx context.Context, project *domain.Project) error {
	record := projectRecord{
		ID:            project.ID(),
		Name:          project.Name(),
		RepositoryURL: project.RepositoryURL(),
	}

	value, err := json.Marshal(record)
	if err != nil {
		return fmt.Errorf("marshal project record: %w", err)
	}

	err = r.db.Update(func(txn *badgerdb.Txn) error {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		return txn.Set(projectKey(project.ID()), value)
	})
	if err != nil {
		return fmt.Errorf("persist project: %w", err)
	}

	return nil
}

func (r *ProjectRepository) List(ctx context.Context) ([]*domain.Project, error) {
	projects := make([]*domain.Project, 0)

	err := r.db.View(func(txn *badgerdb.Txn) error {
		iteratorOptions := badgerdb.DefaultIteratorOptions
		iteratorOptions.PrefetchValues = true

		iterator := txn.NewIterator(iteratorOptions)
		defer iterator.Close()

		prefix := []byte(projectKeyPrefix)

		for iterator.Seek(prefix); iterator.ValidForPrefix(prefix); iterator.Next() {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
			}

			item := iterator.Item()

			err := item.Value(func(value []byte) error {
				var record projectRecord

				err := json.Unmarshal(value, &record)
				if err != nil {
					return fmt.Errorf("unmarshal project record: %w", err)
				}

				projects = append(
					projects,
					domain.NewRepository(record.ID, record.Name, record.RepositoryURL),
				)

				return nil
			})
			if err != nil {
				return fmt.Errorf("read project record: %w", err)
			}
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("load projects: %w", err)
	}

	return projects, nil
}

func projectKey(id string) []byte {
	return []byte(projectKeyPrefix + id)
}
