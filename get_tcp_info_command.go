package ravendb

import (
	"net/http"
)

var (
	_ RavenCommand = &GetTcpInfoCommand{}
)

// GetTcpInfoCommand describes "get tcp info" command
type GetTcpInfoCommand struct {
	RavenCommandBase

	tag           string
	dbName        string
	requestedNode *ServerNode

	Result *TcpConnectionInfo
}

// NewGetTcpInfoCommand returns new GetTcpInfoCommand
// dbName is optional
func NewGetTcpInfoCommand(tag, dbName string) *GetTcpInfoCommand {
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
		url = node.URL + "/info/tcp?tcp=" + c.tag
	} else {
		url = node.URL + "/databases/" + c.dbName + "/info/tcp?tag=" + c.tag
	}
	c.requestedNode = node
	return newHttpGet(url)
}

func (c *GetTcpInfoCommand) SetResponse(response []byte, fromCache bool) error {
	if len(response) == 0 {
		return throwInvalidResponse()
	}

	return jsonUnmarshal(response, &c.Result)
}
