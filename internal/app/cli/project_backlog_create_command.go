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
	Project         string `arg:""                           help:"Project name."       name:"project"`
	Title           string `arg:""                           help:"Backlog item title." name:"title"`
	Description     string `help:"Backlog item description." name:"description"         xor:"description_input"`
	DescriptionFile string `help:"File path."                name:"description-file"    xor:"description_input"`
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
	if c.DescriptionFile == "" {
		return c.Description, nil
	}

	content, err := os.ReadFile(c.DescriptionFile)
	if err != nil {
		return "", fmt.Errorf("read description file: %w", err)
	}

	return string(content), nil
}
