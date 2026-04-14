package gitter_test

import (
	"context"
	"testing"

	"github.com/neatflowcv/worker/internal/pkg/gitter"
	"github.com/stretchr/testify/require"
)

func TestGitterInterfaceCanBeImplemented(t *testing.T) {
	t.Parallel()

	// Arrange
	var client gitter.Gitter = stubGitter{}

	// Act
	cloneErr := client.CloneRepository(t.Context(), "https://example.com/repo.git", "/tmp/repo")
	pullErr := client.PullRepository(t.Context(), "/tmp/repo")

	// Assert
	require.NoError(t, cloneErr)
	require.NoError(t, pullErr)
}

type stubGitter struct{}

func (stubGitter) CloneRepository(_ context.Context, _ string, _ string) error {
	return nil
}

func (stubGitter) PullRepository(_ context.Context, _ string) error {
	return nil
}
