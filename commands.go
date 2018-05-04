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

// InternalServerError is retruned when server returns 500 Internal Server response
type InternalServerError struct {
	URL      string `json:"Url"`
	Type     string `json:"Type"`
	Message  string `json:"Message"`
	ErrorStr string `json:"Error"`
}

// Error makes it conform to error interface
func (e *InternalServerError) Error() string {
	return fmt.Sprintf(`Server returned 500 Internal Server for URL '%s'
Type: %s
Message: %s
Error: %s`, e.URL, e.Type, e.Message, e.ErrorStr)
}

// ServiceUnavailableError is returned when server returns 501 Service Unavailable
// response. This is additional information sent by the server.
type ServiceUnavailableError struct {
	Type    string `json:"Type"`
	Message string `json:"Message"`
}

// Error makes it conform to error interface
func (e *ServiceUnavailableError) Error() string {
	return fmt.Sprintf(`Server returned 501 Service Unavailable'
Type: %s
Message: %s`, e.Type, e.Message)
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
	if rsp.StatusCode == http.StatusBadRequest {
		var res BadRequestError
		err = decodeJSONFromReader(rsp.Body, &res)
		if err != nil {
			return nil, err
		}
		return nil, &res
	}

	// convert 503 Service Unavailable to ServiceUnavailableError
	if rsp.StatusCode == http.StatusServiceUnavailable {
		var res ServiceUnavailableError
		err = decodeJSONFromReader(rsp.Body, &res)
		if err != nil {
			return nil, err
		}
		return nil, &res
	}

	// convert 500 Internal Server to InternalServerError
	if rsp.StatusCode == http.StatusInternalServerError {
		var res InternalServerError
		err = decodeJSONFromReader(rsp.Body, &res)
		if err != nil {
			return nil, err
		}
		return nil, &res
	}

	// TODO: handle other server errors
	panicIf(rsp.StatusCode != http.StatusOK, "not handled status %d", rsp.StatusCode)

	return rsp, nil
}

// MakeSimpleExecutor creates a command executor talking to a given node
func MakeSimpleExecutor(n *ServerNode) CommandExecutorFunc {
	fn := func(cmd *RavenCommand, shouldRetry bool) (*http.Response, error) {
		return simpleExecutor(n, cmd, shouldRetry)
	}
	return fn
}

func excuteCmdAndJSONDecode(exec CommandExecutorFunc, cmd *RavenCommand, shouldRetry bool, v interface{}) error {
	rsp, err := ExecuteCommand(exec, cmd, shouldRetry)
	if err != nil {
		return err
	}
	defer rsp.Body.Close()

	if rsp.StatusCode == 200 {
		return decodeJSONFromReader(rsp.Body, v)
	}

	return nil
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

// NewGetClusterTopologyCommand creates a new GetClusterTopologyCommand
func NewGetClusterTopologyCommand() *RavenCommand {
	res := &RavenCommand{
		Method:        http.MethodGet,
		IsReadRequest: true,
		URLTemplate:   "{url}/cluster/topology",
	}
	return res
}

// ExecuteGetClusterTopologyCommand executes GetClusterTopologyCommand
func ExecuteGetClusterTopologyCommand(exec CommandExecutorFunc, cmd *RavenCommand, shouldRetry bool) (*ClusterTopologyResponse, error) {
	var res ClusterTopologyResponse
	err := excuteCmdAndJSONDecode(exec, cmd, shouldRetry, &res)
	if err != nil {
		return nil, err
	}
	return &res, nil
}

// DatabaseStatistics describes a result of GetStatisticsCommand
// https://sourcegraph.com/github.com/ravendb/ravendb-jvm-client@v4.0/-/blob/src/main/java/net/ravendb/client/documents/operations/DatabaseStatistics.java#L8:14
type DatabaseStatistics struct {
	LastDocEtag               int64 `json:"LastDocEtag"`
	CountOfIndexes            int64 `json:"CountOfIndexes"`
	CountOfDocuments          int64 `json:"CountOfDocuments"`
	CountOfRevisionDocuments  int64 `json:"CountOfRevisionDocuments"` // TODO: present in Java, not seen in JSON
	CountOfDocumentsConflicts int64 `json:"CountOfDocumentsConflicts"`
	CountOfTombstones         int64 `json:"CountOfTombstones"`
	CountOfConflicts          int64 `json:"CountOfConflicts"`
	CountOfAttachments        int64 `json:"CountOfAttachments"`
	CountOfUniqueAttachments  int64 `json:"CountOfUniqueAttachments"`

	Indexes []interface{} `json:"Indexes"` // TODO: this is []IndexInformation

	DatabaseChangeVector                     string      `json:"DatabaseChangeVector"`
	DatabaseID                               string      `json:"DatabaseId"`
	Is64Bit                                  bool        `json:"Is64Bit"`
	Pager                                    string      `json:"Pager"`
	LastIndexingTime                         interface{} `json:"LastIndexingTime"` // TODO: this is time, can be null so must be a pointer
	SizeOnDisk                               SizeOnDisk  `json:"SizeOnDisk"`
	NumberOfTransactionMergerQueueOperations int64       `json:"NumberOfTransactionMergerQueueOperations"`
}

// SizeOnDisk describes size of entity on disk
type SizeOnDisk struct {
	HumaneSize  string `json:"HumaneSize"`
	SizeInBytes int64  `json:"SizeInBytes"`
}

// TODO: add IndexInformation

// NewGetStatisticsCommand creates a new GetStatisticsCommand
// https://sourcegraph.com/github.com/ravendb/RavenDB-Python-Client@v4.0/-/blob/pyravendb/commands/raven_commands.py#L322
// https://sourcegraph.com/github.com/ravendb/ravendb-jvm-client@v4.0/-/blob/src/main/java/net/ravendb/client/documents/operations/GetStatisticsOperation.java#L12
func NewGetStatisticsCommand(debugTag string) *RavenCommand {
	url := "{url}/databases/{db}/stats"
	if debugTag != "" {
		url += "?" + debugTag
	}

	res := &RavenCommand{
		Method:      http.MethodGet,
		URLTemplate: url,
	}
	return res
}

// ExecuteGetStatisticsCommand executes GetStatisticsCommand
func ExecuteGetStatisticsCommand(exec CommandExecutorFunc, cmd *RavenCommand, shouldRetry bool) (*DatabaseStatistics, error) {
	var res DatabaseStatistics
	err := excuteCmdAndJSONDecode(exec, cmd, shouldRetry, &res)
	if err != nil {
		return nil, err
	}
	return &res, nil
}

// ServerNode describes a single server node
type ServerNode struct {
	URL        string `json:"Url"`
	ClusterTag string `json:"ClusterTag"`
	ServerRole string `json:"ServerRole"`
	Database   string `json:"Database"`
}

// Topology describes server nodes
// Result of
// {"Nodes":[{"Url":"http://localhost:9999","ClusterTag":"A","Database":"PyRavenDB","ServerRole":"Rehab"}],"Etag":10}
type Topology struct {
	Nodes []ServerNode `json:"Nodes"`
	Etag  int          `json:"Etag"`
}

// NewGetTopologyCommand creates a new GetClusterTopologyCommand
func NewGetTopologyCommand() *RavenCommand {
	res := &RavenCommand{
		Method:        http.MethodGet,
		IsReadRequest: true,
		URLTemplate:   "{url}/topology?name={db}",
	}
	return res
}

// ExecuteGetTopologyCommand executes GetClusterTopologyCommand
func ExecuteGetTopologyCommand(exec CommandExecutorFunc, cmd *RavenCommand, shouldRetry bool) (*Topology, error) {
	var res Topology
	err := excuteCmdAndJSONDecode(exec, cmd, shouldRetry, &res)
	if err != nil {
		return nil, err
	}
	return &res, nil
}

// GetDatabaseNamesResponse describes response of GetDatabaseNames command
type GetDatabaseNamesResponse struct {
	Databases []string `json:"Databases"`
}

// NewGetDatabaseNamesCommand creates a new GetClusterTopologyCommand
func NewGetDatabaseNamesCommand(start, pageSize int) *RavenCommand {
	url := fmt.Sprintf("{url}/databases?start=%d&pageSize=%d&namesOnly=true", start, pageSize)
	res := &RavenCommand{
		Method:        http.MethodGet,
		IsReadRequest: true,
		URLTemplate:   url,
	}
	return res
}

// ExecuteGetDatabaseNamesCommand executes GetClusterTopologyCommand
func ExecuteGetDatabaseNamesCommand(exec CommandExecutorFunc, cmd *RavenCommand, shouldRetry bool) (*GetDatabaseNamesResponse, error) {
	var res GetDatabaseNamesResponse
	err := excuteCmdAndJSONDecode(exec, cmd, shouldRetry, &res)
	if err != nil {
		return nil, err
	}
	return &res, nil
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
