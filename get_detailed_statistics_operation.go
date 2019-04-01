package ravendb

import (
	"net/http"
)

var _ IMaintenanceOperation = &GetDetailedStatisticsOperation{}

type GetDetailedStatisticsOperation struct {
	_debugTag string

	Command *GetDetailedStatisticsCommand
}

func NewGetDetailedStatisticsOperation(debugTag string) *GetDetailedStatisticsOperation {
	return &GetDetailedStatisticsOperation{
		_debugTag: debugTag,
	}
}

func (o *GetDetailedStatisticsOperation) GetCommand(conventions *DocumentConventions) (RavenCommand, error) {
	o.Command = NewGetDetailedStatisticsCommand(o._debugTag)
	return o.Command, nil
}

var (
	_ RavenCommand = &GetDetailedStatisticsCommand{}
)

type GetDetailedStatisticsCommand struct {
	RavenCommandBase

	_debugTag string

	Result *DetailedDatabaseStatistics
}

func NewGetDetailedStatisticsCommand(debugTag string) *GetDetailedStatisticsCommand {
	res := &GetDetailedStatisticsCommand{
		RavenCommandBase: NewRavenCommandBase(),
		_debugTag:        debugTag,
	}
	res.IsReadRequest = true
	return res
}

func (c *GetDetailedStatisticsCommand) createRequest(node *ServerNode) (*http.Request, error) {
	url := node.URL + "/databases/" + node.Database + "/stats/detailed"
	if c._debugTag != "" {
		url += "?" + c._debugTag
	}

	return newHttpGet(url)
}

func (c *GetDetailedStatisticsCommand) setResponse(response []byte, fromCache bool) error {
	return jsonUnmarshal(response, &c.Result)
}
