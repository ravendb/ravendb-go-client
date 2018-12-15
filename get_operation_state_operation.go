package ravendb

import (
	"net/http"
	"strconv"
)

type GetOperationStateOperation struct {
	_id int
}

func (o *GetOperationStateOperation) GetCommand(conventions *DocumentConventions) *GetOperationStateCommand {
	return NewGetOperationStateCommand(DocumentConventions_defaultConventions(), o._id)
}

type GetOperationStateCommand struct {
	RavenCommandBase

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

func (c *GetOperationStateCommand) CreateRequest(node *ServerNode) (*http.Request, error) {
	url := node.GetUrl() + "/databases/" + node.GetDatabase() + "/operations/state?id=" + strconv.Itoa(c._id)
	return NewHttpGet(url)
}

func (c *GetOperationStateCommand) SetResponse(response []byte, fromCache bool) error {
	if len(response) == 0 {
		return nil
	}

	return jsonUnmarshal(response, &c.Result)
}
