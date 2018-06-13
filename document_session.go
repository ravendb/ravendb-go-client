package ravendb

import (
	"encoding/json"
	"io"
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
	command := NewHeadDocumentCommand(id, nil)

	err := s._requestExecutor.executeCommandWithSessionInfo(command, s.sessionInfo)
	if err != nil {
		return false, err
	}

	ok := command.exists()
	return ok, nil
}

// TODO:    public <T> void refresh(T entity) {
// TODO:    protected String generateId(Object entity) {
// TODO:    public ResponseTimeInformation executeAllPendingLazyOperations() {
// TODO:    private boolean executeLazyOperationsSingleStep(ResponseTimeInformation responseTimeInformation, List<GetRequest> requests) {
// TODO:    public ILoaderWithInclude include(String path) {
// TODO:    public <T> Lazy<T> addLazyOperation(Class<T> clazz, ILazyOperation operation, Consumer<T> onEval) {
// TODO:    protected Lazy<Integer> addLazyCountOperation(ILazyOperation operation) {
// TODO:    public <T> Lazy<Map<String, T>> lazyLoadInternal(Class<T> clazz, String[] ids, String[] includes, Consumer<Map<String, T>> onEval)

func (s *DocumentSession) load(clazz reflect.Type, id string) (interface{}, error) {
	if id == "" {
		return Defaults_defaultValue(clazz), nil
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

func (s *DocumentSession) loadMulti(clazz reflect.Type, ids []string) (map[string]interface{}, error) {
	loadOperation := NewLoadOperation(s.InMemoryDocumentSessionOperations)
	err := s.loadInternalWithOperation(ids, loadOperation, nil)
	if err != nil {
		return nil, err
	}
	return loadOperation.getDocuments(clazz)
}

func (s *DocumentSession) loadInternalWithOperation(ids []string, operation *LoadOperation, stream io.Writer) error {
	operation.byIds(ids)

	command := operation.createRequest()
	if command != nil {
		s._requestExecutor.executeCommandWithSessionInfo(command, s.sessionInfo)

		if stream != nil {
			result := command.Result
			// TODO: serialize directly to stream
			d, err := json.Marshal(result)
			panicIf(err != nil, "json.Marshal() failed with %s", err)
			_, err = stream.Write(d)
			if err != nil {
				return err
			}
		} else {
			operation.setResult(command.Result)
		}
	}
	return nil
}

func (s *DocumentSession) loadInternalMulti(clazz reflect.Type, ids []string, includes []string) (map[string]interface{}, error) {
	loadOperation := NewLoadOperation(s.InMemoryDocumentSessionOperations)
	loadOperation.byIds(ids)
	loadOperation.withIncludes(includes)

	command := loadOperation.createRequest()
	if command != nil {
		err := s._requestExecutor.executeCommandWithSessionInfo(command, s.sessionInfo)
		if err != nil {
			return nil, err
		}
		loadOperation.setResult(command.Result)
	}

	return loadOperation.getDocuments(clazz)
}

// TODO: more
