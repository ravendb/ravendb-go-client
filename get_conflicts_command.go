package ravendb

import (
	"net/http"
)

var (
	_ RavenCommand = &GetConflictsCommand{}
)

type GetConflictsCommand struct {
	RavenCommandBase

	_id string

	Result *GetConflictsResult
}

func NewGetConflictsCommand(id string) *GetConflictsCommand {
	panicIf(id == "", "id cannot be empty")
	cmd := &GetConflictsCommand{
		RavenCommandBase: NewRavenCommandBase(),

		_id: id,
	}
	cmd.IsReadRequest = true

	return cmd
}

func (c *GetConflictsCommand) CreateRequest(node *ServerNode) (*http.Request, error) {
	url := node.URL + "/databases/" + node.Database + "/replication/conflicts?docId=" + c._id

	return newHttpGet(url)
}

func (c *GetConflictsCommand) SetResponse(response []byte, fromCache bool) error {
	if len(response) == 0 {
		return throwInvalidResponse()
	}
	return jsonUnmarshal(response, &c.Result)
}
