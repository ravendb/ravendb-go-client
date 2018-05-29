package ravendb

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

// JSONArrayResult represents result of BatchCommand, which is array of JSON objects
// it's a type alias so that it doesn't need casting when json marshalling
type JSONArrayResult = []ObjectNode

// RavenCommand represents data needed to issue an HTTP command to the server
type RavenCommand struct {
	Method        string // GET, PUT etc.
	IsReadRequest bool
	// to create a full url, replace {url} and {db} with ServerNode.URL and
	// ServerNode.Database
	URLTemplate string
	// additional HTTP request headers
	Headers map[string]string
	Data    []byte

	failedNodes []*ServerNode
}

func (c *RavenCommand) addFailedNode(node *ServerNode) {
	c.failedNodes = append(c.failedNodes, node)
}

func (c *RavenCommand) isFailedWithNode(node *ServerNode) bool {
	for _, n := range c.failedNodes {
		if n == node {
			return true
		}
	}
	return false
}

// BuildFullURL creates a full url by filling in server address and database name
func (c *RavenCommand) BuildFullURL(n *ServerNode) string {
	url := strings.Replace(c.URLTemplate, "{url}", n.URL, -1)
	return strings.Replace(url, "{db}", n.Database, -1)
}

// CommandExecutorFunc takes RavenCommand, sends it over HTTP to the server and
// returns raw HTTP response
type CommandExecutorFunc func(cmd *RavenCommand) (*http.Response, error)

// ExecuteCommand executes RavenCommand with a given executor function
func ExecuteCommand(exec CommandExecutorFunc, cmd *RavenCommand) (*http.Response, error) {
	return exec(cmd)
}

// MakeSimpleExecutor creates a command executor talking to a given node
func MakeSimpleExecutor(n *ServerNode) CommandExecutorFunc {
	fn := func(cmd *RavenCommand) (*http.Response, error) {
		return simpleExecutor(n, cmd)
	}
	return fn
}

func excuteCmdWithEmptyResult(exec CommandExecutorFunc, cmd *RavenCommand) error {
	rsp, err := ExecuteCommand(exec, cmd)
	if err != nil {
		return err
	}
	rsp.Body.Close()

	// expectes 204 No Content
	if rsp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("Returned unexpected status code %d (expected 204)", rsp.StatusCode)
	}
	return nil

}

func excuteCmdAndJSONDecode(exec CommandExecutorFunc, cmd *RavenCommand, v interface{}) error {
	rsp, err := ExecuteCommand(exec, cmd)
	if err != nil {
		return err
	}
	if rsp == nil {
		return nil
	}

	// ok: 200, created: 201
	if rsp.StatusCode == http.StatusOK || rsp.StatusCode == http.StatusCreated {
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
func ExecuteGetClusterTopologyCommand(exec CommandExecutorFunc, cmd *RavenCommand) (*ClusterTopologyResponse, error) {
	var res ClusterTopologyResponse
	err := excuteCmdAndJSONDecode(exec, cmd, &res)
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
func ExecuteGetStatisticsCommand(exec CommandExecutorFunc, cmd *RavenCommand) (*DatabaseStatistics, error) {
	var res DatabaseStatistics
	err := excuteCmdAndJSONDecode(exec, cmd, &res)
	if err != nil {
		return nil, err
	}
	return &res, nil
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
func ExecuteGetTopologyCommand(exec CommandExecutorFunc, cmd *RavenCommand) (*Topology, error) {
	var res Topology
	err := excuteCmdAndJSONDecode(exec, cmd, &res)
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
func ExecuteGetDatabaseNamesCommand(exec CommandExecutorFunc, cmd *RavenCommand) (*GetDatabaseNamesResponse, error) {
	var res GetDatabaseNamesResponse
	err := excuteCmdAndJSONDecode(exec, cmd, &res)
	if err != nil {
		return nil, err
	}
	return &res, nil
}

// DeleteDatabaseResult represents result of Delete Database command
// https://sourcegraph.com/github.com/ravendb/ravendb-jvm-client@v4.0/-/blob/src/main/java/net/ravendb/client/serverwide/operations/DeleteDatabaseResult.java#L3
type DeleteDatabaseResult struct {
	RaftCommandIndex int `json:"RaftCommandIndex"`
	//PendingDeletes   []string `json:"PendingDeletes"` // TODO: not send back in python test
}

// NewDeleteDatabaseCommand creates a new DeleteDatabaseCommand.
// fromNode can be empty string
// https://sourcegraph.com/github.com/ravendb/RavenDB-Python-Client@v4.0/-/blob/pyravendb/raven_operations/server_operations.py#L79
// TODO: add timeToWaitForConfirmation time.Duration? Don't know the format,
// python example sends null
func NewDeleteDatabaseCommand(dbName string, hardDelete bool, fromNode string) *RavenCommand {

	opts := map[string]interface{}{}
	opts["DatabaseNames"] = []string{dbName}
	opts["HardDelete"] = hardDelete
	if fromNode != "" {
		opts["FromNodes"] = []string{fromNode}
	}

	data, err := json.Marshal(opts)
	must(err)
	res := &RavenCommand{
		Method:      http.MethodDelete,
		URLTemplate: "{url}/admin/databases",
		Data:        data,
	}
	return res
}

// ExecuteDeleteDatabaseCommand executes CreateDatabaseCommand
func ExecuteDeleteDatabaseCommand(exec CommandExecutorFunc, cmd *RavenCommand) (*DeleteDatabaseResult, error) {
	var res DeleteDatabaseResult
	err := excuteCmdAndJSONDecode(exec, cmd, &res)
	if err != nil {
		return nil, err
	}
	return &res, nil
}

// GetOperationStateCommandResult describes a result of GetOperationsState
type GetOperationStateCommandResult struct {
	ErrorStr string `json:"Error"`
	Status   string `json:"Status"`
}

// NewGetOperationStateCommand creates GetOperationsState command
// https://sourcegraph.com/github.com/ravendb/RavenDB-Python-Client/-/blob/pyravendb/commands/raven_commands.py#L371
// https://sourcegraph.com/github.com/ravendb/ravendb-jvm-client/-/blob/src/main/java/net/ravendb/client/documents/operations/GetOperationStateOperation.java#L14
// TODO: add isServerStoreOp bool? Is in Python, not in Java
func NewGetOperationStateCommand(opID string) *RavenCommand {
	res := &RavenCommand{
		Method:      http.MethodGet,
		URLTemplate: "{url}/operations/state?id=" + opID,
	}
	return res
}

// ExecuteGetOperationStateCommand executes GetOperationsState command
func ExecuteGetOperationStateCommand(exec CommandExecutorFunc, cmd *RavenCommand) (*GetOperationStateCommandResult, error) {
	var res GetOperationStateCommandResult
	err := excuteCmdAndJSONDecode(exec, cmd, &res)
	if err != nil {
		return nil, err
	}
	return &res, nil
}

// PutResult describes result of PutDocumentCommand
// https://sourcegraph.com/github.com/ravendb/ravendb-jvm-client@v4.0/-/blob/src/main/java/net/ravendb/client/documents/commands/batches/PutResult.java#L6
type PutResult struct {
	ID           string `json:"Id"`
	ChangeVector string `json:"ChangeVector"`
}

func addChangeVectorIfNotEmpty(cmd *RavenCommand, changeVector string) {
	if changeVector != "" {
		if cmd.Headers == nil {
			cmd.Headers = map[string]string{}
		}
		cmd.Headers["If-Match"] = fmt.Sprintf(`"%s"`, changeVector)
	}
}

// NewPutDocumentJSONCommand creates a command for PutDocument operation
// TODO: should I validatte js is a valid json?
func NewPutDocumentJSONCommand(key string, js []byte, changeVector string) *RavenCommand {
	panicIf(key == "", "key can't be empty string")
	res := &RavenCommand{
		Method:      http.MethodPut,
		URLTemplate: "{url}/databases/{db}/docs?id=" + quoteKey(key),
	}
	addChangeVectorIfNotEmpty(res, changeVector)
	res.Data = js
	return res
}

// NewPutDocumentRawCommand creates a command for PutDocument operation
func NewPutDocumentRawCommand(key string, doc map[string]interface{}, changeVector string) *RavenCommand {
	js, err := json.Marshal(doc)
	must(err) // TODO: return an error
	return NewPutDocumentJSONCommand(key, js, changeVector)
}

// ExecutePutDocumentRawCommand executes PutDocument command
func ExecutePutDocumentRawCommand(exec CommandExecutorFunc, cmd *RavenCommand) (*PutResult, error) {
	var res PutResult
	err := excuteCmdAndJSONDecode(exec, cmd, &res)
	if err != nil {
		return nil, err
	}
	return &res, nil
}

func isGetDocumentPost(keys []string) bool {
	maxKeySize := 1024
	size := 0
	for _, key := range keys {
		size += len(key)
		if size > maxKeySize {
			return true
		}
	}
	return false
}

// NewGetDocumentCommand creates a command for GetDocument operation
// https://sourcegraph.com/github.com/ravendb/RavenDB-Python-Client@v4.0/-/blob/pyravendb/commands/raven_commands.py#L52:7
//https://sourcegraph.com/github.com/ravendb/ravendb-jvm-client@v4.0/-/blob/src/main/java/net/ravendb/client/documents/commands/GetDocumentsCommand.java#L37
// TODO: java has start/pageSize
func NewGetDocumentCommand(keys []string, includes []string, metadataOnly bool) *RavenCommand {
	panicIf(len(keys) == 0, "must provide at least one key") // TODO: return an error?
	res := &RavenCommand{
		Method: http.MethodGet,
	}
	path := "docs?"
	for _, s := range includes {
		path += "&include=" + quoteKey(s)
	}
	if metadataOnly {
		path += "&metadataOnly=true"
	}

	if isGetDocumentPost(keys) {
		res.Method = http.MethodPost
		js, err := json.Marshal(keys)
		must(err)
		res.Data = js
	} else {
		for _, s := range keys {
			path += "&id=" + quoteKey(s)
		}
	}
	res.URLTemplate = "{url}/databases/{db}/" + path
	return res
}

// GetDocumentResult is a result of GetDocument command
// https://sourcegraph.com/github.com/ravendb/ravendb-jvm-client@v4.0/-/blob/src/main/java/net/ravendb/client/documents/commands/GetDocumentsResult.java#L6:14
type GetDocumentResult struct {
	Includes      map[string]ObjectNode `json:"Includes"`
	Results       JSONArrayResult       `json:"Results"`
	NextPageStart int                   `json:"NextPageStart"`
}

// ExecuteGetDocumentCommand executes GetDocument command
func ExecuteGetDocumentCommand(exec CommandExecutorFunc, cmd *RavenCommand) (*GetDocumentResult, error) {
	var res GetDocumentResult
	err := excuteCmdAndJSONDecode(exec, cmd, &res)
	if err != nil {
		return nil, err
	}
	return &res, nil
}

// NewDeleteDocumentCommand creates DeleteDocument command
func NewDeleteDocumentCommand(key string, changeVector string) *RavenCommand {
	url := fmt.Sprintf("{url}/databases/{db}/docs?id=%s", quoteKey(key))
	res := &RavenCommand{
		Method:      http.MethodDelete,
		URLTemplate: url,
	}
	addChangeVectorIfNotEmpty(res, changeVector)
	return res
}

// ExecuteDeleteDocumentCommand executes DeleteDocument command
func ExecuteDeleteDocumentCommand(exec CommandExecutorFunc, cmd *RavenCommand) error {
	return excuteCmdWithEmptyResult(exec, cmd)
}

// NewBatchCommand creates a new batch command
// https://sourcegraph.com/github.com/ravendb/RavenDB-Python-Client@v4.0/-/blob/pyravendb/commands/raven_commands.py#L172
func NewBatchCommand(commands []*CommandData) *RavenCommand {
	var data []map[string]interface{}
	for _, command := range commands {
		if command.typ == "AttachmentPUT" {
			// TODO: handle AttachmentPUT and set files
			panicIf(true, "NYI")
		}
		data = append(data, command.json)
	}
	v := map[string]interface{}{
		"Commands": data,
	}
	js, err := json.Marshal(v)
	must(err)
	res := &RavenCommand{
		Method:      http.MethodPost,
		URLTemplate: "{url}/databases/{db}/bulk_docs",
		Data:        js,
	}
	return res
}

// BatchCommandResult describes server's JSON response to batch command
type BatchCommandResult struct {
	Results JSONArrayResult `json:"Results"`
}

// ExecuteBatchCommand executes batch command
// https://sourcegraph.com/github.com/ravendb/RavenDB-Python-Client@v4.0/-/blob/pyravendb/commands/raven_commands.py#L196
// TODO: maybe more
func ExecuteBatchCommand(exec CommandExecutorFunc, cmd *RavenCommand) (JSONArrayResult, error) {
	var res BatchCommandResult
	err := excuteCmdAndJSONDecode(exec, cmd, &res)
	if err != nil {
		return nil, err
	}
	return res.Results, nil
}

/* Done:
GetDocumentCommand
DeleteDocumentCommand
PutDocumentCommand

GetStatisticsCommand
GetTopologyCommand
GetClusterTopologyCommand
GetOperationStateCommand

// server_operations.py
_CreateDatabaseCommand
_DeleteDatabaseCommand
_GetDatabaseNamesCommand

// hilo_generator.py
HiLoReturnCommand
NextHiLoCommand

// raven_commands.py
BatchCommand

*/

/*
PutCommandData
DeleteCommandData
PatchCommandData
PutAttachmentCommandData
DeleteAttachmentCommandData

Commands to implement:

// raven_commands.py
DeleteIndexCommand
PatchCommand
QueryCommand
PutAttachmentCommand
GetFacetsCommand
MultiGetCommand
GetDatabaseRecordCommand
WaitForRaftIndexCommand - maybe not, only in python client
GetTcpInfoCommand
QueryStreamCommand

CreateSubscriptionCommand
DeleteSubscriptionCommand
DropSubscriptionConnectionCommand
GetSubscriptionsCommand
GetSubscriptionStateCommand

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
_GetCertificateCommand
_CreateClientCertificateCommand
_PutClientCertificateCommand
_DeleteCertificateCommand
*/
