package ravendb

import (
	"net/http"
)

var (
	_ RavenCommand = &GetClusterTopologyCommand{}
)

type GetClusterTopologyCommand struct {
	RavenCommandBase
	Result *ClusterTopologyResponse
}

func NewGetClusterTopologyCommand() *GetClusterTopologyCommand {
	cmd := &GetClusterTopologyCommand{
		RavenCommandBase: NewRavenCommandBase(),
	}
	cmd.IsReadRequest = true
	return cmd
}

func (c *GetClusterTopologyCommand) createRequest(node *ServerNode) (*http.Request, error) {
	url := node.URL + "/cluster/topology"
	return newHttpGet(url)
}

func (c *GetClusterTopologyCommand) setResponse(response []byte, fromCache bool) error {
	if len(response) == 0 {
		return throwInvalidResponse()
	}
	return jsonUnmarshal(response, &c.Result)
}
