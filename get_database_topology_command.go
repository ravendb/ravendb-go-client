package ravendb

import (
	"net/http"
	"strings"
)

var (
	_ RavenCommand = &GetDatabaseTopologyCommand{}
)

type GetDatabaseTopologyCommand struct {
	RavenCommandBase

	Result *Topology
}

func NewGetDatabaseTopologyCommand() *GetDatabaseTopologyCommand {
	cmd := &GetDatabaseTopologyCommand{
		RavenCommandBase: NewRavenCommandBase(),
	}
	cmd.IsReadRequest = true
	return cmd
}

func (c *GetDatabaseTopologyCommand) createRequest(node *ServerNode) (*http.Request, error) {
	url := node.URL + "/topology?name=" + node.Database
	if strings.Contains(strings.ToLower(node.URL), ".fiddler") {
		// we want to keep the '.fiddler' stuff there so we'll keep tracking request
		// so we are going to ask the server to respect it
		url += "&localUrl=" + urlUtilsEscapeDataString(node.URL)
	}
	return newHttpGet(url)
}

func (c *GetDatabaseTopologyCommand) setResponse(response []byte, fromCache bool) error {
	return jsonUnmarshal(response, &c.Result)
}
