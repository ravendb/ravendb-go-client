package ravendb

import (
	"fmt"
	"net/http"
	"time"
)

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

// ExecuteNewNextHiLoCommand executes NextHiLoResult command
func ExecuteNewNextHiLoCommand(exec CommandExecutorFunc, cmd *RavenCommand) (*HiLoResult, error) {
	var res HiLoResult
	err := excuteCmdAndJSONDecode(exec, cmd, &res)
	if err != nil {
		return nil, err
	}
	return &res, nil
}
