package ravendb

import (
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

func (o *GetIndexStatisticsOperation) GetCommand(conventions *DocumentConventions) RavenCommand {
	o.Command = NewGetIndexStatisticsCommand(o._indexName)
	return o.Command
}

var (
	_ RavenCommand = &GetIndexStatisticsCommand{}
)

type GetIndexStatisticsCommand struct {
	RavenCommandBase

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
	url := node.URL + "/databases/" + node.Database + "/indexes/stats?name=" + urlUtilsEscapeDataString(c._indexName)

	return NewHttpGet(url)
}

func (c *GetIndexStatisticsCommand) SetResponse(response []byte, fromCache bool) error {
	if response == nil {
		return throwInvalidResponse()
	}

	var res struct {
		Results []*IndexStats `json:"Results"`
	}

	err := jsonUnmarshal(response, &res)
	if err != nil {
		return err
	}
	if len(res.Results) == 0 {
		return throwInvalidResponse()
	}
	c.Result = res.Results[0]
	return nil
}
