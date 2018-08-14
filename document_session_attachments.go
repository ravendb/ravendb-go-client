package ravendb

type DocumentSessionAttachments struct {
	*DocumentSessionAttachmentsBase
}

func NewDocumentSessionAttachments(session *InMemoryDocumentSessionOperations) *DocumentSessionAttachments {
	res := &DocumentSessionAttachments{}
	res.DocumentSessionAttachmentsBase = NewDocumentSessionAttachmentsBase(session)
	return res
}

func (s *DocumentSessionAttachments) exists(documentId string, name string) (bool, error) {
	command := NewHeadAttachmentCommand(documentId, name, nil)
	err := s.requestExecutor.executeCommandWithSessionInfo(command, s.sessionInfo)
	if err != nil {
		return false, err
	}
	res := command.Result != ""
	return res, nil
}

func (s *DocumentSessionAttachments) get(documentId string, name string) (*CloseableAttachmentResult, error) {
	operation := NewGetAttachmentOperation(documentId, name, AttachmentType_DOCUMENT, "", nil)
	err := s.session.getOperations().SendWithSessionInfo(operation, s.sessionInfo)
	if err != nil {
		return nil, err
	}
	res := operation.Command.Result
	return res, nil
}

func (s *DocumentSessionAttachments) getEntity(entity Object, name string) (*CloseableAttachmentResult, error) {
	document := s.documentsByEntity[entity]
	if document == nil {
		return nil, throwEntityNotInSession(entity)
	}
	return s.get(document.getId(), name)
}

func (s *DocumentSessionAttachments) getRevision(documentId string, name string, changeVector *string) (*CloseableAttachmentResult, error) {
	operation := NewGetAttachmentOperation(documentId, name, AttachmentType_REVISION, "", changeVector)
	err := s.session.getOperations().SendWithSessionInfo(operation, s.sessionInfo)
	if err != nil {
		return nil, err
	}
	res := operation.Command.Result
	return res, nil
}
