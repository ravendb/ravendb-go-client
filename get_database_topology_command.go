package ravendb

import (
	"encoding/json"
	"net/http"
	"strings"
)

var (
	_ RavenCommand = &GetDatabaseTopologyCommand{}
)

type GetDatabaseTopologyCommand struct {
	*RavenCommandBase

	Result *Topology
}

func NewGetDatabaseTopologyCommand() *GetDatabaseTopologyCommand {
	cmd := &GetDatabaseTopologyCommand{
		RavenCommandBase: NewRavenCommandBase(),
	}
	cmd.IsReadRequest = true
	return cmd
}

func (c *GetDatabaseTopologyCommand) CreateRequest(node *ServerNode) (*http.Request, error) {
	url := node.getUrl() + "/topology?name=" + node.getDatabase()
	if strings.Contains(strings.ToLower(node.getUrl()), ".fiddler") {
		// we want to keep the '.fiddler' stuff there so we'll keep tracking request
		// so we are going to ask the server to respect it
		url += "&localUrl=" + UrlUtils_escapeDataString(node.getUrl())
	}
	return NewHttpGet(url)
}

func (c *GetDatabaseTopologyCommand) SetResponse(response []byte, fromCache bool) error {
	var res Topology
	err := json.Unmarshal(response, &res)
	if err != nil {
		return err
	}
	c.Result = &res
	return nil
}
