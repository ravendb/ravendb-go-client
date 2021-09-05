package operations

import (
	"encoding/json"
	"github.com/ravendb/ravendb-go-client"
	"net/http"
)

type OperationPromoteClusterNode struct {
	Node string `json:"Node"`
}

func NewOperationPromoteClusterNode(node string) *OperationPromoteClusterNode {
	return &OperationPromoteClusterNode{
		Node: node,
	}
}
func (operation *OperationPromoteClusterNode) GetCommand(conventions *ravendb.DocumentConventions) (ravendb.RavenCommand, error) {
	return &promoteNodeCommand{
		RaftCommandBase: ravendb.RaftCommandBase{
			RavenCommandBase: ravendb.RavenCommandBase{
				ResponseType: ravendb.RavenCommandResponseTypeObject,
			},
		},
		parent: operation,
	}, nil
}

type promoteNodeCommand struct {
	ravendb.RaftCommandBase
	parent *OperationPromoteClusterNode
}

func (c *promoteNodeCommand) CreateRequest(node *ravendb.ServerNode) (*http.Request, error) {
	url := node.URL + "/admin/cluster/promote?nodeTag=" + c.parent.Node
	return http.NewRequest(http.MethodPost, url, nil)
}

func (c *promoteNodeCommand) SetResponse(response []byte, fromCache bool) error {
	return json.Unmarshal(response, c.parent)
}
