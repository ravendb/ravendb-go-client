package ravendb

import (
	"encoding/json"
	"fmt"
	"io"
	"reflect"
	"strconv"
	"time"
)

// Note: Java's IDocumentSessionImpl is DocumentSession

// TODO: decide if we want to return ErrNotFound or nil if the value is not found
// Java returns nil (which, I guess, is default value for reference (i.e. all) types)
// var ErrNotFound = errors.New("Not found")
// var ErrNotFound = error(nil)

// DocumentSession is a Unit of Work for accessing RavenDB server
type DocumentSession struct {
	*InMemoryDocumentSessionOperations

	attachments *AttachmentsSessionOperations
	revisions   *RevisionsSessionOperations
	valsCount   int
	customCount int
}

func (s *DocumentSession) Advanced() *AdvancedSessionOperations {
	return &AdvancedSessionOperations{
		s: s,
	}
}

func (s *DocumentSession) Lazily() *LazySessionOperations {
	return newLazySessionOperations(s)
}

type EagerSessionOperations struct {
	s *DocumentSession
}

func (s *EagerSessionOperations) ExecuteAllPendingLazyOperations() (*ResponseTimeInformation, error) {
	return s.s.executeAllPendingLazyOperations()
}

func (s *DocumentSession) Eagerly() *EagerSessionOperations {
	return &EagerSessionOperations{
		s: s,
	}
}

func (s *DocumentSession) Attachments() *AttachmentsSessionOperations {
	if s.attachments == nil {
		s.attachments = NewDocumentSessionAttachments(s.InMemoryDocumentSessionOperations)
	}
	return s.attachments
}

func (s *DocumentSession) Revisions() *RevisionsSessionOperations {
	return s.revisions
}

// NewDocumentSession creates a new DocumentSession
func NewDocumentSession(dbName string, documentStore *DocumentStore, id string, re *RequestExecutor) *DocumentSession {
	res := &DocumentSession{
		InMemoryDocumentSessionOperations: newInMemoryDocumentSessionOperations(dbName, documentStore, re, id),
	}

	res.InMemoryDocumentSessionOperations.session = res

	// TODO: this must be delayed until Attachments() or else attachments_session_test.go fail. Why?
	//res.attachments = NewDocumentSessionAttachments(res.InMemoryDocumentSessionOperations)
	res.revisions = newDocumentSessionRevisions(res.InMemoryDocumentSessionOperations)

	return res
}

// SaveChanges saves changes queued in memory to the database
func (s *DocumentSession) SaveChanges() error {
	saveChangeOperation := NewBatchOperation(s.InMemoryDocumentSessionOperations)

	command, err := saveChangeOperation.createRequest()
	if err != nil {
		return err
	}
	if command == nil {
		return nil
	}
	defer func() {
		_ = command.Close()
	}()
	err = s.requestExecutor.ExecuteCommand(command, s.sessionInfo)
	if err != nil {
		return err
	}
	result := command.Result
	return saveChangeOperation.setResult(result.Results)
}

// Exists returns true if an entity with a given id exists in the database
func (s *DocumentSession) Exists(id string) (bool, error) {
	if id == "" {
		return false, newIllegalArgumentError("id cannot be empty string")
	}

	if stringArrayContainsNoCase(s.knownMissingIds, id) {
		return false, nil
	}

	if s.documentsByID.getValue(id) != nil {
		return true, nil
	}
	command := NewHeadDocumentCommand(id, nil)

	if err := s.requestExecutor.ExecuteCommand(command, s.sessionInfo); err != nil {
		return false, err
	}

	ok := command.Exists()
	return ok, nil
}

// Refresh reloads information about a given entity in the session from the database
func (s *DocumentSession) Refresh(entity interface{}) error {
	if err := checkValidEntityIn(entity, "entity"); err != nil {
		return err
	}
	documentInfo := getDocumentInfoByEntity(s.documents, entity)
	if documentInfo == nil {
		return newIllegalStateError("Cannot refresh a transient instance")
	}
	if err := s.incrementRequestCount(); err != nil {
		return err
	}

	command, err := NewGetDocumentsCommand([]string{documentInfo.id}, nil, false)
	if err != nil {
		return err
	}
	if err = s.requestExecutor.ExecuteCommand(command, s.sessionInfo); err != nil {
		return err
	}
	return s.refreshInternal(entity, command, documentInfo)
}

// TODO:    protected string generateID(Object entity) {


func (s *DocumentSession) executeAllPendingLazyOperations() (*ResponseTimeInformation, error) {
	var requests []*getRequest
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
	if err := s.incrementRequestCount(); err != nil {
		return nil, err
	}

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
		onLazyEval := s.onEvaluateLazy[pendingLazyOperation]
		if onLazyEval != nil {
			err := pendingLazyOperation.getResult(onLazyEval.result)
			if err != nil {
				return nil, err
			}
			onLazyEval.fn()
		}
	}

	dur := time.Since(sw)

	responseTimeDuration.totalClientDuration = dur
	return responseTimeDuration, nil
}

func (s *DocumentSession) executeLazyOperationsSingleStep(responseTimeInformation *ResponseTimeInformation, requests []*getRequest) (bool, error) {
	multiGetOperation := &MultiGetOperation{
		session: s.InMemoryDocumentSessionOperations,
	}
	multiGetCommand := multiGetOperation.createRequest(requests)

	err := s.GetRequestExecutor().ExecuteCommand(multiGetCommand, s.sessionInfo)
	if err != nil {
		return false, err
	}
	responses := multiGetCommand.Result
	for i, op := range s.pendingLazyOperations {
		response := responses[i]
		tempReqTime := response.Headers[headersRequestTime]
		totalTime, _ := strconv.Atoi(tempReqTime)
		uri := requests[i].getUrlAndQuery()
		dur := time.Millisecond * time.Duration(totalTime)
		timeItem := ResponseTimeItem{
			URL:      uri,
			Duration: dur,
		}
		responseTimeInformation.durationBreakdown = append(responseTimeInformation.durationBreakdown, timeItem)
		if response.requestHasErrors() {
			return false, newIllegalStateError("Got an error from server, status code: %d\n%s", response.StatusCode, response.Result)
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

func (s *DocumentSession) addLazyOperation(operation ILazyOperation, onEval func(), onEvalResult interface{}) *Lazy {
	s.pendingLazyOperations = append(s.pendingLazyOperations, operation)

	fn := func(result interface{}) error {
		_, err := s.executeAllPendingLazyOperations()
		if err != nil {
			return err
		}
		err = operation.getResult(result)
		return err
	}
	lazyValue := newLazy(fn)

	if onEval != nil {
		if s.onEvaluateLazy == nil {
			s.onEvaluateLazy = map[ILazyOperation]*onLazyEval{}
		}
		// TODO: make sure this is tested
		s.onEvaluateLazy[operation] = &onLazyEval{
			fn: onEval,
			result: onEvalResult,
		}
	}

	return lazyValue
}

func (s *DocumentSession) addLazyCountOperation(operation ILazyOperation) *Lazy {
	s.pendingLazyOperations = append(s.pendingLazyOperations, operation)

	fn := func(result interface{}) error {
		_, err := s.executeAllPendingLazyOperations()
		if err != nil {
			return err
		}
		count := result.(*int)
		*count = operation.getQueryResult().TotalResults
		return nil
	}
	return newLazy(fn)
}

func (s *DocumentSession) lazyLoadInternal(ids []string, includes []string, onEval func(), onEvalResult interface{}) *Lazy {
	if s.checkIfIdAlreadyIncluded(ids, includes) {
		fn := func(results interface{}) error {
			// res should be the same as results
			err := s.LoadMulti(results, ids)
			return err
		}
		return newLazy(fn)
	}

	loadOperation := NewLoadOperation(s.InMemoryDocumentSessionOperations)
	loadOperation = loadOperation.byIds(ids)
	loadOperation = loadOperation.withIncludes(includes)

	lazyOp := newLazyLoadOperation(s.InMemoryDocumentSessionOperations, loadOperation)
	lazyOp = lazyOp.byIds(ids)
	lazyOp = lazyOp.withIncludes(includes)

	return s.addLazyOperation(lazyOp, onEval, onEvalResult)
}

func checkIsPtrPtrStruct(v interface{}, argName string) error {
	if v == nil {
		return newIllegalArgumentError("%s can't be nil", argName)
	}
	tp := reflect.TypeOf(v)

	if tp.Kind() == reflect.Struct {
		// possibly a common mistake, so try to provide a helpful error message
		typeGot := fmt.Sprintf("%T", v)
		typeExpect := "**" + typeGot
		return newIllegalArgumentError("%s can't be of type %s, try passing %s", argName, typeGot, typeExpect)
	}

	if tp.Kind() != reflect.Ptr {
		return newIllegalArgumentError("%s can't be of type %T", argName, v)
	}

	if tp.Elem().Kind() == reflect.Struct {
		// possibly a common mistake, so try to provide a helpful error message
		typeGot := fmt.Sprintf("%T", v)
		typeExpect := "*" + typeGot
		return newIllegalArgumentError("%s can't be of type %s, try passing %s", argName, typeGot, typeExpect)
	}

	if tp.Elem().Kind() != reflect.Ptr {
		return newIllegalArgumentError("%s can't be of type %T", argName, v)
	}

	// we only allow pointer to struct
	if tp.Elem().Elem().Kind() == reflect.Struct {
		return nil
	}
	return newIllegalArgumentError("%s can't be of type %T", argName, v)
}

// check if v is a valid argument to Load().
// it must be *<type> where <type> is *struct or map[string]interface{}
func checkValidLoadArg(v interface{}, argName string) error {
	if v == nil {
		return newIllegalArgumentError("%s can't be nil", argName)
	}

	if _, ok := v.(**map[string]interface{}); ok {
		return nil
	}

	// TODO: better error message for *map[string]interface{} and map[string]interface{}
	/* TODO: allow map as an argument
	if _, ok := v.(map[string]interface{}); ok {
		if reflect.ValueOf(v).IsNil() {
			return newIllegalArgumentError("%s can't be a nil map", argName)
		}
		return nil
	}
	*/
	return checkIsPtrPtrStruct(v, argName)
}

func (s *DocumentSession) Load(result interface{}, id string) error {
	if id == "" {
		return newIllegalArgumentError("id cannot be empty string")
	}
	if err := checkValidLoadArg(result, "result"); err != nil {
		return err
	}
	loadOperation := NewLoadOperation(s.InMemoryDocumentSessionOperations)

	loadOperation.byID(id)

	command, err := loadOperation.createRequest()
	if err != nil {
		return err
	}

	if command != nil {
		err := s.requestExecutor.ExecuteCommand(command, s.sessionInfo)
		if err != nil {
			return err
		}
		result := command.Result
		loadOperation.setResult(result)
	}

	return loadOperation.getDocument(result)
}

// check if v is a valid argument to LoadMulti().
// it must be map[string]*<type> where <type> is struct
func checkValidLoadMultiArg(v interface{}, argName string) error {
	if v == nil {
		return newIllegalArgumentError("%s can't be nil", argName)
	}
	tp := reflect.TypeOf(v)
	if tp.Kind() != reflect.Map {
		typeGot := fmt.Sprintf("%T", v)
		return newIllegalArgumentError("%s can't be of type %s, must be map[string]<type>", argName, typeGot)
	}
	if tp.Key().Kind() != reflect.String {
		typeGot := fmt.Sprintf("%T", v)
		return newIllegalArgumentError("%s can't be of type %s, must be map[string]<type>", argName, typeGot)
	}
	// type of the map element, must be *struct
	// TODO: also accept map[string]interface{} as type of map element
	tp = tp.Elem()
	if tp.Kind() != reflect.Ptr || tp.Elem().Kind() != reflect.Struct {
		typeGot := fmt.Sprintf("%T", v)
		return newIllegalArgumentError("%s can't be of type %s, must be map[string]<type>", argName, typeGot)
	}

	if reflect.ValueOf(v).IsNil() {
		return newIllegalArgumentError("%s can't be a nil map", argName)
	}
	return nil
}

// LoadMulti loads multiple values with given ids into results, which should
// be a map from string (id) to pointer to struct
func (s *DocumentSession) LoadMulti(results interface{}, ids []string) error {
	if len(ids) == 0 {
		return newIllegalArgumentError("ids cannot be empty array")
	}
	if err := checkValidLoadMultiArg(results, "results"); err != nil {
		return err
	}
	loadOperation := NewLoadOperation(s.InMemoryDocumentSessionOperations)
	err := s.loadInternalWithOperation(ids, loadOperation, nil)
	if err != nil {
		return err
	}
	return loadOperation.getDocuments(results)
}

func (s *DocumentSession) loadInternalWithOperation(ids []string, operation *LoadOperation, stream io.Writer) error {
	operation.byIds(ids)

	command, err := operation.createRequest()
	if err != nil {
		return err
	}
	if command != nil {
		err := s.requestExecutor.ExecuteCommand(command, s.sessionInfo)
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
	if len(ids) == 0 {
		return newIllegalArgumentError("ids cannot be empty array")
	}

	loadOperation := NewLoadOperation(s.InMemoryDocumentSessionOperations)
	loadOperation.byIds(ids)
	loadOperation.withIncludes(includes)

	command, err := loadOperation.createRequest()
	if err != nil {
		return err
	}
	if command != nil {
		err := s.requestExecutor.ExecuteCommand(command, s.sessionInfo)
		if err != nil {
			return err
		}
		loadOperation.setResult(command.Result)
	}

	return loadOperation.getDocuments(results)
}

func (s *DocumentSession) LoadStartingWith(results interface{}, args *StartsWithArgs) error {
	// TODO: early validation of results
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
	if output == nil {
		return newIllegalArgumentError("Output cannot be null")
	}
	if args.StartsWith == "" {
		return newIllegalArgumentError("args.StartsWith cannot be empty string")
	}
	loadStartingWithOperation := NewLoadStartingWithOperation(s.InMemoryDocumentSessionOperations)
	if args.PageSize == 0 {
		args.PageSize = 25
	}
	_, err := s.loadStartingWithInternal(args.StartsWith, loadStartingWithOperation, output, args.Matches, args.Start, args.PageSize, args.Exclude, args.StartAfter)
	return err
}

func (s *DocumentSession) loadStartingWithInternal(idPrefix string, operation *LoadStartingWithOperation, stream io.Writer,
	matches string, start int, pageSize int, exclude string, startAfter string) (*GetDocumentsCommand, error) {

	operation.withStartWithFull(idPrefix, matches, start, pageSize, exclude, startAfter)

	command, err := operation.createRequest()
	if err != nil {
		return nil, err
	}
	if command != nil {
		err := s.requestExecutor.ExecuteCommand(command, s.sessionInfo)
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

// LoadIntoStream loads entities identified by ids and writes them (in JSON form)
// to output
func (s *DocumentSession) LoadIntoStream(ids []string, output io.Writer) error {
	if len(ids) == 0 {
		return newIllegalArgumentError("Ids cannot be empty")
	}

	op := NewLoadOperation(s.InMemoryDocumentSessionOperations)
	return s.loadInternalWithOperation(ids, op, output)
}

// Increment increments member identified by path in an entity by a given
// valueToAdd (can be negative, to subtract)
func (s *DocumentSession) Increment(entity interface{}, path string, valueToAdd interface{}) error {
	if path == "" {
		return newIllegalArgumentError("path can't be empty string")
	}
	if valueToAdd == nil {
		return newIllegalArgumentError("valueToAdd can't be nil")
	}
	metadata, err := s.GetMetadataFor(entity)
	if err != nil {
		return err
	}
	id, _ := metadata.Get(MetadataID)
	return s.IncrementByID(id.(string), path, valueToAdd)
}

// IncrementByID increments member identified by path in an entity identified by id by a given
// valueToAdd (can be negative, to subtract)
func (s *DocumentSession) IncrementByID(id string, path string, valueToAdd interface{}) error {
	if id == "" {
		return newIllegalArgumentError("id can't be empty string")
	}
	if path == "" {
		return newIllegalArgumentError("path can't be empty string")
	}
	if valueToAdd == nil {
		return newIllegalArgumentError("valueToAdd can't be nil")
	}
	// TODO: check that valueToAdd is numeric?
	patchRequest := &PatchRequest{}

	valsCountStr := strconv.Itoa(s.valsCount)
	variable := "this." + path
	value := "args.val_" + valsCountStr

	patchRequest.Script = variable + " = " + variable + " ? " + variable + " + " + value + " : " + value + ";"

	m := map[string]interface{}{
		"val_" + valsCountStr: valueToAdd,
	}
	patchRequest.Values = m

	s.valsCount++

	if !s.tryMergePatches(id, patchRequest) {
		cmdData := NewPatchCommandData(id, nil, patchRequest, nil)
		s.Defer(cmdData)
	}
	return nil
}

// Patch updates entity by changing part identified by path to a given value
func (s *DocumentSession) Patch(entity interface{}, path string, value interface{}) error {
	if path == "" {
		return newIllegalArgumentError("path can't be empty string")
	}
	if value == nil {
		return newIllegalArgumentError("value can't be nil")
	}
	metadata, err := s.GetMetadataFor(entity)
	if err != nil {
		return err
	}
	id, _ := metadata.Get(MetadataID)
	return s.PatchByID(id.(string), path, value)
}

// PatchByID updates entity identified by id by changing part identified by path to a given value
func (s *DocumentSession) PatchByID(id string, path string, value interface{}) error {
	if id == "" {
		return newIllegalArgumentError("id can't be empty string")
	}
	if path == "" {
		return newIllegalArgumentError("path can't be empty string")
	}
	if value == nil {
		return newIllegalArgumentError("value can't be nil")
	}
	patchRequest := &PatchRequest{}
	valsCountStr := strconv.Itoa(s.valsCount)
	patchRequest.Script = "this." + path + " = args.val_" + valsCountStr + ";"
	m := map[string]interface{}{
		"val_" + valsCountStr: value,
	}
	patchRequest.Values = m

	s.valsCount++

	if !s.tryMergePatches(id, patchRequest) {
		cmdData := NewPatchCommandData(id, nil, patchRequest, nil)
		s.Defer(cmdData)
	}
	return nil
}

// PatchArray updates an array value of document under a given path. Modify
// the array inside arrayAdder function
func (s *DocumentSession) PatchArray(entity interface{}, pathToArray string, arrayAdder func(*JavaScriptArray)) error {
	if pathToArray == "" {
		return newIllegalArgumentError("pathToArray can't be empty string")
	}
	if arrayAdder == nil {
		return newIllegalArgumentError("arrayAdder can't be nil")
	}
	metadata, err := s.GetMetadataFor(entity)
	if err != nil {
		return err
	}
	id, ok := metadata.Get(MetadataID)
	if !ok {
		return newIllegalStateError("entity doesn't have an ID")
	}
	return s.PatchArrayByID(id.(string), pathToArray, arrayAdder)
}

func (s *DocumentSession) PatchArrayByID(id string, pathToArray string, arrayAdder func(*JavaScriptArray)) error {
	if id == "" {
		return newIllegalArgumentError("id can't be empty string")
	}
	if pathToArray == "" {
		return newIllegalArgumentError("pathToArray can't be empty string")
	}
	if arrayAdder == nil {
		return newIllegalArgumentError("arrayAdder can't be nil")
	}
	s.customCount++
	scriptArray := NewJavaScriptArray(s.customCount, pathToArray)

	arrayAdder(scriptArray)

	patchRequest := &PatchRequest{}
	patchRequest.Script = scriptArray.GetScript()
	patchRequest.Values = scriptArray.Parameters

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
	idType := newIDTypeAndName(id, CommandPatch, "")
	command := s.deferredCommandsMap[idType]
	if command == nil {
		return false
	}

	s.deferredCommands = removeDeferredCommand(s.deferredCommands, command)

	// We'll overwrite the deferredCommandsMap when calling Defer
	// No need to call deferredCommandsMap.remove((id, CommandType.PATCH, null));

	oldPatch := command.(*PatchCommandData)
	newScript := oldPatch.patch.Script + "\n" + patchRequest.Script
	newVals := cloneMapStringObject(oldPatch.patch.Values)

	for k, v := range patchRequest.Values {
		newVals[k] = v
	}

	newPatchRequest := &PatchRequest{}
	newPatchRequest.Script = newScript
	newPatchRequest.Values = newVals

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

// DocumentQuery* seem redundant with Query* functions
// I assume in Java it was done to avoid conflicts in IAdvancedSessionOperations
// and other interfaces
/*
func (s *DocumentSession) DocumentQueryIndex(indexName string) *DocumentQuery {
	opts := &DocumentQueryOptions{
		IndexName: indexName,
		session:   s.InMemoryDocumentSessionOperations,
	}
	q, _ := newDocumentQuery(opts)
	return q
}

func (s *DocumentSession) DocumentQueryCollection(collectionName string) *DocumentQuery {
	opts := &DocumentQueryOptions{
		CollectionName: collectionName,
		session:        s.InMemoryDocumentSessionOperations,
	}
	q, _ := newDocumentQuery(opts)
	return q
}

func (s *DocumentSession) DocumentQueryCollectionForType(clazz reflect.Type) (*DocumentQuery, error) {
	panicIf(s.InMemoryDocumentSessionOperations.session != s, "must have session")
	indexName, collectionName, err := s.processQueryParameters(clazz, "", "", s.GetConventions())
	panicIfErr(err)
	opts := &DocumentQueryOptions{
		IndexName:      indexName,
		CollectionName: collectionName,
		Type:           clazz,
		session:        s.InMemoryDocumentSessionOperations,
	}
	return newDocumentQuery(opts)
}

// DocumentQuery starts a new DocumentQuery
func (s *DocumentSession) DocumentQuery() *DocumentQuery {
	return s.DocumentQueryAll("", "", false)
}

func (s *DocumentSession) DocumentQueryAll(indexName string, collectionName string, isMapReduce bool) *DocumentQuery {
	panicIf(s.InMemoryDocumentSessionOperations.session != s, "must have session")
	opts := &DocumentQueryOptions{
		IndexName:      indexName,
		CollectionName: collectionName,
		IsMapReduce:    isMapReduce,
		session:        s.InMemoryDocumentSessionOperations,
	}
	q, _ := newDocumentQuery(opts)
	return q
}

func (s *DocumentSession) DocumentQueryAllOld(clazz reflect.Type, indexName string, collectionName string, isMapReduce bool) *DocumentQuery {
	panicIf(s.InMemoryDocumentSessionOperations.session != s, "must have session")
	var err error
	indexName, collectionName, err = s.processQueryParameters(clazz, indexName, collectionName, s.GetConventions())
	panicIfErr(err)
	opts := &DocumentQueryOptions{
		Type:           clazz,
		session:        s.InMemoryDocumentSessionOperations,
		IndexName:      indexName,
		CollectionName: collectionName,
		IsMapReduce:    isMapReduce,
	}
	q, _ := newDocumentQuery(opts)
	return q
}
*/

// RawQuery returns new DocumentQuery representing a raw query
func (s *DocumentSession) RawQuery(rawQuery string) *RawDocumentQuery {
	opts := &DocumentQueryOptions{
		session:  s.InMemoryDocumentSessionOperations,
		rawQuery: rawQuery,
	}
	aq := newAbstractDocumentQuery(opts)
	return &RawDocumentQuery{
		abstractDocumentQuery: aq,
	}
}

// Query return a new DocumentQuery
func (s *DocumentSession) Query(opts *DocumentQueryOptions) *DocumentQuery {
	if opts == nil {
		opts = &DocumentQueryOptions{}
	}
	opts.session = s.InMemoryDocumentSessionOperations
	opts.conventions = s.GetConventions()
	return newDocumentQuery(opts)
}

// QueryCollection creates a new query over documents of a given collection
func (s *DocumentSession) QueryCollection(collectionName string) *DocumentQuery {
	opts := &DocumentQueryOptions{
		CollectionName: collectionName,
		session:        s.InMemoryDocumentSessionOperations,
		conventions:    s.GetConventions(),
	}
	res := newDocumentQuery(opts)
	if res.err != nil {
		return res
	}

	if collectionName == "" {
		res.err = newIllegalArgumentError("collectionName cannot be empty")
		return res
	}
	res.err = throwIfInvalidCollectionName(collectionName)
	return res
}

// QueryCollectionForType creates a new query over documents of a given type
func (s *DocumentSession) QueryCollectionForType(typ reflect.Type) *DocumentQuery {
	opts := &DocumentQueryOptions{
		Type:        typ,
		session:     s.InMemoryDocumentSessionOperations,
		conventions: s.GetConventions(),
	}
	res := newDocumentQuery(opts)
	if res.err == nil {
		if typ == nil {
			res.err = newIllegalArgumentError("typ cannot be nil")
		}
	}
	return res
}

// QueryIndex creates a new query in a index with a given name
func (s *DocumentSession) QueryIndex(indexName string) *DocumentQuery {
	opts := &DocumentQueryOptions{
		IndexName:   indexName,
		session:     s.InMemoryDocumentSessionOperations,
		conventions: s.GetConventions(),
	}
	res := newDocumentQuery(opts)
	if res.err != nil {
		return res
	}
	if indexName == "" {
		res.err = newIllegalArgumentError("indexName cannot be empty")
	}
	return res
}

// StreamQuery starts a streaming query and returns iterator for results.
// If streamQueryStats is provided, it'll be filled with information about query statistics.
func (s *DocumentSession) StreamQuery(query *DocumentQuery, streamQueryStats *StreamQueryStatistics) (*StreamIterator, error) {
	streamOperation := NewStreamOperation(s.InMemoryDocumentSessionOperations, streamQueryStats)
	q, err := query.GetIndexQuery()
	if err != nil {
		return nil, err
	}
	command, err := streamOperation.createRequestForIndexQuery(q)
	if err != nil {
		return nil, err
	}
	err = s.GetRequestExecutor().ExecuteCommand(command, s.sessionInfo)
	if err != nil {
		return nil, err
	}
	result, err := streamOperation.setResult(command.Result)
	if err != nil {
		return nil, err
	}
	onNextItem := func(res map[string]interface{}) {
		query.invokeAfterStreamExecuted(res)
	}
	return newStreamIterator(s, result, query.fieldsToFetchToken, onNextItem), nil
}

// StreamRawQuery starts a raw streaming query and returns iterator for results.
// If streamQueryStats is provided, it'll be filled with information about query statistics.
func (s *DocumentSession) StreamRawQuery(query *RawDocumentQuery, streamQueryStats *StreamQueryStatistics) (*StreamIterator, error) {
	streamOperation := NewStreamOperation(s.InMemoryDocumentSessionOperations, streamQueryStats)
	q, err := query.GetIndexQuery()
	if err != nil {
		return nil, err
	}
	command, err := streamOperation.createRequestForIndexQuery(q)
	if err != nil {
		return nil, err
	}
	err = s.GetRequestExecutor().ExecuteCommand(command, s.sessionInfo)
	if err != nil {
		return nil, err
	}
	result, err := streamOperation.setResult(command.Result)
	if err != nil {
		return nil, err
	}
	onNextItem := func(res map[string]interface{}) {
		query.invokeAfterStreamExecuted(res)
	}
	return newStreamIterator(s, result, query.fieldsToFetchToken, onNextItem), nil
}

// StreamRawQueryInto starts a raw streaming query that will write the results
// (in JSON format) to output
func (s *DocumentSession) StreamRawQueryInto(query *RawDocumentQuery, output io.Writer) error {
	streamOperation := NewStreamOperation(s.InMemoryDocumentSessionOperations, nil)
	q, err := query.GetIndexQuery()
	if err != nil {
		return err
	}
	command, err := streamOperation.createRequestForIndexQuery(q)
	if err != nil {
		return err
	}
	err = s.GetRequestExecutor().ExecuteCommand(command, s.sessionInfo)
	if err != nil {
		return err
	}
	stream := command.Result.Stream
	_, err = io.Copy(output, stream)
	return err
}

// StreamQueryInto starts a streaming query that will write the results
// (in JSON format) to output
func (s *DocumentSession) StreamQueryInto(query *DocumentQuery, output io.Writer) error {
	streamOperation := NewStreamOperation(s.InMemoryDocumentSessionOperations, nil)
	q, err := query.GetIndexQuery()
	if err != nil {
		return err
	}

	command, err := streamOperation.createRequestForIndexQuery(q)
	if err != nil {
		return err
	}
	err = s.GetRequestExecutor().ExecuteCommand(command, s.sessionInfo)
	if err != nil {
		return err
	}
	stream := command.Result.Stream
	_, err = io.Copy(output, stream)
	return err
}

func (s *DocumentSession) createStreamResult(v interface{}, document map[string]interface{}, fieldsToFetch *fieldsToFetchToken) (*StreamResult, error) {
	// we expect v to be **Foo. We deserialize into *Foo and assign it to v
	rt := reflect.TypeOf(v)
	if rt.Kind() != reflect.Ptr {
		return nil, newIllegalArgumentError("v should be a pointer to a pointer to  struct, is %T. rt: %s", v, rt)
	}
	rt = rt.Elem()
	if rt.Kind() != reflect.Ptr {
		return nil, newIllegalArgumentError("v should be a pointer to a pointer to  struct, is %T. rt: %s", v, rt)
	}

	metadataV, ok := document[MetadataKey]
	if !ok {
		return nil, newIllegalStateError("Document must have a metadata")
	}
	metadata, ok := metadataV.(map[string]interface{})
	if !ok {
		return nil, newIllegalStateError("Document metadata should be map[string]interface{} but is %T", metadataV)
	}

	changeVector := jsonGetAsTextPointer(metadata, MetadataChangeVector)
	if changeVector == nil {
		return nil, newIllegalStateError("Document must have a Change Vector")
	}

	// MapReduce indexes return reduce results that don't have @id property
	id, _ := jsonGetAsString(metadata, MetadataID)

	err := queryOperationDeserialize(v, id, document, metadata, fieldsToFetch, true, s.InMemoryDocumentSessionOperations)
	if err != nil {
		return nil, err
	}
	meta := NewMetadataAsDictionaryWithSource(metadata)
	entity := reflect.ValueOf(v).Elem().Interface()
	streamResult := &StreamResult{
		ID:           id,
		ChangeVector: changeVector,
		Document:     entity,
		Metadata:     meta,
	}
	return streamResult, nil
}

// Stream starts an iteration and returns StreamIterator
func (s *DocumentSession) Stream(args *StartsWithArgs) (*StreamIterator, error) {
	streamOperation := NewStreamOperation(s.InMemoryDocumentSessionOperations, nil)

	command := streamOperation.createRequest(args.StartsWith, args.Matches, args.Start, args.PageSize, "", args.StartAfter)
	err := s.GetRequestExecutor().ExecuteCommand(command, s.sessionInfo)
	if err != nil {
		return nil, err
	}

	cmdResult := command.Result
	result, err := streamOperation.setResult(cmdResult)
	if err != nil {
		return nil, err
	}
	return newStreamIterator(s, result, nil, nil), nil
}

// StreamIterator represents iterator of stream query
type StreamIterator struct {
	session            *DocumentSession
	innerIterator      *yieldStreamResults
	fieldsToFetchToken *fieldsToFetchToken
	onNextItem         func(map[string]interface{})
}

func newStreamIterator(session *DocumentSession, innerIterator *yieldStreamResults, fieldsToFetchToken *fieldsToFetchToken, onNextItem func(map[string]interface{})) *StreamIterator {
	return &StreamIterator{
		session:            session,
		innerIterator:      innerIterator,
		fieldsToFetchToken: fieldsToFetchToken,
		onNextItem:         onNextItem,
	}
}

// Next returns next result in a streaming query.
func (i *StreamIterator) Next(v interface{}) (*StreamResult, error) {
	nextValue, err := i.innerIterator.nextJSONObject()
	if err != nil {
		return nil, err
	}
	if i.onNextItem != nil {
		i.onNextItem(nextValue)
	}
	return i.session.createStreamResult(v, nextValue, i.fieldsToFetchToken)
}

// Close closes an iterator
func (i *StreamIterator) Close() error {
	return i.innerIterator.close()
}
