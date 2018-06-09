package ravendb

import (
	"encoding/json"
	"net/http"
)

type _GetTcpInfoCommand struct {
	tag           String
	dbName        String
	requestedNode *ServerNode
}

func NewGetTcpInfoCommand(tag string) *RavenCommand {
	return NewGetTcpInfoCommandWithDatbase(tag, "")
}

func NewGetTcpInfoCommandWithDatbase(tag, dbName string) *RavenCommand {
	data := &_GetTcpInfoCommand{
		tag:    tag,
		dbName: dbName,
	}
	cmd := NewRavenCommand()
	cmd.IsReadRequest = true
	cmd.data = data
	cmd.createRequestFunc = GetTcpInfoCommand_createRequest
	cmd.setResponseFunc = GetTcpInfoCommand_setResponse

	return cmd
}

func GetTcpInfoCommand_createRequest(cmd *RavenCommand, node *ServerNode) (*http.Request, error) {
	data := cmd.data.(*_GetTcpInfoCommand)

	url := ""
	if data.dbName == "" {
		url = node.getUrl() + "/info/tcp?tcp=" + data.tag
	} else {
		url = node.getUrl() + "/databases/" + data.dbName + "/info/tcp?tag=" + data.tag
	}
	data.requestedNode = node
	return NewHttpGet(url)
}

func GetTcpInfoCommand_setResponse(cmd *RavenCommand, response String, fromCache bool) error {
	if response == "" {
		return throwInvalidResponse()
	}
	var res TcpConnectionInfo
	err := json.Unmarshal([]byte(response), &res)
	if err != nil {
		return err
	}
	cmd.result = &res
	return nil
}
