package ravendb

import (
	"net/http"
)

var _ IVoidMaintenanceOperation = &DisableIndexOperation{}

type DisableIndexOperation struct {
	_indexName string

	Command *DisableIndexCommand
}

func NewDisableIndexOperation(indexName string) *DisableIndexOperation {
	panicIf(indexName == "", "Index name connot be empty")
	return &DisableIndexOperation{
		_indexName: indexName,
	}
}

func (o *DisableIndexOperation) GetCommand(conventions *DocumentConventions) RavenCommand {
	o.Command = NewDisableIndexCommand(o._indexName)
	return o.Command
}

var (
	_ RavenCommand = &DisableIndexCommand{}
)

type DisableIndexCommand struct {
	*RavenCommandBase

	_indexName string
}

func NewDisableIndexCommand(indexName string) *DisableIndexCommand {
	panicIf(indexName == "", "Index name connot be empty")

	cmd := &DisableIndexCommand{
		RavenCommandBase: NewRavenCommandBase(),

		_indexName: indexName,
	}
	cmd.responseType = RavenCommandResponseType_EMPTY
	return cmd
}

func (c *DisableIndexCommand) CreateRequest(node *ServerNode) (*http.Request, error) {
	url := node.getUrl() + "/databases/" + node.getDatabase() + "/admin/indexes/disable?name=" + UrlUtils_escapeDataString(c._indexName)

	return NewHttpPost(url, nil)
}
