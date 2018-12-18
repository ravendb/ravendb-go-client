package ravendb

type AfterSaveChangesEventArgs struct {
	_documentMetadata *MetadataAsDictionary

	Session    *InMemoryDocumentSessionOperations
	DocumentID string
	Entity     interface{}
}

func NewAfterSaveChangesEventArgs(session *InMemoryDocumentSessionOperations, documentID string, entity interface{}) *AfterSaveChangesEventArgs {
	return &AfterSaveChangesEventArgs{
		Session:    session,
		DocumentID: documentID,
		Entity:     entity,
	}
}

func (a *AfterSaveChangesEventArgs) GetDocumentMetadata() *MetadataAsDictionary {
	if a._documentMetadata == nil {
		a._documentMetadata, _ = a.Session.GetMetadataFor(a.Entity)
	}

	return a._documentMetadata
}
