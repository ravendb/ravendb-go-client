package ravendb

import (
	"fmt"
	"net/http"
)

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
