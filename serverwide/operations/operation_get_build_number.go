package operations

import (
	"encoding/json"
	"github.com/ravendb/ravendb-go-client"
	"net/http"
)

type OperationGetBuildNumber struct {
	BuildVersion   int    `json:"BuildVersion"`
	ProductVersion string `json:"ProductVersion"`
	CommitHash     string `json:"CommitHash"`
	FullVersion    string `json:"FullVersion"`
}

func (operation *OperationGetBuildNumber) GetCommand(conventions *ravendb.DocumentConventions) (ravendb.RavenCommand, error) {
	return &getBuildNumber{
		RavenCommandBase: ravendb.RavenCommandBase{
			ResponseType: ravendb.RavenCommandResponseTypeObject,
		},
		parent: operation,
	}, nil
}

type getBuildNumber struct {
	ravendb.RavenCommandBase
	parent *OperationGetBuildNumber
}

func (c *getBuildNumber) createRequest(node *ravendb.ServerNode) (*http.Request, error) {
	url := node.URL + "/build/version"
	return http.NewRequest(http.MethodGet, url, nil)
}

func (c *getBuildNumber) setResponse(response []byte, fromCache bool) error {
	return json.Unmarshal(response, c.parent)
}
