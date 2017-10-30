package http

import (
	"github.com/ravendb-go-client/http/server_nodes"
	"encoding/json"
)

type Topology struct{
	Etag int64
	Nodes []server_nodes.IServerNode
}

func NewTopology(etag int64, nodes []server_nodes.IServerNode) (*Topology, error){
	return &Topology{Etag:etag, Nodes:nodes}, nil
}

func CreateFromJSON(data []byte) (*Topology, error){
	var topology Topology
	err := json.Unmarshal(data, topology)
	return &topology, err
}