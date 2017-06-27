package http

import (
	"../data"
	"net/http"
	"../tools"
)

type RequestExecutor struct{

	url string
	ServerNode ServerNode
	database string
	apiKey string
	convention data.DocumentConvention
	topology Topology
	IsFirstTryToLoadFromTopologyCache bool
	VersionInfo string
	Headers []http.Header
	TopologyChangeCounter uint
	RequestCount uint
	authenticator tools.Authenticator
}