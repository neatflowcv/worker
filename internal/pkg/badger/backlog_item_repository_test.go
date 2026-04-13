package badger_test

import (
	"testing"

	"github.com/neatflowcv/worker/internal/pkg/badger"
	"github.com/neatflowcv/worker/internal/pkg/domain"
	"github.com/neatflowcv/worker/internal/pkg/repository"
	"github.com/stretchr/testify/require"
)

func TestProjectRepository_CreateBacklogItemPersistsBacklogItem(t *testing.T) {
	t.Parallel()

	// Arrange
	dir := t.TempDir()
	repo, database := newBacklogItemRepositoryAt(t, dir)

	// Act
	err := repo.CreateBacklogItem(
		t.Context(),
		mustNewBacklogItem(t, "backlog-1", "project-1", "First", "desc", ""),
	)
	require.NoError(t, err)
	require.NoError(t, database.Close())

	reopened, _ := newBacklogItemRepositoryAt(t, dir)
	items, err := reopened.ListBacklogItems(t.Context(), "project-1", "", 10)

	// Assert
	require.NoError(t, err)
	require.Len(t, items, 1)
	require.Equal(t, "backlog-1", items[0].ID())
	require.Equal(t, "project-1", items[0].ProjectID())
	require.Equal(t, "First", items[0].Title())
	require.Equal(t, "desc", items[0].Description())
	require.Equal(t, domain.BacklogItemStatusOpen, items[0].Status())
	require.Equal(t, "000000000001", items[0].OrderKey())
}

func TestProjectRepository_GetBacklogItemReturnsPersistedItem(t *testing.T) {
	t.Parallel()

	// Arrange
	repo := newBacklogItemRepository(t)
	require.NoError(
		t,
		repo.CreateBacklogItem(
			t.Context(),
			mustNewBacklogItemWithStatus(
				t,
				"backlog-1",
				"project-1",
				"First",
				"desc",
				domain.BacklogItemStatusRunning,
				"000000000001",
			),
		),
	)

	// Act
	item, err := repo.GetBacklogItem(t.Context(), "backlog-1")

	// Assert
	require.NoError(t, err)
	require.Equal(t, "backlog-1", item.ID())
	require.Equal(t, "project-1", item.ProjectID())
	require.Equal(t, "First", item.Title())
	require.Equal(t, "desc", item.Description())
	require.Equal(t, domain.BacklogItemStatusRunning, item.Status())
	require.Equal(t, "000000000001", item.OrderKey())
}

func TestProjectRepository_GetBacklogItemReturnsErrorWhenMissing(t *testing.T) {
	t.Parallel()

	// Arrange
	repo := newBacklogItemRepository(t)

	// Act
	item, err := repo.GetBacklogItem(t.Context(), "missing")

	// Assert
	require.ErrorIs(t, err, repository.ErrBacklogItemNotFound)
	require.Nil(t, item)
}

func TestProjectRepository_ListBacklogItemsReturnsItemsInOrderKeyOrder(t *testing.T) {
	t.Parallel()

	// Arrange
	repo := newBacklogItemRepository(t)
	require.NoError(
		t,
		repo.CreateBacklogItem(
			t.Context(),
			mustNewBacklogItem(t, "backlog-2", "project-1", "Second", "", "000000000002"),
		),
	)
	require.NoError(
		t,
		repo.CreateBacklogItem(
			t.Context(),
			mustNewBacklogItem(t, "backlog-1", "project-1", "First", "", "000000000001"),
		),
	)
	require.NoError(
		t,
		repo.CreateBacklogItem(
			t.Context(),
			mustNewBacklogItem(t, "backlog-3", "project-1", "Third", "", "000000000003"),
		),
	)

	// Act
	items, err := repo.ListBacklogItems(t.Context(), "project-1", "", 10)

	// Assert
	require.NoError(t, err)
	require.Len(t, items, 3)
	require.Equal(t, "backlog-1", items[0].ID())
	require.Equal(t, "backlog-2", items[1].ID())
	require.Equal(t, "backlog-3", items[2].ID())
}

func TestProjectRepository_ListBacklogItemsRespectsAfterIDAndLimit(t *testing.T) {
	t.Parallel()

	// Arrange
	repo := newBacklogItemRepository(t)
	require.NoError(
		t,
		repo.CreateBacklogItem(
			t.Context(),
			mustNewBacklogItem(t, "backlog-1", "project-1", "First", "", "000000000001"),
		),
	)
	require.NoError(
		t,
		repo.CreateBacklogItem(
			t.Context(),
			mustNewBacklogItem(t, "backlog-2", "project-1", "Second", "", "000000000002"),
		),
	)
	require.NoError(
		t,
		repo.CreateBacklogItem(
			t.Context(),
			mustNewBacklogItem(t, "backlog-3", "project-1", "Third", "", "000000000003"),
		),
	)

	// Act
	items, err := repo.ListBacklogItems(t.Context(), "project-1", "backlog-1", 1)

	// Assert
	require.NoError(t, err)
	require.Len(t, items, 1)
	require.Equal(t, "backlog-2", items[0].ID())
}

func TestProjectRepository_ListBacklogItemsReturnsErrorWhenAfterIDIsMissing(t *testing.T) {
	t.Parallel()

	// Arrange
	repo := newBacklogItemRepository(t)
	require.NoError(
		t,
		repo.CreateBacklogItem(
			t.Context(),
			mustNewBacklogItem(t, "backlog-1", "project-1", "First", "", ""),
		),
	)

	// Act
	items, err := repo.ListBacklogItems(t.Context(), "project-1", "missing", 10)

	// Assert
	require.ErrorIs(t, err, repository.ErrBacklogItemNotFound)
	require.Nil(t, items)
}

func TestProjectRepository_ListBacklogItemsReturnsErrorWhenAfterIDBelongsToAnotherProject(t *testing.T) {
	t.Parallel()

	// Arrange
	repo := newBacklogItemRepository(t)
	require.NoError(
		t,
		repo.CreateBacklogItem(
			t.Context(),
			mustNewBacklogItem(t, "backlog-1", "project-1", "First", "", ""),
		),
	)
	require.NoError(
		t,
		repo.CreateBacklogItem(
			t.Context(),
			mustNewBacklogItem(t, "backlog-2", "project-2", "Second", "", ""),
		),
	)

	// Act
	items, err := repo.ListBacklogItems(t.Context(), "project-1", "backlog-2", 10)

	// Assert
	require.ErrorIs(t, err, repository.ErrBacklogItemNotFound)
	require.Nil(t, items)
}

func TestProjectRepository_UpdateBacklogItemPersistsUpdatedFields(t *testing.T) {
	t.Parallel()

	// Arrange
	repo := newBacklogItemRepository(t)
	require.NoError(
		t,
		repo.CreateBacklogItem(
			t.Context(),
			mustNewBacklogItemWithStatus(
				t,
				"backlog-1",
				"project-1",
				"Before",
				"old",
				domain.BacklogItemStatusRunning,
				"000000000001",
			),
		),
	)

	updatedItem, err := domain.NewBacklogItem(
		"backlog-1",
		"project-1",
		"Before",
		"old",
		domain.BacklogItemStatusRunning,
		"000000000001",
	)
	require.NoError(t, err)

	updatedItem, err = updatedItem.SetTitle("After")
	require.NoError(t, err)

	updatedItem = updatedItem.SetDescription("new")

	// Act
	err = repo.UpdateBacklogItem(t.Context(), updatedItem)
	require.NoError(t, err)
	item, err := repo.GetBacklogItem(t.Context(), "backlog-1")

	// Assert
	require.NoError(t, err)
	require.Equal(t, "After", item.Title())
	require.Equal(t, "new", item.Description())
	require.Equal(t, domain.BacklogItemStatusRunning, item.Status())
	require.Equal(t, "000000000001", item.OrderKey())
}

func TestProjectRepository_UpdateBacklogItemReturnsErrorWhenMissing(t *testing.T) {
	t.Parallel()

	// Arrange
	repo := newBacklogItemRepository(t)

	// Act
	err := repo.UpdateBacklogItem(
		t.Context(),
		mustNewBacklogItem(t, "missing", "project-1", "After", "new", "000000000001"),
	)

	// Assert
	require.ErrorIs(t, err, repository.ErrBacklogItemNotFound)
}

func newBacklogItemRepository(t *testing.T) *badger.BacklogItemRepository {
	t.Helper()

	repo, _ := newBacklogItemRepositoryAt(t, t.TempDir())

	return repo
}

func newBacklogItemRepositoryAt(
	t *testing.T,
	dir string,
) (*badger.BacklogItemRepository, *badger.Database) {
	t.Helper()

	database, err := badger.NewDatabase(dir)
	require.NoError(t, err)

	t.Cleanup(func() {
		require.NoError(t, database.Close())
	})

	return badger.NewBacklogItemRepository(database), database
}

func mustNewBacklogItem(
	t *testing.T,
	id string,
	projectID string,
	title string,
	description string,
	orderKey string,
) *domain.BacklogItem {
	t.Helper()

	return mustNewBacklogItemWithStatus(
		t,
		id,
		projectID,
		title,
		description,
		domain.BacklogItemStatusOpen,
		orderKey,
	)
}

func mustNewBacklogItemWithStatus(
	t *testing.T,
	id string,
	projectID string,
	title string,
	description string,
	status domain.BacklogItemStatus,
	orderKey string,
) *domain.BacklogItem {
	t.Helper()

	item, err := domain.NewBacklogItem(
		id,
		projectID,
		title,
		description,
		status,
		orderKey,
	)
	require.NoError(t, err)

	return item
}
