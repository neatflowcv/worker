package cli

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/neatflowcv/worker/internal/app/flow"
)

const refineOutputFileMode = 0o600

type projectBacklogRefineCommand struct {
	Project string `arg:""            help:"Project name."     name:"project"`
	ID      string `arg:""            help:"Backlog item ID."  name:"id"`
	Output  string `default:"plan.md" help:"Output file path." name:"output"`
}

func (c *projectBacklogRefineCommand) Run(ctx context.Context, service *flow.Service, stdout io.Writer) error {
	_ = stdout

	item, err := service.RefineBacklogItem(ctx, c.Project, c.ID)
	if err != nil {
		return fmt.Errorf("refine backlog item: %w", err)
	}

	err = os.WriteFile(c.Output, []byte(item.Description()), refineOutputFileMode)
	if err != nil {
		return fmt.Errorf("write refined backlog item file: %w", err)
	}

	return nil
}
