package commands

import (
	"github.com/ravendb-go-client/http/server_nodes"
	"net/http"
	"fmt"
	"io/ioutil"
)

type GetTopologyCommand struct{
	command *Command
}

func NewGetTopologyCommand() (*GetTopologyCommand, error){
	command, err := NewRavenCommand()
	command.SetMethod("GET")
	return &GetTopologyCommand{command: command}, err
}

func (command GetTopologyCommand) CreateRequest(node server_nodes.IServerNode){
	command.SetUrl(fmt.Sprintf("%s/topology?name=%s", node.GetUrl(), node.GetDatabase()))
}

func (command GetTopologyCommand) SetResponse(resp *http.Response) ([]byte, error){
	if resp.StatusCode == 200{
		data, err := ioutil.ReadAll(resp.Body)
		if err != nil{
			return []byte{}, err
		}
		return data, err
	}
	return []byte{}, nil
}


func (command GetTopologyCommand) SetHeaders(headers map[string]string){
	command.command.SetHeaders(headers)
}

func (command GetTopologyCommand) GetHeaders() map[string]string{
	return command.command.GetHeaders()
}

func (command GetTopologyCommand) GetUrl() string{
	return command.command.GetUrl()
}

func (command GetTopologyCommand) SetUrl(url string){
	command.command.SetUrl(url)
}

func (command GetTopologyCommand) GetMethod() string{
	return command.command.GetMethod()
}

func (command GetTopologyCommand) SetMethod(method string){
	command.command.SetMethod(method)
}

func (command GetTopologyCommand) GetData() interface{}{
	return command.command.GetData()
}

func (command GetTopologyCommand) SetData(data interface{}){
	command.command.SetData(data)
}

func (command GetTopologyCommand) GetFailedNodes() []server_nodes.IServerNode{
	return command.command.GetFailedNodes()
}

func (command GetTopologyCommand) AddFailedNode(nodes server_nodes.IServerNode, err error){
	command.command.AddFailedNode(nodes, err)
}

func (command GetTopologyCommand) HasFailedWithNode(node server_nodes.IServerNode) bool{
	return command.command.HasFailedWithNode(node)
}
