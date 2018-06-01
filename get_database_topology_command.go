package ravendb

import (
	"encoding/json"
	"net/http"
	"strings"
)

func NewGetDatabaseTopologyCommand() *RavenCommand {
	cmd := NewRavenCommand()
	cmd.IsReadRequest = true
	cmd.setResponseFunc = GetDatabaseTopologyCommand_setResponse
	cmd.createRequestFunc = GetDatabaseTopologyCommand_createRequest
	return cmd
}

func GetDatabaseTopologyCommand_createRequest(cmd *RavenCommand, node *ServerNode) (*http.Request, error) {

	url := node.getUrl() + "/topology?name=" + node.getDatabase()
	if strings.Contains(strings.ToLower(node.getUrl()), ".fiddler") {
		// we want to keep the '.fiddler' stuff there so we'll keep tracking request
		// so we are going to ask the server to respect it
		url += "&localUrl=" + UrlUtils_escapeDataString(node.getUrl())
	}
	return NewHttpGet(url)
}

func GetDatabaseTopologyCommand_setResponse(cmd *RavenCommand, response String, fromCache bool) error {
	var res Topology
	err := json.Unmarshal([]byte(response), &res)
	if err != nil {
		return err
	}
	cmd.result = &res
	return nil
}
