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
	fmt.Printf("NewGetClusterTopologyCommand()\n")
	cmd := &GetClusterTopologyCommand{
		RavenCommandBase: NewRavenCommandBase(),
	}
	cmd.IsReadRequest = true
	return cmd
}

func (c *GetClusterTopologyCommand) createRequest(node *ServerNode) (*http.Request, error) {
	fmt.Printf("NewGetClusterTopologyCommand.createRequest()\n")
	url := node.getUrl() + "/cluster/topology"
	return NewHttpGet(url)
}

func (c *GetClusterTopologyCommand) setResponse(response []byte, fromCache bool) error {
	fmt.Printf("NewGetClusterTopologyCommand.setResponse()\n")
	if len(response) == 0 {
		fmt.Printf("NewGetClusterTopologyCommand.setResponse(): len(response)==0\n")
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
