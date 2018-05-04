package ravendb

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// RavenCommand represents data needed to issue an HTTP command to the server
type RavenCommand struct {
	Method        string // GET, PUT etc.
	IsReadRequest bool
	// to create a full url, replace {url} and {db} with ServerNode.URL and
	// ServerNode.Database
	URLTemplate string
	// additional HTTP request headers
	Headers map[string]string
}

// BadRequestError is returned when server returns 400 Bad Request response
// This is additional information sent by the server
type BadRequestError struct {
	URL      string `json:"Url"`
	Type     string `json:"Type"`
	Message  string `json:"Message"`
	ErrorStr string `json:"Error"`
}

// Error makes it conform to error interface
func (e *BadRequestError) Error() string {
	return fmt.Sprintf(`Server returned 400 Bad Request for URL '%s'
Type: %s
Message: %s
Error: %s`, e.URL, e.Type, e.Message, e.ErrorStr)
}

// NewGetClusterTopologyCommand creates a new GetClusterTopologyCommand
func NewGetClusterTopologyCommand() *RavenCommand {
	res := &RavenCommand{
		Method:        http.MethodGet,
		IsReadRequest: true,
		URLTemplate:   "{url}/cluster/topology",
	}
	return res
}

// CommandExecutorFunc takes RavenCommand, sends it over HTTP to the server and
// returns raw HTTP response
type CommandExecutorFunc func(cmd *RavenCommand, shouldRetry bool) (*http.Response, error)

// ExecuteCommand executes RavenCommand with a given executor function
func ExecuteCommand(exec CommandExecutorFunc, cmd *RavenCommand, shouldRetry bool) (*http.Response, error) {
	return exec(cmd, shouldRetry)
}

func decodeJSONFromReader(r io.Reader, v interface{}) error {
	return json.NewDecoder(r).Decode(v)
}

// ExecuteGetClusterTopologyCommand executes GetClusterTopologyCommand
func ExecuteGetClusterTopologyCommand(exec CommandExecutorFunc, cmd *RavenCommand, shouldRetry bool) (*ClusterTopologyResponse, error) {
	rsp, err := ExecuteCommand(exec, cmd, shouldRetry)
	if err != nil {
		return nil, err
	}
	defer rsp.Body.Close()

	if rsp.StatusCode == 200 {
		var res ClusterTopologyResponse
		err = decodeJSONFromReader(rsp.Body, &res)
		if err != nil {
			return nil, err
		}
		return &res, nil
	}

	return nil, nil
}

// TODO: do I need to explicitly enable compression or does the client does
// it by default? It seems to send Accept-Encoding: gzip by default
func simpleExecutor(n *ServerNode, cmd *RavenCommand, shouldRetry bool) (*http.Response, error) {
	client := &http.Client{
		Timeout: time.Second * 5,
	}
	url := strings.Replace(cmd.URLTemplate, "{url}", n.URL, -1)
	url = strings.Replace(url, "{db}", n.Database, -1)
	req, err := http.NewRequest(cmd.Method, url, nil)
	if err != nil {
		return nil, err
	}
	for k, v := range cmd.Headers {
		req.Header.Add(k, v)
	}
	req.Header.Add("User-Agent", "ravendb-go-client/1.0")
	req.Header.Add("Raven-Client-Version", "4.0.0.0")
	req.Header.Add("Accept", "application/json")
	panicIf(n.ClusterTag == "", "ClusterTag is empty string in %v", n)
	// TODO: do I need to quote the tag? Python client
	etag := fmt.Sprintf(`"%s"`, n.ClusterTag)
	req.Header.Add("Topology-Etag", etag)
	rsp, err := client.Do(req)
	// this is for network-level errors when we don't get response
	if err != nil {
		return rsp, err
	}
	// we have response but it could be one of the error server response

	// convert 400 Bad Request response to BadReqeustError
	if rsp.StatusCode == 400 {
		var res BadRequestError
		err = decodeJSONFromReader(rsp.Body, &res)
		if err != nil {
			return nil, err
		}
		return nil, &res
	}
	return rsp, nil
}

// MakeSimpleExecutor creates a command executor talking to a given node
func MakeSimpleExecutor(n *ServerNode) CommandExecutorFunc {
	fn := func(cmd *RavenCommand, shouldRetry bool) (*http.Response, error) {
		return simpleExecutor(n, cmd, shouldRetry)
	}
	return fn
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

	// Those map name like A to server url like http://localhost:9999
	Members     map[string]string
	Promotables map[string]string
	Watchers    map[string]string
}

// GetAllNodes returns all nodes
// https://sourcegraph.com/github.com/ravendb/ravendb-jvm-client@v4.0/-/blob/src/main/java/net/ravendb/client/http/ClusterTopology.java#L46
func (t *ClusterTopology) GetAllNodes() map[string]string {
	res := map[string]string{}
	for name, uri := range t.Members {
		res[name] = uri
	}
	for name, uri := range t.Promotables {
		res[name] = uri
	}
	for name, uri := range t.Watchers {
		res[name] = uri
	}
	return res
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
	// see https://app.quicktype.io?share=pzquGxXJcXyMncfA9JPa for fuller definition
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
