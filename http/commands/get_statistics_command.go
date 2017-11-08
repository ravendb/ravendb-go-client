package commands

import (
	"github.com/ravendb-go-client/http/server_nodes"
	"net/http"
	"fmt"
	"io/ioutil"
)

type GetStatisticsCommand struct{
	command *Command
}

func NewGetStatisticsCommand() (*GetStatisticsCommand, error){
	command, err := NewRavenCommand()
	command.SetMethod("GET")
	return &GetStatisticsCommand{command: command}, err
}

func (command *GetStatisticsCommand) CreateRequest(node server_nodes.IServerNode){
	command.SetUrl(fmt.Sprintf("%s/database/%s/stats", node.GetUrl(), node.GetDatabase()))
}

func (command *GetStatisticsCommand) GetResponseRaw(resp *http.Response) ([]byte, error){
	if resp.StatusCode == 200{
		data, err := ioutil.ReadAll(resp.Body)
		if err != nil{
			return []byte{}, err
		}
		return data, err
	}
	return []byte{}, nil
}


func (command *GetStatisticsCommand) SetHeaders(headers map[string]string){
	command.command.SetHeaders(headers)
}

func (command *GetStatisticsCommand) GetHeaders() map[string]string{
	return command.command.GetHeaders()
}

func (command *GetStatisticsCommand) GetUrl() string{
	return command.command.GetUrl()
}

func (command *GetStatisticsCommand) SetUrl(url string){
	command.command.SetUrl(url)
}

func (command *GetStatisticsCommand) GetMethod() string{
	return command.command.GetMethod()
}

func (command *GetStatisticsCommand) SetMethod(method string){
	command.command.SetMethod(method)
}

func (command *GetStatisticsCommand) GetData() interface{}{
	return command.command.GetData()
}

func (command *GetStatisticsCommand) SetData(data interface{}){
	command.command.SetData(data)
}

func (command *GetStatisticsCommand) GetFailedNodes() []server_nodes.IServerNode{
	return command.command.GetFailedNodes()
}

func (command *GetStatisticsCommand) AddFailedNode(nodes server_nodes.IServerNode, err error){
	command.command.AddFailedNode(nodes, err)
}

func (command *GetStatisticsCommand) HasFailedWithNode(node server_nodes.IServerNode) bool{
	return command.command.HasFailedWithNode(node)
}