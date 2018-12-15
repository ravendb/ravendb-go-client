package ravendb

import (
	"net/http"
)

var (
	_ IMaintenanceOperation = &GetCollectionStatisticsOperation{}
)

type GetCollectionStatisticsOperation struct {
	Command *GetCollectionStatisticsCommand
}

func NewGetCollectionStatisticsOperation() *GetCollectionStatisticsOperation {
	return &GetCollectionStatisticsOperation{}
}

func (o *GetCollectionStatisticsOperation) GetCommand(conventions *DocumentConventions) RavenCommand {
	o.Command = NewGetCollectionStatisticsCommand()
	return o.Command
}

var _ RavenCommand = &GetCollectionStatisticsCommand{}

type GetCollectionStatisticsCommand struct {
	RavenCommandBase

	Result *CollectionStatistics
}

func NewGetCollectionStatisticsCommand() *GetCollectionStatisticsCommand {
	cmd := &GetCollectionStatisticsCommand{
		RavenCommandBase: NewRavenCommandBase(),
	}
	cmd.IsReadRequest = true
	return cmd
}

func (c *GetCollectionStatisticsCommand) CreateRequest(node *ServerNode) (*http.Request, error) {
	url := node.GetUrl() + "/databases/" + node.GetDatabase() + "/collections/stats"
	return NewHttpGet(url)
}

func (c *GetCollectionStatisticsCommand) SetResponse(response []byte, fromCache bool) error {
	if len(response) == 0 {
		return throwInvalidResponse()
	}

	return jsonUnmarshal(response, &c.Result)
}
