package ravendb

import (
	"net/http"
	"strconv"
)

var (
	_ IOperation = &GetAttachmentOperation{}
)

type GetAttachmentOperation struct {
	Command *GetAttachmentCommand

	_documentID   string
	_name         string
	_type         AttachmentType
	_changeVector *string
}

func NewGetAttachmentOperation(documentID string, name string, typ AttachmentType, contentType string, changeVector *string) *GetAttachmentOperation {
	return &GetAttachmentOperation{
		_documentID:   documentID,
		_name:         name,
		_type:         typ,
		_changeVector: changeVector,
	}
}

func (o *GetAttachmentOperation) GetCommand(store *DocumentStore, conventions *DocumentConventions, cache *httpCache) (RavenCommand, error) {
	var err error
	o.Command, err = NewGetAttachmentCommand(o._documentID, o._name, o._type, o._changeVector)
	return o.Command, err
}

var _ RavenCommand = &GetAttachmentCommand{}

type GetAttachmentCommand struct {
	RavenCommandBase

	_documentID   string
	_name         string
	_type         AttachmentType
	_changeVector *string

	Result *AttachmentResult
}

// TODO: should stream be io.ReadCloser? Who owns closing the attachment
func NewGetAttachmentCommand(documentID string, name string, typ AttachmentType, changeVector *string) (*GetAttachmentCommand, error) {
	if stringIsBlank(documentID) {
		return nil, newIllegalArgumentError("DocumentId cannot be null or empty")
	}

	if stringIsBlank(name) {
		return nil, newIllegalArgumentError("Name cannot be null or empty")
	}

	if typ != AttachmentDocument && changeVector == nil {
		return nil, newIllegalArgumentError("Change vector cannot be null for attachment type " + typ)
	}

	cmd := &GetAttachmentCommand{
		RavenCommandBase: NewRavenCommandBase(),

		_documentID:   documentID,
		_name:         name,
		_type:         typ,
		_changeVector: changeVector,
	}
	cmd.IsReadRequest = true
	return cmd, nil
}

func (c *GetAttachmentCommand) CreateRequest(node *ServerNode) (*http.Request, error) {
	url := node.URL + "/databases/" + node.Database + "/attachments?id=" + urlUtilsEscapeDataString(c._documentID) + "&name=" + urlUtilsEscapeDataString(c._name)

	if c._type == AttachmentRevision {
		m := map[string]interface{}{
			"Type":         "Revision",
			"ChangeVector": c._changeVector,
		}
		d, err := jsonMarshal(m)
		if err != nil {
			return nil, err
		}
		return NewHttpPost(url, d)
	}

	return newHttpGet(url)
}

func (c *GetAttachmentCommand) processResponse(cache *httpCache, response *http.Response, url string) (responseDisposeHandling, error) {
	contentType := response.Header.Get("Content-Type")
	changeVector := gttpExtensionsGetEtagHeader(response)
	hash := response.Header.Get("Attachment-Hash")
	size := int64(0)
	sizeHeader := response.Header.Get("Attachment-Size")
	if sizeHeader != "" {
		size, _ = strconv.ParseInt(sizeHeader, 10, 64)
	}

	attachmentDetails := &AttachmentDetails{
		AttachmentName: AttachmentName{
			Name:        c._name,
			ContentType: contentType,
			Hash:        hash,
			Size:        size,
		},
		ChangeVector: changeVector,
		DocumentID:   c._documentID,
	}
	c.Result = newAttachmentResult(response, attachmentDetails)
	return responseDisposeHandlingManually, nil
}
