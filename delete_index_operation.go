package ravendb

import (
	"net/http"
)

var _ IVoidMaintenanceOperation = &DeleteIndexOperation{}

type DeleteIndexOperation struct {
	_indexName string

	Command *DeleteIndexCommand
}

func NewDeleteIndexOperation(indexName string) *DeleteIndexOperation {
	panicIf(indexName == "", "indexName cannot be empty")

	return &DeleteIndexOperation{
		_indexName: indexName,
	}
}

func (o *DeleteIndexOperation) GetCommand(conventions *DocumentConventions) (RavenCommand, error) {
	var err error
	o.Command, err = NewDeleteIndexCommand(o._indexName)
	if err != nil {
		return nil, err
	}
	return o.Command, nil
}

var (
	_ RavenCommand = &DeleteIndexCommand{}
)

type DeleteIndexCommand struct {
	RavenCommandBase

	_indexName string
}

func NewDeleteIndexCommand(indexName string) (*DeleteIndexCommand, error) {
	if indexName == "" {
		return nil, newIllegalArgumentError("indexName cannot be empty")
	}
	cmd := &DeleteIndexCommand{
		RavenCommandBase: NewRavenCommandBase(),

		_indexName: indexName,
	}
	cmd.ResponseType = RavenCommandResponseTypeEmpty
	return cmd, nil
}

func (c *DeleteIndexCommand) CreateRequest(node *ServerNode) (*http.Request, error) {
	url := node.URL + "/databases/" + node.Database + "/indexes?name=" + urlUtilsEscapeDataString(c._indexName)

	return newHttpDelete(url, nil)
}
