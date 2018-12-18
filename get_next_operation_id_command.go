package ravendb

import (
	"net/http"
)

var (
	_ RavenCommand = &GetNextOperationIDCommand{}
)

type _GetNextOperationIDCommandResponse struct {
	ID int `json:"Id"`
}

// GetNextOperationIDCommand represents command for getting next
// id from the server
type GetNextOperationIDCommand struct {
	RavenCommandBase

	Result int
}

// NewGetNextOperationIDCommand returns GetNextOperationIDCommand
func NewGetNextOperationIDCommand() *GetNextOperationIDCommand {
	cmd := &GetNextOperationIDCommand{
		RavenCommandBase: NewRavenCommandBase(),
	}
	return cmd
}

// CreateRequest creates a new request
func (c *GetNextOperationIDCommand) CreateRequest(node *ServerNode) (*http.Request, error) {
	url := node.GetUrl() + "/databases/" + node.GetDatabase() + "/operations/next-operation-id"
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
