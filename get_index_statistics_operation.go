package ravendb

import (
	"encoding/json"
	"net/http"
)

var _ IMaintenanceOperation = &GetIndexStatisticsOperation{}

type GetIndexStatisticsOperation struct {
	_indexName string

	Command *GetIndexStatisticsCommand
}

func NewGetIndexStatisticsOperation(indexName string) *GetIndexStatisticsOperation {
	panicIf(indexName == "", "Index name connot be empty")
	return &GetIndexStatisticsOperation{
		_indexName: indexName,
	}
}

func (o *GetIndexStatisticsOperation) getCommand(conventions *DocumentConventions) RavenCommand {
	o.Command = NewGetIndexStatisticsCommand(o._indexName)
	return o.Command
}

var (
	_ RavenCommand = &GetIndexStatisticsCommand{}
)

type GetIndexStatisticsCommand struct {
	*RavenCommandBase

	_indexName string

	Result *IndexStats
}

func NewGetIndexStatisticsCommand(indexName string) *GetIndexStatisticsCommand {
	panicIf(indexName == "", "Index name connot be empty")

	res := &GetIndexStatisticsCommand{
		RavenCommandBase: NewRavenCommandBase(),

		_indexName: indexName,
	}
	res.IsReadRequest = true
	return res
}

func (c *GetIndexStatisticsCommand) CreateRequest(node *ServerNode) (*http.Request, error) {
	url := node.getUrl() + "/databases/" + node.getDatabase() + "/indexes/stats?name=" + UrlUtils_escapeDataString(c._indexName)

	return NewHttpGet(url)
}

func (c *GetIndexStatisticsCommand) SetResponse(response []byte, fromCache bool) error {
	if response == nil {
		return throwInvalidResponse()
	}

	var res struct {
		Results []*IndexStats `json:"Results"`
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
