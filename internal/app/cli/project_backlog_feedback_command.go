package cli

import (
	"context"
	"fmt"
	"io"

	"github.com/neatflowcv/worker/internal/app/flow"
)

type projectBacklogFeedbackCommand struct {
	Project string `arg:"" help:"Project name."     name:"project"`
	ID      string `arg:"" help:"Backlog item ID."  name:"id"`
	Message string `arg:"" help:"Feedback message." name:"message"`
}

func (c *projectBacklogFeedbackCommand) Run(ctx context.Context, service *flow.Service, stdout io.Writer) error {
	item, err := service.FeedbackBacklogItem(ctx, c.Project, c.ID, c.Message)
	if err != nil {
		return fmt.Errorf("feedback backlog item: %w", err)
	}

	_, err = fmt.Fprintf(
		stdout,
		"sent feedback to backlog item %s (%s) in project %s\n",
		item.Title(),
		item.ID(),
		c.Project,
	)
	if err != nil {
		return fmt.Errorf("write backlog item output: %w", err)
	}

	return nil
}
