package operations

import (
	"encoding/json"
	"github.com/ravendb/ravendb-go-client"
	"net/http"
	"net/url"
)

type OperationAddDatabaseNode struct {
	Name string `json:"Name"`
	Node string `json:"Node"`
}

func (operation *OperationAddDatabaseNode) GetCommand(conventions *ravendb.DocumentConventions) (ravendb.RavenCommand, error) {
	return &addDatabaseOperation{
		RaftCommandBase: ravendb.RaftCommandBase{
			RavenCommandBase: ravendb.RavenCommandBase{
				ResponseType: ravendb.RavenCommandResponseTypeObject,
			},
		},
		parent: operation,
	}, nil
}

type addDatabaseOperation struct {
	ravendb.RaftCommandBase
	parent *OperationAddDatabaseNode
}

func (o *addDatabaseOperation) CreateRequest(node *ravendb.ServerNode) (*http.Request, error) {
	base, err := url.Parse(node.URL + "/admin/databases/node?name=")
	if err != nil {
		return nil, err
	}
	params := url.Values{}
	params.Add("name", o.parent.Name)
	params.Add("node", o.parent.Node)
	base.RawQuery = params.Encode()

	return http.NewRequest(http.MethodPut, base.String(), nil)
}

func (c *addDatabaseOperation) SetResponse(response []byte, fromCache bool) error {
	return json.Unmarshal(response, c.parent)
}
