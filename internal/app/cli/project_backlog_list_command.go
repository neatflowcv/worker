package cli

import (
	"context"
	"io"

	"github.com/neatflowcv/worker/internal/app/flow"
)

type projectBacklogListCommand struct {
	Project string `arg:""                help:"Project name." name:"project"`
	After   string `help:"List after ID." name:"after"`
	Limit   int    `help:"Maximum items." name:"limit"`
}

func (c *projectBacklogListCommand) Run(ctx context.Context, service *flow.Service, stdout io.Writer) error {
	_ = c
	_ = ctx
	_ = service
	_ = stdout

	return errNotImplemented
}
