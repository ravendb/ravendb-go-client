package ravendb

import (
	"net/http"
)

type OperationPromoteClusterNode struct {
	Node string
}

func NewOperationPromoteClusterNode(node string) *OperationPromoteClusterNode {
	return &OperationPromoteClusterNode{
		Node: node,
	}
}
func (operation *OperationPromoteClusterNode) GetCommand(conventions *DocumentConventions) (RavenCommand, error) {
	return &promoteNodeCommand{
		op: operation,
	}, nil
}

type promoteNodeCommand struct {
	RaftCommandBase
	op *OperationPromoteClusterNode
}

func (c *promoteNodeCommand) createRequest(node *ServerNode) (*http.Request, error) {
	url := node.URL + "/admin/cluster/promote?nodeTag=" + c.op.Node
	return http.NewRequest(http.MethodPost, url, nil)
}
