package ravendb

import "net/http"

var (
	_ RavenCommand = &KillOperationCommand{}
)

// KillOperationCommand represents "kill operation" command
type KillOperationCommand struct {
	RavenCommandBase

	id string
}

// NewKillOperationCommand returns new KillOperationCommand
func NewKillOperationCommand(id string) (*KillOperationCommand, error) {
	if id == "" {
		return nil, newIllegalArgumentError("id cannot be empty")
	}
	cmd := &KillOperationCommand{
		RavenCommandBase: NewRavenCommandBase(),

		id: id,
	}
	cmd.ResponseType = RavenCommandResponseTypeEmpty

	return cmd, nil
}

func (c *KillOperationCommand) CreateRequest(node *ServerNode) (*http.Request, error) {
	url := node.URL + "/databases/" + node.Database + "/operations/kill?id=" + c.id

	return NewHttpPost(url, nil)
}
