package ravendb

import (
	"encoding/json"
	"net/http"
)

var (
	_ RavenCommand = &CreateSubscriptionCommand{}
)

// CreateSubscriptionCommand represents "create subscription" command
type CreateSubscriptionCommand struct {
	RavenCommandBase

	conventions *DocumentConventions
	options     *SubscriptionCreationOptions
	id          string

	Result *CreateSubscriptionResult
}

func newCreateSubscriptionCommand(conventions *DocumentConventions, options *SubscriptionCreationOptions, id string) *CreateSubscriptionCommand {
	return &CreateSubscriptionCommand{
		RavenCommandBase: NewRavenCommandBase(),

		conventions: conventions,
		options:     options,
		id:          id,
	}
}

func (c *CreateSubscriptionCommand) createRequest(node *ServerNode) (*http.Request, error) {
	uri := node.URL + "/databases/" + node.Database + "/subscriptions"

	if c.id != "" {
		uri += "?id=" + c.id
	}

	d, err := json.Marshal(c.options)
	if err != nil {
		return nil, err
	}

	return NewHttpPut(uri, d)
}

// SetResponse sets result based on http response
func (c *CreateSubscriptionCommand) SetResponse(response []byte, fromCache bool) error {
	if len(response) == 0 {
		return throwInvalidResponse()
	}
	return jsonUnmarshal(response, &c.Result)
}
