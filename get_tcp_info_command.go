package ravendb

import (
	"encoding/json"
	"net/http"
)

var (
	_ RavenCommand = &GetTcpInfoCommand{}
)

type GetTcpInfoCommand struct {
	*RavenCommandBase

	tag           string
	dbName        string
	requestedNode *ServerNode

	Result *TcpConnectionInfo
}

func NewGetTcpInfoCommand(tag string) *GetTcpInfoCommand {
	return NewGetTcpInfoCommandWithDatbase(tag, "")
}

func NewGetTcpInfoCommandWithDatbase(tag, dbName string) *GetTcpInfoCommand {
	cmd := &GetTcpInfoCommand{
		RavenCommandBase: NewRavenCommandBase(),

		tag:    tag,
		dbName: dbName,
	}
	cmd.IsReadRequest = true
	return cmd
}

func (c *GetTcpInfoCommand) CreateRequest(node *ServerNode) (*http.Request, error) {
	url := ""
	if c.dbName == "" {
		url = node.GetUrl() + "/info/tcp?tcp=" + c.tag
	} else {
		url = node.GetUrl() + "/databases/" + c.dbName + "/info/tcp?tag=" + c.tag
	}
	c.requestedNode = node
	return NewHttpGet(url)
}

func (c *GetTcpInfoCommand) SetResponse(response []byte, fromCache bool) error {
	if len(response) == 0 {
		return throwInvalidResponse()
	}

	return json.Unmarshal(response, &c.Result)
}
