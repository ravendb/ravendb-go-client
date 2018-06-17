package ravendb

type AfterSaveChangesEventArgs struct {
	_documentMetadata *IMetadataDictionary

	session    *InMemoryDocumentSessionOperations
	documentId string
	entity     Object
}

func NewAfterSaveChangesEventArgs(session *InMemoryDocumentSessionOperations, documentId string, entity Object) *AfterSaveChangesEventArgs {
	return &AfterSaveChangesEventArgs{
		session:    session,
		documentId: documentId,
		entity:     entity,
	}
}

func (a *AfterSaveChangesEventArgs) getSession() *InMemoryDocumentSessionOperations {
	return a.session
}

func (a *AfterSaveChangesEventArgs) getDocumentId() string {
	return a.documentId
}

func (a *AfterSaveChangesEventArgs) getEntity() Object {
	return a.entity
}

func (a *AfterSaveChangesEventArgs) getDocumentMetadata() *IMetadataDictionary {
	if a._documentMetadata == nil {
		a._documentMetadata, _ = a.session.getMetadataFor(a.entity)
	}

	return a._documentMetadata
}
