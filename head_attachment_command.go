package ravendb

import (
	"net/http"
)

var (
	_ RavenCommand = &HeadAttachmentCommand{}
)

type HeadAttachmentCommand struct {
	RavenCommandBase

	_documentID   string
	_name         string
	_changeVector *string

	Result string // TODO: should this be *string?
}

func NewHeadAttachmentCommand(documentID string, name string, changeVector *string) (*HeadAttachmentCommand, error) {
	if stringIsBlank(documentID) {
		return nil, newIllegalArgumentError("DocumentId cannot be null or empty")
	}
	if stringIsBlank(name) {
		return nil, newIllegalArgumentError("Name cannot be null or empty")
	}

	cmd := &HeadAttachmentCommand{
		RavenCommandBase: NewRavenCommandBase(),

		_documentID:   documentID,
		_name:         name,
		_changeVector: changeVector,
	}
	return cmd, nil
}

func (c *HeadAttachmentCommand) CreateRequest(node *ServerNode) (*http.Request, error) {
	url := node.URL + "/databases/" + node.Database + "/attachments?id=" + urlUtilsEscapeDataString(c._documentID) + "&name=" + urlUtilsEscapeDataString(c._name)

	request, err := NewHttpGet(url)
	if err != nil {
		return nil, err
	}

	if c._changeVector != nil {
		request.Header.Set(headersIfNoneMatch, *c._changeVector)
	}

	return request, nil
}

func (c *HeadAttachmentCommand) processResponse(cache *HttpCache, response *http.Response, url string) (responseDisposeHandling, error) {
	if response.StatusCode == http.StatusNotModified {
		if c._changeVector != nil {
			c.Result = *c._changeVector
		}
		return responseDisposeHandlingAutomatic, nil
	}

	if response.StatusCode == http.StatusNotFound {
		c.Result = ""
		return responseDisposeHandlingAutomatic, nil
	}

	res, err := gttpExtensionsGetRequiredEtagHeader(response)
	if err != nil {
		return responseDisposeHandlingAutomatic, err
	}
	if res != nil {
		c.Result = *res
	}
	return responseDisposeHandlingAutomatic, nil
}

func (c *HeadAttachmentCommand) SetResponse(response []byte, fromCache bool) error {
	if response != nil {
		return throwInvalidResponse()
	}
	c.Result = ""
	return nil
}
