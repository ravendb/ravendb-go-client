package ravendb

import (
	"fmt"
	"reflect"
	"sync/atomic"
	"time"
)

var (
	clientSessionIDCounter int32 = 1
)

func newClientSessionID() int {
	newID := atomic.AddInt32(&clientSessionIDCounter, 1)
	return int(newID)
}

type onLazyEval struct {
	fn     func()
	result interface{}
}

// InMemoryDocumentSessionOperations represents database operations queued
// in memory
type InMemoryDocumentSessionOperations struct {
	clientSessionID             int
	deletedEntities             *objectSet
	requestExecutor             *RequestExecutor
	operationExecutor           *OperationExecutor
	pendingLazyOperations       []ILazyOperation
	onEvaluateLazy              map[ILazyOperation]*onLazyEval
	generateDocumentKeysOnStore bool
	sessionInfo                 *SessionInfo
	saveChangesOptions          *BatchOptions
	transactionMode             TransactionMode
	isDisposed                  bool

	// Note: skipping unused isDisposed
	id string

	onBeforeStore      []func(*BeforeStoreEventArgs)
	onAfterSaveChanges []func(*AfterSaveChangesEventArgs)

	onBeforeDelete []func(*BeforeDeleteEventArgs)
	onBeforeQuery  []func(*BeforeQueryEventArgs)

	// ids of entities that were deleted
	knownMissingIds []string // case insensitive

	// Note: skipping unused externalState

	documentsByID *documentsByID

	// Translate between an ID and its associated entity
	// TODO: ignore case for keys
	includedDocumentsByID map[string]*documentInfo

	// hold the data required to manage the data for RavenDB's Unit of Work
	// Note: in Java it's LinkedHashMap where iteration order is same
	// as insertion order. In Go map has random iteration order so we must
	// use an array
	documentsByEntity []*documentInfo

	documentStore *DocumentStore

	DatabaseName string

	numberOfRequests int

	Conventions *DocumentConventions

	maxNumberOfRequestsPerSession int
	useOptimisticConcurrency      bool

	deferredCommands []ICommandData

	noTracking bool

	// Note: using value type so that lookups are based on value
	deferredCommandsMap map[idTypeAndName]ICommandData

	generateEntityIDOnTheClient *generateEntityIDOnTheClient
	entityToJSON                *entityToJSON

	// Note: in java DocumentSession inherits from InMemoryDocumentSessionOperations
	// so we can upcast/downcast between them
	// In Go we need a backlink to reach DocumentSession
	session *DocumentSession
}

func newInMemoryDocumentSessionOperations(store *DocumentStore, id string, options *SessionOptions) *InMemoryDocumentSessionOperations {
	clientSessionID := newClientSessionID()
	databaseName := options.Database
	if databaseName == "" {
		databaseName = store.GetDatabase()
	}
	re := options.RequestExecutor
	if re == nil {
		re = store.GetRequestExecutor(databaseName)
	}

	res := &InMemoryDocumentSessionOperations{
		id:                            id,
		DatabaseName:                  databaseName,
		documentStore:                 store,
		requestExecutor:               re,
		noTracking:                    options.NoTracking,
		useOptimisticConcurrency:      re.conventions.UseOptimisticConcurrency,
		maxNumberOfRequestsPerSession: re.conventions.MaxNumberOfRequestsPerSession,
		sessionInfo:                   &SessionInfo{SessionID: clientSessionID},
		transactionMode:               options.TransactionMode,

		deletedEntities:             newObjectSet(),
		generateDocumentKeysOnStore: true,
		documentsByID:               newDocumentsByID(),
		includedDocumentsByID:       map[string]*documentInfo{},
		documentsByEntity:           []*documentInfo{},
		deferredCommandsMap:         map[idTypeAndName]ICommandData{},
	}

	genIDFunc := func(entity interface{}) (string, error) {
		return res.GenerateID(entity)
	}
	res.generateEntityIDOnTheClient = newGenerateEntityIDOnTheClient(re.conventions, genIDFunc)
	res.entityToJSON = newEntityToJSON(res)
	return res
}

func (s *InMemoryDocumentSessionOperations) GetCurrentSessionNode() (*ServerNode, error) {
	var result *CurrentIndexAndNode
	readBalance := s.documentStore.GetConventions().ReadBalanceBehavior
	var err error
	switch readBalance {
	case ReadBalanceBehaviorNone:
		result, err = s.requestExecutor.getPreferredNode()
	case ReadBalanceBehaviorRoundRobin:
		result, err = s.requestExecutor.getNodeBySessionID(s.clientSessionID)
	case ReadBalanceBehaviorFastestNode:
		result, err = s.requestExecutor.getFastestNode()
	default:
		return nil, newIllegalArgumentError("unknown readBalance value %s", readBalance)
	}
	if err != nil {
		return nil, err
	}
	return result.currentNode, nil
}

// GetDeferredCommandsCount returns number of deferred commands
func (s *InMemoryDocumentSessionOperations) GetDeferredCommandsCount() int {
	return len(s.deferredCommands)
}

// AddBeforeStoreStoreListener registers a function that will be called before storing an entity.
// Returns listener id that can be passed to RemoveBeforeStoreListener to unregister
// the listener.
func (s *InMemoryDocumentSessionOperations) AddBeforeStoreListener(handler func(*BeforeStoreEventArgs)) int {
	s.onBeforeStore = append(s.onBeforeStore, handler)
	return len(s.onBeforeStore) - 1
}

// RemoveBeforeStoreListener removes a listener given id returned by AddBeforeStoreListener
func (s *InMemoryDocumentSessionOperations) RemoveBeforeStoreListener(handlerID int) {
	s.onBeforeStore[handlerID] = nil
}

// AddAfterSaveChangesListener registers a function that will be called before saving changes.
// Returns listener id that can be passed to RemoveAfterSaveChangesListener to unregister
// the listener.
func (s *InMemoryDocumentSessionOperations) AddAfterSaveChangesListener(handler func(*AfterSaveChangesEventArgs)) int {
	s.onAfterSaveChanges = append(s.onAfterSaveChanges, handler)
	return len(s.onAfterSaveChanges) - 1
}

// RemoveAfterSaveChangesListener removes a listener given id returned by AddAfterSaveChangesListener
func (s *InMemoryDocumentSessionOperations) RemoveAfterSaveChangesListener(handlerID int) {
	s.onAfterSaveChanges[handlerID] = nil
}

// AddBeforeDeleteListener registers a function that will be called before deleting an entity.
// Returns listener id that can be passed to RemoveBeforeDeleteListener to unregister
// the listener.
func (s *InMemoryDocumentSessionOperations) AddBeforeDeleteListener(handler func(*BeforeDeleteEventArgs)) int {
	s.onBeforeDelete = append(s.onBeforeDelete, handler)
	return len(s.onBeforeDelete) - 1
}

// RemoveBeforeDeleteListener removes a listener given id returned by AddBeforeDeleteListener
func (s *InMemoryDocumentSessionOperations) RemoveBeforeDeleteListener(handlerID int) {
	s.onBeforeDelete[handlerID] = nil
}

// AddBeforeQueryListener registers a function that will be called before running a query.
// It allows customizing query via DocumentQueryCustomization.
// Returns listener id that can be passed to RemoveBeforeQueryListener to unregister
// the listener.
func (s *InMemoryDocumentSessionOperations) AddBeforeQueryListener(handler func(*BeforeQueryEventArgs)) int {
	s.onBeforeQuery = append(s.onBeforeQuery, handler)
	return len(s.onBeforeQuery) - 1
}

// RemoveBeforeQueryListener removes a listener given id returned by AddBeforeQueryListener
func (s *InMemoryDocumentSessionOperations) RemoveBeforeQueryListener(handlerID int) {
	s.onBeforeQuery[handlerID] = nil
}

func (s *InMemoryDocumentSessionOperations) getEntityToJSON() *entityToJSON {
	return s.entityToJSON
}

// GetNumberOfEntitiesInUnitOfWork returns number of entities
func (s *InMemoryDocumentSessionOperations) GetNumberOfEntitiesInUnitOfWork() int {
	return len(s.documentsByEntity)
}

// GetConventions returns DocumentConventions
func (s *InMemoryDocumentSessionOperations) GetConventions() *DocumentConventions {
	return s.requestExecutor.conventions
}

func (s *InMemoryDocumentSessionOperations) GenerateID(entity interface{}) (string, error) {
	return s.GetConventions().GenerateDocumentID(s.DatabaseName, entity)
}

func (s *InMemoryDocumentSessionOperations) GetDocumentStore() *DocumentStore {
	return s.documentStore
}

func (s *InMemoryDocumentSessionOperations) GetRequestExecutor() *RequestExecutor {
	return s.requestExecutor
}

func (s *InMemoryDocumentSessionOperations) GetOperations() *OperationExecutor {
	if s.operationExecutor == nil {
		dbName := s.DatabaseName
		s.operationExecutor = s.GetDocumentStore().Operations().ForDatabase(dbName)
	}
	return s.operationExecutor
}

// GetNumberOfRequests returns number of requests sent to the server
func (s *InMemoryDocumentSessionOperations) GetNumberOfRequests() int {
	return s.numberOfRequests
}

// GetMetadataFor gets the metadata for the specified entity.
// TODO: should we make the API more robust by accepting **struct as well as
// *struct and doing the necessary tweaking automatically? It looks like
// GetMetadataFor(&foo) might be used reflexively and it might not be easy
// to figure out why it fails. Alternatively, error out early with informative
// error message
func (s *InMemoryDocumentSessionOperations) GetMetadataFor(instance interface{}) (*MetadataAsDictionary, error) {
	err := checkValidEntityIn(instance, "instance")
	if err != nil {
		return nil, err
	}

	documentInfo, err := s.getDocumentInfo(instance)
	if err != nil {
		return nil, err
	}
	if documentInfo.metadataInstance != nil {
		return documentInfo.metadataInstance, nil
	}

	metadataAsJSON := documentInfo.metadata
	metadata := NewMetadataAsDictionaryWithSource(metadataAsJSON)
	documentInfo.metadataInstance = metadata
	return metadata, nil
}

// GetChangeVectorFor returns metadata for a given instance
// empty string means there is not change vector
func (s *InMemoryDocumentSessionOperations) GetChangeVectorFor(instance interface{}) (*string, error) {
	err := checkValidEntityIn(instance, "instance")
	if err != nil {
		return nil, err
	}

	documentInfo, err := s.getDocumentInfo(instance)
	if err != nil {
		return nil, err
	}
	changeVector := jsonGetAsTextPointer(documentInfo.metadata, MetadataChangeVector)
	return changeVector, nil
}

// GetLastModifiedFor returns last modified time for a given instance
func (s *InMemoryDocumentSessionOperations) GetLastModifiedFor(instance interface{}) (*time.Time, error) {
	err := checkValidEntityIn(instance, "instance")
	if err != nil {
		return nil, err
	}

	documentInfo, err := s.getDocumentInfo(instance)
	if err != nil {
		return nil, err
	}
	lastModified, ok := jsonGetAsString(documentInfo.metadata, MetadataLastModified)
	if !ok {
		return nil, nil
	}
	t, err := ParseTime(lastModified)
	if err != nil {
		return nil, err
	}
	return &t, err
}

func getDocumentInfoByEntity(docs []*documentInfo, entity interface{}) *documentInfo {
	for _, doc := range docs {
		if doc.entity == entity {
			return doc
		}
	}
	return nil
}

// adds or replaces documentInfo in a list by entity
func setDocumentInfo(docsRef *[]*documentInfo, toAdd *documentInfo) {
	docs := *docsRef
	entity := toAdd.entity
	for i, doc := range docs {
		if doc.entity == entity {
			docs[i] = toAdd
			return
		}
	}
	*docsRef = append(docs, toAdd)
}

// returns deleted documentInfo
func deleteDocumentInfoByEntity(docsRef *[]*documentInfo, entity interface{}) *documentInfo {
	docs := *docsRef
	for i, doc := range docs {
		if doc.entity == entity {
			docs = append(docs[:i], docs[i+1:]...)
			*docsRef = docs
			return doc
		}
	}
	return nil
}

// getDocumentInfo returns documentInfo for a given instance
// Returns nil if not found
func (s *InMemoryDocumentSessionOperations) getDocumentInfo(instance interface{}) (*documentInfo, error) {
	documentInfo := getDocumentInfoByEntity(s.documentsByEntity, instance)
	if documentInfo != nil {
		return documentInfo, nil
	}

	id, ok := s.generateEntityIDOnTheClient.tryGetIDFromInstance(instance)
	if !ok {
		return nil, newIllegalStateError("Could not find the document id for %s", instance)
	}

	if err := s.assertNoNonUniqueInstance(instance, id); err != nil {
		return nil, err
	}

	err := fmt.Errorf("Document %#v doesn't exist in the session", instance)
	return nil, err
}

// IsLoaded returns true if document with this id is loaded
func (s *InMemoryDocumentSessionOperations) IsLoaded(id string) bool {
	return s.IsLoadedOrDeleted(id)
}

// IsLoadedOrDeleted returns true if document with this id is loaded
func (s *InMemoryDocumentSessionOperations) IsLoadedOrDeleted(id string) bool {
	documentInfo := s.documentsByID.getValue(id)
	if documentInfo != nil && documentInfo.document != nil {
		// is loaded
		return true
	}
	if s.IsDeleted(id) {
		return true
	}
	_, found := s.includedDocumentsByID[id]
	return found
}

// IsDeleted returns true if document with this id is deleted in this session
func (s *InMemoryDocumentSessionOperations) IsDeleted(id string) bool {
	return stringArrayContainsNoCase(s.knownMissingIds, id)
}

// GetDocumentID returns id of a given instance
func (s *InMemoryDocumentSessionOperations) GetDocumentID(instance interface{}) string {
	if instance == nil {
		return ""
	}
	value := getDocumentInfoByEntity(s.documentsByEntity, instance)
	if value == nil {
		return ""
	}
	return value.id
}

// IncrementRequestCount increments requests count
func (s *InMemoryDocumentSessionOperations) incrementRequestCount() error {
	s.numberOfRequests++
	if s.numberOfRequests > s.maxNumberOfRequestsPerSession {
		return newIllegalStateError("exceeded max number of requests per session of %d", s.maxNumberOfRequestsPerSession)
	}
	return nil
}

// result is a pointer to expected value
func (s *InMemoryDocumentSessionOperations) TrackEntityInDocumentInfo(result interface{}, documentFound *documentInfo) error {
	return s.TrackEntity(result, documentFound.id, documentFound.document, documentFound.metadata, false)
}

// TrackEntity tracks a given object
// result is a pointer to a decoded value (e.g. **Foo) and will be set with
// value decoded from JSON (e.g. *result = &Foo{})
func (s *InMemoryDocumentSessionOperations) TrackEntity(result interface{}, id string, document map[string]interface{}, metadata map[string]interface{}, noTracking bool) error {
	if id == "" {
		return s.deserializeFromTransformer(result, "", document)
	}

	docInfo := s.documentsByID.getValue(id)
	if docInfo != nil {
		// the local instance may have been changed, we adhere to the current Unit of Work
		// instance, and return that, ignoring anything new.

		if docInfo.entity == nil {
			err := s.entityToJSON.convertToEntity2(result, id, document)
			if err != nil {
				return err
			}
			docInfo.setEntity(result)
		} else {
			err := setInterfaceToValue(result, docInfo.entity)
			if err != nil {
				return err
			}
		}

		if !noTracking {
			delete(s.includedDocumentsByID, id)
			setDocumentInfo(&s.documentsByEntity, docInfo)
		}
		return nil
	}

	docInfo = s.includedDocumentsByID[id]
	if docInfo != nil {
		// TODO: figure out a test case that fails if I invert setResultToDocEntity
		setResultToDocEntity := true
		if docInfo.entity == nil {
			err := s.entityToJSON.convertToEntity2(result, id, document)
			if err != nil {
				return err
			}
			docInfo.setEntity(result)
			setResultToDocEntity = false
		}

		if !noTracking {
			delete(s.includedDocumentsByID, id)
			s.documentsByID.add(docInfo)
			setDocumentInfo(&s.documentsByEntity, docInfo)
		}

		if setResultToDocEntity {
			return setInterfaceToValue(result, docInfo.entity)
		}
		return nil
	}

	err := s.entityToJSON.convertToEntity2(result, id, document)
	if err != nil {
		return err
	}

	changeVector := jsonGetAsTextPointer(metadata, MetadataChangeVector)
	if changeVector == nil {
		return newIllegalStateError("Document %s must have Change Vector", id)
	}

	if !noTracking {
		newDocumentInfo := &documentInfo{}
		newDocumentInfo.id = id
		newDocumentInfo.document = document
		newDocumentInfo.metadata = metadata
		newDocumentInfo.setEntity(result)
		newDocumentInfo.changeVector = changeVector

		s.documentsByID.add(newDocumentInfo)
		setDocumentInfo(&s.documentsByEntity, newDocumentInfo)
	}

	return nil
}

// will convert **Foo => *Foo if tp is *Foo and o is **Foo
// TODO: probably there's a better way
// Test case: TestCachingOfDocumentInclude.cofi_can_avoid_using_server_for_multiload_with_include_if_everything_is_in_session_cache
func matchValueToType(o interface{}, tp reflect.Type) interface{} {
	vt := reflect.TypeOf(o)
	if vt == tp {
		return o
	}
	panicIf(vt.Kind() != reflect.Ptr, "couldn't match type ov v (%T) to %s\n", o, tp)
	vt = vt.Elem()
	panicIf(vt != tp, "couldn't match type ov v (%T) to %s\n", o, tp)
	v := reflect.ValueOf(o)
	v = v.Elem()
	return v.Interface()
}

// Delete marks the specified entity for deletion. The entity will be deleted when SaveChanges is called.
func (s *InMemoryDocumentSessionOperations) Delete(entity interface{}) error {
	err := checkValidEntityIn(entity, "entity")
	if err != nil {
		return err
	}

	value := getDocumentInfoByEntity(s.documentsByEntity, entity)
	if value == nil {
		return newIllegalStateError("%#v is not associated with the session, cannot delete unknown entity instance", entity)
	}

	s.deletedEntities.add(entity)
	delete(s.includedDocumentsByID, value.id)
	s.knownMissingIds = append(s.knownMissingIds, value.id)
	return nil
}

// DeleteByID marks the specified entity for deletion. The entity will be deleted when SaveChanges is called.
// WARNING: This method will not call beforeDelete listener!
func (s *InMemoryDocumentSessionOperations) DeleteByID(id string, expectedChangeVector string) error {
	if id == "" {
		return newIllegalArgumentError("id cannot be empty")
	}

	var changeVector string
	documentInfo := s.documentsByID.getValue(id)
	if documentInfo != nil {
		newObj := convertEntityToJSON(documentInfo.entity, documentInfo)
		if documentInfo.entity != nil && s.entityChanged(newObj, documentInfo, nil) {
			return newIllegalStateError("Can't delete changed entity using identifier. Use delete(Class clazz, T entity) instead.")
		}

		if documentInfo.entity != nil {
			deleteDocumentInfoByEntity(&s.documentsByEntity, documentInfo.entity)
		}

		s.documentsByID.remove(id)
		if documentInfo.changeVector != nil {
			changeVector = *documentInfo.changeVector
		}
	}

	s.knownMissingIds = append(s.knownMissingIds, id)
	if !s.useOptimisticConcurrency {
		changeVector = ""
	}
	cmdData := NewDeleteCommandData(id, firstNonEmptyString(expectedChangeVector, changeVector))
	s.Defer(cmdData)
	return nil
}

// checks if entity is of valid type for operations like Store(), Delete(), GetMetadataFor() etc.
// We support non-nil values of *struct and *map[string]interface{}
// see handling_maps.md for why *map[string]interface{} and not map[string]interface{}
func checkValidEntityIn(v interface{}, argName string) error {
	if v == nil {
		return newIllegalArgumentError("%s can't be nil", argName)
	}

	if _, ok := v.(map[string]interface{}); ok {
		// possibly a common mistake, so try to provide a helpful error message
		typeGot := fmt.Sprintf("%T", v)
		typeExpect := "*" + typeGot
		return newIllegalArgumentError("%s can't be of type %s, try passing %s", argName, typeGot, typeExpect)
	}

	if _, ok := v.(*map[string]interface{}); ok {
		rv := reflect.ValueOf(v)
		if rv.IsNil() {
			return newIllegalArgumentError("%s can't be a nil pointer to a map", argName)
		}
		rv = rv.Elem()
		if rv.IsNil() {
			return newIllegalArgumentError("%s can't be a pointer to a nil map", argName)
		}
		return nil
	}

	tp := reflect.TypeOf(v)
	if tp.Kind() == reflect.Struct {
		// possibly a common mistake, so try to provide a helpful error message
		typeGot := fmt.Sprintf("%T", v)
		typeExpect := "*" + typeGot
		return newIllegalArgumentError("%s can't be of type %s, try passing %s", argName, typeGot, typeExpect)
	}

	if tp.Kind() != reflect.Ptr {
		return newIllegalArgumentError("%s can't be of type %T", argName, v)
	}

	// at this point it's a pointer to some type
	if reflect.ValueOf(v).IsNil() {
		return newIllegalArgumentError("%s of type %T can't be nil", argName, v)
	}

	// we only allow pointer to struct
	elem := tp.Elem()
	if elem.Kind() == reflect.Struct {
		return nil
	}

	if elem.Kind() == reflect.Ptr {
		// possibly a common mistake, so try to provide a helpful error message
		typeGot := fmt.Sprintf("%T", v)
		typeExpect := typeGot[1:]
		for len(typeExpect) > 0 && typeExpect[0] == '*' {
			typeExpect = typeExpect[1:]
		}
		typeExpect = "*" + typeExpect
		return newIllegalArgumentError("%s can't be of type %s, try passing %s", argName, typeGot, typeExpect)

	}

	return newIllegalArgumentError("%s can't be of type %T", argName, v)
}

// Store stores entity in the session. The entity will be saved when SaveChanges is called.
func (s *InMemoryDocumentSessionOperations) Store(entity interface{}) error {
	err := checkValidEntityIn(entity, "entity")
	if err != nil {
		return err
	}

	_, hasID := s.generateEntityIDOnTheClient.tryGetIDFromInstance(entity)
	concu := ConcurrencyCheckAuto
	if !hasID {
		concu = ConcurrencyCheckForced
	}
	return s.storeInternal(entity, "", "", concu)
}

// StoreWithID stores  entity in the session, explicitly specifying its Id. The entity will be saved when SaveChanges is called.
func (s *InMemoryDocumentSessionOperations) StoreWithID(entity interface{}, id string) error {
	err := checkValidEntityIn(entity, "entity")
	if err != nil {
		return err
	}

	return s.storeInternal(entity, "", id, ConcurrencyCheckAuto)
}

// StoreWithChangeVectorAndID stores entity in the session, explicitly specifying its id and change vector. The entity will be saved when SaveChanges is called.
func (s *InMemoryDocumentSessionOperations) StoreWithChangeVectorAndID(entity interface{}, changeVector string, id string) error {
	err := checkValidEntityIn(entity, "entity")
	if err != nil {
		return err
	}

	concurr := ConcurrencyCheckDisabled
	if changeVector != "" {
		concurr = ConcurrencyCheckForced
	}

	return s.storeInternal(entity, changeVector, id, concurr)
}

func (s *InMemoryDocumentSessionOperations) rememberEntityForDocumentIdGeneration(entity interface{}) error {
	return newNotImplementedError("You cannot set GenerateDocumentIDsOnStore to false without implementing rememberEntityForDocumentIdGeneration")
}

func (s *InMemoryDocumentSessionOperations) storeInternal(entity interface{}, changeVector string, id string, forceConcurrencyCheck ConcurrencyCheckMode) error {
	value := getDocumentInfoByEntity(s.documentsByEntity, entity)
	if value != nil {
		if changeVector != "" {
			value.changeVector = &changeVector
		}
		value.concurrencyCheckMode = forceConcurrencyCheck
		return nil
	}

	var err error
	if id == "" {
		if s.generateDocumentKeysOnStore {
			if id, err = s.generateEntityIDOnTheClient.generateDocumentKeyForStorage(entity); err != nil {
				return err
			}
		} else {
			if err = s.rememberEntityForDocumentIdGeneration(entity); err != nil {
				return err
			}
		}
	} else {
		// Store it back into the Id field so the client has access to it
		s.generateEntityIDOnTheClient.trySetIdentity(entity, id)
	}

	tmp := newIDTypeAndName(id, CommandClientAnyCommand, "")
	if _, ok := s.deferredCommandsMap[tmp]; ok {
		return newIllegalStateError("Can't Store document, there is a deferred command registered for this document in the session. Document id: %s", id)
	}

	if s.deletedEntities.contains(entity) {
		return newIllegalStateError("Can't Store object, it was already deleted in this session.  Document id: %s", id)
	}

	// we make the check here even if we just generated the ID
	// users can override the ID generation behavior, and we need
	// to detect if they generate duplicates.

	if err := s.assertNoNonUniqueInstance(entity, id); err != nil {
		return err
	}

	collectionName := s.requestExecutor.GetConventions().getCollectionName(entity)
	metadata := map[string]interface{}{}
	if collectionName != "" {
		metadata[MetadataCollection] = collectionName
	}
	goType := s.requestExecutor.GetConventions().getGoTypeName(entity)
	if goType != "" {
		metadata[MetadataRavenGoType] = goType
	}
	if id != "" {
		s.knownMissingIds = stringArrayRemoveNoCase(s.knownMissingIds, id)
	}
	var changeVectorPtr *string
	if changeVector != "" {
		changeVectorPtr = &changeVector
	}
	s.storeEntityInUnitOfWork(id, entity, changeVectorPtr, metadata, forceConcurrencyCheck)
	return nil
}

func (s *InMemoryDocumentSessionOperations) storeEntityInUnitOfWork(id string, entity interface{}, changeVector *string, metadata map[string]interface{}, forceConcurrencyCheck ConcurrencyCheckMode) {
	s.deletedEntities.remove(entity)
	if id != "" {
		s.knownMissingIds = stringArrayRemoveNoCase(s.knownMissingIds, id)
	}
	documentInfo := &documentInfo{}
	documentInfo.id = id
	documentInfo.metadata = metadata
	documentInfo.changeVector = changeVector
	documentInfo.concurrencyCheckMode = forceConcurrencyCheck
	documentInfo.setEntity(entity)
	documentInfo.newDocument = true
	documentInfo.document = nil

	setDocumentInfo(&s.documentsByEntity, documentInfo)
	if id != "" {
		s.documentsByID.add(documentInfo)
	}
}

func (s *InMemoryDocumentSessionOperations) assertNoNonUniqueInstance(entity interface{}, id string) error {
	nLastChar := len(id) - 1
	if len(id) == 0 || id[nLastChar] == '|' || id[nLastChar] == '/' {
		return nil
	}
	info := s.documentsByID.getValue(id)
	if info == nil || info.entity == entity {
		return nil
	}

	return newNonUniqueObjectError("Attempted to associate a different object with id '" + id + "'.")
}

func (s *InMemoryDocumentSessionOperations) prepareForSaveChanges() (*saveChangesData, error) {
	result := newSaveChangesData(s)

	s.deferredCommands = nil
	s.deferredCommandsMap = make(map[idTypeAndName]ICommandData)

	err := s.prepareForEntitiesDeletion(result, nil)
	if err != nil {
		return nil, err
	}
	err = s.prepareForEntitiesPuts(result)
	if err != nil {
		return nil, err
	}

	if len(s.deferredCommands) > 0 {
		// this allow OnBeforeStore to call Defer during the call to include
		// additional values during the same SaveChanges call
		result.deferredCommands = append(result.deferredCommands, s.deferredCommands...)
		for k, v := range s.deferredCommandsMap {
			result.deferredCommandsMap[k] = v
		}
		s.deferredCommands = nil
		s.deferredCommandsMap = nil
	}
	return result, nil
}

func (s *InMemoryDocumentSessionOperations) UpdateMetadataModifications(documentInfo *documentInfo) bool {
	dirty := false
	metadataInstance := documentInfo.metadataInstance
	metadata := documentInfo.metadata
	if metadataInstance != nil {
		if metadataInstance.IsDirty() {
			dirty = true
		}
		props := metadataInstance.KeySet()
		for _, prop := range props {
			propValue, ok := metadataInstance.Get(prop)
			if !ok {
				dirty = true
				continue
			}
			if d, ok := propValue.(*MetadataAsDictionary); ok {
				if d.IsDirty() {
					dirty = true
				}
			}
			metadata[prop] = propValue
		}
	}
	return dirty
}

func (s *InMemoryDocumentSessionOperations) prepareForEntitiesDeletion(result *saveChangesData, changes map[string][]*DocumentsChanges) error {
	for deletedEntity := range s.deletedEntities.items {
		documentInfo := getDocumentInfoByEntity(s.documentsByEntity, deletedEntity)
		if documentInfo == nil {
			continue
		}
		if changes != nil {
			docChanges := []*DocumentsChanges{}
			change := &DocumentsChanges{
				FieldNewValue: "",
				FieldOldValue: "",
				Change:        DocumentChangeDocumentDeleted,
			}

			docChanges = append(docChanges, change)
			changes[documentInfo.id] = docChanges
		} else {
			idType := newIDTypeAndName(documentInfo.id, CommandClientAnyCommand, "")
			command := result.deferredCommandsMap[idType]
			if command != nil {
				err := s.throwInvalidDeletedDocumentWithDeferredCommand(command)
				if err != nil {
					return err
				}
			}

			var changeVector *string
			documentInfo = s.documentsByID.getValue(documentInfo.id)

			if documentInfo != nil {
				changeVector = documentInfo.changeVector

				if documentInfo.entity != nil {
					deleteDocumentInfoByEntity(&s.documentsByEntity, documentInfo.entity)
					result.addEntity(documentInfo.entity)
				}

				s.documentsByID.remove(documentInfo.id)
			}

			if !s.useOptimisticConcurrency {
				changeVector = nil
			}

			beforeDeleteEventArgs := newBeforeDeleteEventArgs(s, documentInfo.id, documentInfo.entity)
			for _, handler := range s.onBeforeDelete {
				if handler != nil {
					handler(beforeDeleteEventArgs)
				}
			}

			cmdData := NewDeleteCommandData(documentInfo.id, stringPtrToString(changeVector))
			result.addSessionCommandData(cmdData)
		}

		if len(changes) == 0 {
			s.deletedEntities.clear()
		}
	}
	return nil
}

func (s *InMemoryDocumentSessionOperations) prepareForEntitiesPuts(result *saveChangesData) error {
	for _, entityValue := range s.documentsByEntity {
		if entityValue.ignoreChanges {
			continue
		}
		entityKey := entityValue.entity

		dirtyMetadata := s.UpdateMetadataModifications(entityValue)

		document := convertEntityToJSON(entityKey, entityValue)

		if !s.entityChanged(document, entityValue, nil) && !dirtyMetadata {
			continue
		}

		idType := newIDTypeAndName(entityValue.id, CommandClientNotAttachment, "")
		command := result.deferredCommandsMap[idType]
		if command != nil {
			err := s.throwInvalidModifiedDocumentWithDeferredCommand(command)
			if err != nil {
				return err
			}
		}

		if len(s.onBeforeStore) > 0 {
			beforeStoreEventArgs := newBeforeStoreEventArgs(s, entityValue.id, entityKey)
			for _, handler := range s.onBeforeStore {
				if handler != nil {
					handler(beforeStoreEventArgs)
				}
			}
			if beforeStoreEventArgs.isMetadataAccessed() {
				s.UpdateMetadataModifications(entityValue)
			}
			if beforeStoreEventArgs.isMetadataAccessed() || s.entityChanged(document, entityValue, nil) {
				document = convertEntityToJSON(entityKey, entityValue)
			}
		}

		entityValue.newDocument = false
		result.addEntity(entityKey)

		if entityValue.id != "" {
			s.documentsByID.remove(entityValue.id)
		}

		entityValue.document = document

		var changeVector *string
		if s.useOptimisticConcurrency {
			if entityValue.concurrencyCheckMode != ConcurrencyCheckDisabled {
				// if the user didn't provide a change vector, we'll test for an empty one
				tmp := ""
				changeVector = firstNonNilString(entityValue.changeVector, &tmp)
			} else {
				changeVector = nil // TODO: redundant
			}
		} else if entityValue.concurrencyCheckMode == ConcurrencyCheckForced {
			changeVector = entityValue.changeVector
		} else {
			changeVector = nil // TODO: redundant
		}
		cmdData := newPutCommandDataWithJSON(entityValue.id, changeVector, document)
		result.addSessionCommandData(cmdData)
	}
	return nil
}

func (s *InMemoryDocumentSessionOperations) throwInvalidModifiedDocumentWithDeferredCommand(resultCommand ICommandData) error {
	err := newIllegalStateError("Cannot perform save because document " + resultCommand.getId() + " has been modified by the session and is also taking part in deferred " + resultCommand.getType() + " command")
	return err
}

func (s *InMemoryDocumentSessionOperations) throwInvalidDeletedDocumentWithDeferredCommand(resultCommand ICommandData) error {
	err := newIllegalStateError("Cannot perform save because document " + resultCommand.getId() + " has been deleted by the session and is also taking part in deferred " + resultCommand.getType() + " command")
	return err
}

func (s *InMemoryDocumentSessionOperations) entityChanged(newObj map[string]interface{}, documentInfo *documentInfo, changes map[string][]*DocumentsChanges) bool {
	return jsonOperationEntityChanged(newObj, documentInfo, changes)
}

func (s *InMemoryDocumentSessionOperations) WhatChanged() (map[string][]*DocumentsChanges, error) {
	changes := map[string][]*DocumentsChanges{}
	err := s.prepareForEntitiesDeletion(nil, changes)
	if err != nil {
		return nil, err
	}
	s.getAllEntitiesChanges(changes)
	return changes, nil
}

// Gets a value indicating whether any of the entities tracked by the session has changes.
func (s *InMemoryDocumentSessionOperations) HasChanges() bool {
	if !s.deletedEntities.isEmpty() {
		return true
	}

	for _, documentInfo := range s.documentsByEntity {
		entity := documentInfo.entity
		document := convertEntityToJSON(entity, documentInfo)
		changed := s.entityChanged(document, documentInfo, nil)
		if changed {
			return true
		}
	}
	return false
}

// HasChanged returns true if an entity has changed.
func (s *InMemoryDocumentSessionOperations) HasChanged(entity interface{}) (bool, error) {
	err := checkValidEntityIn(entity, "entity")
	if err != nil {
		return false, err
	}
	documentInfo := getDocumentInfoByEntity(s.documentsByEntity, entity)

	if documentInfo == nil {
		return false, nil
	}

	document := convertEntityToJSON(entity, documentInfo)
	return s.entityChanged(document, documentInfo, nil), nil
}

func (s *InMemoryDocumentSessionOperations) WaitForReplicationAfterSaveChanges(options func(*ReplicationWaitOptsBuilder)) {
	// TODO: what does it do? looks like a no-op
	builder := &ReplicationWaitOptsBuilder{}
	options(builder)

	builderOptions := builder.getOptions()
	if builderOptions.replicationOptions.waitForReplicasTimeout == 0 {
		builderOptions.replicationOptions.waitForReplicasTimeout = time.Second * 15
	}
	builderOptions.replicationOptions.waitForReplicas = true
}

func (s *InMemoryDocumentSessionOperations) WaitForIndexesAfterSaveChanges(options func(*IndexesWaitOptsBuilder)) {
	// TODO: what does it do? looks like a no-op
	builder := &IndexesWaitOptsBuilder{}
	options(builder)

	builderOptions := builder.getOptions()
	if builderOptions.indexOptions.waitForIndexesTimeout == 0 {
		builderOptions.indexOptions.waitForIndexesTimeout = time.Second * 15
	}
	builderOptions.indexOptions.waitForIndexes = true
}

func (s *InMemoryDocumentSessionOperations) getAllEntitiesChanges(changes map[string][]*DocumentsChanges) {
	for _, docInfo := range s.documentsByID.inner {
		s.UpdateMetadataModifications(docInfo)
		entity := docInfo.entity
		newObj := convertEntityToJSON(entity, docInfo)
		s.entityChanged(newObj, docInfo, changes)
	}
}

// IgnoreChangesFor marks the entity as one that should be ignore for change tracking purposes,
// it still takes part in the session, but is ignored for SaveChanges.
func (s *InMemoryDocumentSessionOperations) IgnoreChangesFor(entity interface{}) error {
	if docInfo, err := s.getDocumentInfo(entity); err != nil {
		return err
	} else {
		docInfo.ignoreChanges = true
		return nil
	}
}

// Evict evicts the specified entity from the session.
// Remove the entity from the delete queue and stops tracking changes for this entity.
func (s *InMemoryDocumentSessionOperations) Evict(entity interface{}) error {
	err := checkValidEntityIn(entity, "entity")
	if err != nil {
		return err
	}

	deleted := deleteDocumentInfoByEntity(&s.documentsByEntity, entity)
	if deleted != nil {
		s.documentsByID.remove(deleted.id)
	}

	s.deletedEntities.remove(entity)
	return nil
}

// Clear clears the session
func (s *InMemoryDocumentSessionOperations) Clear() {
	s.documentsByEntity = nil
	s.deletedEntities.clear()
	s.documentsByID = nil
	s.knownMissingIds = nil
	s.includedDocumentsByID = nil
}

// Defer defers commands to be executed on SaveChanges()
func (s *InMemoryDocumentSessionOperations) Defer(commands ...ICommandData) {
	for _, cmd := range commands {
		s.deferredCommands = append(s.deferredCommands, cmd)
		s.deferInternal(cmd)
	}
}

func (s *InMemoryDocumentSessionOperations) deferInternal(command ICommandData) {
	idType := newIDTypeAndName(command.getId(), command.getType(), command.getName())
	s.deferredCommandsMap[idType] = command
	idType = newIDTypeAndName(command.getId(), CommandClientAnyCommand, "")
	s.deferredCommandsMap[idType] = command

	cmdType := command.getType()
	isAttachmentCmd := (cmdType == CommandAttachmentPut) || (cmdType == CommandAttachmentDelete)
	if !isAttachmentCmd {
		idType = newIDTypeAndName(command.getId(), CommandClientNotAttachment, "")
		s.deferredCommandsMap[idType] = command
	}
}

func (s *InMemoryDocumentSessionOperations) _close(isDisposing bool) {
	if s.isDisposed {
		return
	}

	s.isDisposed = true

	// nothing more to do for now
}

// Close performs application-defined tasks associated with freeing, releasing, or resetting unmanaged resources.
func (s *InMemoryDocumentSessionOperations) Close() {
	s._close(true)
}

func (s *InMemoryDocumentSessionOperations) registerMissing(id string) {
	s.knownMissingIds = append(s.knownMissingIds, id)
}

func (s *InMemoryDocumentSessionOperations) unregisterMissing(id string) {
	s.knownMissingIds = stringArrayRemoveNoCase(s.knownMissingIds, id)
}

func (s *InMemoryDocumentSessionOperations) registerIncludes(includes map[string]interface{}) {
	if includes == nil {
		return
	}

	// Java's ObjectNode fieldNames are keys of map[string]interface{}
	for _, fieldValue := range includes {
		// TODO: this needs to check if value inside is nil
		if fieldValue == nil {
			continue
		}
		json, ok := fieldValue.(map[string]interface{})
		panicIf(!ok, "fieldValue of unsupported type %T", fieldValue)
		newDocumentInfo := getNewDocumentInfo(json)
		if tryGetConflict(newDocumentInfo.metadata) {
			continue
		}

		s.includedDocumentsByID[newDocumentInfo.id] = newDocumentInfo
	}
}

func (s *InMemoryDocumentSessionOperations) registerMissingIncludes(results []map[string]interface{}, includes map[string]interface{}, includePaths []string) {
	if len(includePaths) == 0 {
		return
	}
	// TODO: ?? This is a no-op in Java
	/*
		for _, result := range results {
			for _, include := range includePaths {
				if include == IndexingFieldNameDocumentID {
					continue
				}
				// TODO: IncludesUtil.include() but it's a no-op in Java code
			}
		}
	*/
}

func (s *InMemoryDocumentSessionOperations) deserializeFromTransformer(result interface{}, id string, document map[string]interface{}) error {
	return s.entityToJSON.convertToEntity2(result, id, document)
}

/*
func (s *InMemoryDocumentSessionOperations) deserializeFromTransformer(clazz reflect.Type, id string, document map[string]interface{}) (interface{}, error) {
	return s.entityToJSON.ConvertToEntity(clazz, id, document)
}
*/

func (s *InMemoryDocumentSessionOperations) checkIfIdAlreadyIncluded(ids []string, includes []string) bool {
	for _, id := range ids {
		if stringArrayContainsNoCase(s.knownMissingIds, id) {
			continue
		}

		// Check if document was already loaded, then check if we've received it through include
		documentInfo := s.documentsByID.getValue(id)
		if documentInfo == nil {
			documentInfo = s.includedDocumentsByID[id]
			if documentInfo == nil {
				return false
			}
		}

		if documentInfo.entity == nil {
			return false
		}

		if len(includes) == 0 {
			continue
		}

		/* TODO: this is no-op in java
		for _, include := range includes {
			hasAll := true

			includesUtilInclude(documentInfo.getDocument(), include, s -> {
				hasAll[0] &= isLoaded(s);
			})

			if !hasAll {
				return false
			}
		}
		*/
	}

	return true
}

func (s *InMemoryDocumentSessionOperations) refreshInternal(entity interface{}, cmd *GetDocumentsCommand, documentInfo *documentInfo) error {
	document := cmd.Result.Results[0]
	if document == nil {
		return newIllegalStateError("Document '%s' no longer exists and was probably deleted", documentInfo.id)
	}

	value := document[MetadataKey]
	meta := value.(map[string]interface{})
	documentInfo.metadata = meta

	if documentInfo.metadata != nil {
		changeVector := jsonGetAsTextPointer(meta, MetadataChangeVector)
		documentInfo.changeVector = changeVector
	}
	documentInfo.document = document
	e, err := s.entityToJSON.convertToEntity(reflect.TypeOf(entity), documentInfo.id, document)
	if err != nil {
		return err
	}

	panicIf(entity != documentInfo.entity, "entity != documentInfo.entity")
	if err = copyValue(documentInfo.entity, e); err != nil {
		return newRuntimeError("Unable to refresh entity: %s", err)
	}

	return nil
}

func (s *InMemoryDocumentSessionOperations) getOperationResult(results interface{}, result interface{}) error {
	return setInterfaceToValue(results, result)
}

func (s *InMemoryDocumentSessionOperations) onAfterSaveChangesInvoke(afterSaveChangesEventArgs *AfterSaveChangesEventArgs) {
	for _, handler := range s.onAfterSaveChanges {
		if handler != nil {
			handler(afterSaveChangesEventArgs)
		}
	}
}

func (s *InMemoryDocumentSessionOperations) onBeforeQueryInvoke(beforeQueryEventArgs *BeforeQueryEventArgs) {
	for _, handler := range s.onBeforeQuery {
		if handler != nil {
			handler(beforeQueryEventArgs)
		}
	}
}

func processQueryParameters(clazz reflect.Type, indexName string, collectionName string, conventions *DocumentConventions) (string, string, error) {
	isIndex := stringIsNotBlank(indexName)
	isCollection := stringIsNotEmpty(collectionName)

	if isIndex && isCollection {
		return "", "", newIllegalStateError("Parameters indexName and collectionName are mutually exclusive. Please specify only one of them.")
	}

	if !isIndex && !isCollection {
		collectionName = conventions.getCollectionName(clazz)
		if collectionName == "" {
			// TODO: what test would exercise this code path?
			collectionName = MetadataAllDocumentsCollection
		}
	}

	return indexName, collectionName, nil
}

type saveChangesData struct {
	deferredCommands    []ICommandData
	deferredCommandsMap map[idTypeAndName]ICommandData
	sessionCommands     []ICommandData
	entities            []interface{}
	options             *BatchOptions
}

func newSaveChangesData(session *InMemoryDocumentSessionOperations) *saveChangesData {
	return &saveChangesData{
		deferredCommands:    copyDeferredCommands(session.deferredCommands),
		deferredCommandsMap: copyDeferredCommandsMap(session.deferredCommandsMap),
		options:             session.saveChangesOptions,
	}
}

func (d *saveChangesData) addSessionCommandData(cmd ICommandData) {
	d.sessionCommands = append(d.sessionCommands, cmd)
}

func (d *saveChangesData) addEntity(entity interface{}) {
	d.entities = append(d.entities, entity)
}

func copyDeferredCommands(in []ICommandData) []ICommandData {
	return append([]ICommandData(nil), in...)
}

func copyDeferredCommandsMap(in map[idTypeAndName]ICommandData) map[idTypeAndName]ICommandData {
	res := map[idTypeAndName]ICommandData{}
	for k, v := range in {
		res[k] = v
	}
	return res
}

type ReplicationWaitOptsBuilder struct {
	saveChangesOptions *BatchOptions
}

func (b *ReplicationWaitOptsBuilder) getOptions() *BatchOptions {
	if b.saveChangesOptions == nil {
		b.saveChangesOptions = NewBatchOptions()
	}
	return b.saveChangesOptions
}

func (b *ReplicationWaitOptsBuilder) WithTimeout(timeout time.Duration) *ReplicationWaitOptsBuilder {
	b.getOptions().replicationOptions.waitForReplicasTimeout = timeout
	return b
}

func (b *ReplicationWaitOptsBuilder) ThrowOnTimeout(shouldThrow bool) *ReplicationWaitOptsBuilder {
	b.getOptions().replicationOptions.throwOnTimeoutInWaitForReplicas = shouldThrow
	return b
}

func (b *ReplicationWaitOptsBuilder) NumberOfReplicas(replicas int) *ReplicationWaitOptsBuilder {
	b.getOptions().replicationOptions.numberOfReplicasToWaitFor = replicas
	return b
}

func (b *ReplicationWaitOptsBuilder) Majority(waitForMajority bool) *ReplicationWaitOptsBuilder {
	b.getOptions().replicationOptions.majority = waitForMajority
	return b
}

type IndexesWaitOptsBuilder struct {
	saveChangesOptions *BatchOptions
}

func (b *IndexesWaitOptsBuilder) getOptions() *BatchOptions {
	if b.saveChangesOptions == nil {
		b.saveChangesOptions = NewBatchOptions()
	}
	return b.saveChangesOptions
}

func (b *IndexesWaitOptsBuilder) WithTimeout(timeout time.Duration) *IndexesWaitOptsBuilder {
	// TODO: most likely a bug and meant waitForIndexesTimeout
	b.getOptions().replicationOptions.waitForReplicasTimeout = timeout
	return b
}

func (b *IndexesWaitOptsBuilder) ThrowOnTimeout(shouldThrow bool) *IndexesWaitOptsBuilder {
	// TODO: most likely a bug and meant throwOnTimeoutInWaitForIndexes
	b.getOptions().replicationOptions.throwOnTimeoutInWaitForReplicas = shouldThrow
	return b
}

func (b *IndexesWaitOptsBuilder) WaitForIndexes(indexes ...string) *IndexesWaitOptsBuilder {
	b.getOptions().indexOptions.waitForSpecificIndexes = indexes
	return b
}
