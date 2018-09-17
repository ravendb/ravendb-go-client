package ravendb

import (
	"encoding/json"
	"net/http"
	"strconv"
)

var (
	_ IOperation = &GetAttachmentOperation{}
)

type GetAttachmentOperation struct {
	Command *GetAttachmentCommand

	_documentId   string
	_name         string
	_type         AttachmentType
	_changeVector *string
}

func NewGetAttachmentOperation(documentId string, name string, typ AttachmentType, contentType string, changeVector *string) *GetAttachmentOperation {
	return &GetAttachmentOperation{
		_documentId:   documentId,
		_name:         name,
		_type:         typ,
		_changeVector: changeVector,
	}
}

func (o *GetAttachmentOperation) GetCommand(store *IDocumentStore, conventions *DocumentConventions, cache *HttpCache) RavenCommand {
	o.Command = NewGetAttachmentCommand(o._documentId, o._name, o._type, o._changeVector)
	return o.Command
}

var _ RavenCommand = &GetAttachmentCommand{}

type GetAttachmentCommand struct {
	RavenCommandBase

	_documentId   string
	_name         string
	_type         AttachmentType
	_changeVector *string

	Result *CloseableAttachmentResult
}

// TODO: should stream be io.ReadCloser? Who owns closing the attachment
func NewGetAttachmentCommand(documentId string, name string, typ AttachmentType, changeVector *string) *GetAttachmentCommand {
	// TODO: validation
	cmd := &GetAttachmentCommand{
		RavenCommandBase: NewRavenCommandBase(),

		_documentId:   documentId,
		_name:         name,
		_type:         typ,
		_changeVector: changeVector,
	}
	cmd.IsReadRequest = true
	return cmd
}

func (c *GetAttachmentCommand) CreateRequest(node *ServerNode) (*http.Request, error) {
	url := node.GetUrl() + "/databases/" + node.GetDatabase() + "/attachments?id=" + UrlUtils_escapeDataString(c._documentId) + "&name=" + UrlUtils_escapeDataString(c._name)

	if c._type == AttachmentType_REVISION {
		m := map[string]interface{}{
			"Type":         "Revision",
			"ChangeVector": c._changeVector,
		}
		d, err := json.Marshal(m)
		if err != nil {
			return nil, err
		}
		return NewHttpPost(url, d)
	}

	return NewHttpGet(url)
}

func (c *GetAttachmentCommand) processResponse(cache *HttpCache, response *http.Response, url string) (ResponseDisposeHandling, error) {
	contentType := response.Header.Get("Content-Type")
	changeVector := HttpExtensions_getEtagHeader(response)
	hash := response.Header.Get("Attachment-Hash")
	size := int64(0)
	sizeHeader := response.Header.Get("Attachment-Size")
	if sizeHeader != "" {
		size, _ = strconv.ParseInt(sizeHeader, 10, 64)
	}

	attachmentDetails := &AttachmentDetails{
		AttachmentName: AttachmentName{
			ContentType: contentType,
			Name:        c._name,
			Hash:        hash,
			Size:        size,
		},
		ChangeVector: changeVector,
		DocumentID:   c._documentId,
	}

	c.Result = NewCloseableAttachmentResult(response, attachmentDetails)
	return ResponseDisposeHandling_MANUALLY, nil
}
