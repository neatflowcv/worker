package domain_test

import (
	"testing"

	"github.com/neatflowcv/worker/internal/pkg/domain"
	"github.com/stretchr/testify/require"
)

func newBacklogItem(t *testing.T, status domain.BacklogItemStatus) *domain.BacklogItem {
	t.Helper()

	item, err := domain.NewBacklogItem(
		"backlog-1",
		"project-1",
		"First",
		"desc",
		status,
		"000000000001",
	)
	require.NoError(t, err)

	return item
}

func assertTransitionSuccess(
	t *testing.T,
	initialStatus domain.BacklogItemStatus,
	expectedStatus domain.BacklogItemStatus,
	transition func(*domain.BacklogItem) (*domain.BacklogItem, error),
) {
	t.Helper()

	// Arrange
	item := newBacklogItem(t, initialStatus)

	// Act
	updatedItem, err := transition(item)

	// Assert
	require.NoError(t, err)
	require.Equal(t, expectedStatus, updatedItem.Status())
	require.Equal(t, initialStatus, item.Status())
}

func assertTransitionFailure(
	t *testing.T,
	initialStatus domain.BacklogItemStatus,
	expectedErr error,
	transition func(*domain.BacklogItem) (*domain.BacklogItem, error),
) {
	t.Helper()

	// Arrange
	item := newBacklogItem(t, initialStatus)

	// Act
	updatedItem, err := transition(item)

	// Assert
	require.ErrorIs(t, err, expectedErr)
	require.Nil(t, updatedItem)
	require.Equal(t, initialStatus, item.Status())
}

func TestNewBacklogItem(t *testing.T) {
	t.Parallel()

	t.Run("creates backlog item with given status", func(t *testing.T) {
		t.Parallel()

		// Arrange
		const status = domain.BacklogItemStatusRunning

		// Act
		item, err := domain.NewBacklogItem(
			"backlog-1",
			"project-1",
			"First",
			"desc",
			status,
			"000000000001",
		)

		// Assert
		require.NoError(t, err)
		require.NotNil(t, item)
		require.Equal(t, status, item.Status())
	})

	t.Run("returns error when status is invalid", func(t *testing.T) {
		t.Parallel()

		// Act
		item, err := domain.NewBacklogItem(
			"backlog-1",
			"project-1",
			"First",
			"desc",
			domain.BacklogItemStatus("invalid"),
			"000000000001",
		)

		// Assert
		require.ErrorIs(t, err, domain.ErrInvalidBacklogItemStatus)
		require.Nil(t, item)
	})
}

func TestBacklogItem_Start(t *testing.T) {
	t.Parallel()

	t.Run("starts item when status is open", func(t *testing.T) {
		t.Parallel()

		assertTransitionSuccess(
			t,
			domain.BacklogItemStatusOpen,
			domain.BacklogItemStatusRunning,
			func(item *domain.BacklogItem) (*domain.BacklogItem, error) {
				return item.Start()
			},
		)
	})

	t.Run("returns error when status is not open", func(t *testing.T) {
		t.Parallel()

		assertTransitionFailure(
			t,
			domain.BacklogItemStatusBlocked,
			domain.ErrBacklogItemCannotStart,
			func(item *domain.BacklogItem) (*domain.BacklogItem, error) {
				return item.Start()
			},
		)
	})
}

func TestBacklogItem_Blocked(t *testing.T) {
	t.Parallel()

	t.Run("blocks item when status is running", func(t *testing.T) {
		t.Parallel()

		assertTransitionSuccess(
			t,
			domain.BacklogItemStatusRunning,
			domain.BacklogItemStatusBlocked,
			func(item *domain.BacklogItem) (*domain.BacklogItem, error) {
				return item.Blocked()
			},
		)
	})

	t.Run("returns error when status is not running", func(t *testing.T) {
		t.Parallel()

		assertTransitionFailure(
			t,
			domain.BacklogItemStatusOpen,
			domain.ErrBacklogItemCannotBlock,
			func(item *domain.BacklogItem) (*domain.BacklogItem, error) {
				return item.Blocked()
			},
		)
	})
}

func TestBacklogItem_Resume(t *testing.T) {
	t.Parallel()

	t.Run("resumes item when status is blocked", func(t *testing.T) {
		t.Parallel()

		assertTransitionSuccess(
			t,
			domain.BacklogItemStatusBlocked,
			domain.BacklogItemStatusRunning,
			func(item *domain.BacklogItem) (*domain.BacklogItem, error) {
				return item.Resume()
			},
		)
	})

	t.Run("returns error when status is not blocked", func(t *testing.T) {
		t.Parallel()

		assertTransitionFailure(
			t,
			domain.BacklogItemStatusOpen,
			domain.ErrBacklogItemCannotResume,
			func(item *domain.BacklogItem) (*domain.BacklogItem, error) {
				return item.Resume()
			},
		)
	})
}
