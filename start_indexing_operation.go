package ravendb

import (
	"net/http"
)

var _ IVoidMaintenanceOperation = &StartIndexingOperation{}

type StartIndexingOperation struct {
	Command *StartIndexingCommand
}

func NewStartIndexingOperation() *StartIndexingOperation {
	return &StartIndexingOperation{}
}

func (o *StartIndexingOperation) GetCommand(conventions *DocumentConventions) (RavenCommand, error) {
	o.Command = NewStartIndexingCommand()
	return o.Command, nil
}

var (
	_ RavenCommand = &StartIndexingCommand{}
)

type StartIndexingCommand struct {
	RavenCommandBase
}

func NewStartIndexingCommand() *StartIndexingCommand {
	cmd := &StartIndexingCommand{
		RavenCommandBase: NewRavenCommandBase(),
	}
	cmd.ResponseType = RavenCommandResponseTypeEmpty
	return cmd
}

func (c *StartIndexingCommand) CreateRequest(node *ServerNode) (*http.Request, error) {
	url := node.URL + "/databases/" + node.Database + "/admin/indexes/start"

	return NewHttpPost(url, nil)
}
