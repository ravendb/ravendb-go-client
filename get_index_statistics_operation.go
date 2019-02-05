package ravendb

import (
	"net/http"
)

var _ IMaintenanceOperation = &GetIndexStatisticsOperation{}

type GetIndexStatisticsOperation struct {
	indexName string

	Command *GetIndexStatisticsCommand
}

func NewGetIndexStatisticsOperation(indexName string) *GetIndexStatisticsOperation {
	panicIf(indexName == "", "Index name connot be empty")
	return &GetIndexStatisticsOperation{
		indexName: indexName,
	}
}

func (o *GetIndexStatisticsOperation) GetCommand(conventions *DocumentConventions) (RavenCommand, error) {
	var err error
	o.Command, err = NewGetIndexStatisticsCommand(o.indexName)
	if err != nil {
		return nil, err
	}
	return o.Command, nil
}

var (
	_ RavenCommand = &GetIndexStatisticsCommand{}
)

type GetIndexStatisticsCommand struct {
	RavenCommandBase

	_indexName string

	Result *IndexStats
}

func NewGetIndexStatisticsCommand(indexName string) (*GetIndexStatisticsCommand, error) {
	if indexName == "" {
		return nil, newIllegalArgumentError("Index name connot be empty")
	}

	res := &GetIndexStatisticsCommand{
		RavenCommandBase: NewRavenCommandBase(),

		_indexName: indexName,
	}
	res.IsReadRequest = true
	return res, nil
}

func (c *GetIndexStatisticsCommand) createRequest(node *ServerNode) (*http.Request, error) {
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
