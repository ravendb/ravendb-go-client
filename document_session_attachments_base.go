package ravendb

import (
	"fmt"
	"io"
	"reflect"
)

type DocumentSessionAttachmentsBase struct {
	*AdvancedSessionExtensionBase
}

func NewDocumentSessionAttachmentsBase(session *InMemoryDocumentSessionOperations) *DocumentSessionAttachmentsBase {
	res := &DocumentSessionAttachmentsBase{}
	res.AdvancedSessionExtensionBase = newAdvancedSessionExtensionBase(session)
	return res
}

func (s *DocumentSessionAttachmentsBase) GetNames(entity interface{}) ([]*AttachmentName, error) {
	err := checkValidEntityIn(entity, "entity")
	if err != nil {
		return nil, err
	}
	document := getDocumentInfoByEntity(s.documents, entity)
	if document == nil {
		return nil, throwEntityNotInSession(entity)
	}
	meta := document.metadata
	attachmentsI, ok := meta[MetadataAttachments]
	if !ok {
		return nil, nil
	}

	attachments, ok := attachmentsI.([]interface{})
	if !ok {
		return nil, fmt.Errorf("meta value '%s' is of type %T, expected []interface{}", MetadataAttachments, attachmentsI)
	}
	n := len(attachments)
	results := make([]*AttachmentName, n)
	clazz := reflect.TypeOf(&AttachmentName{})
	for i := 0; i < n; i++ {
		jsonNode := attachments[i]
		resI, err := convertValue(jsonNode, clazz)
		if err != nil {
			return nil, err
		}
		res := resI.(*AttachmentName)
		results[i] = res
	}
	return results, nil
}

// contentType is optional
func (s *DocumentSessionAttachmentsBase) StoreByID(documentID string, name string, stream io.Reader, contentType string) error {
	if stringIsBlank(documentID) {
		return newIllegalArgumentError("documentID can't be an empty string")
	}
	if stringIsBlank(name) {
		return newIllegalArgumentError("name can't be an empty string")
	}
	if stream == nil {
		return newIllegalArgumentError("stream can't be nil")
	}

	deferredCommandsMap := s.deferredCommandsMap

	key := newIDTypeAndName(documentID, CommandDelete, "")
	if _, ok := deferredCommandsMap[key]; ok {
		return newIllegalStateError("Cannot Store attachment" + name + " of document " + documentID + ", there is a deferred command registered for this document to be deleted")
	}

	key = newIDTypeAndName(documentID, CommandAttachmentPut, name)
	if _, ok := deferredCommandsMap[key]; ok {
		return newIllegalStateError("Cannot Store attachment" + name + " of document " + documentID + ", there is a deferred command registered to create an attachment with the same name.")
	}

	key = newIDTypeAndName(documentID, CommandAttachmentDelete, name)
	if _, ok := deferredCommandsMap[key]; ok {
		return newIllegalStateError("Cannot Store attachment" + name + " of document " + documentID + ", there is a deferred command registered to delete an attachment with the same name.")
	}

	documentInfo := s.documentsByID.getValue(documentID)
	if documentInfo != nil && s.deletedEntities.contains(documentInfo.entity) {
		return newIllegalStateError("Cannot Store attachment " + name + " of document " + documentID + ", the document was already deleted in this session.")
	}

	cmdData, err := NewPutAttachmentCommandData(documentID, name, stream, contentType, nil)
	if err != nil {
		return err
	}
	s.Defer(cmdData)
	return nil
}

// Store stores an entity
func (s *DocumentSessionAttachmentsBase) Store(entity interface{}, name string, stream io.Reader, contentType string) error {
	document := getDocumentInfoByEntity(s.documents, entity)
	if document == nil {
		return throwEntityNotInSession(entity)
	}

	return s.StoreByID(document.id, name, stream, contentType)
}

// Delete deletes a given entity
// TODO: support **struct or return good error message
func (s *DocumentSessionAttachmentsBase) Delete(entity interface{}, name string) error {
	document := getDocumentInfoByEntity(s.documents, entity)
	if document == nil {
		return throwEntityNotInSession(entity)
	}

	return s.DeleteByID(document.id, name)
}

// Delete deletes entity with a given i
func (s *DocumentSessionAttachmentsBase) DeleteByID(documentID string, name string) error {
	if stringIsBlank(documentID) {
		return newIllegalArgumentError("DocumentId cannot be null")
	}

	if stringIsBlank(name) {
		return newIllegalArgumentError("Name cannot be null")
	}

	deferredCommandsMap := s.deferredCommandsMap

	key := newIDTypeAndName(documentID, CommandDelete, "")
	if _, ok := deferredCommandsMap[key]; ok {
		return nil // no-op
	}

	key = newIDTypeAndName(documentID, CommandAttachmentDelete, name)
	if _, ok := deferredCommandsMap[key]; ok {
		return nil // no-op
	}

	documentInfo := s.documentsByID.getValue(documentID)
	if documentInfo != nil && s.deletedEntities.contains(documentInfo.entity) {
		return nil //no-op
	}

	key = newIDTypeAndName(documentID, CommandAttachmentPut, name)
	if _, ok := deferredCommandsMap[key]; ok {
		return newIllegalStateError("Cannot delete attachment " + name + " of document " + documentID + ", there is a deferred command registered to create an attachment with the same name.")
	}

	cmdData, err := NewDeleteAttachmentCommandData(documentID, name, nil)
	if err != nil {
		return err
	}
	s.Defer(cmdData)
	return nil
}

func throwEntityNotInSession(entity interface{}) *IllegalArgumentError {
	return newIllegalArgumentError("%v is not associated with the session. Use documentID instead or track the entity in the session.", entity)
}
