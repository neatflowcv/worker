package cli

import (
	"context"
	"fmt"
	"io"

	"github.com/neatflowcv/worker/internal/app/flow"
)

type projectListCommand struct{}

func (c *projectListCommand) Run(ctx context.Context, service *flow.Service, stdout io.Writer) error {
	projects, err := service.ListProjects(ctx)
	if err != nil {
		return fmt.Errorf("list projects: %w", err)
	}

	for _, project := range projects {
		_, err = fmt.Fprintf(
			stdout,
			"%s (%s) %s\n",
			project.Name(),
			project.ID(),
			project.RepositoryURL(),
		)
		if err != nil {
			return fmt.Errorf("write project output: %w", err)
		}
	}

	return nil
}
