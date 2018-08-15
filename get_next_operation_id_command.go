package ravendb

import (
	"encoding/json"
	"net/http"
)

var (
	_ RavenCommand = &GetNextOperationIdCommand{}
)

type _GetNextOperationIdCommandResponse struct {
	ID int `json:"Id"`
}

type GetNextOperationIdCommand struct {
	*RavenCommandBase

	Result int
}

func NewGetNextOperationIdCommand() *GetNextOperationIdCommand {
	cmd := &GetNextOperationIdCommand{
		RavenCommandBase: NewRavenCommandBase(),
	}
	return cmd
}

func (c *GetNextOperationIdCommand) CreateRequest(node *ServerNode) (*http.Request, error) {
	url := node.GetUrl() + "/databases/" + node.GetDatabase() + "/operations/next-operation-id"
	return NewHttpGet(url)
}

func (c *GetNextOperationIdCommand) SetResponse(response []byte, fromCache bool) error {
	var res _GetNextOperationIdCommandResponse
	err := json.Unmarshal(response, &res)
	if err != nil {
		return err
	}
	c.Result = res.ID
	return nil
}
