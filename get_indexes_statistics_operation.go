package ravendb

import (
	"encoding/json"
	"net/http"
)

var _ IMaintenanceOperation = &GetIndexesStatisticsOperation{}

type GetIndexesStatisticsOperation struct {
	Command *GetIndexesStatisticsCommand
}

func NewGetIndexesStatisticsOperationWithPageSize() *GetIndexesStatisticsOperation {
	return &GetIndexesStatisticsOperation{}
}

func (o *GetIndexesStatisticsOperation) getCommand(conventions *DocumentConventions) RavenCommand {
	o.Command = NewGetIndexesStatisticsCommand()
	return o.Command
}

var (
	_ RavenCommand = &GetIndexesStatisticsCommand{}
)

type GetIndexesStatisticsCommand struct {
	*RavenCommandBase

	Result []*IndexStats
}

func NewGetIndexesStatisticsCommand() *GetIndexesStatisticsCommand {

	res := &GetIndexesStatisticsCommand{
		RavenCommandBase: NewRavenCommandBase(),
	}
	res.IsReadRequest = true
	return res
}

func (c *GetIndexesStatisticsCommand) createRequest(node *ServerNode) (*http.Request, error) {
	url := node.getUrl() + "/databases/" + node.getDatabase() + "/indexes/stats"

	return NewHttpGet(url)
}

func (c *GetIndexesStatisticsCommand) setResponse(response []byte, fromCache bool) error {
	if response == nil {
		return throwInvalidResponse()
	}

	var res []*IndexStats
	err := json.Unmarshal(response, &res)
	if err != nil {
		return err
	}
	c.Result = res
	return nil
}
