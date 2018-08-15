package ravendb

import (
	"net/http"
)

var (
	_ IOperation = &DeleteAttachmentOperation{}
)

type DeleteAttachmentOperation struct {
	Command *DeleteAttachmentCommand

	_documentId   string
	_name         string
	_changeVector *string
}

func NewDeleteAttachmentOperation(documentId string, name string, changeVector *string) *DeleteAttachmentOperation {
	return &DeleteAttachmentOperation{
		_documentId:   documentId,
		_name:         name,
		_changeVector: changeVector,
	}
}

func (o *DeleteAttachmentOperation) GetCommand(store *IDocumentStore, conventions *DocumentConventions, cache *HttpCache) RavenCommand {
	o.Command = NewDeleteAttachmentCommand(o._documentId, o._name, o._changeVector)
	return o.Command
}

var _ RavenCommand = &DeleteAttachmentCommand{}

type DeleteAttachmentCommand struct {
	*RavenCommandBase

	_documentId   string
	_name         string
	_changeVector *string
}

func NewDeleteAttachmentCommand(documentId string, name string, changeVector *string) *DeleteAttachmentCommand {
	// TODO: validation
	cmd := &DeleteAttachmentCommand{
		RavenCommandBase: NewRavenCommandBase(),
		_documentId:      documentId,
		_name:            name,
		_changeVector:    changeVector,
	}
	cmd.RavenCommandBase.responseType = RavenCommandResponseType_EMPTY
	return cmd
}

func (c *DeleteAttachmentCommand) CreateRequest(node *ServerNode) (*http.Request, error) {
	url := node.getUrl() + "/databases/" + node.getDatabase() + "/attachments?id=" + UrlUtils_escapeDataString(c._documentId) + "&name=" + UrlUtils_escapeDataString(c._name)

	request, err := NewHttpDelete(url, nil)
	if err != nil {
		return nil, err
	}
	addChangeVectorIfNotNull(c._changeVector, request)
	return request, err
}
