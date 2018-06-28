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

func (o *StopIndexOperation) getCommand(conventions *DocumentConventions) RavenCommand {
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
	cmd.responseType = RavenCommandResponseType_EMPTY
	return cmd
}

func (c *StopIndexCommand) createRequest(node *ServerNode) (*http.Request, error) {
	url := node.getUrl() + "/databases/" + node.getDatabase() + "/admin/indexes/stop?name=" + UrlUtils_escapeDataString(c._indexName)

	return NewHttpPost(url, nil)
}