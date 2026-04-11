package cli

import (
	"context"
	"fmt"
	"io"

	"github.com/neatflowcv/worker/internal/app/flow"
)

type projectBacklogRefineCommand struct {
	Project string `arg:"" help:"Project name."    name:"project"`
	ID      string `arg:"" help:"Backlog item ID." name:"id"`
}

func (c *projectBacklogRefineCommand) Run(ctx context.Context, service *flow.Service, stdout io.Writer) error {
	item, err := service.RefineBacklogItem(ctx, c.Project, c.ID)
	if err != nil {
		return fmt.Errorf("refine backlog item: %w", err)
	}

	_, err = fmt.Fprintln(stdout, item.Description())
	if err != nil {
		return fmt.Errorf("write refined backlog item output: %w", err)
	}

	return nil
}
