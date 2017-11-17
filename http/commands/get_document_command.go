package commands

import (
	"github.com/ravendb-go-client/http/server_nodes"
	"net/http"
	"fmt"
	"io/ioutil"
	"net/url"
	"strings"
)

type GetDocumentCommand struct{
	command *RavenCommand
	Document interface{}
	metadataOnly bool
	includes []string
	Key string
}

func NewGetDocumentCommand(key string, includes []string, metadataOnly bool) (*GetDocumentCommand, error){
	command := NewRavenCommand()
	command.SetMethod("GET")
	return &GetDocumentCommand{command: command, Key:key, includes:includes, metadataOnly:metadataOnly}, err
}

func (command *GetDocumentCommand) CreateRequest(node server_nodes.IServerNode){
	path := "docs?"
	if len(command.includes) > 0{
		includes := make([]string, len(command.includes))
		for _, include := range command.includes{
			includes = append(includes, "&include=" + include)
		}
		path += strings.Join(includes, ",")
	}
	command.SetData(command.Document)
	command.SetUrl(fmt.Sprintf("%s/databases/%s/docs?id=%s", node.GetUrl(), node.GetDatabase(), url.QueryEscape(command.Key)))
}

func (command GetDocumentCommand) GetResponseRaw(resp *http.Response) ([]byte, error){
	if resp.StatusCode == 200{
		data, err := ioutil.ReadAll(resp.Body)
		if err != nil{
			return []byte{}, err
		}
		return data, err
	}
	return []byte{}, nil
}


func (command *GetDocumentCommand) SetHeaders(headers map[string]string){
	command.command.SetHeaders(headers)
}

func (command *GetDocumentCommand) GetHeaders() map[string]string{
	return command.command.GetHeaders()
}

func (command *GetDocumentCommand) GetUrl() string{
	return command.command.GetUrl()
}

func (command *GetDocumentCommand) SetUrl(url string){
	command.command.SetUrl(url)
}

func (command *GetDocumentCommand) GetMethod() string{
	return command.command.GetMethod()
}

func (command *GetDocumentCommand) SetMethod(method string){
	command.command.SetMethod(method)
}

func (command *GetDocumentCommand) SetData(data interface{}){
	command.command.SetData(data)
}

func (command *GetDocumentCommand) GetData() interface{}{
	return command.command.GetData()
}

func (command *GetDocumentCommand) GetFailedNodes() []server_nodes.IServerNode{
	return command.command.GetFailedNodes()
}

func (command *GetDocumentCommand) AddFailedNode(nodes server_nodes.IServerNode, err error){
	command.command.AddFailedNode(nodes, err)
}

func (command *GetDocumentCommand) HasFailedWithNode(node server_nodes.IServerNode) bool{
	return command.command.HasFailedWithNode(node)
}
