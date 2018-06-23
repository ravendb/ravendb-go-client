package ravendb

import (
	"encoding/json"
	"net/http"
	"strconv"
)

type GetOperationStateOperation struct {
	_id int
}

func (o *GetOperationStateOperation) getCommand(conventions *DocumentConventions) *GetOperationStateCommand {
	return NewGetOperationStateCommand(DocumentConventions_defaultConventions(), o._id)
}

type GetOperationStateCommand struct {
	*RavenCommandBase

	_conventions *DocumentConventions
	_id          int

	Result ObjectNode
}

func NewGetOperationStateCommand(conventions *DocumentConventions, id int) *GetOperationStateCommand {
	cmd := &GetOperationStateCommand{
		RavenCommandBase: NewRavenCommandBase(),

		_conventions: conventions,
		_id:          id,
	}
	cmd.IsReadRequest = true

	return cmd
}

func (c *GetOperationStateCommand) createRequest(node *ServerNode) (*http.Request, error) {
	url := node.getUrl() + "/databases/" + node.getDatabase() + "/operations/state?id=" + strconv.Itoa(c._id)
	return NewHttpGet(url)
}

func (c *GetOperationStateCommand) setResponse(response []byte, fromCache bool) error {
	if len(response) == 0 {
		return nil
	}

	var res ObjectNode
	err := json.Unmarshal(response, &res)
	if err != nil {
		return err
	}
	c.Result = res
	return nil
}
