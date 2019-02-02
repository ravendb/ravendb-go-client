package ravendb

import (
	"net/http"
)

var (
	_ IMaintenanceOperation = &GetIndexErrorsOperation{}
)

type GetIndexErrorsOperation struct {
	indexNames []string

	Command *GetIndexErrorsCommand
}

func NewGetIndexErrorsOperation(indexNames []string) *GetIndexErrorsOperation {
	return &GetIndexErrorsOperation{
		indexNames: indexNames,
	}
}

func (o *GetIndexErrorsOperation) GetCommand(conventions *DocumentConventions) (RavenCommand, error) {
	o.Command = NewGetIndexErrorsCommand(o.indexNames)
	return o.Command, nil
}

var _ RavenCommand = &GetIndexErrorsCommand{}

type GetIndexErrorsCommand struct {
	RavenCommandBase

	indexNames []string

	Result []*IndexErrors
}

func NewGetIndexErrorsCommand(indexNames []string) *GetIndexErrorsCommand {
	res := &GetIndexErrorsCommand{
		RavenCommandBase: NewRavenCommandBase(),

		indexNames: indexNames,
	}
	res.IsReadRequest = true
	return res
}

func (c *GetIndexErrorsCommand) CreateRequest(node *ServerNode) (*http.Request, error) {
	url := node.URL + "/databases/" + node.Database + "/indexes/errors"

	if len(c.indexNames) > 0 {
		url += "?"

		for _, indexName := range c.indexNames {
			url += "&name=" + indexName
		}
	}

	return NewHttpGet(url)
}

func (c *GetIndexErrorsCommand) SetResponse(response []byte, fromCache bool) error {
	if len(response) == 0 {
		return throwInvalidResponse()
	}
	var res struct {
		Results []*IndexErrors
	}
	err := jsonUnmarshal(response, &res)
	if err != nil {
		return err
	}
	c.Result = res.Results
	return nil
}
