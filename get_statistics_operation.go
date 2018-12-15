package ravendb

import (
	"net/http"
)

var (
	_ IMaintenanceOperation = &GetStatisticsOperation{}
)

type GetStatisticsOperation struct {
	_debugTag string

	Command *GetStatisticsCommand
}

func NewGetStatisticsOperation() *GetStatisticsOperation {
	return NewGetStatisticsOperationWithDebugTag("")
}

func NewGetStatisticsOperationWithDebugTag(debugTag string) *GetStatisticsOperation {
	return &GetStatisticsOperation{
		_debugTag: debugTag,
	}
}

func (o *GetStatisticsOperation) GetCommand(conventions *DocumentConventions) RavenCommand {
	o.Command = NewGetStatisticsCommandWithDebugTag(o._debugTag)
	return o.Command
}

var (
	_ RavenCommand = &GetStatisticsCommand{}
)

type GetStatisticsCommand struct {
	RavenCommandBase

	debugTag string

	Result *DatabaseStatistics
}

func NewGetStatisticsCommand() *GetStatisticsCommand {
	return NewGetStatisticsCommandWithDebugTag("")
}

func NewGetStatisticsCommandWithDebugTag(debugTag string) *GetStatisticsCommand {
	cmd := &GetStatisticsCommand{
		RavenCommandBase: NewRavenCommandBase(),

		debugTag: debugTag,
	}
	cmd.IsReadRequest = true
	return cmd
}

func (c *GetStatisticsCommand) CreateRequest(node *ServerNode) (*http.Request, error) {
	url := node.GetUrl() + "/databases/" + node.GetDatabase() + "/stats"
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
