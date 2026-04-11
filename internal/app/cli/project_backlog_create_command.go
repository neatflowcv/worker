package cli

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/neatflowcv/worker/internal/app/flow"
)

var errNotImplemented = errors.New("not implemented")

type projectBacklogCreateCommand struct {
	Project     string `arg:""                                        help:"Project name."       name:"project"`
	Title       string `arg:""                                        help:"Backlog item title." name:"title"`
	Description string `help:"Path to backlog item description file." name:"description"`
}

func (c *projectBacklogCreateCommand) Run(ctx context.Context, service *flow.Service, stdout io.Writer) error {
	description, err := c.resolveDescription()
	if err != nil {
		return err
	}

	item, err := service.CreateBacklogItem(ctx, c.Project, c.Title, description)
	if err != nil {
		return fmt.Errorf("create backlog item: %w", err)
	}

	_, err = fmt.Fprintf(
		stdout,
		"created backlog item %s (%s) in project %s\n",
		item.Title(),
		item.ID(),
		c.Project,
	)
	if err != nil {
		return fmt.Errorf("write backlog item output: %w", err)
	}

	return nil
}

func (c *projectBacklogCreateCommand) resolveDescription() (string, error) {
	if c.Description == "" {
		return "", nil
	}

	content, err := os.ReadFile(c.Description)
	if err != nil {
		return "", fmt.Errorf("read description file: %w", err)
	}

	return string(content), nil
}
