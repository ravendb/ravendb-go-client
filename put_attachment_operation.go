package ravendb

import (
	"encoding/json"
	"io"
	"net/http"
)

var (
	_ IOperation = &PutAttachmentOperation{}
)

type PutAttachmentOperation struct {
	Command *PutAttachmentCommand

	_documentId   string
	_name         string
	_stream       io.Reader
	_contentType  string
	_changeVector *string
}

func NewPutAttachmentOperation(documentId string, name string, stream io.Reader, contentType string, changeVector *string) *PutAttachmentOperation {
	return &PutAttachmentOperation{
		_documentId:   documentId,
		_name:         name,
		_stream:       stream,
		_contentType:  contentType,
		_changeVector: changeVector,
	}
}

func (o *PutAttachmentOperation) getCommand(store *IDocumentStore, conventions *DocumentConventions, cache *HttpCache) RavenCommand {
	o.Command = NewPutAttachmentCommand(o._documentId, o._name, o._stream, o._contentType, o._changeVector)
	return o.Command
}

var _ RavenCommand = &PutAttachmentCommand{}

type PutAttachmentCommand struct {
	*RavenCommandBase

	_documentId   string
	_name         string
	_stream       io.Reader
	_contentType  string
	_changeVector *string

	Result *AttachmentDetails
}

// TODO: should stream be io.ReadCloser? Who owns closing the attachment
func NewPutAttachmentCommand(documentId string, name string, stream io.Reader, contentType string, changeVector *string) *PutAttachmentCommand {
	// TODO: validation
	cmd := &PutAttachmentCommand{
		RavenCommandBase: NewRavenCommandBase(),

		_documentId:   documentId,
		_name:         name,
		_stream:       stream,
		_contentType:  contentType,
		_changeVector: changeVector,
	}
	return cmd
}

func (c *PutAttachmentCommand) createRequest(node *ServerNode) (*http.Request, error) {
	url := node.getUrl() + "/databases/" + node.getDatabase() + "/attachments?id=" + UrlUtils_escapeDataString(c._documentId) + "&name=" + UrlUtils_escapeDataString(c._name)

	if StringUtils_isNotEmpty(c._contentType) {
		url += "&contentType=" + UrlUtils_escapeDataString(c._contentType)
	}

	return NewHttpPutReader(url, c._stream)

}

func (c *PutAttachmentCommand) setResponse(response []byte, fromCache bool) error {
	var res AttachmentDetails
	err := json.Unmarshal(response, &res)
	if err != nil {
		return err
	}
	c.Result = &res
	return nil
}
