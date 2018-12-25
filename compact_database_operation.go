package ravendb

import (
	"net/http"
)

var (
	_ IServerOperation = &CompactDatabaseOperation{}
)

// CompactDatabaseOperation describes "compact database" operation
type CompactDatabaseOperation struct {
	_compactSettings *CompactSettings

	Command *CompactDatabaseCommand
}

// NewCompactDatabaseOperation returns new CompactDatabaseOperation
func NewCompactDatabaseOperation(compactSettings *CompactSettings) *CompactDatabaseOperation {
	return &CompactDatabaseOperation{
		_compactSettings: compactSettings,
	}
}

// GetCommand returns a comman
func (o *CompactDatabaseOperation) GetCommand(conventions *DocumentConventions) RavenCommand {
	o.Command = NewCompactDatabaseCommand(conventions, o._compactSettings)
	return o.Command
}

var _ RavenCommand = &CompactDatabaseCommand{}

// CompactDatabaseCommand describes "compact database" command
type CompactDatabaseCommand struct {
	RavenCommandBase

	_compactSettings []byte // CompactSettings serialized to json

	Result *OperationIDResult
}

//NewCompactDatabaseCommand returns new CompactDatabaseCommand
func NewCompactDatabaseCommand(conventions *DocumentConventions, compactSettings *CompactSettings) *CompactDatabaseCommand {
	panicIf(conventions == nil, "Conventions cannot be null")
	panicIf(compactSettings == nil, "CompactSettings cannot be null")

	d, err := jsonMarshal(compactSettings)
	panicIf(err != nil, "jsonMarshal failed with '%s'", err)
	res := &CompactDatabaseCommand{
		RavenCommandBase: NewRavenCommandBase(),

		_compactSettings: d,
	}
	return res
}

// CreateRequest creates a request
func (c *CompactDatabaseCommand) CreateRequest(node *ServerNode) (*http.Request, error) {
	url := node.GetUrl() + "/admin/compact"
	return NewHttpPost(url, c._compactSettings)
}

// SetResponse sets a reponse
func (c *CompactDatabaseCommand) SetResponse(response []byte, fromCache bool) error {
	if len(response) == 0 {
		return throwInvalidResponse()
	}
	return jsonUnmarshal(response, &c.Result)
}
