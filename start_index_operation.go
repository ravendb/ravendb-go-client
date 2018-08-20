package ravendb

import (
	"net/http"
)

var _ IVoidMaintenanceOperation = &StartIndexOperation{}

type StartIndexOperation struct {
	_indexName string

	Command *StartIndexCommand
}

func NewStartIndexOperation(indexName string) *StartIndexOperation {
	panicIf(indexName == "", "Index name connot be empty")
	return &StartIndexOperation{
		_indexName: indexName,
	}
}

func (o *StartIndexOperation) GetCommand(conventions *DocumentConventions) RavenCommand {
	o.Command = NewStartIndexCommand(o._indexName)
	return o.Command
}

var (
	_ RavenCommand = &StartIndexCommand{}
)

type StartIndexCommand struct {
	RavenCommandBase

	_indexName string
}

func NewStartIndexCommand(indexName string) *StartIndexCommand {
	panicIf(indexName == "", "Index name connot be empty")

	cmd := &StartIndexCommand{
		RavenCommandBase: NewRavenCommandBase(),

		_indexName: indexName,
	}
	cmd.ResponseType = RavenCommandResponseType_EMPTY
	return cmd
}

func (c *StartIndexCommand) CreateRequest(node *ServerNode) (*http.Request, error) {
	url := node.GetUrl() + "/databases/" + node.GetDatabase() + "/admin/indexes/start?name=" + UrlUtils_escapeDataString(c._indexName)

	return NewHttpPost(url, nil)
}
