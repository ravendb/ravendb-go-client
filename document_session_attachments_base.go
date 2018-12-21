package ravendb

import (
	"fmt"
	"io"
	"reflect"
)

type DocumentSessionAttachmentsBase struct {
	*AdvancedSessionExtentionBase
}

func NewDocumentSessionAttachmentsBase(session *InMemoryDocumentSessionOperations) *DocumentSessionAttachmentsBase {
	res := &DocumentSessionAttachmentsBase{}
	res.AdvancedSessionExtentionBase = NewAdvancedSessionExtentionBase(session)
	return res
}

func (s *DocumentSessionAttachmentsBase) GetNames(entity interface{}) ([]*AttachmentName, error) {
	if entity == nil {
		return nil, nil
	}
	document := getDocumentInfoByEntity(s.documents, entity)
	if document == nil {
		return nil, throwEntityNotInSession(entity)
	}
	meta := document.metadata
	attachmentsI, ok := meta[Constants_Documents_Metadata_ATTACHMENTS]
	if !ok {
		return nil, nil
	}

	attachments, ok := attachmentsI.([]interface{})
	if !ok {
		return nil, fmt.Errorf("meta value '%s' is of type %T, expected []interface{}", Constants_Documents_Metadata_ATTACHMENTS, attachmentsI)
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
// TODO: maybe split into Store() and storeWithContentType()
func (s *DocumentSessionAttachmentsBase) Store(documentID string, name string, stream io.Reader, contentType string) error {
	// TODO: validate args

	deferredCommandsMap := s.deferredCommandsMap

	key := newIDTypeAndName(documentID, CommandType_DELETE, "")
	if _, ok := deferredCommandsMap[key]; ok {
		return newIllegalStateError("Cannot Store attachment" + name + " of document " + documentID + ", there is a deferred command registered for this document to be deleted")
	}

	key = newIDTypeAndName(documentID, CommandType_ATTACHMENT_PUT, name)
	if _, ok := deferredCommandsMap[key]; ok {
		return newIllegalStateError("Cannot Store attachment" + name + " of document " + documentID + ", there is a deferred command registered to create an attachment with the same name.")
	}

	key = newIDTypeAndName(documentID, CommandType_ATTACHMENT_DELETE, name)
	if _, ok := deferredCommandsMap[key]; ok {
		return newIllegalStateError("Cannot Store attachment" + name + " of document " + documentID + ", there is a deferred command registered to delete an attachment with the same name.")
	}

	documentInfo := s.documentsByID.getValue(documentID)
	if documentInfo != nil && s.deletedEntities.contains(documentInfo.entity) {
		return newIllegalStateError("Cannot Store attachment " + name + " of document " + documentID + ", the document was already deleted in this session.")
	}

	cmdData := NewPutAttachmentCommandData(documentID, name, stream, contentType, nil)
	s.DeferMany([]ICommandData{cmdData})
	return nil
}

// StoreEntity stores an entity
func (s *DocumentSessionAttachmentsBase) StoreEntity(entity interface{}, name string, stream io.Reader, contentType string) error {
	document := getDocumentInfoByEntity(s.documents, entity)
	if document == nil {
		return throwEntityNotInSession(entity)
	}

	return s.Store(document.id, name, stream, contentType)
}

// DeleteEntity deletes a given entity
func (s *DocumentSessionAttachmentsBase) DeleteEntity(entity interface{}, name string) error {
	document := getDocumentInfoByEntity(s.documents, entity)
	if document == nil {
		return throwEntityNotInSession(entity)
	}

	return s.Delete(document.id, name)
}

// Delete deletes entity with a given i
func (s *DocumentSessionAttachmentsBase) Delete(documentID string, name string) error {
	// TODO: validate args

	deferredCommandsMap := s.deferredCommandsMap

	key := newIDTypeAndName(documentID, CommandType_DELETE, "")
	if _, ok := deferredCommandsMap[key]; ok {
		return nil // no-op
	}

	key = newIDTypeAndName(documentID, CommandType_ATTACHMENT_DELETE, name)
	if _, ok := deferredCommandsMap[key]; ok {
		return nil // no-op
	}

	documentInfo := s.documentsByID.getValue(documentID)
	if documentInfo != nil && s.deletedEntities.contains(documentInfo.entity) {
		return nil //no-op
	}

	key = newIDTypeAndName(documentID, CommandType_ATTACHMENT_PUT, name)
	if _, ok := deferredCommandsMap[key]; ok {
		return newIllegalStateError("Cannot delete attachment " + name + " of document " + documentID + ", there is a deferred command registered to create an attachment with the same name.")
	}

	cmdData := NewDeleteAttachmentCommandData(documentID, name, nil)
	s.DeferMany([]ICommandData{cmdData})
	return nil
}

func throwEntityNotInSession(entity interface{}) *IllegalArgumentError {
	return newIllegalArgumentError("%v is not associated with the session. Use documentID instead or track the entity in the session.", entity)
}
