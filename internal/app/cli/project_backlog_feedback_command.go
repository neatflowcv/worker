package cli

import (
	"context"
	"errors"
	"io"

	"github.com/neatflowcv/worker/internal/app/flow"
)

var errFeedbackMessageRequired = errors.New("exactly one of --message or --message-file is required")

type projectBacklogFeedbackCommand struct {
	Project     string `arg:""                   help:"Project name."    name:"project"`
	ID          string `arg:""                   help:"Backlog item ID." name:"id"`
	Message     string `help:"Feedback message." name:"message"          xor:"message_input"`
	MessageFile string `help:"Message file."     name:"message-file"     xor:"message_input"`
}

func (c *projectBacklogFeedbackCommand) Validate() error {
	hasMessage := c.Message != ""
	hasMessageFile := c.MessageFile != ""

	if hasMessage == hasMessageFile {
		return errFeedbackMessageRequired
	}

	return nil
}

func (c *projectBacklogFeedbackCommand) Run(ctx context.Context, service *flow.Service, stdout io.Writer) error {
	_ = c
	_ = ctx
	_ = service
	_ = stdout

	return errNotImplemented
}
