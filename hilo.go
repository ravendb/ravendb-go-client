package ravendb

import (
	"fmt"
	"net/http"
	"time"
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

const (
	// Python does "0001-01-01 00:00:00"
	// Java sends more complicated format https://sourcegraph.com/github.com/ravendb/ravendb-jvm-client@v4.0/-/blob/src/main/java/net/ravendb/client/primitives/NetISO8601Utils.java#L8
	timeFormat = "2006-02-01 15:04:05"
)

// NewNextHiLoCommand creates a NextHiLoCommand
func NewNextHiLoCommand(tag string, lastBatchSize int, lastRangeAt time.Time, identityPartsSeparator string, lastRangeMax int) *RavenCommand {
	lastRangeAtStr := quoteKey(lastRangeAt.Format(timeFormat))
	path := fmt.Sprintf("hilo/next?tag=%s&lastBatchSize=%d&lastRangeAt=%s&identityPartsSeparator=%s&lastMax=%d", tag, lastBatchSize, lastRangeAtStr, identityPartsSeparator, lastRangeMax)
	url := "{url}/databases/{db}/" + path
	res := &RavenCommand{
		Method:      http.MethodGet,
		URLTemplate: url,
	}
	return res
}

// NextHiLoResult is a result of NextHiLoResult command
type NextHiLoResult struct {
	Prefix      string `json:"Prefix"`
	Low         int    `json:"Low"`
	High        int    `json:"High"`
	LastSize    int    `json:"LastSize"`
	ServerTag   string `json:"ServerTag"`
	LastRangeAt string `json:"LastRangeAt"`
}

const (
	// time format returned by the server
	// 2018-05-08T05:20:31.5233900Z
	serverTimeFormat = "2006-01-02T15:04:05.999999999Z"
)

// GetLastRangeAt parses LastRangeAt which is in a format:
// 2018-05-08T05:20:31.5233900Z
func (r *NextHiLoResult) GetLastRangeAt() time.Time {
	t, err := time.Parse(serverTimeFormat, r.LastRangeAt)
	must(err) // TODO: should silently fail? return an error?
	return t
}

// ExecuteNewNextHiLoCommand executes NextHiLoResult command
func ExecuteNewNextHiLoCommand(exec CommandExecutorFunc, cmd *RavenCommand) (*NextHiLoResult, error) {
	var res NextHiLoResult
	err := excuteCmdAndJSONDecode(exec, cmd, &res)
	if err != nil {
		return nil, err
	}
	return &res, nil
}
