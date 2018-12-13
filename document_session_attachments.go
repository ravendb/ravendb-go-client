package ravendb

type IAttachmentsSessionOperations = DocumentSessionAttachments

type DocumentSessionAttachments struct {
	*DocumentSessionAttachmentsBase
}

func NewDocumentSessionAttachments(session *InMemoryDocumentSessionOperations) *DocumentSessionAttachments {
	res := &DocumentSessionAttachments{}
	res.DocumentSessionAttachmentsBase = NewDocumentSessionAttachmentsBase(session)
	return res
}

func (s *DocumentSessionAttachments) Exists(documentId string, name string) (bool, error) {
	command := NewHeadAttachmentCommand(documentId, name, nil)
	err := s.requestExecutor.ExecuteCommandWithSessionInfo(command, s.sessionInfo)
	if err != nil {
		return false, err
	}
	res := command.Result != ""
	return res, nil
}

func (s *DocumentSessionAttachments) Get(documentId string, name string) (*CloseableAttachmentResult, error) {
	operation := NewGetAttachmentOperation(documentId, name, AttachmentType_DOCUMENT, "", nil)
	err := s.session.GetOperations().SendWithSessionInfo(operation, s.sessionInfo)
	if err != nil {
		return nil, err
	}
	res := operation.Command.Result
	return res, nil
}

func (s *DocumentSessionAttachments) GetEntity(entity interface{}, name string) (*CloseableAttachmentResult, error) {
	document := getDocumentInfoByEntity(s.documents, entity)
	if document == nil {
		return nil, throwEntityNotInSession(entity)
	}
	return s.Get(document.id, name)
}

func (s *DocumentSessionAttachments) GetRevision(documentId string, name string, changeVector *string) (*CloseableAttachmentResult, error) {
	operation := NewGetAttachmentOperation(documentId, name, AttachmentType_REVISION, "", changeVector)
	err := s.session.GetOperations().SendWithSessionInfo(operation, s.sessionInfo)
	if err != nil {
		return nil, err
	}
	res := operation.Command.Result
	return res, nil
}
