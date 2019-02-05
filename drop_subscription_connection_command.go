package ravendb

import (
	"net/http"
)

var (
	_ RavenCommand = &DropSubscriptionConnectionCommand{}
)

// DropSubscriptionConnectionCommand describes "drop subscription" command
type DropSubscriptionConnectionCommand struct {
	RavenCommandBase

	name string
}

func newDropSubscriptionConnectionCommand(name string) *DropSubscriptionConnectionCommand {
	cmd := &DropSubscriptionConnectionCommand{
		RavenCommandBase: NewRavenCommandBase(),

		name: name,
	}
	cmd.ResponseType = RavenCommandResponseTypeEmpty
	return cmd
}

func (c *DropSubscriptionConnectionCommand) createRequest(node *ServerNode) (*http.Request, error) {
	url := node.URL + "/databases/" + node.Database + "/subscriptions/drop?name=" + urlUtilsEscapeDataString(c.name)

	return NewHttpPost(url, nil)
}
