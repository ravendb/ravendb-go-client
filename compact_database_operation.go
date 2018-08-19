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

func (o *CompactDatabaseOperation) GetCommand(conventions *DocumentConventions) RavenCommand {
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

func (c *CompactDatabaseCommand) CreateRequest(node *ServerNode) (*http.Request, error) {
	url := node.GetUrl() + "/admin/compact"
	return NewHttpPost(url, c._compactSettings)
}

func (c *CompactDatabaseCommand) SetResponse(response []byte, fromCache bool) error {
	if len(response) == 0 {
		return throwInvalidResponse()
	}
	return json.Unmarshal(response, &c.Result)
}
