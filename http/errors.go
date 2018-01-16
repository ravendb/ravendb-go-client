package http

import (
	"fmt"
	"github.com/ravendb/ravendb-go-client/http/server_nodes"
	"encoding/json"
	"io/ioutil"
	"net/http"
)

type AllTopologyNodesDownError struct{
	error string
	Topology Topology
}

func NewAllTopologyNodesDownError(error string, topology Topology) (*AllTopologyNodesDownError, error){
	return &AllTopologyNodesDownError{Topology: topology, error: error}, nil
}

func (err AllTopologyNodesDownError) Error() string{
	return err.error
}

type TopologyUpdateError struct{
	error string
	ErrorList map[string]error
}

func (topologyUpdateError TopologyUpdateError) Error() string{
	return topologyUpdateError.error
}

type ErrorResponseError struct{
}

func NewErrorResponseError() (*ErrorResponseError, error){
	return &ErrorResponseError{}, nil
}

func (err ErrorResponseError) Error() string{
	return fmt.Sprintf("Failed to put document in the database please check the connection to the server")
}

type UnsuccessfulRequestError struct{
	Url string
	FailedNode server_nodes.IServerNode
}

func NewUnsuccessfulRequestError(url string, node server_nodes.IServerNode) (*UnsuccessfulRequestError, error){
	return &UnsuccessfulRequestError{FailedNode: node, Url: url}, nil
}

func (err UnsuccessfulRequestError) Error() string{
	return fmt.Sprintf("Request to %s on node %s failed", err.Url, err.FailedNode.GetClusterTag())
}

type ServerError struct{
	Url, Type, Message, Error string
}

func NewServerError(response http.Response) (*ServerError, error){
	var servErr ServerError
	body, err := ioutil.ReadAll(response.Body)
	if err != nil{
		return nil, err
	}
	err = json.Unmarshal(body, servErr)
	if err != nil{
		return nil, err
	}
	return &servErr, nil
}