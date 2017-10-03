package errors

import (
	ravenHttp "../http"
)

type AllTopologyNodesDownError struct{
	Topology ravenHttp.Topology
	error string
}

func NewAllTopologyNodesDownError(error string, topology ravenHttp.Topology) (*AllTopologyNodesDownError, error){
	return &AllTopologyNodesDownError{Topology: topology, error: error}, nil
}

func (err AllTopologyNodesDownError) Error() string{
	return err.error
}