package operations

import (
	"encoding/json"
	"github.com/ravendb/ravendb-go-client"
	"net/http"
)

type OperationBootstrap struct {
}

func (o *OperationBootstrap) GetCommand(conventions *ravendb.DocumentConventions) (ravendb.RavenCommand, error) {
	return &bootstrapCommand{
		RaftCommandBase: ravendb.RaftCommandBase{
			RavenCommandBase: ravendb.RavenCommandBase{
				ResponseType: ravendb.RavenCommandResponseTypeObject,
			},
		},
		parent: o,
	}, nil
}

type bootstrapCommand struct {
	ravendb.RaftCommandBase
	parent *OperationBootstrap
}

func (c *bootstrapCommand) createRequest(node *ravendb.ServerNode) (*http.Request, error) {
	url := node.URL + "/admin/cluster/bootstrap"
	return ravendb.NewHttpPost(url, []byte{})
}

func (c *bootstrapCommand) setResponse(response []byte, fromCache bool) error {
	return json.Unmarshal(response, c.parent)
}
