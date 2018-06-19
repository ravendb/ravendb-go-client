package ravendb

import (
	"encoding/json"
	"net/http"
	"strconv"
)

var (
	_ RavenCommand = &GetRevisionsBinEntryCommand{}
)

type GetRevisionsBinEntryCommand struct {
	*RavenCommandBase

	_etag     int
	_pageSize int

	Result *ArrayNode
}

func NewGetRevisionsBinEntryCommand(etag int, pageSize int) *GetRevisionsBinEntryCommand {
	cmd := &GetRevisionsBinEntryCommand{
		RavenCommandBase: NewRavenCommandBase(),

		_etag:     etag,
		_pageSize: pageSize,
	}
	cmd.IsReadRequest = true
	return cmd
}

func (c *GetRevisionsBinEntryCommand) createRequest(node *ServerNode) (*http.Request, error) {
	url := node.getUrl() + "/databases/" + node.getDatabase() + "/revisions/bin?etag=" + strconv.Itoa(c._etag)

	if c._pageSize != 0 {
		url += "&pageSize=" + strconv.Itoa(c._pageSize)
	}

	return NewHttpGet(url)
}

func (c *GetRevisionsBinEntryCommand) setResponse(response []byte, fromCache bool) error {
	if len(response) == 0 {
		return throwInvalidResponse()
	}

	var res ArrayNode
	err := json.Unmarshal(response, &res)
	if err != nil {
		return err
	}
	c.Result = &res
	return nil
}
