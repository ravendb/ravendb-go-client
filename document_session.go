package ravendb

import (
	"encoding/json"
	"io"
	"reflect"
)

// DocumentSession is a Unit of Work for accessing RavenDB server
type DocumentSession struct {
	*InMemoryDocumentSessionOperations

	_attachments *IAttachmentsSessionOperations
	_revisions   *IRevisionsSessionOperations
	_valsCount   int
	_customCount int
}

// TODO: consider exposing it as IAdvancedSessionOperations interface, like in Java
func (s *DocumentSession) advanced() *DocumentSession {
	return s
}

//    public ILazySessionOperations lazily() {
//    public IEagerSessionOperations eagerly() {

func (s *DocumentSession) attachments() *IAttachmentsSessionOperations {
	if s._attachments == nil {
		s._attachments = NewDocumentSessionAttachments(s.InMemoryDocumentSessionOperations)
	}
	return s._attachments
}

func (s *DocumentSession) revisions() *IRevisionsSessionOperations {
	return s._revisions
}

// NewDocumentSession creates a new DocumentSession
func NewDocumentSession(dbName string, documentStore *DocumentStore, id string, re *RequestExecutor) *DocumentSession {
	res := &DocumentSession{
		InMemoryDocumentSessionOperations: NewInMemoryDocumentSessionOperations(dbName, documentStore, re, id),
	}

	//TODO: res._attachments: NewDocumentSessionAttachments(res)
	res._revisions = NewDocumentSessionRevisions(res.InMemoryDocumentSessionOperations)

	return res
}

func (s *DocumentSession) SaveChanges() error {
	saveChangeOperation := NewBatchOperation(s.InMemoryDocumentSessionOperations)

	command := saveChangeOperation.createRequest()
	if command == nil {
		return nil
	}
	defer command.Close()
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

func (s *DocumentSession) refresh(entity Object) error {
	documentInfo := s.documentsByEntity[entity]
	if documentInfo == nil {
		return NewIllegalStateException("Cannot refresh a transient instance")
	}
	if err := s.incrementRequestCount(); err != nil {
		return err
	}

	command := NewGetDocumentsCommand([]string{documentInfo.getId()}, nil, false)
	err := s._requestExecutor.executeCommandWithSessionInfo(command, s.sessionInfo)
	if err != nil {
		return err
	}
	return s.refreshInternal(entity, command, documentInfo)
}

// TODO:    protected string generateId(Object entity) {
// TODO:    public ResponseTimeInformation executeAllPendingLazyOperations() {
// TODO:    private boolean executeLazyOperationsSingleStep(ResponseTimeInformation responseTimeInformation, List<GetRequest> requests) {

func (s *DocumentSession) include(path string) ILoaderWithInclude {
	return NewMultiLoaderWithInclude(s).include(path)
}

// TODO:    public <T> Lazy<T> addLazyOperation(Class<T> clazz, ILazyOperation operation, Consumer<T> onEval) {
// TODO:    protected Lazy<Integer> addLazyCountOperation(ILazyOperation operation) {
// TODO:    public <T> Lazy<Map<string, T>> lazyLoadInternal(Class<T> clazz, string[] ids, string[] includes, Consumer<Map<string, T>> onEval)

func (s *DocumentSession) load(clazz reflect.Type, id string) (interface{}, error) {
	if id == "" {
		return Defaults_defaultValue(clazz), nil
	}

	loadOperation := NewLoadOperation(s.InMemoryDocumentSessionOperations)

	loadOperation.byId(id)

	command := loadOperation.createRequest()

	if command != nil {
		err := s._requestExecutor.executeCommandWithSessionInfo(command, s.sessionInfo)
		if err != nil {
			return nil, err
		}
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
		err := s._requestExecutor.executeCommandWithSessionInfo(command, s.sessionInfo)
		if err != nil {
			return err
		}

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

func (s *DocumentSession) loadStartingWith(clazz reflect.Type, idPrefix string) ([]interface{}, error) {
	return s.loadStartingWithFull(clazz, idPrefix, "", 0, 25, "", "")
}

func (s *DocumentSession) loadStartingWithFull(clazz reflect.Type, idPrefix string, matches string, start int, pageSize int, exclude string, startAfter string) ([]interface{}, error) {
	loadStartingWithOperation := NewLoadStartingWithOperation(s.InMemoryDocumentSessionOperations)
	_, err := s.loadStartingWithInternal(idPrefix, loadStartingWithOperation, nil, matches, start, pageSize, exclude, startAfter)
	if err != nil {
		return nil, err
	}
	return loadStartingWithOperation.getDocuments(clazz)
}

// public void loadStartingWithIntoStream(String idPrefix, OutputStream output) {
// public void loadStartingWithIntoStream(String idPrefix, OutputStream output, String matches) {
// public void loadStartingWithIntoStream(String idPrefix, OutputStream output, String matches, int start)
// public void loadStartingWithIntoStream(String idPrefix, OutputStream output, String matches, int start, int pageSize) {
// public void loadStartingWithIntoStream(String idPrefix, OutputStream output, String matches, int start, int pageSize, String exclude) {
// public void loadStartingWithIntoStream(String idPrefix, OutputStream output, String matches, int start, int

func (s *DocumentSession) loadStartingWithInternal(idPrefix string, operation *LoadStartingWithOperation, stream io.Writer,
	matches string, start int, pageSize int, exclude string, startAfter string) (*GetDocumentsCommand, error) {

	operation.withStartWithFull(idPrefix, matches, start, pageSize, exclude, startAfter)

	command := operation.createRequest()
	if command != nil {
		err := s._requestExecutor.executeCommandWithSessionInfo(command, s.sessionInfo)
		if err != nil {
			return nil, err
		}

		if stream != nil {
			panicIf(true, "NYI")
			/*
				try {
					GetDocumentsResult result = command.getResult();
					JsonExtensions.getDefaultMapper().writeValue(stream, result);
				} catch (IOException e) {
					throw new RuntimeException("Unable to serialize returned value into stream" + e.getMessage(), e);
				}
			*/
		} else {
			operation.setResult(command.Result)
		}
	}
	return command, nil
}

// public void loadIntoStream(Collection<String> ids, OutputStream output) {
// public <T, U> void increment(T entity, String path, U valueToAdd) {
// public <T, U> void increment(String id, String path, U valueToAdd) {
// public <T, U> void patch(T entity, String path, U value) {
// public <T, U> void patch(String id, String path, U value) {
// public <T, U> void patch(T entity, String pathToArray, Consumer<JavaScriptArray<U>> arrayAdder) {
// public <T, U> void patch(String id, String pathToArray, Consumer<JavaScriptArray<U>> arrayAdder) {
// private boolean tryMergePatches(String id, PatchRequest patchRequest) {
// public <T, TIndex extends AbstractIndexCreationTask> IDocumentQuery<T> documentQuery(Class<T> clazz, Class<TIndex> indexClazz) {
// public <T> IDocumentQuery<T> documentQuery(Class<T> clazz) {
// public <T> IDocumentQuery<T> documentQuery(Class<T> clazz, String indexName, String collectionName, boolean
// public <T> IRawDocumentQuery<T> rawQuery(Class<T> clazz, String query) {
// public <T> IDocumentQuery<T> query(Class<T> clazz) {
// public <T> IDocumentQuery<T> query(Class<T> clazz, Query collectionOrIndexName) {
// public <T, TIndex extends AbstractIndexCreationTask> IDocumentQuery<T> query(Class<T> clazz, Class<TIndex>
// public <T> CloseableIterator<StreamResult<T>> stream(IDocumentQuery<T> query) {
// public <T> CloseableIterator<StreamResult<T>> stream(IDocumentQuery<T> query, Reference<StreamQueryStatistics> streamQueryStats) {
// public <T> CloseableIterator<StreamResult<T>> stream(IRawDocumentQuery<T> query) {
// public <T> CloseableIterator<StreamResult<T>> stream(IRawDocumentQuery<T> query, Reference<StreamQueryStatistics> streamQueryStats) {
// private <T> CloseableIterator<StreamResult<T>> yieldResults(AbstractDocumentQuery query, CloseableIterator<ObjectNode> enumerator) {
// public <T> void streamInto(IRawDocumentQuery<T> query, OutputStream output) {
// public <T> void streamInto(IDocumentQuery<T> query, OutputStream output) {
// private <T> StreamResult<T> createStreamResult(Class<T> clazz, ObjectNode json, FieldsToFetchToken fieldsToFetch) throws IOException {
// public <T> CloseableIterator<StreamResult<T>> stream(Class<T> clazz, String startsWith) {
// public <T> CloseableIterator<StreamResult<T>> stream(Class<T> clazz, String startsWith, String matches) {
// public <T> CloseableIterator<StreamResult<T>> stream(Class<T> clazz, String startsWith, String matches, int start) {
// public <T> CloseableIterator<StreamResult<T>> stream(Class<T> clazz, String startsWith, String matches, int start, int pageSize) {
// public <T> CloseableIterator<StreamResult<T>> stream(Class<T> clazz, String startsWith, String matches, int start, int pageSize, String startAfter) {
