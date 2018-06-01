package ravendb

// ObjectNode is an alias for a json document represented as a map
// Name comes from Java implementation
type ObjectNode = map[string]interface{}

type JsonNodeType = interface{}

// TODO: remove it, it only exists to make initial porting faster
type Object = interface{}

// TODO: remove it, it only exists to make initial porting faster
type String = string

// DocumentSession is a Unit of Work for accessing RavenDB server
// https://sourcegraph.com/github.com/ravendb/RavenDB-Python-Client/-/blob/pyravendb/store/document_session.py#L18
// https://sourcegraph.com/github.com/ravendb/RavenDB-Python-Client/-/blob/pyravendb/store/document_session.py#L18
type DocumentSession struct {
	*InMemoryDocumentSessionOperations

	// _attachments *IAttachmentsSessionOperations
	// _revisions *IRevisionsSessionOperations
}

//    public IAdvancedSessionOperations advanced() {
//    public ILazySessionOperations lazily() {
//    public IEagerSessionOperations eagerly() {
//    public IAttachmentsSessionOperations attachments() {
//    public IRevisionsSessionOperations revisions() {

// NewDocumentSession creates a new DocumentSession
func NewDocumentSession(dbName string, store *DocumentStore, id string, re *RequestExecutor) *DocumentSession {
	res := &DocumentSession{
		InMemoryDocumentSessionOperations: NewInMemoryDocumentSessionOperations(dbName, store, re, id),
	}

	//TODO: res._attachments: NewDocumentSessionAttachments(res)
	//TODO: res._revisions = NewDocumentSessionRevisions(res)

	return res
}

func (s *DocumentSession) SaveChanges() error {
	saveChangeOperation := NewBatchOperation(s.InMemoryDocumentSessionOperations)

	command := saveChangeOperation.createRequest()
	if command == nil {
		return nil
	}
	err := s.RequestExecutor.executeCommandWithSessionInfo(command, &s.sessionInfo)
	if err != nil {
		return err
	}
	result := command.result.(ArrayNode)
	saveChangeOperation.setResult(result)
	return nil
}
