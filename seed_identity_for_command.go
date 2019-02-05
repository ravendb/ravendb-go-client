package ravendb

import (
	"net/http"
)

var (
	_ RavenCommand = &SeedIdentityForCommand{}
)

type SeedIdentityForCommand struct {
	RavenCommandBase

	id     string
	value  int64
	forced bool

	Result int
}

func NewSeedIdentityForCommand(id string, value int64, forced bool) (*SeedIdentityForCommand, error) {
	if id == "" {
		return nil, newIllegalArgumentError("Id cannot be null")
	}

	res := &SeedIdentityForCommand{
		RavenCommandBase: NewRavenCommandBase(),

		id:     id,
		value:  value,
		forced: forced,
	}
	return res, nil
}

func (c *SeedIdentityForCommand) createRequest(node *ServerNode) (*http.Request, error) {
	err := ensureIsNotNullOrString(c.id, "ID")
	if err != nil {
		return nil, err
	}

	url := node.URL + "/databases/" + node.Database + "/identity/seed?name=" + urlEncode(c.id) + "&value=" + i64toa(c.value)

	if c.forced {
		url += "&force=true"
	}

	return NewHttpPost(url, nil)
}

func (c *SeedIdentityForCommand) SetResponse(response []byte, fromCache bool) error {
	if len(response) == 0 {
		return throwInvalidResponse()
	}

	var jsonNode map[string]interface{}
	err := jsonUnmarshal(response, &jsonNode)
	if err != nil {
		return err
	}
	n, ok := jsonGetAsInt(jsonNode, "NewSeedValue")
	if !ok {
		return throwInvalidResponse()
	}
	c.Result = n
	return nil
}
