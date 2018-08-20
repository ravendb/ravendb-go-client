package ravendb

import (
	"encoding/json"
	"net/http"
	"strconv"
)

var (
	_ RavenCommand = &SeedIdentityForCommand{}
)

type SeedIdentityForCommand struct {
	RavenCommandBase

	_id     string
	_value  int
	_forced bool

	Result int
}

func NewSeedIdentityForCommand(id string, value int) *SeedIdentityForCommand {
	return NewSeedIdentityForCommandWithForced(id, value, false)
}

func NewSeedIdentityForCommandWithForced(id string, value int, forced bool) *SeedIdentityForCommand {
	panicIf(id == "", "Id cannot be null")

	res := &SeedIdentityForCommand{
		RavenCommandBase: NewRavenCommandBase(),

		_id:     id,
		_value:  value,
		_forced: forced,
	}
	return res
}

func (c *SeedIdentityForCommand) CreateRequest(node *ServerNode) (*http.Request, error) {
	err := ensureIsNotNullOrString(c._id, "ID")
	if err != nil {
		return nil, err
	}

	url := node.GetUrl() + "/databases/" + node.GetDatabase() + "/identity/seed?name=" + urlEncode(c._id) + "&value=" + strconv.Itoa(c._value)

	if c._forced {
		url += "&force=true"
	}

	return NewHttpPost(url, nil)
}

func (c *SeedIdentityForCommand) SetResponse(response []byte, fromCache bool) error {
	if len(response) == 0 {
		return throwInvalidResponse()
	}

	var jsonNode map[string]interface{}
	err := json.Unmarshal(response, &jsonNode)
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
