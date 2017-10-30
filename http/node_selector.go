package http

import "github.com/ravendb-go-client/http/server_nodes"

type NodeSelector struct{
	Topology Topology
	CurrentNode server_nodes.IServerNode
	CurrentNodeIndex int
}

func NewNodeSelector(topology *Topology) (*NodeSelector, error){
	return &NodeSelector{*topology, topology.Nodes[0], 0}, nil
}

func (selector NodeSelector) OnUpdateTopology(topology *Topology) bool{
	//todo
	return true
}

func (selector NodeSelector) OnFailedRequest(node server_nodes.IServerNode) bool{
	//todo
	return true
}

func (selector NodeSelector) GetCurrentNode() server_nodes.IServerNode{
	return selector.CurrentNode
}

func (selector NodeSelector) RestoreNodeIndex(index int) error{
	//todo
	return nil
}

