package ravendb

import (
	"net/http"
)

var (
	_ IServerOperation = &CompactDatabaseOperation{}
)

// CompactDatabaseOperation describes "compact database" operation
type CompactDatabaseOperation struct {
	compactSettings *CompactSettings

	Command *CompactDatabaseCommand
}

// NewCompactDatabaseOperation returns new CompactDatabaseOperation
func NewCompactDatabaseOperation(compactSettings *CompactSettings) *CompactDatabaseOperation {
	return &CompactDatabaseOperation{
		compactSettings: compactSettings,
	}
}

// GetCommand returns a command
func (o *CompactDatabaseOperation) GetCommand(conventions *DocumentConventions) (RavenCommand, error) {
	var err error
	o.Command, err = NewCompactDatabaseCommand(conventions, o.compactSettings)
	if err != nil {
		return nil, err
	}
	return o.Command, nil
}

var _ RavenCommand = &CompactDatabaseCommand{}

// CompactDatabaseCommand describes "compact database" command
type CompactDatabaseCommand struct {
	RavenCommandBase

	compactSettings []byte // CompactSettings serialized to json

	Result *OperationIDResult
}

//NewCompactDatabaseCommand returns new CompactDatabaseCommand
func NewCompactDatabaseCommand(conventions *DocumentConventions, compactSettings *CompactSettings) (*CompactDatabaseCommand, error) {
	if conventions == nil {
		return nil, newIllegalArgumentError("Conventions cannot be null")
	}
	if compactSettings == nil {
		return nil, newIllegalArgumentError("CompactSettings cannot be null")
	}

	d, err := jsonMarshal(compactSettings)
	if err != nil {
		return nil, err
	}
	res := &CompactDatabaseCommand{
		RavenCommandBase: NewRavenCommandBase(),

		compactSettings: d,
	}
	return res, nil
}

// CreateRequest creates a request
func (c *CompactDatabaseCommand) CreateRequest(node *ServerNode) (*http.Request, error) {
	url := node.URL + "/admin/compact"
	return NewHttpPost(url, c.compactSettings)
}

// SetResponse sets a response
func (c *CompactDatabaseCommand) SetResponse(response []byte, fromCache bool) error {
	if len(response) == 0 {
		return throwInvalidResponse()
	}
	return jsonUnmarshal(response, &c.Result)
}
