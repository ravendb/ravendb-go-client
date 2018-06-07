package ravendb

import (
	"encoding/json"
	"net/http"
	"strconv"
)

type GetOperationStateOperation struct {
	_id int
}

func (o *GetOperationStateOperation) getCommand(conventions *DocumentConventions) *RavenCommand {
	return NewGetOperationStateCommand(DocumentConventions_defaultConventions, o._id)
}

type _GetOperationStateCommand struct {
	_conventions *DocumentConventions
	_id          int
}

func NewGetOperationStateCommand(conventions *DocumentConventions, id int) *RavenCommand {
	data := &_GetOperationStateCommand{
		_conventions: conventions,
		_id:          id,
	}

	cmd := NewRavenCommand()
	cmd.IsReadRequest = true
	cmd.data = data
	cmd.createRequestFunc = GetOperationStateCommand_createRequest
	cmd.setResponseFunc = GetOperationStateCommand_setResponse

	return cmd
}

func GetOperationStateCommand_createRequest(cmd *RavenCommand, node *ServerNode) (*http.Request, error) {
	data := cmd.data.(*_GetOperationStateCommand)
	url := node.getUrl() + "/databases/" + node.getDatabase() + "/operations/state?id=" + strconv.Itoa(data._id)
	return NewHttpGet(url)
}

func GetOperationStateCommand_setResponse(cmd *RavenCommand, response String, fromCache bool) error {
	if response == "" {
		return nil
	}

	var res ObjectNode
	err := json.Unmarshal([]byte(response), &res)
	if err != nil {
		return err
	}
	cmd.result = &res
	return nil
}
