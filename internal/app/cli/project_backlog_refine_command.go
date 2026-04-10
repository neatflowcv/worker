package cli

import (
	"context"
	"io"

	"github.com/neatflowcv/worker/internal/app/flow"
)

type projectBacklogRefineCommand struct {
	Project string `arg:"" help:"Project name."    name:"project"`
	ID      string `arg:"" help:"Backlog item ID." name:"id"`
}

func (c *projectBacklogRefineCommand) Run(ctx context.Context, service *flow.Service, stdout io.Writer) error {
	_ = c
	_ = ctx
	_ = service
	_ = stdout

	return errNotImplemented
}
