package ravendb

import (
	"net/http"
	"strconv"
)

type GetServerWideOperationStateOperation struct {
	_id int64
}

func (o *GetServerWideOperationStateOperation) GetCommand(conventions *DocumentConventions) *GetServerWideOperationStateCommand {
	return NewGetServerWideOperationStateCommand(getDefaultConventions(), o._id)
}

type GetServerWideOperationStateCommand struct {
	RavenCommandBase

	_conventions *DocumentConventions
	_id          int64

	Result map[string]interface{}
}

func NewGetServerWideOperationStateCommand(conventions *DocumentConventions, id int64) *GetServerWideOperationStateCommand {
	cmd := &GetServerWideOperationStateCommand{
		RavenCommandBase: NewRavenCommandBase(),

		_conventions: conventions,
		_id:          id,
	}
	cmd.IsReadRequest = true

	return cmd
}

func (c *GetServerWideOperationStateCommand) CreateRequest(node *ServerNode) (*http.Request, error) {
	url := node.URL + "/operations/state?id=" + strconv.FormatInt(c._id, 10)
	return NewHttpGet(url)
}

func (c *GetServerWideOperationStateCommand) SetResponse(response []byte, fromCache bool) error {
	if len(response) == 0 {
		return nil
	}

	return jsonUnmarshal(response, &c.Result)
}
