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

func (o *StartIndexOperation) getCommand(conventions *DocumentConventions) RavenCommand {
	o.Command = NewStartIndexCommand(o._indexName)
	return o.Command
}

var (
	_ RavenCommand = &StartIndexCommand{}
)

type StartIndexCommand struct {
	*RavenCommandBase

	_indexName string
}

func NewStartIndexCommand(indexName string) *StartIndexCommand {
	panicIf(indexName == "", "Index name connot be empty")

	return &StartIndexCommand{
		RavenCommandBase: NewRavenCommandBase(),

		_indexName: indexName,
	}
}

func (c *StartIndexCommand) createRequest(node *ServerNode) (*http.Request, error) {
	url := node.getUrl() + "/databases/" + node.getDatabase() + "/admin/indexes/start?name=" + UrlUtils_escapeDataString(c._indexName)

	return NewHttpPost(url, nil)
}
