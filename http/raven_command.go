package http

import (
	"net/http"
	"time"
	"strconv"
)

type ResponseType uint8

const (
	EMPTY ResponseType = iota
	OBJECT
	RAW
)

type RavenRequestable interface{
	CreateRequest(IServerNode, *string) (http.Request, error)
	GetTimeout() time.Duration
	Send(http.Client, *http.Request) (*http.Response, error)
	GetFailedNodes() map[IServerNode]error
	SetFailedNode(IServerNode, error)
	SetStatusCode(int)
	ShouldRefreshTopology() bool
	ProcessResponse(*http.Response, string)
}

type RavenCommand struct{
	ResponseType ResponseType
	FailedNodes map[IServerNode]error
	Result interface{}
	timeout time.Duration
	RefreshTopology bool
	statusCode int
}

func NewRavenCommand() (*RavenCommand, error){
	return &RavenCommand{ResponseType: OBJECT, timeout:0}, nil
}

func (command RavenCommand) Send(client http.Client, request *http.Request) (*http.Response, error){
	return client.Do(request)
}

func (command RavenCommand) GetTimeout() time.Duration{
	return command.timeout
}

func (command RavenCommand) GetFailedNodes() map[IServerNode]error{
	if command.FailedNodes == nil{
		command.FailedNodes = make(map[IServerNode]error)
	}
}

func (command RavenCommand) SetStatusCode(code int){
	command.statusCode = code
}

func (command RavenCommand) ShouldRefreshTopology() bool{
	return command.RefreshTopology
}

func (command RavenCommand) ProcessResponse(response http.Response, url string){
	if response.ContentLength == 0{
		return
	}
	refreshTopologyHeaderVal := response.Header.Get("Refresh-Topology")
	refreshTopology, err := strconv.ParseBool(refreshTopologyHeaderVal)
	if err != nil{
		refreshTopology = false
	}
	command.RefreshTopology = refreshTopology
	command.setResponseRaw(response)
}