package http

import (
	"net/http"
	"time"
)

type ResponseType uint8

const (
	EMPTY ResponseType = iota
	OBJECT
	RAW
)

type RavenRequestable interface{
	CreateRequest(ServerNode, *string) http.Request
	GetTimeout() time.Duration
	Send(http.Client, http.Request) (http.Response, error)
	GetFailedNodes() map[ServerNode]error
	SetFailedNode(ServerNode, error)
}

type RavenCommand struct{
	ResponseType ResponseType
	FailedNodes map[ServerNode]error
	Result interface{}
	timeout time.Duration

}

func NewRavenCommand() (*RavenCommand, error){
	return &RavenCommand{OBJECT, timeout:0}, nil
}

func (command RavenCommand) Send(client http.Client, request http.Request) (http.Response, error){
	return client.Do(request)
}

func (command RavenCommand) GetTimeout() time.Duration{
	return command.timeout
}

func (command RavenCommand) GetFailedNodes() map[ServerNode]error{
	if command.FailedNodes == nil{
		command.FailedNodes = make(map[ServerNode]error)
	}
}