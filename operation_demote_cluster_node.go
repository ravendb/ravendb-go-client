package ravendb

import (
	"net/http"
)

type OperationDemoteClusterNode struct {
	Node string
}

func (operation *OperationDemoteClusterNode) GetCommand(conventions *DocumentConventions) (RavenCommand, error) {
	return &demoteNodeCommand{
		op: operation,
	}, nil
}

type demoteNodeCommand struct {
	RaftCommandBase
	op *OperationDemoteClusterNode
}

func (c *demoteNodeCommand) createRequest(node *ServerNode) (*http.Request, error) {
	url := node.URL + "/admin/cluster/demote?nodeTag=" + c.op.Node
	return http.NewRequest(http.MethodPost, url, nil)
}
