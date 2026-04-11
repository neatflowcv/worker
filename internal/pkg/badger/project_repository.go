package badger

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	badgerdb "github.com/dgraph-io/badger/v4"
	"github.com/neatflowcv/worker/internal/pkg/domain"
	"github.com/neatflowcv/worker/internal/pkg/repository"
)

const projectKeyPrefix = "project/"
const projectNameKeyPrefix = "project_name/"

var _ repository.ProjectRepository = (*ProjectRepository)(nil)

type ProjectRepository struct {
	db *badgerdb.DB
}

type projectRecord struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	RepositoryURL string `json:"repositoryUrl"`
}

func NewProjectRepository(database *Database) *ProjectRepository {
	return &ProjectRepository{
		db: database.db,
	}
}

func (r *ProjectRepository) CreateProject(ctx context.Context, project *domain.Project) error {
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

		err := txn.Set(projectKey(project.ID()), value)
		if err != nil {
			return fmt.Errorf("persist project record: %w", err)
		}

		err = txn.Set(projectNameKey(project.Name()), []byte(project.ID()))
		if err != nil {
			return fmt.Errorf("persist project name index: %w", err)
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("persist project: %w", err)
	}

	return nil
}

func (r *ProjectRepository) ListProjects(ctx context.Context) ([]*domain.Project, error) {
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
					domain.NewProject(record.ID, record.Name, record.RepositoryURL),
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

func (r *ProjectRepository) GetProjectByName(ctx context.Context, name string) (*domain.Project, error) {
	project, err := r.loadProjectByName(ctx, name)
	if err != nil {
		if errors.Is(err, repository.ErrProjectNotFound) {
			return nil, repository.ErrProjectNotFound
		}

		return nil, err
	}

	return project, nil
}

func projectKey(id string) []byte {
	return []byte(projectKeyPrefix + id)
}

func projectNameKey(name string) []byte {
	return []byte(projectNameKeyPrefix + name)
}

func (r *ProjectRepository) loadProjectByName(
	ctx context.Context,
	name string,
) (*domain.Project, error) {
	var project *domain.Project

	err := r.db.View(func(txn *badgerdb.Txn) error {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		projectID, err := readProjectIDByName(txn, name)
		if err != nil {
			return err
		}

		project, err = readProject(txn, projectID)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("load project by name: %w", err)
	}

	return project, nil
}

func readProjectIDByName(txn *badgerdb.Txn, name string) (string, error) {
	item, err := txn.Get(projectNameKey(name))
	if err != nil {
		if errors.Is(err, badgerdb.ErrKeyNotFound) {
			return "", repository.ErrProjectNotFound
		}

		return "", fmt.Errorf("read project name index: %w", err)
	}

	var projectID []byte

	err = item.Value(func(value []byte) error {
		projectID = append(projectID, value...)

		return nil
	})
	if err != nil {
		return "", fmt.Errorf("read project name index value: %w", err)
	}

	return string(projectID), nil
}

func readProject(txn *badgerdb.Txn, id string) (*domain.Project, error) {
	item, err := txn.Get(projectKey(id))
	if err != nil {
		if errors.Is(err, badgerdb.ErrKeyNotFound) {
			return nil, repository.ErrProjectNotFound
		}

		return nil, fmt.Errorf("read project record: %w", err)
	}

	var project *domain.Project

	err = item.Value(func(value []byte) error {
		var record projectRecord

		err := json.Unmarshal(value, &record)
		if err != nil {
			return fmt.Errorf("unmarshal project record: %w", err)
		}

		project = domain.NewProject(record.ID, record.Name, record.RepositoryURL)

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("read project record value: %w", err)
	}

	return project, nil
}
