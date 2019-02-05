package ravendb

import (
	"net/http"
)

var (
	_ RavenCommand = &HeadDocumentCommand{}
)

// HeadDocumentCommand describes "head document" command
type HeadDocumentCommand struct {
	RavenCommandBase

	id           string
	changeVector *string

	Result *string // change vector
}

// NewHeadDocumentCommand returns new HeadDocumentCommand
func NewHeadDocumentCommand(id string, changeVector *string) *HeadDocumentCommand {
	panicIf(id == "", "id cannot be empty")
	cmd := &HeadDocumentCommand{
		RavenCommandBase: NewRavenCommandBase(),

		id:           id,
		changeVector: changeVector,
	}

	return cmd
}

func (c *HeadDocumentCommand) createRequest(node *ServerNode) (*http.Request, error) {
	url := node.URL + "/databases/" + node.Database + "/docs?id=" + urlUtilsEscapeDataString(c.id)

	request, err := NewHttpHead(url)
	if err != nil {
		return nil, err
	}

	if c.changeVector != nil {
		request.Header.Set(headersIfNoneMatch, *c.changeVector)
	}

	return request, nil
}

// ProcessResponse processes HTTP response
func (c *HeadDocumentCommand) ProcessResponse(cache *HttpCache, response *http.Response, url string) (responseDisposeHandling, error) {
	statusCode := response.StatusCode
	if statusCode == http.StatusNotModified {
		c.Result = c.changeVector
		return responseDisposeHandlingAutomatic, nil
	}

	if statusCode == http.StatusNotFound {
		c.Result = nil
		return responseDisposeHandlingAutomatic, nil
	}

	var err error
	c.Result, err = gttpExtensionsGetRequiredEtagHeader(response)
	return responseDisposeHandlingAutomatic, err
}

// SetResponse sets the response
func (c *HeadDocumentCommand) SetResponse(response []byte, fromCache bool) error {
	if len(response) != 0 {
		return throwInvalidResponse()
	}
	// This is called from handleUnsuccessfulResponse() to mark the command
	// as having empty result
	c.Result = nil
	return nil
}

// Exists returns true if the command has a result
func (c *HeadDocumentCommand) Exists() bool {
	return c.Result != nil
}
