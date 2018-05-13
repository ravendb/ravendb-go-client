package http

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/ravendb/ravendb-go-client/http/server_nodes"
)

type AllTopologyNodesDownError struct {
	error    string
	Topology Topology
}

func NewAllTopologyNodesDownError(error string, topology Topology) (*AllTopologyNodesDownError, error) {
	return &AllTopologyNodesDownError{Topology: topology, error: error}, nil
}

func (err AllTopologyNodesDownError) Error() string {
	return err.error
}

type TopologyUpdateError struct {
	error     string
	ErrorList map[string]error
}

func (topologyUpdateError TopologyUpdateError) Error() string {
	return topologyUpdateError.error
}

type ErrorResponseError struct {
}

func NewErrorResponseError() (*ErrorResponseError, error) {
	return &ErrorResponseError{}, nil
}

func (err ErrorResponseError) Error() string {
	return fmt.Sprintf("Failed to put document in the database please check the connection to the server")
}

// UnsuccessfulRequestError describes unsuccessful request error
type UnsuccessfulRequestError struct {
	Url        string
	FailedNode server_nodes.IServerNode
}

// NewUnsuccessfulRequestError creates new UnsuccessfulRequestError
func NewUnsuccessfulRequestError(url string, node server_nodes.IServerNode) (*UnsuccessfulRequestError, error) {
	return &UnsuccessfulRequestError{FailedNode: node, Url: url}, nil
}

func (err UnsuccessfulRequestError) Error() string {
	return fmt.Sprintf("Request to %s on node %s failed", err.Url, err.FailedNode.GetClusterTag())
}

// ServerError describes and error json returned by the server
type ServerError struct {
	URL      string `json:"Url"`
	Type     string `json:"Type"`
	Message  string `json:"Message"`
	ErrorStr string `json:"Error"`
}

// NewServerError creates ServerError from server HTTP response
func NewServerError(rsp http.Response) (*ServerError, error) {
	var servErr ServerError
	body, err := ioutil.ReadAll(rsp.Body)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(body, servErr)
	if err != nil {
		return nil, err
	}
	return &servErr, nil
}
