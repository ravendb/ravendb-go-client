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

	id            string
	start         int
	pageSize      int
	metadataOnly  bool
	changeVector  string
	changeVectors []string

	Result *JSONArrayResult
}

func NewGetRevisionsCommand(changeVectors []string, metadataOnly bool) *GetRevisionsCommand {
	cmd := &GetRevisionsCommand{
		RavenCommandBase: NewRavenCommandBase(),

		changeVectors: changeVectors,
		metadataOnly:  metadataOnly,
	}
	cmd.IsReadRequest = true
	return cmd
}

func NewGetRevisionsCommandRange(id string, start int, pageSize int, metadataOnly bool) *GetRevisionsCommand {
	cmd := &GetRevisionsCommand{
		RavenCommandBase: NewRavenCommandBase(),

		id:           id,
		start:        start,
		pageSize:     pageSize,
		metadataOnly: metadataOnly,
	}
	cmd.IsReadRequest = true
	return cmd
}

func (c *GetRevisionsCommand) GetChangeVectors() []string {
	return c.changeVectors
}

func (c *GetRevisionsCommand) createRequest(node *ServerNode) (*http.Request, error) {
	url := node.URL + "/databases/" + node.Database + "/revisions?"

	if c.id != "" {
		url += "&id=" + urlUtilsEscapeDataString(c.id)
	} else if c.changeVector != "" {
		url += "&changeVector=" + urlUtilsEscapeDataString(c.changeVector)
	} else if c.changeVectors != nil {
		for _, changeVector := range c.changeVectors {
			url += "&changeVector=" + urlUtilsEscapeDataString(changeVector)
		}
	}

	if c.start > 0 {
		url += "&start=" + strconv.Itoa(c.start)
	}

	if c.pageSize > 0 {
		url += "&pageSize=" + strconv.Itoa(c.pageSize)
	}

	if c.metadataOnly {
		url += "&metadataOnly=true"
	}

	return newHttpGet(url)
}

func (c *GetRevisionsCommand) setResponse(response []byte, fromCache bool) error {
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
