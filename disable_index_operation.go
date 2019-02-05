package ravendb

import (
	"net/http"
)

var _ IVoidMaintenanceOperation = &DisableIndexOperation{}

type DisableIndexOperation struct {
	_indexName string

	Command *DisableIndexCommand
}

func NewDisableIndexOperation(indexName string) (*DisableIndexOperation, error) {
	if indexName == "" {
		return nil, newIllegalArgumentError("Index name connot be empty")
	}
	return &DisableIndexOperation{
		_indexName: indexName,
	}, nil
}

func (o *DisableIndexOperation) GetCommand(conventions *DocumentConventions) (RavenCommand, error) {
	var err error
	o.Command, err = NewDisableIndexCommand(o._indexName)
	if err != nil {
		return nil, err
	}
	return o.Command, nil
}

var (
	_ RavenCommand = &DisableIndexCommand{}
)

type DisableIndexCommand struct {
	RavenCommandBase

	_indexName string
}

func NewDisableIndexCommand(indexName string) (*DisableIndexCommand, error) {
	if indexName == "" {
		return nil, newIllegalArgumentError("Index name connot be empty")
	}

	cmd := &DisableIndexCommand{
		RavenCommandBase: NewRavenCommandBase(),

		_indexName: indexName,
	}
	cmd.ResponseType = RavenCommandResponseTypeEmpty
	return cmd, nil
}

func (c *DisableIndexCommand) createRequest(node *ServerNode) (*http.Request, error) {
	url := node.URL + "/databases/" + node.Database + "/admin/indexes/disable?name=" + urlUtilsEscapeDataString(c._indexName)

	return newHttpPost(url, nil)
}
