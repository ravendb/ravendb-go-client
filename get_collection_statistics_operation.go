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

func (o *GetCollectionStatisticsOperation) GetCommand(conventions *DocumentConventions) (RavenCommand, error) {
	o.Command = NewGetCollectionStatisticsCommand()
	return o.Command, nil
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

func (c *GetCollectionStatisticsCommand) createRequest(node *ServerNode) (*http.Request, error) {
	url := node.URL + "/databases/" + node.Database + "/collections/stats"
	return newHttpGet(url)
}

func (c *GetCollectionStatisticsCommand) setResponse(response []byte, fromCache bool) error {
	if len(response) == 0 {
		return throwInvalidResponse()
	}

	return jsonUnmarshal(response, &c.Result)
}
