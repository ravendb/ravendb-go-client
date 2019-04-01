package ravendb

import (
	"net/http"
)

var (
	_ IMaintenanceOperation = &PutIndexesOperation{}
)

// PutIndexesOperation represents put indexes operation
type PutIndexesOperation struct {
	indexToAdd            []*IndexDefinition
	_allJavaScriptIndexes bool

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

	indexToAdd            []map[string]interface{}
	_allJavaScriptIndexes bool

	Result []*PutIndexResult
}

// NewPutIndexesCommand returns new PutIndexesCommand
func NewPutIndexesCommand(conventions *DocumentConventions, indexesToAdd []*IndexDefinition) (*PutIndexesCommand, error) {
	if conventions == nil {
		return nil, newIllegalArgumentError("conventions cannot be nil")
	}
	if len(indexesToAdd) == 0 {
		return nil, newIllegalArgumentError("indexesToAdd cannot be empty")
	}

	cmd := &PutIndexesCommand{
		RavenCommandBase:      NewRavenCommandBase(),
		_allJavaScriptIndexes: true,
	}

	for _, indexToAdd := range indexesToAdd {
		// We validate on the server that it is indeed a javascript index.

		// Note: unlike java, Type is not calculated on demand. This is a decent
		// place to ensure it. Assumes that indexToAdd will not be modified
		// between now an createRequest()
		indexToAdd.updateIndexTypeAndMaps()

		if !IndexTypeExtensions_isJavaScript(indexToAdd.IndexType) {
			cmd._allJavaScriptIndexes = false

		}
		if indexToAdd.Name == "" {
			return nil, newIllegalArgumentError("Index name cannot be empty")
		}
		objectNode := valueToTree(indexToAdd)
		cmd.indexToAdd = append(cmd.indexToAdd, objectNode)
	}

	return cmd, nil
}

func (c *PutIndexesCommand) createRequest(node *ServerNode) (*http.Request, error) {
	url := node.URL + "/databases/" + node.Database
	if c._allJavaScriptIndexes {
		url += "/indexes"
	} else {
		url += "/admin/indexes"
	}

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
