package ravendb

type BeforeStoreEventArgs struct {
	_documentMetadata *IMetadataDictionary

	session    *InMemoryDocumentSessionOperations
	documentId string
	entity     Object
}

func NewBeforeStoreEventArgs(session *InMemoryDocumentSessionOperations, documentId string, entity Object) *BeforeStoreEventArgs {
	return &BeforeStoreEventArgs{
		session:    session,
		documentId: documentId,
		entity:     entity,
	}
}

func (a *BeforeStoreEventArgs) getSession() *InMemoryDocumentSessionOperations {
	return a.session
}

func (a *BeforeStoreEventArgs) GetDocumentID() string {
	return a.documentId
}

func (a *BeforeStoreEventArgs) getEntity() Object {
	return a.entity
}

func (a *BeforeStoreEventArgs) isMetadataAccessed() bool {
	return a._documentMetadata != nil
}

func (a *BeforeStoreEventArgs) getDocumentMetadata() *IMetadataDictionary {
	if a._documentMetadata == nil {
		a._documentMetadata, _ = a.session.GetMetadataFor(a.entity)
	}

	return a._documentMetadata
}
