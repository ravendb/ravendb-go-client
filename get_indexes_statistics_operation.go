package ravendb

import (
	"net/http"
)

var _ IMaintenanceOperation = &GetIndexesStatisticsOperation{}

type GetIndexesStatisticsOperation struct {
	Command *GetIndexesStatisticsCommand
}

func NewGetIndexesStatisticsOperation() *GetIndexesStatisticsOperation {
	return &GetIndexesStatisticsOperation{}
}

func (o *GetIndexesStatisticsOperation) GetCommand(conventions *DocumentConventions) (RavenCommand, error) {
	o.Command = NewGetIndexesStatisticsCommand()
	return o.Command, nil
}

var (
	_ RavenCommand = &GetIndexesStatisticsCommand{}
)

type GetIndexesStatisticsCommand struct {
	RavenCommandBase

	Result []*IndexStats
}

func NewGetIndexesStatisticsCommand() *GetIndexesStatisticsCommand {

	res := &GetIndexesStatisticsCommand{
		RavenCommandBase: NewRavenCommandBase(),
	}
	res.IsReadRequest = true
	return res
}

func (c *GetIndexesStatisticsCommand) CreateRequest(node *ServerNode) (*http.Request, error) {
	url := node.URL + "/databases/" + node.Database + "/indexes/stats"

	return NewHttpGet(url)
}

func (c *GetIndexesStatisticsCommand) SetResponse(response []byte, fromCache bool) error {
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
	c.Result = res.Results
	return nil
}
