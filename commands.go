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
