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

	etag     int64
	pageSize int

	Result *JSONArrayResult
}

func NewGetRevisionsBinEntryCommand(etag int64, pageSize int) *GetRevisionsBinEntryCommand {
	cmd := &GetRevisionsBinEntryCommand{
		RavenCommandBase: NewRavenCommandBase(),

		etag:     etag,
		pageSize: pageSize,
	}
	cmd.IsReadRequest = true
	return cmd
}

func (c *GetRevisionsBinEntryCommand) CreateRequest(node *ServerNode) (*http.Request, error) {
	etagStr := strconv.FormatInt(c.etag, 10)
	url := node.URL + "/databases/" + node.Database + "/revisions/bin?etag=" + etagStr

	if c.pageSize > 0 {
		url += "&pageSize=" + strconv.Itoa(c.pageSize)
	}

	return NewHttpGet(url)
}

func (c *GetRevisionsBinEntryCommand) SetResponse(response []byte, fromCache bool) error {
	if len(response) == 0 {
		return throwInvalidResponse()
	}

	return jsonUnmarshal(response, &c.Result)
}
