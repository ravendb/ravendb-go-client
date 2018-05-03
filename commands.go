package ravendb

import (
	"errors"
	"net/http"
)

// RavenCommand represents data needed to issue an HTTP command to the server
type RavenCommand struct {
}

// CommandExecutorFunc takes RavenCommand, sends it over HTTP to the server and
// returns raw HTTP response
type CommandExecutorFunc func(cmd *RavenCommand, shouldRetry bool) (*http.Response, error)

// ExecuteCommand executes RavenCommand with a given executor function
func ExecuteCommand(exec CommandExecutorFunc, cmd *RavenCommand, shouldRetry bool) (*http.Response, error) {
	return exec(cmd, shouldRetry)
}

// ExecuteGetClusterTopologyCommand executes GetClusterTopologyCommand
func ExecuteGetClusterTopologyCommand(exec CommandExecutorFunc, cmd *RavenCommand, shouldRetry bool) (*ClusterTopologyResponse, error) {
	return nil, errors.New("NYI")
}

// ServerNode describes a single server node
type ServerNode struct {
	URL        string `json:"Url"`
	ClusterTag string
	Database   string
}

// Topology describes server nodes
// Result of
// {"Nodes":[{"Url":"http://localhost:9999","ClusterTag":"A","Database":"PyRavenDB","ServerRole":"Rehab"}],"Etag":10}
type Topology struct {
	Nodes      []ServerNode
	ServerRole string
	Etag       int
}

// ClusterTopology is a part of ClusterTopologyResponse
// https://sourcegraph.com/github.com/ravendb/ravendb-jvm-client@v4.0/-/blob/src/main/java/net/ravendb/client/http/ClusterTopology.java#L6
type ClusterTopology struct {
	LastNodeID string `json:"LastNodeId"`
	TopologyID string `json:"TopologyId"`

	Members     map[string]string
	Promotables map[string]string
	Watchers    map[string]string
}

// GetAllNodes returns all nodes
// https://sourcegraph.com/github.com/ravendb/ravendb-jvm-client@v4.0/-/blob/src/main/java/net/ravendb/client/http/ClusterTopology.java#L46
func (t *ClusterTopology) GetAllNodes() map[string]string {
	// TODO: implement me
	panicIf(true, "NYI")
	return nil
}

// ClusterTopologyResponse is a response of GetClusterTopologyCommand
// https://sourcegraph.com/github.com/ravendb/ravendb-jvm-client@v4.0/-/blob/src/main/java/net/ravendb/client/http/ClusterTopologyResponse.java#L3
// Sample response:
// {"Topology":{"TopologyId":"8bf47de1-601e-4fff-b300-2e2c07ab6822","AllNodes":{"A":"http://localhost:9999"},"Members":{"A":"http://localhost:9999"},"Promotables":{},"Watchers":{},"LastNodeId":"A"},"Leader":"A","LeaderShipDuration":61407928,"CurrentState":"Leader","NodeTag":"A","CurrentTerm":4,"NodeLicenseDetails":{"A":{"UtilizedCores":3,"NumberOfCores":8,"InstalledMemoryInGb":16.0,"UsableMemoryInGb":16.0}},"LastStateChangeReason":"Leader, I'm the only one in thecluster, so I'm the leader (at 5/2/18 7:49:23 AM)","Status":{}}
type ClusterTopologyResponse struct {
	Topology *ClusterTopology `json:"Topology"`
	Leader   string           `json:"Leader"`
	NodeTag  string           `json:"NodeTag"`
	// note: the response returns more info
}

/*
PutCommandData
DeleteCommandData
PatchCommandData
PutAttachmentCommandData
DeleteAttachmentCommandData


Commands to implement:

// raven_commands.py
GetDocumentCommand
DeleteDocumentCommand
PutDocumentCommand
BatchCommand
DeleteIndexCommand
PatchCommand
QueryCommand
GetStatisticsCommand
GetTopologyCommand
GetClusterTopologyCommand
GetOperationStateCommand
PutAttachmentCommand
GetFacetsCommand
MultiGetCommand
GetDatabaseRecordCommand
WaitForRaftIndexCommand
GetTcpInfoCommand
QueryStreamCommand

CreateSubscriptionCommand
DeleteSubscriptionCommand
DropSubscriptionConnectionCommand
GetSubscriptionsCommand
GetSubscriptionStateCommand

// hilo_generator.py
HiLoReturnCommand
NextHiLoCommand

// maintenance_operations.py
_DeleteIndexCommand
_GetIndexCommand
_GetIndexNamesCommand
_PutIndexesCommand

// operations.py
_DeleteAttachmentCommand
_PatchByQueryCommand
_DeleteByQueryCommand
_GetAttachmentCommand
_GetMultiFacetsCommand

// server_operations.py
_CreateDatabaseCommand
_DeleteDatabaseCommand
_GetDatabaseNamesCommand

_GetCertificateCommand
_CreateClientCertificateCommand
_PutClientCertificateCommand
_DeleteCertificateCommand

*/
