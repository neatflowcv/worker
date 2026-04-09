package cli

import (
	"context"
	"fmt"
	"io"

	"github.com/neatflowcv/worker/internal/app/flow"
)

type projectCreateCommand struct {
	Name string `arg:"" help:"Project name."   name:"name"`
	URL  string `arg:"" help:"Repository URL." name:"url"`
}

func (c *projectCreateCommand) Run(ctx context.Context, service *flow.Service, stdout io.Writer) error {
	project, err := service.CreateProject(ctx, c.Name, c.URL)
	if err != nil {
		return fmt.Errorf("create project: %w", err)
	}

	_, err = fmt.Fprintf(
		stdout,
		"created project %s (%s) from %s\n",
		project.Name(),
		project.ID(),
		project.RepositoryURL(),
	)
	if err != nil {
		return fmt.Errorf("write project output: %w", err)
	}

	return nil
}
