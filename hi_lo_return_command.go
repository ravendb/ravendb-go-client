package ravendb

import (
	"fmt"
	"net/http"
)

// NewHiLoReturnCommand creates a HiLoReturn command
func NewHiLoReturnCommand(tag string, last, end int) *RavenCommand {
	path := fmt.Sprintf("hilo/return?tag=%s&end=%d&last=%d", tag, end, last)
	url := "{url}/databases/{db}/" + path
	res := &RavenCommand{
		Method:      http.MethodPut,
		URLTemplate: url,
	}
	return res
}

// ExecuteHiLoReturnCommand executes HiLoReturnCommand
func ExecuteHiLoReturnCommand(exec CommandExecutorFunc, cmd *RavenCommand) error {
	return excuteCmdWithEmptyResult(exec, cmd)
}
