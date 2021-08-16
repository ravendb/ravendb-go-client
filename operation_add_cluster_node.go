package ravendb

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

type OperationAddClusterNode struct {
	Url     string
	Tag     string
	Watcher bool
}

func (operation *OperationAddClusterNode) GetCommand(conventions *DocumentConventions) (RavenCommand, error) {
	return &addNodeCommand{
		op: operation,
	}, nil
}

type addNodeCommand struct {
	RavenCommandBase
	op *OperationAddClusterNode
}

func (c *addNodeCommand) createRequest(node *ServerNode) (*http.Request, error) {
	url := node.URL + "/admin/cluster/node?url=" + url.QueryEscape(c.op.Url) + "&watcher=" + strconv.FormatBool(c.op.Watcher)

	if len(strings.TrimSpace(c.op.Tag)) == 0 {
		url += fmt.Sprintf("&tag=%s", c.op.Tag)
	}
	return http.NewRequest(http.MethodPut, url, nil)
}
