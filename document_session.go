package ravendb

import (
	"reflect"
)

// DocumentSession is a Unit of Work for accessing RavenDB server
// https://sourcegraph.com/github.com/ravendb/RavenDB-Python-Client/-/blob/pyravendb/store/document_session.py#L18
// https://sourcegraph.com/github.com/ravendb/RavenDB-Python-Client/-/blob/pyravendb/store/document_session.py#L18
type DocumentSession struct {
	*InMemoryDocumentSessionOperations

	// _attachments *IAttachmentsSessionOperations
	// _revisions *IRevisionsSessionOperations
}

// NewDocumentSession creates a new DocumentSession
func NewDocumentSession(dbName string, documentStore *DocumentStore, id string, re *RequestExecutor) *DocumentSession {
	res := &DocumentSession{
		InMemoryDocumentSessionOperations: NewInMemoryDocumentSessionOperations(dbName, documentStore, re, id),
	}

	//TODO: res._attachments: NewDocumentSessionAttachments(res)
	//TODO: res._revisions = NewDocumentSessionRevisions(res)

	return res
}

// TODO: consider exposing it as IAdvancedSessionOperations interface, like in Java
func (s *DocumentSession) advanced() *DocumentSession {
	return s
}

//    public ILazySessionOperations lazily() {
//    public IEagerSessionOperations eagerly() {
//    public IAttachmentsSessionOperations attachments() {
//    public IRevisionsSessionOperations revisions() {

func (s *DocumentSession) SaveChanges() error {
	saveChangeOperation := NewBatchOperation(s.InMemoryDocumentSessionOperations)

	command := saveChangeOperation.createRequest()
	if command == nil {
		return nil
	}
	err := s._requestExecutor.executeCommandWithSessionInfo(command, s.sessionInfo)
	if err != nil {
		return err
	}
	result := command.Result
	saveChangeOperation.setResult(result.Results)
	return nil
}

func (s *DocumentSession) exists(id string) (bool, error) {
	if s.documentsById.getValue(id) != nil {
		return true, nil
	}
	return false, nil
	/*

		command := NewHeadDocumentCommand(id, nil)

		err := s._requestExecutor.executeCommand(command, s.sessionInfo)
		if err != nil {
			return false, err
		}

		ok := command.getResult().(bool)
		return ok, nil
	*/
}

// TODO:    public <T> void refresh(T entity) {
// TODO:    protected String generateId(Object entity) {
// TODO:    public ResponseTimeInformation executeAllPendingLazyOperations() {
// TODO:    private boolean executeLazyOperationsSingleStep(ResponseTimeInformation responseTimeInformation, List<GetRequest> requests) {
// TODO:    public ILoaderWithInclude include(String path) {
// TODO:    public <T> Lazy<T> addLazyOperation(Class<T> clazz, ILazyOperation operation, Consumer<T> onEval) {
// TODO:    protected Lazy<Integer> addLazyCountOperation(ILazyOperation operation) {
// TODO:    public <T> Lazy<Map<String, T>> lazyLoadInternal(Class<T> clazz, String[] ids, String[] includes, Consumer<Map<String, T>> onEval)

func (s *DocumentSession) load(clazz reflect.Type, id string) interface{} {
	if id == "" {
		return nil
	}
	loadOperation := NewLoadOperation(s.InMemoryDocumentSessionOperations)

	loadOperation.byId(id)

	command := loadOperation.createRequest()

	if command != nil {
		s._requestExecutor.executeCommandWithSessionInfo(command, s.sessionInfo)
		result := command.Result
		loadOperation.setResult(result)
	}

	return loadOperation.getDocument(clazz)
}

// TODO: more
