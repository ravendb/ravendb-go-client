package ravendb

import (
	"net/http"
)

var _ IVoidMaintenanceOperation = &StopIndexOperation{}

type StopIndexOperation struct {
	_indexName string

	Command *StopIndexCommand
}

func NewStopIndexOperation(indexName string) *StopIndexOperation {
	panicIf(indexName == "", "Index name connot be empty")
	return &StopIndexOperation{
		_indexName: indexName,
	}
}

func (o *StopIndexOperation) GetCommand(conventions *DocumentConventions) RavenCommand {
	o.Command = NewStopIndexCommand(o._indexName)
	return o.Command
}

var (
	_ RavenCommand = &StopIndexCommand{}
)

type StopIndexCommand struct {
	*RavenCommandBase

	_indexName string
}

func NewStopIndexCommand(indexName string) *StopIndexCommand {
	panicIf(indexName == "", "Index name connot be empty")

	cmd := &StopIndexCommand{
		RavenCommandBase: NewRavenCommandBase(),

		_indexName: indexName,
	}
	cmd.ResponseType = RavenCommandResponseType_EMPTY
	return cmd
}

func (c *StopIndexCommand) CreateRequest(node *ServerNode) (*http.Request, error) {
	url := node.GetUrl() + "/databases/" + node.GetDatabase() + "/admin/indexes/stop?name=" + UrlUtils_escapeDataString(c._indexName)

	return NewHttpPost(url, nil)
}
