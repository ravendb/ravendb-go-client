package ravendb

import (
	"net/http"
)

type OperationBootstrap struct {
}

func (o* OperationBootstrap) GetCommand(conventions *DocumentConventions) (RavenCommand, error) {
	return &bootstrapCommand{}, nil
}

type bootstrapCommand struct{
	RavenCommandBase

}

func (c *bootstrapCommand) createRequest(node *ServerNode) (*http.Request, error) {
	url := node.URL + "/admin/cluster/bootstrap"
	return newHttpPost(url, []byte{})
}
