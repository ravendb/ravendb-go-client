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

	start             int
	pageSize          int
	continuationToken string

	Result *JSONArrayResult
}

func NewGetRevisionsBinEntryCommand(start int, pageSize int) *GetRevisionsBinEntryCommand {
	cmd := &GetRevisionsBinEntryCommand{
		RavenCommandBase: NewRavenCommandBase(),

		start:    start,
		pageSize: pageSize,
	}
	cmd.IsReadRequest = true
	return cmd
}

func NewGetRevisionsBinEntryCommandWithContinuationToken(continuationToken string) *GetRevisionsBinEntryCommand {
	cmd := &GetRevisionsBinEntryCommand{
		RavenCommandBase: NewRavenCommandBase(),

		continuationToken: continuationToken,
	}
	cmd.IsReadRequest = true
	return cmd
}

func (c *GetRevisionsBinEntryCommand) CreateRequest(node *ServerNode) (*http.Request, error) {
	url := node.URL + "/databases/" + node.Database + "/revisions/bin?start=" + strconv.Itoa(c.start)

	if c.pageSize > 0 {
		url += "&pageSize=" + strconv.Itoa(c.pageSize)
	}

	if c.continuationToken != "" {
		url += "&continuationToken=" + urlUtilsEscapeDataString(c.continuationToken)
	}

	return newHttpGet(url)
}

func (c *GetRevisionsBinEntryCommand) SetResponse(response []byte, fromCache bool) error {
	if len(response) == 0 {
		return throwInvalidResponse()
	}

	return jsonUnmarshal(response, &c.Result)
}
