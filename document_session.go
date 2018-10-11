package ravendb

import (
	"encoding/json"
	"io"
	"reflect"
	"strconv"
)

// type IDocumentSessionImpl = DocumentSession

// TODO: decide if we want to return ErrNotFound or nil if the value is not found
// Java returns nil (which, I guess, is default value for reference (i.e. all) types)
// var ErrNotFound = errors.New("Not found")
var ErrNotFound = error(nil)

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

func (s *DocumentSession) Lazily() *ILazySessionOperations {
	return NewLazySessionOperations(s)
}

// TODO: remove in API cleanup phase
type IEagerSessionOperations = DocumentSession

func (s *DocumentSession) Eagerly() *IEagerSessionOperations {
	return s
}

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

	command := NewGetDocumentsCommand([]string{documentInfo.id}, nil, false)
	err := s._requestExecutor.ExecuteCommandWithSessionInfo(command, s.sessionInfo)
	if err != nil {
		return err
	}
	return s.refreshInternal(entity, command, documentInfo)
}

// TODO:    protected string generateId(Object entity) {

func (s *DocumentSession) Include(path string) *MultiLoaderWithInclude {
	return NewMultiLoaderWithInclude(s).Include(path)
}

// TODO:    public <T> Lazy<T> addLazyOperation(reflect.Type clazz, ILazyOperation operation, Consumer<T> onEval) {
// TODO:    protected Lazy<Integer> addLazyCountOperation(ILazyOperation operation) {
// TODO:    public <T> Lazy<Map<string, T>> lazyLoadInternal(reflect.Type clazz, string[] ids, string[] includes, Consumer<Map<string, T>> onEval)

func (s *DocumentSession) Load(result interface{}, id string) error {
	if id == "" {
		// TODO: or should it return default value?
		return ErrNotFound
	}
	loadOperation := NewLoadOperation(s.InMemoryDocumentSessionOperations)

	loadOperation.byId(id)

	command := loadOperation.CreateRequest()

	if command != nil {
		err := s._requestExecutor.ExecuteCommandWithSessionInfo(command, s.sessionInfo)
		if err != nil {
			return err
		}
		result := command.Result
		loadOperation.setResult(result)
	}

	return loadOperation.getDocument(result)
}

// LoadMulti loads multiple values with given ids into results, which should
// be a map from string (id) to pointer to struct
func (s *DocumentSession) LoadMulti(results interface{}, ids []string) error {
	loadOperation := NewLoadOperation(s.InMemoryDocumentSessionOperations)
	err := s.loadInternalWithOperation(ids, loadOperation, nil)
	if err != nil {
		return err
	}
	return loadOperation.getDocuments(results)
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

// results should be map[string]*struct
func (s *DocumentSession) loadInternalMulti(results interface{}, ids []string, includes []string) error {
	loadOperation := NewLoadOperation(s.InMemoryDocumentSessionOperations)
	loadOperation.byIds(ids)
	loadOperation.withIncludes(includes)

	command := loadOperation.CreateRequest()
	if command != nil {
		err := s._requestExecutor.ExecuteCommandWithSessionInfo(command, s.sessionInfo)
		if err != nil {
			return err
		}
		loadOperation.setResult(command.Result)
	}

	return loadOperation.getDocuments(results)
}

func (s *DocumentSession) LoadStartingWith(results interface{}, idPrefix string) error {
	return s.LoadStartingWithFull(results, idPrefix, "", 0, 25, "", "")
}

func (s *DocumentSession) LoadStartingWithFull(results interface{}, idPrefix string, matches string, start int, pageSize int, exclude string, startAfter string) error {
	loadStartingWithOperation := NewLoadStartingWithOperation(s.InMemoryDocumentSessionOperations)
	_, err := s.loadStartingWithInternal(idPrefix, loadStartingWithOperation, nil, matches, start, pageSize, exclude, startAfter)
	if err != nil {
		return err
	}
	return loadStartingWithOperation.getDocuments(results)
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

func (s *DocumentSession) IncrementEntity(entity interface{}, path string, valueToAdd interface{}) error {
	metadata, err := s.GetMetadataFor(entity)
	if err != nil {
		return err
	}
	// TODO: return an error if no id or id not string
	id, _ := metadata.Get(Constants_Documents_Metadata_ID)
	return s.IncrementByID(id.(string), path, valueToAdd)
}

func (s *DocumentSession) IncrementByID(id string, path string, valueToAdd interface{}) error {
	patchRequest := NewPatchRequest()

	valsCountStr := strconv.Itoa(s._valsCount)
	patchRequest.SetScript("this." + path + " += args.val_" + valsCountStr + ";")

	m := map[string]interface{}{
		"val_" + valsCountStr: valueToAdd,
	}
	patchRequest.SetValues(m)

	s._valsCount++

	if !s.tryMergePatches(id, patchRequest) {
		cmdData := NewPatchCommandData(id, nil, patchRequest, nil)
		s.Defer(cmdData)
	}
	return nil
}

func (s *DocumentSession) PatchEntity(entity interface{}, path string, value interface{}) error {
	metadata, err := s.GetMetadataFor(entity)
	if err != nil {
		return err
	}
	// TODO: return an error if no id or id not string
	id, _ := metadata.Get(Constants_Documents_Metadata_ID)
	return s.PatchByID(id.(string), path, value)
}

func (s *DocumentSession) PatchByID(id string, path string, value interface{}) error {
	patchRequest := NewPatchRequest()
	valsCountStr := strconv.Itoa(s._valsCount)
	patchRequest.SetScript("this." + path + " = args.val_" + valsCountStr + ";")
	m := map[string]interface{}{
		"val_" + valsCountStr: value,
	}
	patchRequest.SetValues(m)

	s._valsCount++

	if !s.tryMergePatches(id, patchRequest) {
		cmdData := NewPatchCommandData(id, nil, patchRequest, nil)
		s.Defer(cmdData)
	}
	return nil
}

func (s *DocumentSession) PatchArrayInEntity(entity interface{}, pathToArray string, arrayAdder func(*JavaScriptArray)) error {
	metadata, err := s.GetMetadataFor(entity)
	if err != nil {
		return err
	}
	// TODO: return an error if no id or id not string
	id, _ := metadata.Get(Constants_Documents_Metadata_ID)
	return s.PatchArrayByID(id.(string), pathToArray, arrayAdder)
}

func (s *DocumentSession) PatchArrayByID(id string, pathToArray string, arrayAdder func(*JavaScriptArray)) error {
	s._customCount++
	scriptArray := NewJavaScriptArray(s._customCount, pathToArray)

	arrayAdder(scriptArray)

	patchRequest := NewPatchRequest()
	patchRequest.SetScript(scriptArray.getScript())
	patchRequest.SetValues(scriptArray.Parameters)

	if !s.tryMergePatches(id, patchRequest) {
		cmdData := NewPatchCommandData(id, nil, patchRequest, nil)
		s.Defer(cmdData)
	}
	return nil
}

func removeDeferredCommand(a []ICommandData, el ICommandData) []ICommandData {
	idx := -1
	n := len(a)
	for i := 0; i < n; i++ {
		if a[i] == el {
			idx = i
			break
		}
	}
	panicIf(idx == -1, "didn't find el in a")
	return append(a[:idx], a[idx+1:]...)
}

func (s *DocumentSession) tryMergePatches(id string, patchRequest *PatchRequest) bool {
	idType := IdTypeAndName_create(id, CommandType_PATCH, "")
	command := s.deferredCommandsMap[idType]
	if command == nil {
		return false
	}

	s.deferredCommands = removeDeferredCommand(s.deferredCommands, command)

	// We'll overwrite the deferredCommandsMap when calling Defer
	// No need to call deferredCommandsMap.remove((id, CommandType.PATCH, null));

	oldPatch := command.(*PatchCommandData)
	newScript := oldPatch.getPatch().GetScript() + "\n" + patchRequest.GetScript()
	newVals := cloneMapStringObject(oldPatch.getPatch().GetValues())

	for k, v := range patchRequest.GetValues() {
		newVals[k] = v
	}

	newPatchRequest := NewPatchRequest()
	newPatchRequest.SetScript(newScript)
	newPatchRequest.SetValues(newVals)

	cmdData := NewPatchCommandData(id, nil, newPatchRequest, nil)
	s.Defer(cmdData)
	return true
}

func cloneMapStringObject(m map[string]Object) map[string]Object {
	res := map[string]Object{}
	for k, v := range m {
		res[k] = v
	}
	return res
}

// public <T, TIndex extends AbstractIndexCreationTask> IDocumentQuery<T> documentQuery(reflect.Type clazz, Class<TIndex> indexClazz) {

// TODO: needs clazz
func (s *DocumentSession) DocumentQueryInIndexOld(clazz reflect.Type, index *AbstractIndexCreationTask) *DocumentQuery {
	return s.DocumentQueryAllOld(clazz, index.GetIndexName(), "", index.IsMapReduce())
}

// TODO: needs clazz
func (s *DocumentSession) DocumentQueryOld(clazz reflect.Type) *DocumentQuery {
	return s.DocumentQueryAllOld(clazz, "", "", false)
}

// TODO: this needs clazz
func (s *DocumentSession) DocumentQueryAllOld(clazz reflect.Type, indexName string, collectionName string, isMapReduce bool) *DocumentQuery {
	indexName, collectionName = s.processQueryParameters(clazz, indexName, collectionName, s.GetConventions())

	return NewDocumentQueryOld(clazz, s.InMemoryDocumentSessionOperations, indexName, collectionName, isMapReduce)
}

func (s *DocumentSession) RawQuery(query string) *IRawDocumentQuery {
	return NewRawDocumentQuery(s.InMemoryDocumentSessionOperations, query)
}

// TODO: needs clazz
func (s *DocumentSession) QueryOld(clazz reflect.Type) *DocumentQuery {
	return s.DocumentQueryAllOld(clazz, "", "", false)
}

func (s *DocumentSession) QueryWithQueryOld(clazz reflect.Type, collectionOrIndexName *Query) *DocumentQuery {
	if stringIsNotEmpty(collectionOrIndexName.getCollection()) {
		return s.DocumentQueryAllOld(clazz, "", collectionOrIndexName.getCollection(), false)
	}

	return s.DocumentQueryAllOld(clazz, collectionOrIndexName.getIndexName(), "", false)
}

func (s *DocumentSession) QueryInIndexOld(clazz reflect.Type, index *AbstractIndexCreationTask) *DocumentQuery {
	return s.DocumentQueryInIndexOld(clazz, index)
}

// public <T> CloseableIterator<StreamResult<T>> stream(IDocumentQuery<T> query) {
// public <T> CloseableIterator<StreamResult<T>> stream(IDocumentQuery<T> query, Reference<StreamQueryStatistics> streamQueryStats) {
// public <T> CloseableIterator<StreamResult<T>> stream(IRawDocumentQuery<T> query) {
// public <T> CloseableIterator<StreamResult<T>> stream(IRawDocumentQuery<T> query, Reference<StreamQueryStatistics> streamQueryStats) {
// private <T> CloseableIterator<StreamResult<T>> yieldResults(AbstractDocumentQuery query, CloseableIterator<ObjectNode> enumerator) {
// public <T> void streamInto(IRawDocumentQuery<T> query, OutputStream output) {
// public <T> void streamInto(IDocumentQuery<T> query, OutputStream output) {

// private <T> StreamResult<T> createStreamResult(reflect.Type clazz, ObjectNode json, fieldsToFetchToken fieldsToFetch) throws IOException {
// public <T> CloseableIterator<StreamResult<T>> stream(reflect.Type clazz, string startsWith) {
// public <T> CloseableIterator<StreamResult<T>> stream(reflect.Type clazz, string startsWith, string matches) {
// public <T> CloseableIterator<StreamResult<T>> stream(reflect.Type clazz, string startsWith, string matches, int start) {
// public <T> CloseableIterator<StreamResult<T>> stream(reflect.Type clazz, string startsWith, string matches, int start, int pageSize) {
// public <T> CloseableIterator<StreamResult<T>> stream(reflect.Type clazz, string startsWith, string matches, int start, int pageSize, string startAfter) {
