package ravendb

import (
	"net/http"
)

var _ IVoidMaintenanceOperation = &ResetIndexOperation{}

type ResetIndexOperation struct {
	_indexName string

	Command *ResetIndexCommand
}

func NewResetIndexOperation(indexName string) *ResetIndexOperation {
	panicIf(indexName == "", "indexName cannot be empty")

	return &ResetIndexOperation{
		_indexName: indexName,
	}
}

func (o *ResetIndexOperation) GetCommand(conventions *DocumentConventions) RavenCommand {
	o.Command = NewResetIndexCommand(o._indexName)
	return o.Command
}

var (
	_ RavenCommand = &ResetIndexCommand{}
)

type ResetIndexCommand struct {
	*RavenCommandBase

	_indexName string
}

func NewResetIndexCommand(indexName string) *ResetIndexCommand {
	panicIf(indexName == "", "indexName cannot be empty")
	cmd := &ResetIndexCommand{
		RavenCommandBase: NewRavenCommandBase(),

		_indexName: indexName,
	}
	cmd.responseType = RavenCommandResponseType_EMPTY
	return cmd
}

func (c *ResetIndexCommand) CreateRequest(node *ServerNode) (*http.Request, error) {
	url := node.getUrl() + "/databases/" + node.getDatabase() + "/indexes?name=" + UrlUtils_escapeDataString(c._indexName)

	return NewHttpReset(url)
}
