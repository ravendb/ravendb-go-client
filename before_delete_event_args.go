package ravendb

// BeforeDeleteEventArgs describes arguments for "before delete" listener
type BeforeDeleteEventArgs struct {
	documentMetadata *MetadataAsDictionary

	Session    *InMemoryDocumentSessionOperations
	DocumentID string
	Entity     interface{}
}

func newBeforeDeleteEventArgs(session *InMemoryDocumentSessionOperations, documentID string, entity interface{}) *BeforeDeleteEventArgs {
	return &BeforeDeleteEventArgs{
		Session:    session,
		DocumentID: documentID,
		Entity:     entity,
	}
}

// GetDocumentMetadata returns metadata for the entity being deleted
func (a *BeforeDeleteEventArgs) GetDocumentMetadata() *MetadataAsDictionary {
	if a.documentMetadata == nil {
		a.documentMetadata, _ = a.Session.GetMetadataFor(a.Entity)
	}

	return a.documentMetadata
}
