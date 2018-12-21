package ravendb

import (
	"net/http"
)

var (
	_ IMaintenanceOperation = &PutIndexesOperation{}
)

// PutIndexesOperation represents put indexes operation
type PutIndexesOperation struct {
	_indexToAdd []*IndexDefinition

	Command *PutIndexesCommand
}

// NewPutIndexesOperation returns new PutIndexesOperation
func NewPutIndexesOperation(indexToAdd ...*IndexDefinition) *PutIndexesOperation {
	return &PutIndexesOperation{
		_indexToAdd: indexToAdd,
	}
}

// GetCommand returns a command for this operation
func (o *PutIndexesOperation) GetCommand(conventions *DocumentConventions) RavenCommand {
	o.Command = NewPutIndexesCommand(conventions, o._indexToAdd)
	return o.Command
}

var _ RavenCommand = &PutIndexesCommand{}

// PutIndexesCommand represents put indexes command
type PutIndexesCommand struct {
	RavenCommandBase

	_indexToAdd []ObjectNode

	Result []*PutIndexResult
}

// NewPutIndexesCommand returns new PutIndexesCommand
func NewPutIndexesCommand(conventions *DocumentConventions, indexesToAdd []*IndexDefinition) *PutIndexesCommand {
	panicIf(conventions == nil, "conventions cannot be nil")
	panicIf(indexesToAdd == nil, "indexesToAdd cannot be nil")

	cmd := &PutIndexesCommand{
		RavenCommandBase: NewRavenCommandBase(),
	}

	for _, indexToAdd := range indexesToAdd {
		// Note: unlike java, Type is not calculated on demand. This is a decent
		// place to ensure it. Assumes that indexToAdd will not be modified
		// between now an CreateRequest()
		indexToAdd.updateIndexTypeAndMaps()

		panicIf(indexToAdd.Name == "", "Index name cannot be empty")
		objectNode := convertEntityToJSON(indexToAdd, nil)
		cmd._indexToAdd = append(cmd._indexToAdd, objectNode)
	}

	return cmd
}

// CreateRequest creates http request for this command
func (c *PutIndexesCommand) CreateRequest(node *ServerNode) (*http.Request, error) {
	url := node.GetUrl() + "/databases/" + node.GetDatabase() + "/admin/indexes"

	m := map[string]interface{}{
		"Indexes": c._indexToAdd,
	}
	d, err := jsonMarshal(m)
	if err != nil {
		return nil, err
	}
	return NewHttpPut(url, d)
}

// SetResponse decodes http response
func (c *PutIndexesCommand) SetResponse(response []byte, fromCache bool) error {
	var res PutIndexesResponse
	err := jsonUnmarshal(response, &res)
	if err != nil {
		return err
	}
	c.Result = res.Results
	return nil
}
