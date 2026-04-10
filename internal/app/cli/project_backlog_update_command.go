package cli

import (
	"context"
	"io"

	"github.com/neatflowcv/worker/internal/app/flow"
)

type projectBacklogUpdateCommand struct {
	Project         string `arg:""                           help:"Project name."    name:"project"`
	ID              string `arg:""                           help:"Backlog item ID." name:"id"`
	Description     string `help:"Backlog item description." name:"description"      xor:"description_input"`
	DescriptionFile string `help:"File path."                name:"description-file" xor:"description_input"`
}

func (c *projectBacklogUpdateCommand) Run(ctx context.Context, service *flow.Service, stdout io.Writer) error {
	_ = c
	_ = ctx
	_ = service
	_ = stdout

	return errNotImplemented
}
