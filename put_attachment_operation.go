package ravendb

import (
	"bytes"
	"io"
	"net/http"
)

var (
	_ IOperation = &PutAttachmentOperation{}
)

type PutAttachmentOperation struct {
	Command *PutAttachmentCommand

	_documentID   string
	_name         string
	_stream       io.Reader
	_contentType  string
	_changeVector *string
}

func NewPutAttachmentOperation(documentID string, name string, stream io.Reader, contentType string, changeVector *string) *PutAttachmentOperation {
	return &PutAttachmentOperation{
		_documentID:   documentID,
		_name:         name,
		_stream:       stream,
		_contentType:  contentType,
		_changeVector: changeVector,
	}
}

func (o *PutAttachmentOperation) GetCommand(store *DocumentStore, conventions *DocumentConventions, cache *HttpCache) RavenCommand {
	o.Command = NewPutAttachmentCommand(o._documentID, o._name, o._stream, o._contentType, o._changeVector)
	return o.Command
}

var _ RavenCommand = &PutAttachmentCommand{}

type PutAttachmentCommand struct {
	RavenCommandBase

	_documentID   string
	_name         string
	_stream       io.Reader
	_contentType  string
	_changeVector *string

	Result *AttachmentDetails
}

// TODO: should stream be io.ReadCloser? Who owns closing the attachment
func NewPutAttachmentCommand(documentID string, name string, stream io.Reader, contentType string, changeVector *string) *PutAttachmentCommand {
	// TODO: validation
	cmd := &PutAttachmentCommand{
		RavenCommandBase: NewRavenCommandBase(),

		_documentID:   documentID,
		_name:         name,
		_stream:       stream,
		_contentType:  contentType,
		_changeVector: changeVector,
	}
	return cmd
}

var noReader = true

func (c *PutAttachmentCommand) CreateRequest(node *ServerNode) (*http.Request, error) {
	url := node.URL + "/databases/" + node.Database + "/attachments?id=" + urlUtilsEscapeDataString(c._documentID) + "&name=" + urlUtilsEscapeDataString(c._name)

	if stringIsNotEmpty(c._contentType) {
		url += "&contentType=" + urlUtilsEscapeDataString(c._contentType)
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
	return jsonUnmarshal(response, &c.Result)
}
