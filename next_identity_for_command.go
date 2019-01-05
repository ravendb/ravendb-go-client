package ravendb

import (
	"net/http"
)

var (
	_ RavenCommand = &NextIdentityForCommand{}
)

type NextIdentityForCommand struct {
	RavenCommandBase

	_id string

	Result int
}

func NewNextIdentityForCommand(id string) *NextIdentityForCommand {
	res := &NextIdentityForCommand{
		RavenCommandBase: NewRavenCommandBase(),

		_id: id,
	}
	panicIf(id == "", "Id cannot be null")

	return res
}

func (c *NextIdentityForCommand) CreateRequest(node *ServerNode) (*http.Request, error) {
	err := ensureIsNotNullOrString(c._id, "ID")
	if err != nil {
		return nil, err
	}

	url := node.URL + "/databases/" + node.Database + "/identity/next?name=" + urlEncode(c._id)

	return NewHttpPost(url, nil)
}

func (c *NextIdentityForCommand) SetResponse(response []byte, fromCache bool) error {
	if len(response) == 0 {
		return throwInvalidResponse()
	}
	var jsonNode map[string]interface{}
	err := jsonUnmarshal(response, &jsonNode)
	if err != nil {
		return err
	}
	n, ok := jsonGetAsInt(jsonNode, "NewIdentityValue")
	if !ok {
		return throwInvalidResponse()
	}
	c.Result = n
	return nil
}
