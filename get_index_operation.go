package ravendb

import (
	"encoding/json"
	"net/http"
)

var _ IMaintenanceOperation = &GetIndexOperation{}

type GetIndexOperation struct {
	_indexName string

	Command *GetIndexCommand
}

func NewGetIndexOperation(indexName string) *GetIndexOperation {
	panicIf(indexName == "", "Index name connot be empty")
	return &GetIndexOperation{
		_indexName: indexName,
	}
}

func (o *GetIndexOperation) GetCommand(conventions *DocumentConventions) RavenCommand {
	o.Command = NewGetIndexCommand(o._indexName)
	return o.Command
}

var (
	_ RavenCommand = &GetIndexCommand{}
)

type GetIndexCommand struct {
	*RavenCommandBase

	_indexName string

	Result *IndexDefinition
}

func NewGetIndexCommand(indexName string) *GetIndexCommand {
	panicIf(indexName == "", "Index name connot be empty")

	res := &GetIndexCommand{
		RavenCommandBase: NewRavenCommandBase(),

		_indexName: indexName,
	}
	res.IsReadRequest = true
	return res
}

func (c *GetIndexCommand) CreateRequest(node *ServerNode) (*http.Request, error) {
	url := node.getUrl() + "/databases/" + node.getDatabase() + "/indexes?name=" + UrlUtils_escapeDataString(c._indexName)

	return NewHttpGet(url)
}

func (c *GetIndexCommand) SetResponse(response []byte, fromCache bool) error {
	if response == nil {
		return throwInvalidResponse()
	}

	var res struct {
		Results []*IndexDefinition `json:"Results"`
	}

	err := json.Unmarshal(response, &res)
	if err != nil {
		return err
	}
	if len(res.Results) == 0 {
		return throwInvalidResponse()
	}
	c.Result = res.Results[0]
	return nil
}
