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

func (o *DeleteAttachmentOperation) GetCommand(store *DocumentStore, conventions *DocumentConventions, cache *HttpCache) (RavenCommand, error) {
	var err error
	o.Command, err = NewDeleteAttachmentCommand(o._documentID, o._name, o._changeVector)
	return o.Command, err
}

var _ RavenCommand = &DeleteAttachmentCommand{}

type DeleteAttachmentCommand struct {
	RavenCommandBase

	_documentID   string
	_name         string
	_changeVector *string
}

func NewDeleteAttachmentCommand(documentID string, name string, changeVector *string) (*DeleteAttachmentCommand, error) {
	if stringIsBlank(documentID) {
		return nil, newIllegalArgumentError("documentId cannot be null")
	}

	if stringIsBlank(name) {
		return nil, newIllegalArgumentError("name cannot be null")
	}

	cmd := &DeleteAttachmentCommand{
		RavenCommandBase: NewRavenCommandBase(),
		_documentID:      documentID,
		_name:            name,
		_changeVector:    changeVector,
	}
	cmd.RavenCommandBase.ResponseType = RavenCommandResponseTypeEmpty
	return cmd, nil
}

func (c *DeleteAttachmentCommand) CreateRequest(node *ServerNode) (*http.Request, error) {
	url := node.URL + "/databases/" + node.Database + "/attachments?id=" + urlUtilsEscapeDataString(c._documentID) + "&name=" + urlUtilsEscapeDataString(c._name)

	request, err := NewHttpDelete(url, nil)
	if err != nil {
		return nil, err
	}
	addChangeVectorIfNotNull(c._changeVector, request)
	return request, err
}
