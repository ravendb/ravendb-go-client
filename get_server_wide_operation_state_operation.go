package ravendb

import (
	"net/http"
)

type GetServerWideOperationStateOperation struct {
	id int64
}

func (o *GetServerWideOperationStateOperation) GetCommand(conventions *DocumentConventions) *GetServerWideOperationStateCommand {
	return NewGetServerWideOperationStateCommand(getDefaultConventions(), o.id)
}

type GetServerWideOperationStateCommand struct {
	RavenCommandBase

	conventions *DocumentConventions
	id          int64

	Result map[string]interface{}
}

func NewGetServerWideOperationStateCommand(conventions *DocumentConventions, id int64) *GetServerWideOperationStateCommand {
	cmd := &GetServerWideOperationStateCommand{
		RavenCommandBase: NewRavenCommandBase(),

		conventions: conventions,
		id:          id,
	}
	cmd.IsReadRequest = true

	return cmd
}

func (c *GetServerWideOperationStateCommand) createRequest(node *ServerNode) (*http.Request, error) {
	url := node.URL + "/operations/state?id=" + i64toa(c.id)
	return NewHttpGet(url)
}

func (c *GetServerWideOperationStateCommand) setResponse(response []byte, fromCache bool) error {
	if len(response) == 0 {
		return nil
	}

	return jsonUnmarshal(response, &c.Result)
}
