package ravendb

import (
	"encoding/json"
	"net/http"
	"strconv"
)

type GetServerWideOperationStateOperation struct {
	_id int
}

func (o *GetServerWideOperationStateOperation) getCommand(conventions *DocumentConventions) *GetServerWideOperationStateCommand {
	return NewGetServerWideOperationStateCommand(DocumentConventions_defaultConventions(), o._id)
}

type GetServerWideOperationStateCommand struct {
	*RavenCommandBase

	_conventions *DocumentConventions
	_id          int

	Result ObjectNode
}

func NewGetServerWideOperationStateCommand(conventions *DocumentConventions, id int) *GetServerWideOperationStateCommand {
	cmd := &GetServerWideOperationStateCommand{
		RavenCommandBase: NewRavenCommandBase(),

		_conventions: conventions,
		_id:          id,
	}
	cmd.IsReadRequest = true

	return cmd
}

func (c *GetServerWideOperationStateCommand) createRequest(node *ServerNode) (*http.Request, error) {
	url := node.getUrl() + "/operations/state?id=" + strconv.Itoa(c._id)
	return NewHttpGet(url)
}

func (c *GetServerWideOperationStateCommand) setResponse(response []byte, fromCache bool) error {
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
