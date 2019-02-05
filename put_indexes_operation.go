package ravendb

import (
	"net/http"
)

var (
	_ IMaintenanceOperation = &PutIndexesOperation{}
)

// PutIndexesOperation represents put indexes operation
type PutIndexesOperation struct {
	indexToAdd []*IndexDefinition

	Command *PutIndexesCommand
}

// NewPutIndexesOperation returns new PutIndexesOperation
func NewPutIndexesOperation(indexToAdd ...*IndexDefinition) *PutIndexesOperation {
	return &PutIndexesOperation{
		indexToAdd: indexToAdd,
	}
}

// GetCommand returns a command for this operation
func (o *PutIndexesOperation) GetCommand(conventions *DocumentConventions) (RavenCommand, error) {
	var err error
	o.Command, err = NewPutIndexesCommand(conventions, o.indexToAdd)
	if err != nil {
		return nil, err
	}
	return o.Command, nil
}

var _ RavenCommand = &PutIndexesCommand{}

// PutIndexesCommand represents put indexes command
type PutIndexesCommand struct {
	RavenCommandBase

	indexToAdd []map[string]interface{}

	Result []*PutIndexResult
}

// NewPutIndexesCommand returns new PutIndexesCommand
func NewPutIndexesCommand(conventions *DocumentConventions, indexesToAdd []*IndexDefinition) (*PutIndexesCommand, error) {
	if conventions == nil {
		return nil, newIllegalArgumentError("conventions cannot be nil")
	}
	if indexesToAdd == nil {
		return nil, newIllegalArgumentError("indexesToAdd cannot be nil")
	}

	cmd := &PutIndexesCommand{
		RavenCommandBase: NewRavenCommandBase(),
	}

	for _, indexToAdd := range indexesToAdd {
		// Note: unlike java, Type is not calculated on demand. This is a decent
		// place to ensure it. Assumes that indexToAdd will not be modified
		// between now an createRequest()
		indexToAdd.updateIndexTypeAndMaps()

		panicIf(indexToAdd.Name == "", "Index name cannot be empty")
		objectNode := convertEntityToJSON(indexToAdd, nil)
		cmd.indexToAdd = append(cmd.indexToAdd, objectNode)
	}

	return cmd, nil
}

func (c *PutIndexesCommand) createRequest(node *ServerNode) (*http.Request, error) {
	url := node.URL + "/databases/" + node.Database + "/admin/indexes"

	m := map[string]interface{}{
		"Indexes": c.indexToAdd,
	}
	d, err := jsonMarshal(m)
	if err != nil {
		return nil, err
	}
	return newHttpPut(url, d)
}

func (c *PutIndexesCommand) setResponse(response []byte, fromCache bool) error {
	var res PutIndexesResponse
	err := jsonUnmarshal(response, &res)
	if err != nil {
		return err
	}
	c.Result = res.Results
	return nil
}
