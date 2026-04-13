package domain_test

import (
	"testing"

	"github.com/neatflowcv/worker/internal/pkg/domain"
	"github.com/stretchr/testify/require"
)

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
