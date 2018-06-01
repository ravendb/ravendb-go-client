package ravendb

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"
)

type _NextHiLoCommand struct {
	_tag                    String
	_lastBatchSize          int
	_lastRangeAt            *time.Time
	_identityPartsSeparator String
	_lastRangeMax           int
}

func NewNextHiLoCommand(tag String, lastBatchSize int, lastRangeAt *time.Time, identityPartsSeparator String, lastRangeMax int) *RavenCommand {
	panicIf(tag == "", "tag cannot be empty")
	panicIf(identityPartsSeparator == "", "identityPartsSeparator cannot be empty")
	data := &_NextHiLoCommand{
		_tag:                    tag,
		_lastBatchSize:          lastBatchSize,
		_lastRangeAt:            lastRangeAt,
		_identityPartsSeparator: identityPartsSeparator,
		_lastRangeMax:           lastRangeMax,
	}

	cmd := NewRavenCommand()
	cmd.IsReadRequest = true
	cmd.data = data
	cmd.setResponseFunc = NextHiLoCommand_setResponse
	cmd.createRequestFunc = NextHiLoCommand_createRequest

	return cmd
}

func NextHiLoCommand_createRequest(cmd *RavenCommand, node *ServerNode) (*http.Request, error) {

	data := cmd.data.(*_NextHiLoCommand)
	date := ""
	if data._lastRangeAt != nil {
		date = (*data._lastRangeAt).Format(serverTimeFormat)
	}
	path := "/hilo/next?tag=" + data._tag + "&lastBatchSize=" + strconv.Itoa(data._lastBatchSize) + "&lastRangeAt=" + date + "&identityPartsSeparator=" + data._identityPartsSeparator + "&lastMax=" + strconv.Itoa(data._lastRangeMax)
	url := node.getUrl() + "/databases/" + node.getDatabase() + path
	return NewHttpGet(url)
}

func NextHiLoCommand_setResponse(cmd *RavenCommand, response String, fromCache bool) error {
	var res HiLoResult
	err := json.Unmarshal([]byte(response), &res)
	if err != nil {
		return err
	}
	cmd.result = &res
	return nil
}
