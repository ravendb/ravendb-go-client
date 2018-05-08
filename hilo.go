package ravendb

import (
	"fmt"
	"net/http"
)

// RangeValue represents an inclusive integer range min to max
type RangeValue struct {
	MinID   int
	MaxID   int
	Current int
}

// NewRangeValue creates a new RangeValue
func NewRangeValue(minID int, maxID int) *RangeValue {
	return &RangeValue{
		MinID:   minID,
		MaxID:   maxID,
		Current: minID - 1,
	}
}

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

// NewHiLoCommand creates a HiLoCommand
func NewHiLoCommand(tag string, lastBatchSize int, lastRangeAt int, identityPartsSeparator string, lastRangeMax int) *RavenCommand {
	url := ""
	res := &RavenCommand{
		Method:      http.MethodGet,
		URLTemplate: url,
	}
	return res
}
