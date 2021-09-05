package ravendb

import (
	"net/http"
)

var _ IVoidMaintenanceOperation = &ResetIndexOperation{}

type ResetIndexOperation struct {
	indexName string

	Command *ResetIndexCommand
}

func NewResetIndexOperation(indexName string) (*ResetIndexOperation, error) {
	if indexName == "" {
		return nil, newIllegalArgumentError("indexName cannot be empty")
	}

	return &ResetIndexOperation{
		indexName: indexName,
	}, nil
}

func (o *ResetIndexOperation) GetCommand(conventions *DocumentConventions) (RavenCommand, error) {
	var err error
	o.Command, err = NewResetIndexCommand(o.indexName)
	if err != nil {
		return nil, err
	}
	return o.Command, nil
}

var (
	_ RavenCommand = &ResetIndexCommand{}
)

type ResetIndexCommand struct {
	RavenCommandBase

	indexName string
}

func NewResetIndexCommand(indexName string) (*ResetIndexCommand, error) {
	if indexName == "" {
		return nil, newIllegalArgumentError("indexName cannot be empty")
	}
	cmd := &ResetIndexCommand{
		RavenCommandBase: NewRavenCommandBase(),

		indexName: indexName,
	}
	cmd.ResponseType = RavenCommandResponseTypeEmpty
	return cmd, nil
}

func (c *ResetIndexCommand) CreateRequest(node *ServerNode) (*http.Request, error) {
	url := node.URL + "/databases/" + node.Database + "/indexes?name=" + urlUtilsEscapeDataString(c.indexName)

	return newHttpReset(url)
}
