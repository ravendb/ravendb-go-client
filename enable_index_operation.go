package ravendb

import (
	"net/http"
)

var _ IVoidMaintenanceOperation = &EnableIndexOperation{}

type EnableIndexOperation struct {
	_indexName string

	Command *EnableIndexCommand
}

func NewEnableIndexOperation(indexName string) *EnableIndexOperation {
	panicIf(indexName == "", "Index name connot be empty")
	return &EnableIndexOperation{
		_indexName: indexName,
	}
}

func (o *EnableIndexOperation) getCommand(conventions *DocumentConventions) RavenCommand {
	o.Command = NewEnableIndexCommand(o._indexName)
	return o.Command
}

var (
	_ RavenCommand = &EnableIndexCommand{}
)

type EnableIndexCommand struct {
	*RavenCommandBase

	_indexName string
}

func NewEnableIndexCommand(indexName string) *EnableIndexCommand {
	panicIf(indexName == "", "Index name connot be empty")

	return &EnableIndexCommand{
		RavenCommandBase: NewRavenCommandBase(),

		_indexName: indexName,
	}
}

func (c *EnableIndexCommand) createRequest(node *ServerNode) (*http.Request, error) {
	url := node.getUrl() + "/databases/" + node.getDatabase() + "/admin/indexes/disable?name=" + UrlUtils_escapeDataString(c._indexName)

	return NewHttpPost(url, nil)
}
