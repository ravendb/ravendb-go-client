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

func (o *StopIndexingOperation) GetCommand(conventions *DocumentConventions) (RavenCommand, error) {
	o.Command = NewStopIndexingCommand()
	return o.Command, nil
}

var (
	_ RavenCommand = &StopIndexingCommand{}
)

type StopIndexingCommand struct {
	RavenCommandBase
}

func NewStopIndexingCommand() *StopIndexingCommand {
	cmd := &StopIndexingCommand{
		RavenCommandBase: NewRavenCommandBase(),
	}
	cmd.ResponseType = RavenCommandResponseTypeEmpty
	return cmd
}

func (c *StopIndexingCommand) createRequest(node *ServerNode) (*http.Request, error) {
	url := node.URL + "/databases/" + node.Database + "/admin/indexes/stop"

	return NewHttpPost(url, nil)
}
