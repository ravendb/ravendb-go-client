package ravendb

import (
	"net/http"
)

var _ IVoidMaintenanceOperation = &StopIndexingOperation{}

type StopIndexingOperation struct {
	Command *StopIndexingCommand
}

func NewStopIndexingOperation() *StopIndexingOperation {
	return &StopIndexingOperation{}
}

func (o *StopIndexingOperation) GetCommand(conventions *DocumentConventions) RavenCommand {
	o.Command = NewStopIndexingCommand()
	return o.Command
}

var (
	_ RavenCommand = &StopIndexingCommand{}
)

type StopIndexingCommand struct {
	*RavenCommandBase
}

func NewStopIndexingCommand() *StopIndexingCommand {
	cmd := &StopIndexingCommand{
		RavenCommandBase: NewRavenCommandBase(),
	}
	cmd.responseType = RavenCommandResponseType_EMPTY
	return cmd
}

func (c *StopIndexingCommand) CreateRequest(node *ServerNode) (*http.Request, error) {
	url := node.getUrl() + "/databases/" + node.getDatabase() + "/admin/indexes/stop"

	return NewHttpPost(url, nil)
}
