package ravendb

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// JSONArrayResult represents result of BatchCommand, which is array of JSON objects
// it's a type alias so that it doesn't need casting when json marshalling
type JSONArrayResult = []ObjectNode

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

/*
func addChangeVectorIfNotEmpty(cmd *RavenCommand, changeVector string) {
	if changeVector != "" {
		if cmd.Headers == nil {
			cmd.Headers = map[string]string{}
		}
		cmd.Headers["If-Match"] = fmt.Sprintf(`"%s"`, changeVector)
	}
}
*/

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
	//addChangeVectorIfNotEmpty(res, changeVector)
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
