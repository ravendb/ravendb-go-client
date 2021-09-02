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
		parent: operation,
	}, nil
}

type promoteNodeCommand struct {
	ravendb.RaftCommandBase
	parent *OperationPromoteClusterNode
}

func (c *promoteNodeCommand) createRequest(node *ravendb.ServerNode) (*http.Request, error) {
	url := node.URL + "/admin/cluster/promote?nodeTag=" + c.parent.Node
	return http.NewRequest(http.MethodPost, url, nil)
}

func (c *promoteNodeCommand) setResponse(response []byte, fromCache bool) error {
	return json.Unmarshal(response, c.parent)
}
