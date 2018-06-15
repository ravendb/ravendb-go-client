package ravendb

import (
	"encoding/json"
	"net/http"
)

var (
	_ IServerOperation = &CompactDatabaseOperation{}
)

type CompactDatabaseOperation struct {
	_compactSettings *CompactSettings

	Command *CompactDatabaseCommand
}

func NewCompactDatabaseOperation(compactSettings *CompactSettings) *CompactDatabaseOperation {
	return &CompactDatabaseOperation{
		_compactSettings: compactSettings,
	}
}

func (o *CompactDatabaseOperation) getCommand(conventions *DocumentConventions) RavenCommand {
	o.Command = NewCompactDatabaseCommand(conventions, o._compactSettings)
	return o.Command
}

var _ RavenCommand = &CompactDatabaseCommand{}

type CompactDatabaseCommand struct {
	*RavenCommandBase

	_compactSettings []byte // CompactSettings serialized to json

	Result *OperationIdResult
}

func NewCompactDatabaseCommand(conventions *DocumentConventions, compactSettings *CompactSettings) *CompactDatabaseCommand {
	panicIf(conventions == nil, "Conventions cannot be null")
	panicIf(compactSettings == nil, "CompactSettings cannot be null")

	d, err := json.Marshal(compactSettings)
	panicIf(err != nil, "json.Marshal failed with '%s'", err)
	res := &CompactDatabaseCommand{
		RavenCommandBase: NewRavenCommandBase(),

		_compactSettings: d,
	}
	return res
}

func (c *CompactDatabaseCommand) createRequest(node *ServerNode) (*http.Request, error) {
	url := node.getUrl() + "/admin/compact"
	return NewHttpPost(url, string(c._compactSettings))
}

func (c *CompactDatabaseCommand) setResponse(response string, fromCache bool) error {
	if response == "" {
		return throwInvalidResponse()
	}
	var res OperationIdResult
	err := json.Unmarshal([]byte(response), &res)
	if err != nil {
		return err
	}
	c.Result = &res
	return nil
}
