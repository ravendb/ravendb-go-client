package ravendb

type BeforeDeleteEventArgs struct {
	_documentMetadata *IMetadataDictionary

	session    *InMemoryDocumentSessionOperations
	documentId string
	entity     interface{}
}

func NewBeforeDeleteEventArgs(session *InMemoryDocumentSessionOperations, documentId string, entity interface{}) *BeforeDeleteEventArgs {
	return &BeforeDeleteEventArgs{
		session:    session,
		documentId: documentId,
		entity:     entity,
	}
}

func (a *BeforeDeleteEventArgs) getSession() *InMemoryDocumentSessionOperations {
	return a.session
}

func (a *BeforeDeleteEventArgs) GetDocumentID() string {
	return a.documentId
}

func (a *BeforeDeleteEventArgs) getEntity() interface{} {
	return a.entity
}

func (a *BeforeDeleteEventArgs) getDocumentMetadata() *IMetadataDictionary {
	if a._documentMetadata == nil {
		a._documentMetadata, _ = a.session.GetMetadataFor(a.entity)
	}

	return a._documentMetadata
}
