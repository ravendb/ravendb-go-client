package ravendb

import (
	"net/http"
	"strconv"
)

var (
	_ RavenCommand = &GetRevisionsCommand{}
)

type GetRevisionsCommand struct {
	RavenCommandBase

	_id            string
	_start         int
	_pageSize      int
	_metadataOnly  bool
	_changeVector  string
	_changeVectors []string

	Result *JSONArrayResult
}

func NewGetRevisionsCommand(changeVectors []string, metadataOnly bool) *GetRevisionsCommand {
	cmd := &GetRevisionsCommand{
		RavenCommandBase: NewRavenCommandBase(),

		_changeVectors: changeVectors,
		_metadataOnly:  metadataOnly,
	}
	cmd.IsReadRequest = true
	return cmd
}

func NewGetRevisionsCommandRange(id string, start int, pageSize int, metadataOnly bool) *GetRevisionsCommand {
	cmd := &GetRevisionsCommand{
		RavenCommandBase: NewRavenCommandBase(),

		_id:           id,
		_start:        start,
		_pageSize:     pageSize,
		_metadataOnly: metadataOnly,
	}
	cmd.IsReadRequest = true
	return cmd
}

func (c *GetRevisionsCommand) GetChangeVectors() []string {
	return c._changeVectors
}

func (c *GetRevisionsCommand) CreateRequest(node *ServerNode) (*http.Request, error) {
	url := node.GetUrl() + "/databases/" + node.GetDatabase() + "/revisions?"

	if c._id != "" {
		url += "&id=" + UrlUtils_escapeDataString(c._id)
	} else if c._changeVector != "" {
		url += "&changeVector=" + UrlUtils_escapeDataString(c._changeVector)
	} else if c._changeVectors != nil {
		for _, changeVector := range c._changeVectors {
			url += "&changeVector=" + UrlUtils_escapeDataString(changeVector)
		}
	}

	if c._start > 0 {
		url += "&start=" + strconv.Itoa(c._start)
	}

	if c._pageSize > 0 {
		url += "&pageSize=" + strconv.Itoa(c._pageSize)
	}

	if c._metadataOnly {
		url += "&metadataOnly=true"
	}

	return NewHttpGet(url)
}

func (c *GetRevisionsCommand) SetResponse(response []byte, fromCache bool) error {
	var res JSONArrayResult
	err := jsonUnmarshal(response, &res)
	if err != nil {
		return err
	}
	if res.Results == nil {
		return throwInvalidResponse()
	}
	c.Result = &res
	return nil
}
