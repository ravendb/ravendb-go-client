package ravendb

import (
	"net/http"
)

var (
	_ RavenCommand = &GetNextOperationIDCommand{}
)

type _GetNextOperationIDCommandResponse struct {
	ID int64 `json:"Id"`
}

// GetNextOperationIDCommand represents command for getting next
// id from the server
type GetNextOperationIDCommand struct {
	RavenCommandBase

	Result int64
}

// NewGetNextOperationIDCommand returns GetNextOperationIDCommand
func NewGetNextOperationIDCommand() *GetNextOperationIDCommand {
	cmd := &GetNextOperationIDCommand{
		RavenCommandBase: NewRavenCommandBase(),
	}
	return cmd
}

func (c *GetNextOperationIDCommand) createRequest(node *ServerNode) (*http.Request, error) {
	url := node.URL + "/databases/" + node.Database + "/operations/next-operation-id"
	return NewHttpGet(url)
}

// SetResponse sets JSON response
func (c *GetNextOperationIDCommand) SetResponse(response []byte, fromCache bool) error {
	var res _GetNextOperationIDCommandResponse
	err := jsonUnmarshal(response, &res)
	if err != nil {
		return err
	}
	c.Result = res.ID
	return nil
}
