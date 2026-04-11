package orderkey_test

import (
	"testing"

	"github.com/neatflowcv/worker/internal/pkg/orderkey"
	"github.com/stretchr/testify/require"
)

func TestFirst(t *testing.T) {
	t.Parallel()

	// Act
	key := orderkey.First()

	// Assert
	require.Equal(t, "000000000001", key)
}

func TestAfterReturnsFirstWhenPreviousKeyIsEmpty(t *testing.T) {
	t.Parallel()

	// Act
	key := orderkey.After("")

	// Assert
	require.Equal(t, orderkey.First(), key)
}

func TestAfterReturnsNextKeyForValidKey(t *testing.T) {
	t.Parallel()

	// Act
	key := orderkey.After(orderkey.First())

	// Assert
	require.Equal(t, "000000000002", key)
}

func TestAfterReturnsLexicographicallyGreaterFallbackForInvalidKey(t *testing.T) {
	t.Parallel()

	// Act
	key := orderkey.After("invalid-key")

	// Assert
	require.Equal(t, "invalid-key0", key)
}

func TestBeforeReturnsFirstWhenNextKeyIsEmpty(t *testing.T) {
	t.Parallel()

	// Act
	key := orderkey.Before("")

	// Assert
	require.Equal(t, orderkey.First(), key)
}

func TestBeforeReturnsPreviousKeyForValidKey(t *testing.T) {
	t.Parallel()

	// Act
	key := orderkey.Before(orderkey.First())

	// Assert
	require.Equal(t, "000000000000", key)
}

func TestBeforeReturnsLexicographicallySmallerFallbackForInvalidKey(t *testing.T) {
	t.Parallel()

	// Act
	key := orderkey.Before("invalid-key")

	// Assert
	require.Equal(t, "-invalid-key", key)
}
