package ravendb

import (
	"encoding/json"
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

func (o *GetStatisticsOperation) getCommand(conventions *DocumentConventions) RavenCommand {
	o.Command = NewGetStatisticsCommandWithDebugTag(o._debugTag)
	return o.Command
}

var (
	_ RavenCommand = &GetStatisticsCommand{}
)

type GetStatisticsCommand struct {
	*RavenCommandBase

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

func (c *GetStatisticsCommand) createRequest(node *ServerNode) (*http.Request, error) {
	url := node.getUrl() + "/databases/" + node.getDatabase() + "/stats"
	if c.debugTag != "" {
		url += "?" + c.debugTag
	}

	return NewHttpGet(url)
}

func (c *GetStatisticsCommand) setResponse(response string, fromCache bool) error {
	if response == "" {
		return throwInvalidResponse()
	}
	var res DatabaseStatistics
	err := json.Unmarshal([]byte(response), &res)
	if err != nil {
		return err
	}
	c.Result = &res
	return nil
}
