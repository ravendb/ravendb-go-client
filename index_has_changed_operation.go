package ravendb

import (
	"net/http"
)

var _ IMaintenanceOperation = &IndexHasChangedOperation{}

type IndexHasChangedOperation struct {
	definition *IndexDefinition

	Command *IndexHasChangedCommand
}

func NewIndexHasChangedOperation(definition *IndexDefinition) *IndexHasChangedOperation {
	return &IndexHasChangedOperation{
		definition: definition,
	}
}

func (o *IndexHasChangedOperation) GetCommand(conventions *DocumentConventions) (RavenCommand, error) {
	var err error
	o.Command, err = NewIndexHasChangedCommand(conventions, o.definition)
	if err != nil {
		return nil, err
	}
	return o.Command, nil
}

var (
	_ RavenCommand = &IndexHasChangedCommand{}
)

type IndexHasChangedCommand struct {
	RavenCommandBase

	definition []byte

	Result bool
}

func NewIndexHasChangedCommand(conventions *DocumentConventions, definition *IndexDefinition) (*IndexHasChangedCommand, error) {
	d, err := jsonMarshal(definition)
	if err != nil {
		return nil, err
	}
	res := &IndexHasChangedCommand{
		RavenCommandBase: NewRavenCommandBase(),

		definition: d,
	}
	return res, nil
}

func (c *IndexHasChangedCommand) createRequest(node *ServerNode) (*http.Request, error) {
	url := node.URL + "/databases/" + node.Database + "/indexes/has-changed"
	return NewHttpPost(url, c.definition)
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
