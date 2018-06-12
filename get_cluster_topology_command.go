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

func (c *GetClusterTopologyCommand) createRequest(node *ServerNode) (*http.Request, error) {
	url := node.getUrl() + "/cluster/topology"
	return NewHttpGet(url)
}

func (c *GetClusterTopologyCommand) setResponse(response String, fromCache bool) error {
	if response == "" {
		return throwInvalidResponse()
	}
	var res ClusterTopologyResponse
	err := json.Unmarshal([]byte(response), &res)
	if err != nil {
		return err
	}
	c.Result = &res
	return nil
}
