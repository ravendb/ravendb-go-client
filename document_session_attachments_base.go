package ravendb

import (
	"fmt"
	"io"
)

type DocumentSessionAttachmentsBase struct {
	*AdvancedSessionExtentionBase
}

func NewDocumentSessionAttachmentsBase(session *InMemoryDocumentSessionOperations) *DocumentSessionAttachmentsBase {
	res := &DocumentSessionAttachmentsBase{}
	res.AdvancedSessionExtentionBase = NewAdvancedSessionExtentionBase(session)
	return res
}

func (s *DocumentSessionAttachmentsBase) GetNames(entity Object) ([]*AttachmentName, error) {
	if entity == nil {
		return nil, nil
	}
	document := s.documentsByEntity[entity]
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
	results := make([]*AttachmentName, n, n)
	clazz := GetTypeOf(&AttachmentName{})
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
func (s *DocumentSessionAttachmentsBase) Store(documentId string, name string, stream io.Reader, contentType string) error {
	// TODO: validate args

	deferredCommandsMap := s.deferredCommandsMap

	key := IdTypeAndName_create(documentId, CommandType_DELETE, "")
	if _, ok := deferredCommandsMap[key]; ok {
		return NewIllegalStateException("Cannot Store attachment" + name + " of document " + documentId + ", there is a deferred command registered for this document to be deleted")
	}

	key = IdTypeAndName_create(documentId, CommandType_ATTACHMENT_PUT, name)
	if _, ok := deferredCommandsMap[key]; ok {
		return NewIllegalStateException("Cannot Store attachment" + name + " of document " + documentId + ", there is a deferred command registered to create an attachment with the same name.")
	}

	key = IdTypeAndName_create(documentId, CommandType_ATTACHMENT_DELETE, name)
	if _, ok := deferredCommandsMap[key]; ok {
		return NewIllegalStateException("Cannot Store attachment" + name + " of document " + documentId + ", there is a deferred command registered to delete an attachment with the same name.")
	}

	documentInfo := s.documentsById.getValue(documentId)
	if documentInfo != nil && s.deletedEntities.contains(documentInfo.entity) {
		return NewIllegalStateException("Cannot Store attachment " + name + " of document " + documentId + ", the document was already deleted in this session.")
	}

	cmdData := NewPutAttachmentCommandData(documentId, name, stream, contentType, nil)
	s.DeferMany([]ICommandData{cmdData})
	return nil
}

func (s *DocumentSessionAttachmentsBase) StoreEntity(entity Object, name string, stream io.Reader, contentType string) error {
	document := s.documentsByEntity[entity]
	if document == nil {
		return throwEntityNotInSession(entity)
	}

	return s.Store(document.id, name, stream, contentType)
}

func (s *DocumentSessionAttachmentsBase) DeleteEntity(entity Object, name string) error {
	document := s.documentsByEntity[entity]
	if document == nil {
		return throwEntityNotInSession(entity)
	}

	return s.Delete(document.id, name)
}

func (s *DocumentSessionAttachmentsBase) Delete(documentId string, name string) error {
	// TODO: validate args

	deferredCommandsMap := s.deferredCommandsMap

	key := IdTypeAndName_create(documentId, CommandType_DELETE, "")
	if _, ok := deferredCommandsMap[key]; ok {
		return nil // no-op
	}

	key = IdTypeAndName_create(documentId, CommandType_ATTACHMENT_DELETE, name)
	if _, ok := deferredCommandsMap[key]; ok {
		return nil // no-op
	}

	documentInfo := s.documentsById.getValue(documentId)
	if documentInfo != nil && s.deletedEntities.contains(documentInfo.entity) {
		return nil //no-op
	}

	key = IdTypeAndName_create(documentId, CommandType_ATTACHMENT_PUT, name)
	if _, ok := deferredCommandsMap[key]; ok {
		return NewIllegalStateException("Cannot delete attachment " + name + " of document " + documentId + ", there is a deferred command registered to create an attachment with the same name.")
	}

	cmdData := NewDeleteAttachmentCommandData(documentId, name, nil)
	s.DeferMany([]ICommandData{cmdData})
	return nil
}

func throwEntityNotInSession(entity Object) *IllegalArgumentException {
	return NewIllegalArgumentException("%v is not associated with the session. Use documentId instead or track the entity in the session.", entity)
}
