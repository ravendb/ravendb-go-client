package ravendb

type AfterSaveChangesEventArgs struct {
	_documentMetadata *MetadataAsDictionary

	Session    *InMemoryDocumentSessionOperations
	DocumentId string
	Entity     interface{}
}

func NewAfterSaveChangesEventArgs(session *InMemoryDocumentSessionOperations, documentId string, entity interface{}) *AfterSaveChangesEventArgs {
	return &AfterSaveChangesEventArgs{
		Session:    session,
		DocumentId: documentId,
		Entity:     entity,
	}
}

func (a *AfterSaveChangesEventArgs) GetDocumentMetadata() *MetadataAsDictionary {
	if a._documentMetadata == nil {
		a._documentMetadata, _ = a.Session.GetMetadataFor(a.Entity)
	}

	return a._documentMetadata
}
