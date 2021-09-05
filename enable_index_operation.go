package ravendb

import (
	"net/http"
)

var _ IVoidMaintenanceOperation = &EnableIndexOperation{}

type EnableIndexOperation struct {
	indexName string

	Command *EnableIndexCommand
}

func NewEnableIndexOperation(indexName string) (*EnableIndexOperation, error) {
	if indexName == "" {
		return nil, newIllegalArgumentError("Index name connot be empty")
	}
	return &EnableIndexOperation{
		indexName: indexName,
	}, nil
}

func (o *EnableIndexOperation) GetCommand(conventions *DocumentConventions) (RavenCommand, error) {
	var err error
	o.Command, err = NewEnableIndexCommand(o.indexName)
	if err != nil {
		return nil, err
	}
	return o.Command, nil
}

var (
	_ RavenCommand = &EnableIndexCommand{}
)

type EnableIndexCommand struct {
	RavenCommandBase

	indexName string
}

func NewEnableIndexCommand(indexName string) (*EnableIndexCommand, error) {
	if indexName == "" {
		return nil, newIllegalArgumentError("Index name connot be empty")
	}

	cmd := &EnableIndexCommand{
		RavenCommandBase: NewRavenCommandBase(),

		indexName: indexName,
	}
	cmd.ResponseType = RavenCommandResponseTypeEmpty
	return cmd, nil
}

func (c *EnableIndexCommand) CreateRequest(node *ServerNode) (*http.Request, error) {
	url := node.URL + "/databases/" + node.Database + "/admin/indexes/enable?name=" + urlUtilsEscapeDataString(c.indexName)

	return NewHttpPost(url, nil)
}
