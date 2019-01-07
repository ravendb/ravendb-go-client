package ravendb

// AfterSaveChangesEventArgs describes arguments for "after save changes" listener
type AfterSaveChangesEventArgs struct {
	documentMetadata *MetadataAsDictionary

	Session    *InMemoryDocumentSessionOperations
	DocumentID string
	Entity     interface{}
}

func newAfterSaveChangesEventArgs(session *InMemoryDocumentSessionOperations, documentID string, entity interface{}) *AfterSaveChangesEventArgs {
	return &AfterSaveChangesEventArgs{
		Session:    session,
		DocumentID: documentID,
		Entity:     entity,
	}
}

// GetDocumentMetadata returns metadata for the entity represented by this event
func (a *AfterSaveChangesEventArgs) GetDocumentMetadata() *MetadataAsDictionary {
	if a.documentMetadata == nil {
		a.documentMetadata, _ = a.Session.GetMetadataFor(a.Entity)
	}

	return a.documentMetadata
}
