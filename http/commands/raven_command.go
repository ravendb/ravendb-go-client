package commands

import (
	"net/http"
	"io/ioutil"
	"github.com/ravendb-go-client/http/server_nodes"
)
// interface for RequestExecutor.Execute method
type IRavenRequestable interface{
	CreateRequest(server_nodes.IServerNode)
	GetResponseRaw(*http.Response) ([]byte, error)
	ICommand
}

type ICommand interface{
	SetHeaders(map[string]string)
	GetHeaders() map[string]string
	GetMethod() string
	SetMethod(string)
	GetUrl() string
	SetUrl(string)
	GetData() interface{}
	SetData(interface{})
	GetFailedNodes() []server_nodes.IServerNode
	AddFailedNode(server_nodes.IServerNode, error)
	HasFailedWithNode(server_nodes.IServerNode) bool
}

type RavenCommand struct{
	Headers map[string]string
	Data interface{}
	Result []byte
	Method, Url string
	FailedNodes []server_nodes.IServerNode
	IsReadRequest, UseStream, ravenCommand bool
}

func NewRavenCommand() (ref *RavenCommand){
	ref = &RavenCommand{}
	return
}

func (command RavenCommand) SetHeaders(headers map[string]string){
	command.Headers = headers
}

func (command RavenCommand) GetHeaders() map[string]string{
	return command.Headers
}

func (command RavenCommand) GetMethod() string{
	return command.Method
}

func (command RavenCommand) SetMethod(method string){
	command.Method = method
}

func (command RavenCommand) GetUrl() string{
	return command.Url
}

func (command RavenCommand) SetUrl(url string){
	command.Url = url
}

func (command RavenCommand) GetData() interface{}{
	return command.Data
}

func (command RavenCommand) SetData(data interface{}){
	command.Data = data
}

func (command RavenCommand) GetFailedNodes() []server_nodes.IServerNode{
	return command.FailedNodes
}
// GetResponseRaw revert response object to JSON slice
func (command RavenCommand) GetResponseRaw(resp *http.Response) ([]byte, error){
	if resp == nil{
		command.Result = []byte{}
		return command.Result, nil
	}
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil{
		return nil, err
	}
	command.Result = data
	return command.Result, err
}

func (command RavenCommand) AddFailedNode(node server_nodes.IServerNode, err error){
	command.FailedNodes = append(command.FailedNodes, node)
}

func (command RavenCommand) HasFailedWithNode(node server_nodes.IServerNode) bool{
	for _, v := range command.FailedNodes {
		if v == node {
			return true
		}
	}
	return false
}
//todo: implement
type BatchCommand struct {
	IRavenRequestable
	commands []string
}
func NewBatchCommand(commans []string) *BatchCommand {
	return &BatchCommand{commands:commans}
}
//todo: implement
type GetOperationStateCommand struct {
	IRavenRequestable
	operationId            string
	isServerStoreOperation bool
}
func NewGetOperationStateCommand(operationId string) *GetOperationStateCommand {
	return &GetOperationStateCommand{operationId: operationId}
}