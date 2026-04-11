package cli

import (
	"context"
	"fmt"
	"io"

	"github.com/neatflowcv/worker/internal/app/flow"
)

type projectBacklogListCommand struct {
	Project string `arg:""                help:"Project name." name:"project"`
	After   string `help:"List after ID." name:"after"`
	Limit   int    `help:"Maximum items." name:"limit"`
}

func (c *projectBacklogListCommand) Run(ctx context.Context, service *flow.Service, stdout io.Writer) error {
	items, err := service.ListBacklogItems(ctx, c.Project, c.After, c.Limit)
	if err != nil {
		return fmt.Errorf("list backlog items: %w", err)
	}

	for _, item := range items {
		_, err = fmt.Fprintf(stdout, "%s %s %s\n", item.ID(), item.Status(), item.Title())
		if err != nil {
			return fmt.Errorf("write backlog item output: %w", err)
		}
	}

	return nil
}
