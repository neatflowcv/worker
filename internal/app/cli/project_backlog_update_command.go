package cli

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/neatflowcv/worker/internal/app/flow"
)

var errBacklogUpdateInputRequired = errors.New(
	"at least one of --title or --description is required",
)

type projectBacklogUpdateCommand struct {
	Project     string  `arg:""                                        help:"Project name."    name:"project"`
	ID          string  `arg:""                                        help:"Backlog item ID." name:"id"`
	Title       *string `help:"Backlog item title."                    name:"title"`
	Description string  `help:"Path to backlog item description file." name:"description"`
}

func (c *projectBacklogUpdateCommand) Validate() error {
	if c.Title == nil && c.Description == "" {
		return errBacklogUpdateInputRequired
	}

	return nil
}

func (c *projectBacklogUpdateCommand) Run(ctx context.Context, service *flow.Service, stdout io.Writer) error {
	description, hasDescription, err := c.resolveDescription()
	if err != nil {
		return err
	}

	var descriptionInput *string

	if hasDescription {
		descriptionInput = &description
	}

	item, err := service.UpdateBacklogItem(
		ctx,
		c.Project,
		c.ID,
		c.Title,
		descriptionInput,
	)
	if err != nil {
		return fmt.Errorf("update backlog item: %w", err)
	}

	_, err = fmt.Fprintf(
		stdout,
		"updated backlog item %s (%s) in project %s\n",
		item.Title(),
		item.ID(),
		c.Project,
	)
	if err != nil {
		return fmt.Errorf("write backlog item output: %w", err)
	}

	return nil
}

func (c *projectBacklogUpdateCommand) resolveDescription() (string, bool, error) {
	if c.Description == "" {
		return "", false, nil
	}

	content, err := os.ReadFile(c.Description)
	if err != nil {
		return "", false, fmt.Errorf("read description file: %w", err)
	}

	return string(content), true, nil
}
