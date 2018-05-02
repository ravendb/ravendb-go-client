package commands

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/ravendb/ravendb-go-client/http/server_nodes"
)

type PutDocumentCommand struct {
	command  *RavenCommand
	Document interface{}
	Key      string
}

func NewPutDocumentCommand(key string, document interface{}) (*PutDocumentCommand, error) {
	command := NewRavenCommand()
	command.SetMethod("PUT")
	if document == nil {
		document = struct{}{}
	}
	return &PutDocumentCommand{command: command, Document: document, Key: key}, nil
}

func (command *PutDocumentCommand) CreateRequest(node server_nodes.IServerNode) {
	command.SetData(command.Document)
	urlv := node.GetUrl()
	database := node.GetDatabase()
	key := command.Key
	command.SetUrl(fmt.Sprintf("%s/databases/%s/docs?id=%s", urlv, database, url.QueryEscape(key)))
}

func (command PutDocumentCommand) GetResponseRaw(resp *http.Response) ([]byte, error) {
	if resp.StatusCode == 200 {
		data, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return []byte{}, err
		}
		return data, err
	}
	return []byte{}, nil
}

func (command *PutDocumentCommand) SetHeaders(headers map[string]string) {
	command.command.SetHeaders(headers)
}

func (command *PutDocumentCommand) GetHeaders() map[string]string {
	return command.command.GetHeaders()
}

func (command *PutDocumentCommand) GetUrl() string {
	return command.command.GetUrl()
}

func (command *PutDocumentCommand) SetUrl(url string) {
	command.command.SetUrl(url)
}

func (command *PutDocumentCommand) GetMethod() string {
	return command.command.GetMethod()
}

func (command *PutDocumentCommand) SetMethod(method string) {
	command.command.SetMethod(method)
}

func (command *PutDocumentCommand) SetData(data interface{}) {
	command.command.SetData(data)
}

func (command *PutDocumentCommand) GetData() interface{} {
	return command.command.GetData()
}

func (command *PutDocumentCommand) GetFailedNodes() []server_nodes.IServerNode {
	return command.command.GetFailedNodes()
}

func (command *PutDocumentCommand) AddFailedNode(nodes server_nodes.IServerNode, err error) {
	command.command.AddFailedNode(nodes, err)
}

func (command *PutDocumentCommand) HasFailedWithNode(node server_nodes.IServerNode) bool {
	return command.command.HasFailedWithNode(node)
}
