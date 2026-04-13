package cli

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/neatflowcv/worker/internal/app/flow"
	"github.com/neatflowcv/worker/internal/pkg/domain"
)

var errProjectAuthRequiresUsernameAndPassword = errors.New(
	"exactly both --username and --password must be provided, or neither",
)

type projectCreateCommand struct {
	Name     string `arg:""                      help:"Project name."   name:"name"`
	URL      string `arg:""                      help:"Repository URL." name:"url"`
	Username string `help:"Repository username." name:"username"`
	Password string `help:"Repository password." name:"password"`
}

func (c *projectCreateCommand) Validate() error {
	hasUsername := c.Username != ""
	hasPassword := c.Password != ""

	if hasUsername != hasPassword {
		return errProjectAuthRequiresUsernameAndPassword
	}

	return nil
}

func (c *projectCreateCommand) Run(ctx context.Context, service *flow.Service, stdout io.Writer) error {
	project, err := service.CreateProject(ctx, c.Name, c.URL, c.auth())
	if err != nil {
		return fmt.Errorf("create project: %w", err)
	}

	_, err = fmt.Fprintf(
		stdout,
		"created project %s (%s) from %s\n",
		project.Name(),
		project.ID(),
		project.RepositoryURL(),
	)
	if err != nil {
		return fmt.Errorf("write project output: %w", err)
	}

	return nil
}

func (c *projectCreateCommand) auth() *domain.Auth {
	if c.Username == "" && c.Password == "" {
		return nil
	}

	return domain.NewAuth(c.Username, c.Password)
}
