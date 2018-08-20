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

func (o *EnableIndexOperation) GetCommand(conventions *DocumentConventions) RavenCommand {
	o.Command = NewEnableIndexCommand(o._indexName)
	return o.Command
}

var (
	_ RavenCommand = &EnableIndexCommand{}
)

type EnableIndexCommand struct {
	RavenCommandBase

	_indexName string
}

func NewEnableIndexCommand(indexName string) *EnableIndexCommand {
	panicIf(indexName == "", "Index name connot be empty")

	cmd := &EnableIndexCommand{
		RavenCommandBase: NewRavenCommandBase(),

		_indexName: indexName,
	}
	cmd.ResponseType = RavenCommandResponseType_EMPTY
	return cmd
}

func (c *EnableIndexCommand) CreateRequest(node *ServerNode) (*http.Request, error) {
	url := node.GetUrl() + "/databases/" + node.GetDatabase() + "/admin/indexes/enable?name=" + UrlUtils_escapeDataString(c._indexName)

	return NewHttpPost(url, nil)
}
