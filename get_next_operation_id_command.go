package ravendb

import (
	"encoding/json"
	"net/http"
)

type _GetNextOperationIdCommandResponse struct {
	ID int `json:"Id"`
}

func NewGetNextOperationIdCommand() *RavenCommand {
	cmd := NewRavenCommand()
	cmd.createRequestFunc = GetNextOperationIdCommand_createRequest
	cmd.setResponseFunc = GetNextOperationIdCommand_setResponse
	return cmd
}

func GetNextOperationIdCommand_createRequest(cmd *RavenCommand, node *ServerNode) (*http.Request, error) {
	url := node.getUrl() + "/databases/" + node.getDatabase() + "/operations/next-operation-id"
	return NewHttpGet(url)
}

func GetNextOperationIdCommand_setResponse(cmd *RavenCommand, response String, fromCache bool) error {
	var res _GetNextOperationIdCommandResponse
	err := json.Unmarshal([]byte(response), &res)
	if err != nil {
		return err
	}
	cmd.result = res.ID
	return nil
}
