package cli

import (
	"context"
	"fmt"
	"io"

	"github.com/neatflowcv/worker/internal/app/flow"
)

type projectBacklogStartCommand struct {
	Project string `arg:"" help:"Project name."    name:"project"`
	ID      string `arg:"" help:"Backlog item ID." name:"id"`
}

func (c *projectBacklogStartCommand) Run(ctx context.Context, service *flow.Service, stdout io.Writer) error {
	item, err := service.StartBacklogItem(ctx, c.Project, c.ID)
	if err != nil {
		return fmt.Errorf("start backlog item: %w", err)
	}

	_, err = fmt.Fprintf(
		stdout,
		"started backlog item %s (%s) in project %s\n",
		item.Title(),
		item.ID(),
		c.Project,
	)
	if err != nil {
		return fmt.Errorf("write backlog item output: %w", err)
	}

	return nil
}
