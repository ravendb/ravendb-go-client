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
func (s *DocumentSession) Advanced() *DocumentSession {
	return s
}

//    public ILazySessionOperations lazily() {
//    public IEagerSessionOperations eagerly() {

func (s *DocumentSession) Attachments() *IAttachmentsSessionOperations {
	if s._attachments == nil {
		s._attachments = NewDocumentSessionAttachments(s.InMemoryDocumentSessionOperations)
	}
	return s._attachments
}

func (s *DocumentSession) Revisions() *IRevisionsSessionOperations {
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

	command, err := saveChangeOperation.CreateRequest()
	if err != nil {
		return err
	}
	if command == nil {
		return nil
	}
	defer command.Close()
	err = s._requestExecutor.ExecuteCommandWithSessionInfo(command, s.sessionInfo)
	if err != nil {
		return err
	}
	result := command.Result
	saveChangeOperation.setResult(result.Results)
	return nil
}

func (s *DocumentSession) Exists(id string) (bool, error) {
	if s.documentsById.getValue(id) != nil {
		return true, nil
	}
	command := NewHeadDocumentCommand(id, nil)

	err := s._requestExecutor.ExecuteCommandWithSessionInfo(command, s.sessionInfo)
	if err != nil {
		return false, err
	}

	ok := command.Exists()
	return ok, nil
}

func (s *DocumentSession) Refresh(entity Object) error {
	documentInfo := s.documentsByEntity[entity]
	if documentInfo == nil {
		return NewIllegalStateException("Cannot refresh a transient instance")
	}
	if err := s.IncrementRequestCount(); err != nil {
		return err
	}

	command := NewGetDocumentsCommand([]string{documentInfo.getId()}, nil, false)
	err := s._requestExecutor.ExecuteCommandWithSessionInfo(command, s.sessionInfo)
	if err != nil {
		return err
	}
	return s.refreshInternal(entity, command, documentInfo)
}

// TODO:    protected string generateId(Object entity) {
// TODO:    public ResponseTimeInformation ExecuteAllPendingLazyOperations() {
// TODO:    private boolean ExecuteLazyOperationsSingleStep(ResponseTimeInformation responseTimeInformation, List<GetRequest> requests) {

func (s *DocumentSession) Include(path string) ILoaderWithInclude {
	return NewMultiLoaderWithInclude(s).Include(path)
}

// TODO:    public <T> Lazy<T> addLazyOperation(reflect.Type clazz, ILazyOperation operation, Consumer<T> onEval) {
// TODO:    protected Lazy<Integer> addLazyCountOperation(ILazyOperation operation) {
// TODO:    public <T> Lazy<Map<string, T>> lazyLoadInternal(reflect.Type clazz, string[] ids, string[] includes, Consumer<Map<string, T>> onEval)

func (s *DocumentSession) Load(clazz reflect.Type, id string) (interface{}, error) {
	if id == "" {
		return Defaults_defaultValue(clazz), nil
	}

	loadOperation := NewLoadOperation(s.InMemoryDocumentSessionOperations)

	loadOperation.byId(id)

	command := loadOperation.CreateRequest()

	if command != nil {
		err := s._requestExecutor.ExecuteCommandWithSessionInfo(command, s.sessionInfo)
		if err != nil {
			return nil, err
		}
		result := command.Result
		loadOperation.setResult(result)
	}

	return loadOperation.getDocument(clazz)
}

func (s *DocumentSession) LoadMulti(clazz reflect.Type, ids []string) (map[string]interface{}, error) {
	loadOperation := NewLoadOperation(s.InMemoryDocumentSessionOperations)
	err := s.loadInternalWithOperation(ids, loadOperation, nil)
	if err != nil {
		return nil, err
	}
	return loadOperation.getDocuments(clazz)
}

func (s *DocumentSession) loadInternalWithOperation(ids []string, operation *LoadOperation, stream io.Writer) error {
	operation.byIds(ids)

	command := operation.CreateRequest()
	if command != nil {
		err := s._requestExecutor.ExecuteCommandWithSessionInfo(command, s.sessionInfo)
		if err != nil {
			return err
		}

		if stream != nil {
			result := command.Result
			enc := json.NewEncoder(stream)
			err = enc.Encode(result)
			panicIf(err != nil, "enc.Encode() failed with %s", err)
		} else {
			operation.setResult(command.Result)
		}
	}
	return nil
}

func (s *DocumentSession) LoadInternalMulti(clazz reflect.Type, ids []string, includes []string) (map[string]interface{}, error) {
	loadOperation := NewLoadOperation(s.InMemoryDocumentSessionOperations)
	loadOperation.byIds(ids)
	loadOperation.withIncludes(includes)

	command := loadOperation.CreateRequest()
	if command != nil {
		err := s._requestExecutor.ExecuteCommandWithSessionInfo(command, s.sessionInfo)
		if err != nil {
			return nil, err
		}
		loadOperation.setResult(command.Result)
	}

	return loadOperation.getDocuments(clazz)
}

func (s *DocumentSession) LoadStartingWith(clazz reflect.Type, idPrefix string) ([]interface{}, error) {
	return s.LoadStartingWithFull(clazz, idPrefix, "", 0, 25, "", "")
}

func (s *DocumentSession) LoadStartingWithFull(clazz reflect.Type, idPrefix string, matches string, start int, pageSize int, exclude string, startAfter string) ([]interface{}, error) {
	loadStartingWithOperation := NewLoadStartingWithOperation(s.InMemoryDocumentSessionOperations)
	_, err := s.loadStartingWithInternal(idPrefix, loadStartingWithOperation, nil, matches, start, pageSize, exclude, startAfter)
	if err != nil {
		return nil, err
	}
	return loadStartingWithOperation.getDocuments(clazz)
}

func (s *DocumentSession) LoadStartingWithIntoStream(idPrefix string, output io.Writer) error {
	return s.LoadStartingWithIntoStreamAll(idPrefix, output, "", 0, 25, "", "")
}

func (s *DocumentSession) LoadStartingWithIntoStream2(idPrefix string, output io.Writer, matches string) error {
	return s.LoadStartingWithIntoStreamAll(idPrefix, output, matches, 0, 25, "", "")
}

func (s *DocumentSession) LoadStartingWithIntoStream3(idPrefix string, output io.Writer, matches string, start int) error {
	return s.LoadStartingWithIntoStreamAll(idPrefix, output, matches, start, 25, "", "")
}

func (s *DocumentSession) LoadStartingWithIntoStream4(idPrefix string, output io.Writer, matches string, start int, pageSize int) error {
	return s.LoadStartingWithIntoStreamAll(idPrefix, output, matches, start, pageSize, "", "")
}

func (s *DocumentSession) LoadStartingWithIntoStream5(idPrefix string, output io.Writer, matches string, start int, pageSize int, exclude string) error {
	return s.LoadStartingWithIntoStreamAll(idPrefix, output, matches, start, pageSize, "", "")
}

func (s *DocumentSession) LoadStartingWithIntoStreamAll(idPrefix string, output io.Writer, matches string, start int, pageSize int, exclude string, startAfter string) error {
	op := NewLoadStartingWithOperation(s.InMemoryDocumentSessionOperations)
	_, err := s.loadStartingWithInternal(idPrefix, op, output, matches, start, pageSize, exclude, startAfter)
	return err
}

func (s *DocumentSession) loadStartingWithInternal(idPrefix string, operation *LoadStartingWithOperation, stream io.Writer,
	matches string, start int, pageSize int, exclude string, startAfter string) (*GetDocumentsCommand, error) {

	operation.withStartWithFull(idPrefix, matches, start, pageSize, exclude, startAfter)

	command := operation.CreateRequest()
	if command != nil {
		err := s._requestExecutor.ExecuteCommandWithSessionInfo(command, s.sessionInfo)
		if err != nil {
			return nil, err
		}

		if stream != nil {
			result := command.Result
			enc := json.NewEncoder(stream)
			err = enc.Encode(result)
			panicIf(err != nil, "enc.Encode() failed with %s", err)
		} else {
			operation.setResult(command.Result)
		}
	}
	return command, nil
}

func (s *DocumentSession) LoadIntoStream(ids []string, output io.Writer) error {
	op := NewLoadOperation(s.InMemoryDocumentSessionOperations)
	return s.loadInternalWithOperation(ids, op, output)
}

// public <T, U> void increment(T entity, string path, U valueToAdd) {
// public <T, U> void increment(string id, string path, U valueToAdd) {

	// public <T, U> void patch(T entity, string path, U value) {
// public <T, U> void patch(string id, string path, U value) {
// public <T, U> void patch(T entity, string pathToArray, Consumer<JavaScriptArray<U>> arrayAdder) {
// public <T, U> void patch(string id, string pathToArray, Consumer<JavaScriptArray<U>> arrayAdder) {
	// private boolean tryMergePatches(string id, PatchRequest patchRequest) {

		// public <T, TIndex extends AbstractIndexCreationTask> IDocumentQuery<T> documentQuery(reflect.Type clazz, Class<TIndex> indexClazz) {

func (s *DocumentSession) DocumentQueryInIndex(clazz reflect.Type, index *AbstractIndexCreationTask) *DocumentQuery {
	return s.DocumentQueryAll(clazz, index.GetIndexName(), "", index.IsMapReduce())
}

func (s *DocumentSession) DocumentQuery(clazz reflect.Type) *DocumentQuery {
	return s.DocumentQueryAll(clazz, "", "", false)
}

func (s *DocumentSession) DocumentQueryAll(clazz reflect.Type, indexName string, collectionName string, isMapReduce bool) *DocumentQuery {
	indexName, collectionName = s.processQueryParameters(clazz, indexName, collectionName, s.GetConventions())

	return NewDocumentQuery(clazz, s.InMemoryDocumentSessionOperations, indexName, collectionName, isMapReduce)
}

func (s *DocumentSession) RawQuery(clazz reflect.Type, query string) *IRawDocumentQuery {
	return NewRawDocumentQuery(clazz, s.InMemoryDocumentSessionOperations, query)
}

func (s *DocumentSession) Query(clazz reflect.Type) *DocumentQuery {
	return s.DocumentQueryAll(clazz, "", "", false)
}

func (s *DocumentSession) QueryWithQuery(clazz reflect.Type, collectionOrIndexName *Query) *DocumentQuery {
	if StringUtils_isNotEmpty(collectionOrIndexName.getCollection()) {
		return s.DocumentQueryAll(clazz, "", collectionOrIndexName.getCollection(), false)
	}

	return s.DocumentQueryAll(clazz, collectionOrIndexName.getIndexName(), "", false)
}

func (s *DocumentSession) QueryInIndex(clazz reflect.Type, index *AbstractIndexCreationTask) *DocumentQuery {
	return s.DocumentQueryInIndex(clazz, index)
}

// public <T> CloseableIterator<StreamResult<T>> stream(IDocumentQuery<T> query) {
// public <T> CloseableIterator<StreamResult<T>> stream(IDocumentQuery<T> query, Reference<StreamQueryStatistics> streamQueryStats) {
// public <T> CloseableIterator<StreamResult<T>> stream(IRawDocumentQuery<T> query) {
// public <T> CloseableIterator<StreamResult<T>> stream(IRawDocumentQuery<T> query, Reference<StreamQueryStatistics> streamQueryStats) {
// private <T> CloseableIterator<StreamResult<T>> yieldResults(AbstractDocumentQuery query, CloseableIterator<ObjectNode> enumerator) {
// public <T> void streamInto(IRawDocumentQuery<T> query, OutputStream output) {
// public <T> void streamInto(IDocumentQuery<T> query, OutputStream output) {

// private <T> StreamResult<T> createStreamResult(reflect.Type clazz, ObjectNode json, FieldsToFetchToken fieldsToFetch) throws IOException {
// public <T> CloseableIterator<StreamResult<T>> stream(reflect.Type clazz, string startsWith) {
// public <T> CloseableIterator<StreamResult<T>> stream(reflect.Type clazz, string startsWith, string matches) {
// public <T> CloseableIterator<StreamResult<T>> stream(reflect.Type clazz, string startsWith, string matches, int start) {
// public <T> CloseableIterator<StreamResult<T>> stream(reflect.Type clazz, string startsWith, string matches, int start, int pageSize) {
// public <T> CloseableIterator<StreamResult<T>> stream(reflect.Type clazz, string startsWith, string matches, int start, int pageSize, string startAfter) {
