package postgresql

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/neatflowcv/worker/internal/pkg/domain"
)

type ProjectModel struct {
	ID            string         `gorm:"primaryKey;type=text"`
	Name          string         `gorm:"type=text;uniqueIndex;not null"`
	RepositoryURL string         `gorm:"type=text;not null"`
	AuthUsername  sql.NullString `gorm:"type=text"`
	AuthPassword  sql.NullString `gorm:"type=text"`
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

func (ProjectModel) TableName() string {
	return "projects"
}

func newProjectModel(project *domain.Project) ProjectModel {
	authUsername, authPassword := toNullAuth(project.Auth())

	return ProjectModel{
		ID:            project.ID(),
		Name:          project.Name(),
		RepositoryURL: project.RepositoryURL(),
		AuthUsername:  authUsername,
		AuthPassword:  authPassword,
		CreatedAt:     time.Time{},
		UpdatedAt:     time.Time{},
	}
}

func (m ProjectModel) toDomain() *domain.Project {
	return domain.NewProject(
		m.ID,
		m.Name,
		m.RepositoryURL,
		toDomainAuth(m.AuthUsername, m.AuthPassword),
	)
}

type BacklogItemModel struct {
	ID          string `gorm:"primaryKey;type=text"`
	ProjectID   string `gorm:"type=text;not null;index:idx_backlog_items_project_order,priority:1"`
	Title       string `gorm:"type=text;not null"`
	Description string `gorm:"type=text;not null"`
	Status      string `gorm:"type=text;not null"`
	OrderKey    string `gorm:"type=text;not null;index:idx_backlog_items_project_order,priority:2"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func (BacklogItemModel) TableName() string {
	return "backlog_items"
}

func newBacklogItemModel(item *domain.BacklogItem) BacklogItemModel {
	return BacklogItemModel{
		ID:          item.ID(),
		ProjectID:   item.ProjectID(),
		Title:       item.Title(),
		Description: item.Description(),
		Status:      string(item.Status()),
		OrderKey:    item.OrderKey(),
		CreatedAt:   time.Time{},
		UpdatedAt:   time.Time{},
	}
}

func (m BacklogItemModel) toDomain() (*domain.BacklogItem, error) {
	item, err := domain.NewBacklogItem(
		m.ID,
		m.ProjectID,
		m.Title,
		m.Description,
		domain.BacklogItemStatus(m.Status),
		m.OrderKey,
	)
	if err != nil {
		return nil, fmt.Errorf("new backlog item: %w", err)
	}

	return item, nil
}

func toNullAuth(auth *domain.Auth) (sql.NullString, sql.NullString) {
	if auth == nil {
		return sql.NullString{
				String: "",
				Valid:  false,
			}, sql.NullString{
				String: "",
				Valid:  false,
			}
	}

	return toNullString(auth.Username()), toNullString(auth.Password())
}

func toDomainAuth(username, password sql.NullString) *domain.Auth {
	if !username.Valid && !password.Valid {
		return nil
	}

	return domain.NewAuth(username.String, password.String)
}

func toNullString(value string) sql.NullString {
	if value == "" {
		return sql.NullString{
			String: "",
			Valid:  false,
		}
	}

	return sql.NullString{
		String: value,
		Valid:  true,
	}
}
