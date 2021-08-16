package ravendb

import (
	"net/http"
)

type RemoveClusterNode struct {
	Node string
}

func NewRemovePromoteClusterNode(node string) *OperationPromoteClusterNode {
	return &OperationPromoteClusterNode{
		Node: node,
	}
}
func (operation *RemoveClusterNode) GetCommand(conventions *DocumentConventions) (RavenCommand, error) {
	return &removeNodeCommand{
		op: operation,
	}, nil
}

type removeNodeCommand struct {
	RavenCommandBase
	op *RemoveClusterNode
}

func (c *removeNodeCommand) createRequest(node *ServerNode) (*http.Request, error) {
	url := node.URL + "/admin/cluster/node?nodeTag=" + c.op.Node
	return http.NewRequest(http.MethodDelete, url, nil)
}
