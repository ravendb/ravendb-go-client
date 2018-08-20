package ravendb

import (
	"bytes"
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

func (o *PutAttachmentOperation) GetCommand(store *IDocumentStore, conventions *DocumentConventions, cache *HttpCache) RavenCommand {
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

var noReader = true

func (c *PutAttachmentCommand) CreateRequest(node *ServerNode) (*http.Request, error) {
	url := node.GetUrl() + "/databases/" + node.GetDatabase() + "/attachments?id=" + UrlUtils_escapeDataString(c._documentId) + "&name=" + UrlUtils_escapeDataString(c._name)

	if StringUtils_isNotEmpty(c._contentType) {
		url += "&contentType=" + UrlUtils_escapeDataString(c._contentType)
	}

	if noReader {
		var buf bytes.Buffer
		_, err := io.Copy(&buf, c._stream)
		if err != nil {
			return nil, err
		}
		req, err := NewHttpPut(url, buf.Bytes())
		if err != nil {
			return nil, err
		}
		req.Header.Del("Content-Type")
		addChangeVectorIfNotNull(c._changeVector, req)
		return req, nil
	}

	req, err := NewHttpPutReader(url, c._stream)
	if err != nil {
		return nil, err
	}
	addChangeVectorIfNotNull(c._changeVector, req)
	return req, nil

}

func (c *PutAttachmentCommand) SetResponse(response []byte, fromCache bool) error {
	return json.Unmarshal(response, &c.Result)
}
