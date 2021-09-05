package operations

import (
	"encoding/json"
	"fmt"
	"github.com/ravendb/ravendb-go-client"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

type OperationAddClusterNode struct {
	Url     string `json:"Url"`
	Tag     string `json:"Tag"`
	Watcher bool   `json:"Watcher"`
}

func (operation *OperationAddClusterNode) GetCommand(conventions *ravendb.DocumentConventions) (ravendb.RavenCommand, error) {
	return &addNodeCommand{
		RaftCommandBase: ravendb.RaftCommandBase{
			RavenCommandBase: ravendb.RavenCommandBase{
				ResponseType: ravendb.RavenCommandResponseTypeObject,
			},
		},
		parent: operation,
	}, nil
}

type addNodeCommand struct {
	ravendb.RaftCommandBase
	parent *OperationAddClusterNode
}

func (c *addNodeCommand) createRequest(node *ravendb.ServerNode) (*http.Request, error) {
	url := node.URL + "/admin/cluster/node?url=" + url.QueryEscape(c.parent.Url) + "&watcher=" + strconv.FormatBool(c.parent.Watcher)

	if len(strings.TrimSpace(c.parent.Tag)) == 0 {
		url += fmt.Sprintf("&tag=%s", c.parent.Tag)
	}
	return http.NewRequest(http.MethodPut, url, nil)
}

func (c *addNodeCommand) setResponse(response []byte, fromCache bool) error {
	return json.Unmarshal(response, c.parent)
}
