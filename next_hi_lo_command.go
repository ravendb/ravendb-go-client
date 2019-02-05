package ravendb

import (
	"net/http"
	"time"
)

var (
	_ RavenCommand = &NextHiLoCommand{}
)

// NextHiLoCommand represents a hi-lo database command
type NextHiLoCommand struct {
	RavenCommandBase

	_tag                    string
	_lastBatchSize          int64
	_lastRangeAt            *time.Time // TODO: our Time?
	_identityPartsSeparator string
	_lastRangeMax           int64

	Result *HiLoResult
}

//NewNextHiLoCommand returns new NextHiLoCommand
func NewNextHiLoCommand(tag string, lastBatchSize int64, lastRangeAt *time.Time, identityPartsSeparator string, lastRangeMax int64) *NextHiLoCommand {
	panicIf(tag == "", "tag cannot be empty")
	panicIf(identityPartsSeparator == "", "identityPartsSeparator cannot be empty")
	cmd := &NextHiLoCommand{
		RavenCommandBase: NewRavenCommandBase(),

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
		date = (*c._lastRangeAt).Format(timeFormat)
	}
	path := "/hilo/next?tag=" + c._tag + "&lastBatchSize=" + i64toa(c._lastBatchSize) + "&lastRangeAt=" + date + "&identityPartsSeparator=" + c._identityPartsSeparator + "&lastMax=" + i64toa(c._lastRangeMax)
	url := node.URL + "/databases/" + node.Database + path
	return newHttpGet(url)
}

func (c *NextHiLoCommand) setResponse(response []byte, fromCache bool) error {
	return jsonUnmarshal(response, &c.Result)
}
