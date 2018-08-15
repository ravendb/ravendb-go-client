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

func (c *GetStatisticsCommand) CreateRequest(node *ServerNode) (*http.Request, error) {
	url := node.getUrl() + "/databases/" + node.getDatabase() + "/stats"
	if c.debugTag != "" {
		url += "?" + c.debugTag
	}

	return NewHttpGet(url)
}

func (c *GetStatisticsCommand) SetResponse(response []byte, fromCache bool) error {
	if len(response) == 0 {
		return throwInvalidResponse()
	}
	//dbg("GetStatisticsCommand: JSON:\n%s\n\n", string(maybePrettyPrintJSON(response)))
	var res DatabaseStatistics
	err := json.Unmarshal(response, &res)
	if err != nil {
		return err
	}
	c.Result = &res
	return nil
}
