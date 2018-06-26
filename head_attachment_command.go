package ravendb

import (
	"net/http"
)

var (
	_ RavenCommand = &HeadAttachmentCommand{}
)

type HeadAttachmentCommand struct {
	*RavenCommandBase

	_documentId   string
	_name         string
	_changeVector *string

	Result string // TODO: should this be *string?
}

func NewHeadAttachmentCommand(documentId string, name string, changeVector *string) *HeadAttachmentCommand {
	// TODO: validation
	cmd := &HeadAttachmentCommand{
		RavenCommandBase: NewRavenCommandBase(),

		_documentId:   documentId,
		_name:         name,
		_changeVector: changeVector,
	}
	return cmd
}

func (c *HeadAttachmentCommand) createRequest(node *ServerNode) (*http.Request, error) {
	url := node.getUrl() + "/databases/" + node.getDatabase() + "/attachments?id=" + UrlUtils_escapeDataString(c._documentId) + "&name=" + UrlUtils_escapeDataString(c._name)

	request, err := NewHttpGet(url)
	if err != nil {
		return nil, err
	}

	if c._changeVector != nil {
		request.Header.Set("If-None-Match", *c._changeVector)
	}

	return request, nil
}

func (c *HeadAttachmentCommand) processResponse(cache *HttpCache, response *http.Response, url string) (ResponseDisposeHandling, error) {
	if response.StatusCode == http.StatusNotModified {
		if c._changeVector != nil {
			c.Result = *c._changeVector
		}
		return ResponseDisposeHandling_AUTOMATIC, nil
	}

	if response.StatusCode == http.StatusNotFound {
		c.Result = ""
		return ResponseDisposeHandling_AUTOMATIC, nil
	}

	res, err := HttpExtensions_getRequiredEtagHeader(response)
	if err != nil {
		return ResponseDisposeHandling_AUTOMATIC, err
	}
	if res != nil {
		c.Result = *res
	}
	return ResponseDisposeHandling_AUTOMATIC, nil
}

func (c *HeadAttachmentCommand) setResponse(response []byte, fromCache bool) error {
	if response != nil {
		return throwInvalidResponse()
	}
	c.Result = ""
	return nil
}
