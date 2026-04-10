package cli

import (
	"context"
	"errors"
	"io"

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
	_ = c
	_ = ctx
	_ = service
	_ = stdout

	return errNotImplemented
}
