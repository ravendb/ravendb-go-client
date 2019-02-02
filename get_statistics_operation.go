package ravendb

import (
	"net/http"
)

var (
	_ IMaintenanceOperation = &GetStatisticsOperation{}
)

type GetStatisticsOperation struct {
	debugTag string

	Command *GetStatisticsCommand
}

func NewGetStatisticsOperation() *GetStatisticsOperation {
	return NewGetStatisticsOperationWithDebugTag("")
}

func NewGetStatisticsOperationWithDebugTag(debugTag string) *GetStatisticsOperation {
	return &GetStatisticsOperation{
		debugTag: debugTag,
	}
}

func (o *GetStatisticsOperation) GetCommand(conventions *DocumentConventions) (RavenCommand, error) {
	o.Command = NewGetStatisticsCommand(o.debugTag)
	return o.Command, nil
}

var (
	_ RavenCommand = &GetStatisticsCommand{}
)

type GetStatisticsCommand struct {
	RavenCommandBase

	debugTag string

	Result *DatabaseStatistics
}

func NewGetStatisticsCommand(debugTag string) *GetStatisticsCommand {
	cmd := &GetStatisticsCommand{
		RavenCommandBase: NewRavenCommandBase(),

		debugTag: debugTag,
	}
	cmd.IsReadRequest = true
	return cmd
}

func (c *GetStatisticsCommand) CreateRequest(node *ServerNode) (*http.Request, error) {
	url := node.URL + "/databases/" + node.Database + "/stats"
	if c.debugTag != "" {
		url += "?" + c.debugTag
	}

	return NewHttpGet(url)
}

func (c *GetStatisticsCommand) SetResponse(response []byte, fromCache bool) error {
	if len(response) == 0 {
		return throwInvalidResponse()
	}

	return jsonUnmarshal(response, &c.Result)
}
