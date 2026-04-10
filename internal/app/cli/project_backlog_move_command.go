package cli

import (
	"context"
	"io"

	"github.com/neatflowcv/worker/internal/app/flow"
)

type projectBacklogMoveCommand struct {
	Project string `arg:""                help:"Project name."    name:"project"`
	ID      string `arg:""                help:"Backlog item ID." name:"id"`
	After   string `help:"Move after ID." name:"after"`
}

func (c *projectBacklogMoveCommand) Run(ctx context.Context, service *flow.Service, stdout io.Writer) error {
	_ = c
	_ = ctx
	_ = service
	_ = stdout

	return errNotImplemented
}
