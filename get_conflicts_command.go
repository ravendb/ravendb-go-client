package ravendb

import (
	"encoding/json"
	"net/http"
)

var (
	_ RavenCommand = &GetConflictsCommand{}
)

type GetConflictsCommand struct {
	*RavenCommandBase

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

func (c *GetConflictsCommand) createRequest(node *ServerNode) (*http.Request, error) {
	url := node.getUrl() + "/databases/" + node.getDatabase() + "/replication/conflicts?docId=" + c._id

	return NewHttpGet(url)
}

func (c *GetConflictsCommand) setResponse(response string, fromCache bool) error {
	if response == "" {
		return throwInvalidResponse()
	}
	var res GetConflictsResult
	err := json.Unmarshal([]byte(response), &res)
	if err != nil {
		return err
	}
	c.Result = &res
	return nil
}
