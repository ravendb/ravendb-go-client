package ravendb

import (
	"encoding/json"
	"fmt"
	"io"
	"reflect"
	"strconv"
	"time"
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

	res.InMemoryDocumentSessionOperations.session = res

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
	if s.documentsByID.getValue(id) != nil {
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

func (s *DocumentSession) Refresh(entity interface{}) error {
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

func (s *DocumentSession) ExecuteAllPendingLazyOperations() (*ResponseTimeInformation, error) {
	var requests []*GetRequest
	var pendingTmp []ILazyOperation
	for _, op := range s.pendingLazyOperations {
		req := op.createRequest()
		if req == nil {
			continue
		}
		pendingTmp = append(pendingTmp, op)
		requests = append(requests, req)
	}
	s.pendingLazyOperations = pendingTmp

	if len(requests) == 0 {
		return &ResponseTimeInformation{}, nil
	}

	sw := time.Now()
	s.IncrementRequestCount()

	defer func() { s.pendingLazyOperations = nil }()

	responseTimeDuration := &ResponseTimeInformation{}
	for {
		shouldRetry, err := s.executeLazyOperationsSingleStep(responseTimeDuration, requests)
		if err != nil {
			return nil, err
		}
		if !shouldRetry {
			break
		}
		time.Sleep(time.Millisecond * 100)
	}
	responseTimeDuration.computeServerTotal()

	for _, pendingLazyOperation := range s.pendingLazyOperations {
		value := s.onEvaluateLazy[pendingLazyOperation]
		if value != nil {
			value(pendingLazyOperation.getResult())
		}
	}

	dur := time.Since(sw)

	responseTimeDuration.totalClientDuration = dur
	return responseTimeDuration, nil
}

func (s *DocumentSession) executeLazyOperationsSingleStep(responseTimeInformation *ResponseTimeInformation, requests []*GetRequest) (bool, error) {
	multiGetOperation := NewMultiGetOperation(s.InMemoryDocumentSessionOperations)
	multiGetCommand := multiGetOperation.createRequest(requests)

	err := s.GetRequestExecutor().ExecuteCommandWithSessionInfo(multiGetCommand, s.sessionInfo)
	if err != nil {
		return false, err
	}
	responses := multiGetCommand.Result
	for i, op := range s.pendingLazyOperations {
		response := responses[i]
		tempReqTime := response.headers[Constants_Headers_REQUEST_TIME]
		totalTime, _ := strconv.Atoi(tempReqTime)
		uri := requests[i].getUrlAndQuery()
		dur := time.Millisecond * time.Duration(totalTime)
		timeItem := ResponseTimeItem{
			url:      uri,
			duration: dur,
		}
		responseTimeInformation.durationBreakdown = append(responseTimeInformation.durationBreakdown, timeItem)
		if response.requestHasErrors() {
			return false, NewIllegalStateException("Got an error from server, status code: %d\n%s", response.statusCode, response.result)
		}
		err = op.handleResponse(response)
		if err != nil {
			return false, err
		}
		if op.isRequiresRetry() {
			return true, nil
		}
	}
	return false, nil
}

func (s *DocumentSession) Include(path string) *MultiLoaderWithInclude {
	return NewMultiLoaderWithInclude(s).Include(path)
}

func (s *DocumentSession) addLazyOperation(clazz reflect.Type, operation ILazyOperation, onEval func(interface{})) *Lazy {
	s.pendingLazyOperations = append(s.pendingLazyOperations, operation)

	fn := func() (interface{}, error) {
		_, err := s.ExecuteAllPendingLazyOperations()
		if err != nil {
			return nil, err
		}
		res := operation.getResult()
		return s.getOperationResult(clazz, res)
	}
	lazyValue := NewLazy(fn)
	if onEval != nil {
		if s.onEvaluateLazy == nil {
			s.onEvaluateLazy = map[ILazyOperation]func(interface{}){}
		}
		fn := func(theResult interface{}) {
			res, _ := s.getOperationResult(clazz, theResult)
			onEval(res)
		}
		s.onEvaluateLazy[operation] = fn
	}

	return lazyValue
}

func (s *DocumentSession) addLazyCountOperation(operation ILazyOperation) *Lazy {
	s.pendingLazyOperations = append(s.pendingLazyOperations, operation)

	fn := func() (interface{}, error) {
		_, err := s.ExecuteAllPendingLazyOperations()
		if err != nil {
			return nil, err
		}
		return operation.getQueryResult().TotalResults, nil
	}
	return NewLazy(fn)
}

// public <T> Lazy<Map<String, T>>
// TODO: change onEval to func(map[string]interface{})
func (s *DocumentSession) lazyLoadInternal(clazz reflect.Type, ids []string, includes []string, onEval func(interface{})) *Lazy {
	if s.checkIfIdAlreadyIncluded(ids, includes) {
		fn := func() (interface{}, error) {
			return s.LoadMultiOld(clazz, ids)
		}
		return NewLazy(fn)
	}

	loadOperation := NewLoadOperation(s.InMemoryDocumentSessionOperations)
	loadOperation = loadOperation.byIds(ids)
	loadOperation = loadOperation.withIncludes(includes)

	lazyOp := NewLazyLoadOperation(clazz, s.InMemoryDocumentSessionOperations, loadOperation)
	lazyOp = lazyOp.byIds(ids)
	lazyOp = lazyOp.withIncludes(includes)

	at := reflect.MapOf(stringType, clazz)

	return s.addLazyOperation(at, lazyOp, onEval)
}

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

func (s *DocumentSession) LoadMultiOld(clazz reflect.Type, ids []string) (map[string]interface{}, error) {
	loadOperation := NewLoadOperation(s.InMemoryDocumentSessionOperations)
	err := s.loadInternalWithOperation(ids, loadOperation, nil)
	if err != nil {
		return nil, err
	}
	return loadOperation.getDocumentsOld(clazz)
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

func (s *DocumentSession) LoadStartingWith(results interface{}, args *StartsWithArgs) error {
	loadStartingWithOperation := NewLoadStartingWithOperation(s.InMemoryDocumentSessionOperations)
	if args.PageSize == 0 {
		args.PageSize = 25
	}
	_, err := s.loadStartingWithInternal(args.StartsWith, loadStartingWithOperation, nil, args.Matches, args.Start, args.PageSize, args.Exclude, args.StartAfter)
	if err != nil {
		return err
	}
	return loadStartingWithOperation.getDocuments(results)
}

func (s *DocumentSession) LoadStartingWithIntoStream(output io.Writer, args *StartsWithArgs) error {
	loadStartingWithOperation := NewLoadStartingWithOperation(s.InMemoryDocumentSessionOperations)
	if args.PageSize == 0 {
		args.PageSize = 25
	}
	_, err := s.loadStartingWithInternal(args.StartsWith, loadStartingWithOperation, output, args.Matches, args.Start, args.PageSize, args.Exclude, args.StartAfter)
	if err != nil {
		return err
	}
	return nil
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

func cloneMapStringObject(m map[string]interface{}) map[string]interface{} {
	res := map[string]interface{}{}
	for k, v := range m {
		res[k] = v
	}
	return res
}

// public <T, TIndex extends AbstractIndexCreationTask> IDocumentQuery<T> documentQuery(reflect.Type clazz, Class<TIndex> indexClazz) {

// TODO: convert to use result interface{} instead of clazz reflect.Type
func (s *DocumentSession) DocumentQueryInIndexOld(clazz reflect.Type, index *AbstractIndexCreationTask) *DocumentQuery {
	return s.DocumentQueryAllOld(clazz, index.GetIndexName(), "", index.IsMapReduce())
}

// TODO: convert to use result interface{} instead of clazz reflect.Type
func (s *DocumentSession) DocumentQueryOld(clazz reflect.Type) *DocumentQuery {
	return s.DocumentQueryAllOld(clazz, "", "", false)
}

// TODO: convert to use result interface{} instead of clazz reflect.Type
func (s *DocumentSession) DocumentQueryAllOld(clazz reflect.Type, indexName string, collectionName string, isMapReduce bool) *DocumentQuery {
	indexName, collectionName = s.processQueryParameters(clazz, indexName, collectionName, s.GetConventions())
	panicIf(s.InMemoryDocumentSessionOperations.session != s, "must have session")
	return NewDocumentQueryOld(clazz, s.InMemoryDocumentSessionOperations, indexName, collectionName, isMapReduce)
}

func (s *DocumentSession) RawQuery(query string) *IRawDocumentQuery {
	return NewRawDocumentQuery(s.InMemoryDocumentSessionOperations, query)
}

// TODO: convert to use result interface{} instead of clazz reflect.Type
func (s *DocumentSession) QueryOld(clazz reflect.Type) *DocumentQuery {
	panicIf(s == nil, "s shouldn't be nil here")
	return s.DocumentQueryAllOld(clazz, "", "", false)
}

// TODO: convert to use result interface{} instead of clazz reflect.Type
func (s *DocumentSession) QueryWithQueryOld(clazz reflect.Type, collectionOrIndexName *Query) *DocumentQuery {
	if stringIsNotEmpty(collectionOrIndexName.getCollection()) {
		return s.DocumentQueryAllOld(clazz, "", collectionOrIndexName.getCollection(), false)
	}

	return s.DocumentQueryAllOld(clazz, collectionOrIndexName.getIndexName(), "", false)
}

// TODO: convert to use result interface{} instead of clazz reflect.Type
func (s *DocumentSession) QueryInIndexOld(clazz reflect.Type, index *AbstractIndexCreationTask) *DocumentQuery {
	return s.DocumentQueryInIndexOld(clazz, index)
}

func (s *DocumentSession) StreamQuery(query *IDocumentQuery, streamQueryStats *StreamQueryStatistics) (*StreamIterator, error) {
	streamOperation := NewStreamOperation(s.InMemoryDocumentSessionOperations, streamQueryStats)
	q := query.GetIndexQuery()
	command := streamOperation.createRequestForIndexQuery(q)
	err := s.GetRequestExecutor().ExecuteCommandWithSessionInfo(command, s.sessionInfo)
	if err != nil {
		return nil, err
	}
	result, err := streamOperation.setResult(command.Result)
	if err != nil {
		return nil, err
	}
	onNextItem := func(res map[string]interface{}) {
		query.InvokeAfterStreamExecuted(res)
	}
	return NewStreamIterator(s, result, query.fieldsToFetchToken, onNextItem), nil
}

func (s *DocumentSession) StreamRawQuery(query *IRawDocumentQuery, streamQueryStats *StreamQueryStatistics) (*StreamIterator, error) {
	streamOperation := NewStreamOperation(s.InMemoryDocumentSessionOperations, streamQueryStats)
	q := query.GetIndexQuery()
	command := streamOperation.createRequestForIndexQuery(q)
	err := s.GetRequestExecutor().ExecuteCommandWithSessionInfo(command, s.sessionInfo)
	if err != nil {
		return nil, err
	}
	result, err := streamOperation.setResult(command.Result)
	if err != nil {
		return nil, err
	}
	onNextItem := func(res map[string]interface{}) {
		query.InvokeAfterStreamExecuted(res)
	}
	return NewStreamIterator(s, result, query.fieldsToFetchToken, onNextItem), nil
}

func (s *DocumentSession) StreamRawQueryInto(query *IRawDocumentQuery, output io.Writer) error {
	streamOperation := NewStreamOperation(s.InMemoryDocumentSessionOperations, nil)
	q := query.GetIndexQuery()
	command := streamOperation.createRequestForIndexQuery(q)
	err := s.GetRequestExecutor().ExecuteCommandWithSessionInfo(command, s.sessionInfo)
	if err != nil {
		return err
	}
	stream := command.Result.Stream
	_, err = io.Copy(output, stream)
	return err
}

func (s *DocumentSession) StreamQueryInto(query *IDocumentQuery, output io.Writer) error {
	streamOperation := NewStreamOperation(s.InMemoryDocumentSessionOperations, nil)
	q := query.GetIndexQuery()
	command := streamOperation.createRequestForIndexQuery(q)
	err := s.GetRequestExecutor().ExecuteCommandWithSessionInfo(command, s.sessionInfo)
	if err != nil {
		return err
	}
	stream := command.Result.Stream
	_, err = io.Copy(output, stream)
	return err
}

func (s *DocumentSession) createStreamResult(v interface{}, document ObjectNode, fieldsToFetch *fieldsToFetchToken) (*StreamResult, error) {
	//fmt.Printf("createStreamResult: document: %#v\n", document)

	// we expect v to be **Foo. We deserialize into *Foo and assign it to v
	rt := reflect.TypeOf(v)
	if rt.Kind() != reflect.Ptr {
		return nil, fmt.Errorf("v should be a pointer to a pointer to  struct, is %T. rt: %s", v, rt)
	}
	rt = rt.Elem()
	if rt.Kind() != reflect.Ptr {
		return nil, fmt.Errorf("v should be a pointer to a pointer to  struct, is %T. rt: %s", v, rt)
	}

	metadataV, ok := document[Constants_Documents_Metadata_KEY]
	if !ok {
		fmt.Printf("document: %#v\n", document)
		// TODO: maybe convert to errors
		panicIf(!ok, "Document must have a metadata")
	}
	metadata, ok := metadataV.(ObjectNode)
	panicIf(!ok, "Document metadata is not a valid type %T", metadataV)

	changeVector := jsonGetAsTextPointer(metadata, Constants_Documents_Metadata_CHANGE_VECTOR)
	// TODO: return an error?
	panicIf(changeVector == nil, "Document must have a Change Vector")

	// MapReduce indexes return reduce results that don't have @id property
	id, _ := jsonGetAsString(metadata, Constants_Documents_Metadata_ID)

	entity, err := QueryOperation_deserialize(rt, id, document, metadata, fieldsToFetch, true, s.InMemoryDocumentSessionOperations)
	if err != nil {
		return nil, err
	}
	setInterfaceToValue(v, entity)

	meta := NewMetadataAsDictionaryWithSource(metadata)
	streamResult := &StreamResult{
		ID:           id,
		changeVector: changeVector,
		document:     entity,
		metadata:     meta,
	}
	return streamResult, nil
}

func (s *DocumentSession) Stream(args *StartsWithArgs) (*StreamIterator, error) {
	streamOperation := NewStreamOperation(s.InMemoryDocumentSessionOperations, nil)

	command := streamOperation.createRequest(args.StartsWith, args.Matches, args.Start, args.PageSize, "", args.StartAfter)
	err := s.GetRequestExecutor().ExecuteCommandWithSessionInfo(command, s.sessionInfo)
	if err != nil {
		return nil, err
	}

	cmdResult := command.Result
	result, err := streamOperation.setResult(cmdResult)
	if err != nil {
		return nil, err
	}
	return NewStreamIterator(s, result, nil, nil), nil
}

type StreamIterator struct {
	_session            *DocumentSession
	_innerIterator      *YieldStreamResults
	_fieldsToFetchToken *fieldsToFetchToken
	_onNextItem         func(ObjectNode)
}

func NewStreamIterator(session *DocumentSession, innerIterator *YieldStreamResults, fieldsToFetchToken *fieldsToFetchToken, onNextItem func(ObjectNode)) *StreamIterator {
	return &StreamIterator{
		_session:            session,
		_innerIterator:      innerIterator,
		_fieldsToFetchToken: fieldsToFetchToken,
		_onNextItem:         onNextItem,
	}
}

func (i *StreamIterator) Next(v interface{}) (*StreamResult, error) {
	nextValue, err := i._innerIterator.NextJSONObject()
	if err != nil {
		return nil, err
	}
	if i._onNextItem != nil {
		i._onNextItem(nextValue)
	}
	return i._session.createStreamResult(v, nextValue, i._fieldsToFetchToken)
}

func (i *StreamIterator) Close() {
	i._innerIterator.Close()
}
