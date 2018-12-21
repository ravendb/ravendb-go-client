package ravendb

// BeforeDeleteEventArgs describes
type BeforeDeleteEventArgs struct {
	documentMetadata *MetadataAsDictionary

	session    *InMemoryDocumentSessionOperations
	DocumentID string
	Entity     interface{}
}

func newBeforeDeleteEventArgs(session *InMemoryDocumentSessionOperations, documentID string, entity interface{}) *BeforeDeleteEventArgs {
	return &BeforeDeleteEventArgs{
		session:    session,
		DocumentID: documentID,
		Entity:     entity,
	}
}

// GetDocumentMetadata returns metadata for the entity being deleted
func (a *BeforeDeleteEventArgs) GetDocumentMetadata() *MetadataAsDictionary {
	if a.documentMetadata == nil {
		a.documentMetadata, _ = a.session.GetMetadataFor(a.Entity)
	}

	return a.documentMetadata
}
