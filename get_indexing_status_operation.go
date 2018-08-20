package ravendb

import (
	"encoding/json"
	"net/http"
)

var _ IMaintenanceOperation = &GetIndexingStatusOperation{}

type GetIndexingStatusOperation struct {
	Command *GetIndexingStatusCommand
}

func NewGetIndexingStatusOperation() *GetIndexingStatusOperation {
	return &GetIndexingStatusOperation{}
}

func (o *GetIndexingStatusOperation) GetCommand(conventions *DocumentConventions) RavenCommand {
	o.Command = NewGetIndexingStatusCommand()
	return o.Command
}

var (
	_ RavenCommand = &GetIndexingStatusCommand{}
)

type GetIndexingStatusCommand struct {
	*RavenCommandBase

	Result *IndexingStatus
}

func NewGetIndexingStatusCommand() *GetIndexingStatusCommand {
	res := &GetIndexingStatusCommand{
		RavenCommandBase: NewRavenCommandBase(),
	}
	res.IsReadRequest = true
	return res
}

func (c *GetIndexingStatusCommand) CreateRequest(node *ServerNode) (*http.Request, error) {
	url := node.GetUrl() + "/databases/" + node.GetDatabase() + "/indexes/status"

	return NewHttpGet(url)
}

func (c *GetIndexingStatusCommand) SetResponse(response []byte, fromCache bool) error {
	if response == nil {
		return throwInvalidResponse()
	}

	return json.Unmarshal(response, &c.Result)
}
