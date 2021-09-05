package operations

import (
	"encoding/json"
	"github.com/ravendb/ravendb-go-client"
	"net/http"
)

type OperationDemoteClusterNode struct {
	Node string `json:"Node"`
}

func (operation *OperationDemoteClusterNode) GetCommand(conventions *ravendb.DocumentConventions) (ravendb.RavenCommand, error) {
	return &demoteNodeCommand{
		RaftCommandBase: ravendb.RaftCommandBase{
			RavenCommandBase: ravendb.RavenCommandBase{
				ResponseType: ravendb.RavenCommandResponseTypeObject,
			},
		},
		parent: operation,
	}, nil
}

type demoteNodeCommand struct {
	ravendb.RaftCommandBase
	parent *OperationDemoteClusterNode
}

func (c *demoteNodeCommand) CreateRequest(node *ravendb.ServerNode) (*http.Request, error) {
	url := node.URL + "/admin/cluster/demote?nodeTag=" + c.parent.Node
	return http.NewRequest(http.MethodPost, url, nil)
}

func (c *demoteNodeCommand) SetResponse(response []byte, fromCache bool) error {
	return json.Unmarshal(response, c.parent)
}
