package ravendb

import (
	"net/http"
)

var _ IVoidMaintenanceOperation = &StopIndexOperation{}

type StopIndexOperation struct {
	indexName string

	Command *StopIndexCommand
}

func NewStopIndexOperation(indexName string) (*StopIndexOperation, error) {
	if indexName == "" {
		return nil, newIllegalArgumentError("Index name connot be empty")
	}
	return &StopIndexOperation{
		indexName: indexName,
	}, nil
}

func (o *StopIndexOperation) GetCommand(conventions *DocumentConventions) (RavenCommand, error) {
	var err error
	o.Command, err = NewStopIndexCommand(o.indexName)
	if err != nil {
		return nil, err
	}
	return o.Command, nil
}

var (
	_ RavenCommand = &StopIndexCommand{}
)

type StopIndexCommand struct {
	RavenCommandBase

	indexName string
}

func NewStopIndexCommand(indexName string) (*StopIndexCommand, error) {
	if indexName == "" {
		return nil, newIllegalArgumentError("Index name connot be empty")
	}

	cmd := &StopIndexCommand{
		RavenCommandBase: NewRavenCommandBase(),

		indexName: indexName,
	}
	cmd.ResponseType = RavenCommandResponseTypeEmpty
	return cmd, nil
}

func (c *StopIndexCommand) createRequest(node *ServerNode) (*http.Request, error) {
	url := node.URL + "/databases/" + node.Database + "/admin/indexes/stop?name=" + urlUtilsEscapeDataString(c.indexName)

	return NewHttpPost(url, nil)
}
