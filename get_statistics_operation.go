package ravendb

import (
	"encoding/json"
	"net/http"
)

type GetStatisticsOperation struct {
	_debugTag String
}

func NewGetStatisticsOperation() *GetStatisticsOperation {
	return NewGetStatisticsOperationWithDebugTag("")
}

func NewGetStatisticsOperationWithDebugTag(debugTag string) *GetStatisticsOperation {
	return &GetStatisticsOperation{
		_debugTag: debugTag,
	}
}

func (s *GetStatisticsOperation) getCommand(conventions *DocumentConventions) *RavenCommand {
	return NewGetStatisticsCommandWithDebugTag(s._debugTag)
}

type GetStatisticsCommandData struct {
	debugTag String
}

func GetStatisticsCommand_createRequest(cmd *RavenCommand, node *ServerNode) (*http.Request, string) {
	url := node.getUrl() + "/databases/" + node.getDatabase() + "/stats"
	data := cmd.data.(*GetStatisticsCommandData)
	if data.debugTag != "" {
		url += "?" + data.debugTag
	}

	return NewHttpGet(), url
}

func GetStatisticsCommand_setResponse(cmd *RavenCommand, response String, fromCache bool) error {
	if response == "" {
		return throwInvalidResponse()
	}
	var res DatabaseStatistics
	err := json.Unmarshal([]byte(response), &res)
	if err != nil {
		return err
	}
	cmd.result = &res
	return nil
}

func NewGetStatisticsCommand() *RavenCommand {
	return NewGetStatisticsCommandWithDebugTag("")
}

func NewGetStatisticsCommandWithDebugTag(debugTag string) *RavenCommand {
	data := &GetStatisticsCommandData{
		debugTag: debugTag,
	}
	cmd := NewRavenCommand()
	cmd.data = data
	cmd.IsReadRequest = true
	cmd.createRequestFunc = GetStatisticsCommand_createRequest
	cmd.setResponseFunc = GetStatisticsCommand_setResponse
	return cmd
}
