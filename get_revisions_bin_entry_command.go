package ravendb

import (
	"net/http"
	"strconv"
)

var (
	_ RavenCommand = &GetRevisionsBinEntryCommand{}
)

type GetRevisionsBinEntryCommand struct {
	RavenCommandBase

	_etag     int
	_pageSize int

	Result []map[string]interface{}
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

func (c *GetRevisionsBinEntryCommand) CreateRequest(node *ServerNode) (*http.Request, error) {
	url := node.URL + "/databases/" + node.Database + "/revisions/bin?etag=" + strconv.Itoa(c._etag)

	if c._pageSize > 0 {
		url += "&pageSize=" + strconv.Itoa(c._pageSize)
	}

	return NewHttpGet(url)
}

func (c *GetRevisionsBinEntryCommand) SetResponse(response []byte, fromCache bool) error {
	if len(response) == 0 {
		return throwInvalidResponse()
	}

	return jsonUnmarshal(response, &c.Result)
}
