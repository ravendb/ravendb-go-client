package ravendb

import (
	"net/http"
)

var _ IMaintenanceOperation = &IndexHasChangedOperation{}

type IndexHasChangedOperation struct {
	_definition *IndexDefinition

	Command *IndexHasChangedCommand
}

func NewIndexHasChangedOperation(definition *IndexDefinition) *IndexHasChangedOperation {
	return &IndexHasChangedOperation{
		_definition: definition,
	}
}

func (o *IndexHasChangedOperation) GetCommand(conventions *DocumentConventions) RavenCommand {
	o.Command = NewIndexHasChangedCommand(conventions, o._definition)
	return o.Command
}

var (
	_ RavenCommand = &IndexHasChangedCommand{}
)

type IndexHasChangedCommand struct {
	RavenCommandBase

	_definition []byte

	Result bool
}

func NewIndexHasChangedCommand(conventions *DocumentConventions, definition *IndexDefinition) *IndexHasChangedCommand {
	d, err := jsonMarshal(definition)
	panicIf(err != nil, "jsonMarshal() failed with %s", err)
	res := &IndexHasChangedCommand{
		RavenCommandBase: NewRavenCommandBase(),

		_definition: d,
	}
	return res
}

func (c *IndexHasChangedCommand) CreateRequest(node *ServerNode) (*http.Request, error) {
	url := node.URL + "/databases/" + node.Database + "/indexes/has-changed"
	return NewHttpPost(url, c._definition)
}

func (c *IndexHasChangedCommand) SetResponse(response []byte, fromCache bool) error {
	if response == nil {
		return throwInvalidResponse()
	}

	var res struct {
		Changed bool `json:"Changed"`
	}
	err := jsonUnmarshalFirst(response, &res)
	if err != nil {
		return err
	}
	c.Result = res.Changed
	return nil
}
