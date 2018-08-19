package ravendb

import (
	"encoding/json"
	"net/http"
)

var (
	_ RavenCommand = &GetClusterTopologyCommand{}
)

type GetClusterTopologyCommand struct {
	*RavenCommandBase
	Result *ClusterTopologyResponse
}

func NewGetClusterTopologyCommand() *GetClusterTopologyCommand {
	cmd := &GetClusterTopologyCommand{
		RavenCommandBase: NewRavenCommandBase(),
	}
	cmd.IsReadRequest = true
	return cmd
}

func (c *GetClusterTopologyCommand) CreateRequest(node *ServerNode) (*http.Request, error) {
	url := node.GetUrl() + "/cluster/topology"
	return NewHttpGet(url)
}

func (c *GetClusterTopologyCommand) SetResponse(response []byte, fromCache bool) error {
	if len(response) == 0 {
		return throwInvalidResponse()
	}
	return json.Unmarshal(response, &c.Result)
}
