package ravendb

type BeforeStoreEventArgs struct {
	_documentMetadata *MetadataAsDictionary

	session    *InMemoryDocumentSessionOperations
	documentID string
	entity     interface{}
}

func NewBeforeStoreEventArgs(session *InMemoryDocumentSessionOperations, documentID string, entity interface{}) *BeforeStoreEventArgs {
	return &BeforeStoreEventArgs{
		session:    session,
		documentID: documentID,
		entity:     entity,
	}
}

func (a *BeforeStoreEventArgs) getSession() *InMemoryDocumentSessionOperations {
	return a.session
}

func (a *BeforeStoreEventArgs) GetDocumentID() string {
	return a.documentID
}

func (a *BeforeStoreEventArgs) getEntity() interface{} {
	return a.entity
}

func (a *BeforeStoreEventArgs) isMetadataAccessed() bool {
	return a._documentMetadata != nil
}

func (a *BeforeStoreEventArgs) getDocumentMetadata() *MetadataAsDictionary {
	if a._documentMetadata == nil {
		a._documentMetadata, _ = a.session.GetMetadataFor(a.entity)
	}

	return a._documentMetadata
}
