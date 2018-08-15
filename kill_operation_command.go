package ravendb

import "net/http"

var (
	_ RavenCommand = &KillOperationCommand{}
)

type KillOperationCommand struct {
	*RavenCommandBase

	_id string
}

func NewKillOperationCommand(id string) *KillOperationCommand {
	panicIf(id == "", "id cannot be empty")
	cmd := &KillOperationCommand{
		RavenCommandBase: NewRavenCommandBase(),

		_id: id,
	}
	cmd.responseType = RavenCommandResponseType_EMPTY

	return cmd
}

func (c *KillOperationCommand) CreateRequest(node *ServerNode) (*http.Request, error) {
	url := node.getUrl() + "/databases/" + node.getDatabase() + "/operations/kill?id=" + c._id

	return NewHttpPost(url, nil)
}
