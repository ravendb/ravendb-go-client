package ravendb

import (
	"net/http"
)

var (
	_ IOperation = &DeleteAttachmentOperation{}
)

type DeleteAttachmentOperation struct {
	Command *DeleteAttachmentCommand

	_documentID   string
	_name         string
	_changeVector *string
}

func NewDeleteAttachmentOperation(documentID string, name string, changeVector *string) *DeleteAttachmentOperation {
	return &DeleteAttachmentOperation{
		_documentID:   documentID,
		_name:         name,
		_changeVector: changeVector,
	}
}

func (o *DeleteAttachmentOperation) GetCommand(store *IDocumentStore, conventions *DocumentConventions, cache *HttpCache) RavenCommand {
	o.Command = NewDeleteAttachmentCommand(o._documentID, o._name, o._changeVector)
	return o.Command
}

var _ RavenCommand = &DeleteAttachmentCommand{}

type DeleteAttachmentCommand struct {
	RavenCommandBase

	_documentID   string
	_name         string
	_changeVector *string
}

func NewDeleteAttachmentCommand(documentID string, name string, changeVector *string) *DeleteAttachmentCommand {
	// TODO: validation
	cmd := &DeleteAttachmentCommand{
		RavenCommandBase: NewRavenCommandBase(),
		_documentID:      documentID,
		_name:            name,
		_changeVector:    changeVector,
	}
	cmd.RavenCommandBase.ResponseType = RavenCommandResponseTypeEmpty
	return cmd
}

func (c *DeleteAttachmentCommand) CreateRequest(node *ServerNode) (*http.Request, error) {
	url := node.GetUrl() + "/databases/" + node.GetDatabase() + "/attachments?id=" + UrlUtils_escapeDataString(c._documentID) + "&name=" + UrlUtils_escapeDataString(c._name)

	request, err := NewHttpDelete(url, nil)
	if err != nil {
		return nil, err
	}
	addChangeVectorIfNotNull(c._changeVector, request)
	return request, err
}
