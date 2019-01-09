package ravendb

// Note: Java's IAttachmentsSessionOperations is DocumentSessionAttachments

// TODO: make a unique wrapper type
type AttachmentsSessionOperations = DocumentSessionAttachments

type DocumentSessionAttachments struct {
	*DocumentSessionAttachmentsBase
}

func NewDocumentSessionAttachments(session *InMemoryDocumentSessionOperations) *DocumentSessionAttachments {
	res := &DocumentSessionAttachments{}
	res.DocumentSessionAttachmentsBase = NewDocumentSessionAttachmentsBase(session)
	return res
}

func (s *DocumentSessionAttachments) Exists(documentID string, name string) (bool, error) {
	command := NewHeadAttachmentCommand(documentID, name, nil)
	err := s.requestExecutor.ExecuteCommandWithSessionInfo(command, s.sessionInfo)
	if err != nil {
		return false, err
	}
	res := command.Result != ""
	return res, nil
}

func (s *DocumentSessionAttachments) Get(documentID string, name string) (*AttachmentResult, error) {
	operation := NewGetAttachmentOperation(documentID, name, AttachmentDocument, "", nil)
	err := s.session.GetOperations().SendWithSessionInfo(operation, s.sessionInfo)
	if err != nil {
		return nil, err
	}
	res := operation.Command.Result
	return res, nil
}

func (s *DocumentSessionAttachments) GetEntity(entity interface{}, name string) (*AttachmentResult, error) {
	document := getDocumentInfoByEntity(s.documents, entity)
	if document == nil {
		return nil, throwEntityNotInSession(entity)
	}
	return s.Get(document.id, name)
}

func (s *DocumentSessionAttachments) GetRevision(documentID string, name string, changeVector *string) (*AttachmentResult, error) {
	operation := NewGetAttachmentOperation(documentID, name, AttachmentRevision, "", changeVector)
	err := s.session.GetOperations().SendWithSessionInfo(operation, s.sessionInfo)
	if err != nil {
		return nil, err
	}
	res := operation.Command.Result
	return res, nil
}
