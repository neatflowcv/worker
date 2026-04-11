package cli

import (
	"context"
	"fmt"
	"io"

	"github.com/neatflowcv/worker/internal/app/flow"
)

type projectBacklogGetCommand struct {
	Project string `arg:"" help:"Project name."    name:"project"`
	ID      string `arg:"" help:"Backlog item ID." name:"id"`
}

func (c *projectBacklogGetCommand) Run(ctx context.Context, service *flow.Service, stdout io.Writer) error {
	item, err := service.GetBacklogItem(ctx, c.Project, c.ID)
	if err != nil {
		return fmt.Errorf("get backlog item: %w", err)
	}

	_, err = fmt.Fprintf(
		stdout,
		"ID: %s\nStatus: %s\nTitle: %s\nDescription:\n%s\n",
		item.ID(),
		item.Status(),
		item.Title(),
		item.Description(),
	)
	if err != nil {
		return fmt.Errorf("write backlog item output: %w", err)
	}

	return nil
}
