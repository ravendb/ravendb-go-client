package ravendb

import (
	"net/http"
)

var _ IMaintenanceOperation = &GetIndexingStatusOperation{}

type GetIndexingStatusOperation struct {
	Command *GetIndexingStatusCommand
}

func NewGetIndexingStatusOperation() *GetIndexingStatusOperation {
	return &GetIndexingStatusOperation{}
}

func (o *GetIndexingStatusOperation) GetCommand(conventions *DocumentConventions) (RavenCommand, error) {
	o.Command = NewGetIndexingStatusCommand()
	return o.Command, nil
}

var (
	_ RavenCommand = &GetIndexingStatusCommand{}
)

type GetIndexingStatusCommand struct {
	RavenCommandBase

	Result *IndexingStatus
}

func NewGetIndexingStatusCommand() *GetIndexingStatusCommand {
	res := &GetIndexingStatusCommand{
		RavenCommandBase: NewRavenCommandBase(),
	}
	res.IsReadRequest = true
	return res
}

func (c *GetIndexingStatusCommand) createRequest(node *ServerNode) (*http.Request, error) {
	url := node.URL + "/databases/" + node.Database + "/indexes/status"

	return NewHttpGet(url)
}

func (c *GetIndexingStatusCommand) SetResponse(response []byte, fromCache bool) error {
	if response == nil {
		return throwInvalidResponse()
	}

	return jsonUnmarshal(response, &c.Result)
}
