package commands

import (
	"net/http"
	"io/ioutil"
	"github.com/ravendb-go-client/http/server_nodes"
	"errors"
	"fmt"
)
// interface for RequestExecutor.Execute method
type IRavenRequestable interface{
	CreateRequest(server_nodes.IServerNode) error
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

func (ref *RavenCommand) SetHeaders(headers map[string]string){
	ref.Headers = headers
}

func (ref *RavenCommand) GetHeaders() map[string]string{
	return ref.Headers
}

func (ref *RavenCommand) GetMethod() string{
	return ref.Method
}

func (ref *RavenCommand) SetMethod(method string){
	ref.Method = method
}

func (ref *RavenCommand) GetUrl() string{
	return ref.Url
}

func (ref *RavenCommand) SetUrl(url string){
	ref.Url = url
}

func (ref *RavenCommand) GetData() interface{}{
	return ref.Data
}

func (ref *RavenCommand) SetData(data interface{}){
	ref.Data = data
}

func (ref *RavenCommand) GetFailedNodes() []server_nodes.IServerNode{
	return ref.FailedNodes
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
	ICommand
	commands []ICommandData
}
func NewBatchCommand(commands []ICommandData) *BatchCommand {
	ravCommand := NewRavenCommand()
	return &BatchCommand{ICommand: ravCommand, commands:commands}
}
func (ref *BatchCommand) CreateRequest(node server_nodes.IServerNode) error{
	var data []map[string]interface{}
	for _, commandData := range ref.commands{
		if !commandData.GetCommand(){
			return errors.New("Not a valid command")
		}
		data = append(data, commandData.ToJson())
	}
	ref.SetUrl(fmt.Sprintf("%s/databases/%s/bulk_docs", node.GetUrl(), node.GetDatabase()))
	ref.SetData(map[string][]map[string]interface{}{"Commands": data})
	return nil
}
func (ref *BatchCommand) GetResponseRaw(resp *http.Response) ([]byte, error){
	return nil, nil
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