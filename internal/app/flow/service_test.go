package flow_test

import (
	"context"
	"testing"

	"github.com/neatflowcv/worker/internal/app/flow"
	"github.com/oklog/ulid/v2"
	"github.com/stretchr/testify/require"
)

func TestService_CreateProject(t *testing.T) {
	t.Parallel()

	service := flow.NewService()

	project, err := service.CreateProject(context.Background(), "worker", "https://github.com/neatflowcv/worker.git")
	require.NoError(t, err)
	require.NotNil(t, project)
	require.Equal(t, "worker", project.Name())
	require.Equal(t, "https://github.com/neatflowcv/worker.git", project.RepositoryURL())

	_, err = ulid.Parse(project.ID())
	require.NoError(t, err)
}
