package ravendb

type BeforeDeleteEventArgs struct {
	_documentMetadata *MetadataAsDictionary

	session    *InMemoryDocumentSessionOperations
	documentID string
	entity     interface{}
}

func NewBeforeDeleteEventArgs(session *InMemoryDocumentSessionOperations, documentID string, entity interface{}) *BeforeDeleteEventArgs {
	return &BeforeDeleteEventArgs{
		session:    session,
		documentID: documentID,
		entity:     entity,
	}
}

func (a *BeforeDeleteEventArgs) getSession() *InMemoryDocumentSessionOperations {
	return a.session
}

func (a *BeforeDeleteEventArgs) GetDocumentID() string {
	return a.documentID
}

func (a *BeforeDeleteEventArgs) getEntity() interface{} {
	return a.entity
}

func (a *BeforeDeleteEventArgs) getDocumentMetadata() *MetadataAsDictionary {
	if a._documentMetadata == nil {
		a._documentMetadata, _ = a.session.GetMetadataFor(a.entity)
	}

	return a._documentMetadata
}
