package ravendb

import (
	"encoding/json"
	"net/http"
)

var (
	_ IMaintenanceOperation = &PutIndexesOperation{}
)

type PutIndexesOperation struct {
	_indexToAdd []*IndexDefinition

	Command *PutIndexesCommand
}

func NewPutIndexesOperation(indexToAdd ...*IndexDefinition) *PutIndexesOperation {
	return &PutIndexesOperation{
		_indexToAdd: indexToAdd,
	}
}

func (o *PutIndexesOperation) GetCommand(conventions *DocumentConventions) RavenCommand {
	o.Command = NewPutIndexesCommand(conventions, o._indexToAdd)
	return o.Command
}

var _ RavenCommand = &PutIndexesCommand{}

type PutIndexesCommand struct {
	RavenCommandBase

	_indexToAdd []ObjectNode

	Result []*PutIndexResult
}

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
		objectNode := EntityToJson_convertEntityToJson(indexToAdd, nil)
		cmd._indexToAdd = append(cmd._indexToAdd, objectNode)
	}

	return cmd
}

func (c *PutIndexesCommand) CreateRequest(node *ServerNode) (*http.Request, error) {
	url := node.GetUrl() + "/databases/" + node.GetDatabase() + "/admin/indexes"

	m := map[string]interface{}{
		"Indexes": c._indexToAdd,
	}
	d, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}
	//fmt.Printf("\nPutIndexesCommand.CreateRequest:\n%s\n\n", string(d))
	return NewHttpPut(url, d)
}

func (c *PutIndexesCommand) SetResponse(response []byte, fromCache bool) error {
	var res PutIndexesResponse
	err := json.Unmarshal(response, &res)
	if err != nil {
		return err
	}
	c.Result = res.Results
	return nil
}
