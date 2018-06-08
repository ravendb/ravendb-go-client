package ravendb

import "errors"

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
func NewDocumentSession(dbName string, documentStore *DocumentStore, id string, re *RequestExecutor) *DocumentSession {
	res := &DocumentSession{
		InMemoryDocumentSessionOperations: NewInMemoryDocumentSessionOperations(dbName, documentStore, re, id),
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
	result := command.result.(*JSONArrayResult)
	saveChangeOperation.setResult(result.Results)
	return nil
}

// TODO:    public boolean exists(String id) {
// TODO:    public <T> void refresh(T entity) {
// TODO:    protected String generateId(Object entity) {
// TODO:    public ResponseTimeInformation executeAllPendingLazyOperations() {
// TODO:    private boolean executeLazyOperationsSingleStep(ResponseTimeInformation responseTimeInformation, List<GetRequest> requests) {
// TODO:    public ILoaderWithInclude include(String path) {
// TODO:    public <T> Lazy<T> addLazyOperation(Class<T> clazz, ILazyOperation operation, Consumer<T> onEval) {
// TODO:    protected Lazy<Integer> addLazyCountOperation(ILazyOperation operation) {
// TODO:    public <T> Lazy<Map<String, T>> lazyLoadInternal(Class<T> clazz, String[] ids, String[] includes, Consumer<Map<String, T>> onEval)

func (s *DocumentSession) load(v interface{}, id string) error {
	return errors.New("NYI")
}

// TODO: much more
