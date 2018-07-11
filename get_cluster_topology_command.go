package ravendb

import (
	"encoding/json"
	"fmt"
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

func (c *GetClusterTopologyCommand) setResponse(response []byte, fromCache bool) error {
	if len(response) == 0 {
		return throwInvalidResponse()
	}
	var res ClusterTopologyResponse
	err := json.Unmarshal(response, &res)
	if err != nil {
		return err
	}
	c.Result = &res
	return nil
}
