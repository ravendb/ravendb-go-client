package ravendb

import (
	"encoding/json"
	"net/http"
)

var (
	_ IMaintenanceOperation = &GetStatisticsOperation{}
)

type GetIndexErrorsOperation struct {
	_indexNames []string

	Command *GetIndexErrorsCommand
}

func NewGetIndexErrorsOperation(indexNames []string) *GetIndexErrorsOperation {
	return &GetIndexErrorsOperation{
		_indexNames: indexNames,
	}
}

func (o *GetIndexErrorsOperation) getCommand(conventions *DocumentConventions) RavenCommand {
	o.Command = NewGetIndexErrorsCommand(o._indexNames)
	return o.Command
}

var _ RavenCommand = &GetIndexErrorsCommand{}

type GetIndexErrorsCommand struct {
	*RavenCommandBase

	_indexNames []string

	Result []*IndexErrors
}

func NewGetIndexErrorsCommand(indexNames []string) *GetIndexErrorsCommand {
	res := &GetIndexErrorsCommand{
		_indexNames: indexNames,
	}
	res.IsReadRequest = true
	return res
}

func (c *GetIndexErrorsCommand) createRequest(node *ServerNode) (*http.Request, error) {
	url := node.getUrl() + "/databases/" + node.getDatabase() + "/indexes/errors"

	if len(c._indexNames) > 0 {
		url += "?"

		for _, indexName := range c._indexNames {
			url += "&name=" + indexName
		}
	}

	return NewHttpGet(url)
}

func (c *GetIndexErrorsCommand) setResponse(response string, fromCache bool) error {
	if response == "" {
		return throwInvalidResponse()
	}
	var res struct {
		Results []*IndexErrors
	}
	err := json.Unmarshal([]byte(response), &res)
	if err != nil {
		return err
	}
	c.Result = res.Results
	return nil
}
