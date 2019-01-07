package ravendb

// BeforeStoreEventArgs describe arguments for "before store" listener
type BeforeStoreEventArgs struct {
	documentMetadata *MetadataAsDictionary

	Session    *InMemoryDocumentSessionOperations
	DocumentID string
	Entity     interface{}
}

func newBeforeStoreEventArgs(session *InMemoryDocumentSessionOperations, documentID string, entity interface{}) *BeforeStoreEventArgs {
	return &BeforeStoreEventArgs{
		Session:    session,
		DocumentID: documentID,
		Entity:     entity,
	}
}

func (a *BeforeStoreEventArgs) isMetadataAccessed() bool {
	return a.documentMetadata != nil
}

// GetDocumentMetadata returns metadata for entity represented by this event
func (a *BeforeStoreEventArgs) GetDocumentMetadata() *MetadataAsDictionary {
	if a.documentMetadata == nil {
		a.documentMetadata, _ = a.Session.GetMetadataFor(a.Entity)
	}

	return a.documentMetadata
}
