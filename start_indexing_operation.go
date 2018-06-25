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

func (o *StartIndexingOperation) getCommand(conventions *DocumentConventions) RavenCommand {
	o.Command = NewStartIndexingCommand()
	return o.Command
}

var (
	_ RavenCommand = &StartIndexingCommand{}
)

type StartIndexingCommand struct {
	*RavenCommandBase
}

func NewStartIndexingCommand() *StartIndexingCommand {
	return &StartIndexingCommand{}
}

func (c *StartIndexingCommand) createRequest(node *ServerNode) (*http.Request, error) {
	url := node.getUrl() + "/databases/" + node.getDatabase() + "/admin/indexes/start"

	return NewHttpPost(url, nil)
}
