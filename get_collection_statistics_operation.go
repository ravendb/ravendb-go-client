package ravendb

import (
	"encoding/json"
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

func (o *GetCollectionStatisticsOperation) getCommand(conventions *DocumentConventions) RavenCommand {
	o.Command = NewGetCollectionStatisticsCommand()
	return o.Command
}

var _ RavenCommand = &GetCollectionStatisticsCommand{}

type GetCollectionStatisticsCommand struct {
	*RavenCommandBase

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
	url := node.getUrl() + "/databases/" + node.getDatabase() + "/collections/stats"
	return NewHttpGet(url)
}

func (c *GetCollectionStatisticsCommand) setResponse(response string, fromCache bool) error {
	if response == "" {
		return throwInvalidResponse()
	}

	var res CollectionStatistics
	err := json.Unmarshal([]byte(response), &res)
	if err != nil {
		return err
	}
	c.Result = &res
	return nil
}
