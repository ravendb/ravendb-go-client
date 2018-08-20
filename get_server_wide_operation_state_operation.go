package ravendb

import (
	"encoding/json"
	"net/http"
	"strconv"
)

type GetServerWideOperationStateOperation struct {
	_id int
}

func (o *GetServerWideOperationStateOperation) GetCommand(conventions *DocumentConventions) *GetServerWideOperationStateCommand {
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

func (c *GetServerWideOperationStateCommand) CreateRequest(node *ServerNode) (*http.Request, error) {
	url := node.GetUrl() + "/operations/state?id=" + strconv.Itoa(c._id)
	return NewHttpGet(url)
}

func (c *GetServerWideOperationStateCommand) SetResponse(response []byte, fromCache bool) error {
	if len(response) == 0 {
		return nil
	}

	return json.Unmarshal(response, &c.Result)
}
