package ravendb

import (
	"net/http"
)

var _ IVoidMaintenanceOperation = &StartIndexOperation{}

type StartIndexOperation struct {
	indexName string

	Command *StartIndexCommand
}

func NewStartIndexOperation(indexName string) (*StartIndexOperation, error) {
	if indexName == "" {
		return nil, newIllegalArgumentError("Index name connot be empty")
	}
	return &StartIndexOperation{
		indexName: indexName,
	}, nil
}

func (o *StartIndexOperation) GetCommand(conventions *DocumentConventions) (RavenCommand, error) {
	var err error
	o.Command, err = NewStartIndexCommand(o.indexName)
	if err != nil {
		return nil, err
	}
	return o.Command, nil
}

var (
	_ RavenCommand = &StartIndexCommand{}
)

type StartIndexCommand struct {
	RavenCommandBase

	indexName string
}

func NewStartIndexCommand(indexName string) (*StartIndexCommand, error) {
	if indexName == "" {
		return nil, newIllegalArgumentError("Index name connot be empty")
	}

	cmd := &StartIndexCommand{
		RavenCommandBase: NewRavenCommandBase(),

		indexName: indexName,
	}
	cmd.ResponseType = RavenCommandResponseTypeEmpty
	return cmd, nil
}

func (c *StartIndexCommand) createRequest(node *ServerNode) (*http.Request, error) {
	url := node.URL + "/databases/" + node.Database + "/admin/indexes/start?name=" + urlUtilsEscapeDataString(c.indexName)

	return NewHttpPost(url, nil)
}
