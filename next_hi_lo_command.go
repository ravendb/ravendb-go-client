package ravendb

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"
)

var (
	_ RavenCommand = &NextHiLoCommand{}
)

type NextHiLoCommand struct {
	*RavenCommandBase
	_tag                    String
	_lastBatchSize          int
	_lastRangeAt            *time.Time
	_identityPartsSeparator String
	_lastRangeMax           int

	Result *HiLoResult
}

func NewNextHiLoCommand(tag String, lastBatchSize int, lastRangeAt *time.Time, identityPartsSeparator String, lastRangeMax int) *NextHiLoCommand {
	panicIf(tag == "", "tag cannot be empty")
	panicIf(identityPartsSeparator == "", "identityPartsSeparator cannot be empty")
	cmd := &NextHiLoCommand{
		RavenCommandBase:        NewRavenCommandBase(),
		_tag:                    tag,
		_lastBatchSize:          lastBatchSize,
		_lastRangeAt:            lastRangeAt,
		_identityPartsSeparator: identityPartsSeparator,
		_lastRangeMax:           lastRangeMax,
	}
	cmd.IsReadRequest = true
	return cmd
}

func (c *NextHiLoCommand) createRequest(node *ServerNode) (*http.Request, error) {

	date := ""
	if c._lastRangeAt != nil && !c._lastRangeAt.IsZero() {
		date = (*c._lastRangeAt).Format(serverTimeFormat)
	}
	path := "/hilo/next?tag=" + c._tag + "&lastBatchSize=" + strconv.Itoa(c._lastBatchSize) + "&lastRangeAt=" + date + "&identityPartsSeparator=" + c._identityPartsSeparator + "&lastMax=" + strconv.Itoa(c._lastRangeMax)
	url := node.getUrl() + "/databases/" + node.getDatabase() + path
	return NewHttpGet(url)
}

func (c *NextHiLoCommand) setResponse(response String, fromCache bool) error {
	var res HiLoResult
	err := json.Unmarshal([]byte(response), &res)
	if err != nil {
		return err
	}
	c.Result = &res
	return nil
}
