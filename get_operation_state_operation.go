package ravendb

import (
	"net/http"
)

type GetOperationStateOperation struct {
	id int64
}

func (o *GetOperationStateOperation) GetCommand(conventions *DocumentConventions) *GetOperationStateCommand {
	return NewGetOperationStateCommand(getDefaultConventions(), o.id)
}

type GetOperationStateCommand struct {
	RavenCommandBase

	conventions *DocumentConventions
	id          int64

	Result map[string]interface{}
}

func NewGetOperationStateCommand(conventions *DocumentConventions, id int64) *GetOperationStateCommand {
	cmd := &GetOperationStateCommand{
		RavenCommandBase: NewRavenCommandBase(),

		conventions: conventions,
		id:          id,
	}
	cmd.IsReadRequest = true

	return cmd
}

func (c *GetOperationStateCommand) CreateRequest(node *ServerNode) (*http.Request, error) {
	url := node.URL + "/databases/" + node.Database + "/operations/state?id=" + i64toa(c.id)
	return NewHttpGet(url)
}

func (c *GetOperationStateCommand) SetResponse(response []byte, fromCache bool) error {
	if len(response) == 0 {
		return nil
	}

	return jsonUnmarshal(response, &c.Result)
}
