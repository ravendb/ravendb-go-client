package ravendb

import (
	"net/http"
)

var (
	_ RavenCommand = &DeleteSubscriptionCommand{}
)

// DeleteSubscriptionCommand describes "delete subscription" command
type DeleteSubscriptionCommand struct {
	RavenCommandBase

	name string
}

func newDeleteSubscriptionCommand(name string) *DeleteSubscriptionCommand {
	cmd := &DeleteSubscriptionCommand{
		RavenCommandBase: NewRavenCommandBase(),

		name: name,
	}
	cmd.ResponseType = RavenCommandResponseTypeEmpty
	return cmd
}

func (c *DeleteSubscriptionCommand) CreateRequest(node *ServerNode) (*http.Request, error) {
	url := node.URL + "/databases/" + node.Database + "/subscriptions?taskName=" + urlUtilsEscapeDataString(c.name)

	return newHttpDelete(url, nil)
}
