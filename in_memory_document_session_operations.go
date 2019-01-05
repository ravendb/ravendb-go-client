package ravendb

import (
	"errors"
	"fmt"
	"reflect"
	"sync/atomic"
	"time"
)

var (
	_clientSessionIDCounter int32 = 1
)

func newClientSessionID() int {
	newID := atomic.AddInt32(&_clientSessionIDCounter, 1)
	return int(newID)
}

// InMemoryDocumentSessionOperations represents database operations queued
// in memory
type InMemoryDocumentSessionOperations struct {
	_clientSessionID            int
	deletedEntities             *objectSet
	_requestExecutor            *RequestExecutor
	_operationExecutor          *OperationExecutor
	pendingLazyOperations       []ILazyOperation
	onEvaluateLazy              map[ILazyOperation]func(interface{})
	generateDocumentKeysOnStore bool
	sessionInfo                 *SessionInfo
	_saveChangesOptions         *BatchOptions
	_isDisposed                 bool

	// Note: skipping unused isDisposed
	id string

	onBeforeStore      []func(interface{}, *BeforeStoreEventArgs)
	onAfterSaveChanges []func(interface{}, *AfterSaveChangesEventArgs)

	onBeforeDelete []func(interface{}, *BeforeDeleteEventArgs)
	onBeforeQuery  []func(interface{}, *BeforeQueryEventArgs)

	// ids of entities that were deleted
	_knownMissingIds []string // case insensitive

	// Note: skipping unused externalState
	// Note: skipping unused getCurrentSessionNode

	documentsByID *documentsByID

	// Translate between an ID and its associated entity
	// TODO: ignore case for keys
	includedDocumentsByID map[string]*documentInfo

	// hold the data required to manage the data for RavenDB's Unit of Work
	// TODO: this uses value semantics, so it works as expected for
	// pointers to structs, but 2 different structs with the same content
	// will match the same object. Should I disallow storing non-pointer structs?
	// convert non-pointer structs to structs?
	documents []*documentInfo

	_documentStore *DocumentStore

	databaseName string

	numberOfRequests int

	Conventions *DocumentConventions

	maxNumberOfRequestsPerSession int
	useOptimisticConcurrency      bool

	deferredCommands []ICommandData

	// Note: using value type so that lookups are based on value
	deferredCommandsMap map[idTypeAndName]ICommandData

	generateEntityIDOnTheClient *generateEntityIDOnTheClient
	entityToJSON                *entityToJSON

	// Note: in java DocumentSession inherits from InMemoryDocumentSessionOperations
	// so we can upcast/downcast between them
	// In Go we need a backlink to reach DocumentSession
	session *DocumentSession
}

// NewInMemoryDocumentSessionOperations creates new InMemoryDocumentSessionOperations
func NewInMemoryDocumentSessionOperations(dbName string, store *DocumentStore, re *RequestExecutor, id string) *InMemoryDocumentSessionOperations {
	clientSessionID := newClientSessionID()
	res := &InMemoryDocumentSessionOperations{
		id:                            id,
		_clientSessionID:              clientSessionID,
		deletedEntities:               newObjectSet(),
		_requestExecutor:              re,
		generateDocumentKeysOnStore:   true,
		sessionInfo:                   &SessionInfo{SessionID: clientSessionID},
		documentsByID:                 newDocumentsByID(),
		includedDocumentsByID:         map[string]*documentInfo{},
		documents:                     []*documentInfo{},
		_documentStore:                store,
		databaseName:                  dbName,
		maxNumberOfRequestsPerSession: re.conventions._maxNumberOfRequestsPerSession,
		useOptimisticConcurrency:      re.conventions.UseOptimisticConcurrency,
		deferredCommandsMap:           make(map[idTypeAndName]ICommandData),
	}

	genIDFunc := func(entity interface{}) string {
		return res.GenerateId(entity)
	}
	res.generateEntityIDOnTheClient = newgenerateEntityIDOnTheClient(re.conventions, genIDFunc)
	res.entityToJSON = newEntityToJSON(res)
	return res
}

func (s *InMemoryDocumentSessionOperations) GetDeferredCommandsCount() int {
	return len(s.deferredCommands)
}

func (s *InMemoryDocumentSessionOperations) AddBeforeStoreListener(handler func(interface{}, *BeforeStoreEventArgs)) int {
	s.onBeforeStore = append(s.onBeforeStore, handler)
	return len(s.onBeforeStore) - 1
}

func (s *InMemoryDocumentSessionOperations) RemoveBeforeStoreListener(handlerIdx int) {
	s.onBeforeStore[handlerIdx] = nil
}

func (s *InMemoryDocumentSessionOperations) AddAfterSaveChangesListener(handler func(interface{}, *AfterSaveChangesEventArgs)) int {
	s.onAfterSaveChanges = append(s.onAfterSaveChanges, handler)
	return len(s.onAfterSaveChanges) - 1
}

func (s *InMemoryDocumentSessionOperations) RemoveAfterSaveChangesListener(handlerIdx int) {
	s.onAfterSaveChanges[handlerIdx] = nil
}

func (s *InMemoryDocumentSessionOperations) AddBeforeDeleteListener(handler func(interface{}, *BeforeDeleteEventArgs)) int {
	s.onBeforeDelete = append(s.onBeforeDelete, handler)
	return len(s.onBeforeDelete) - 1
}

func (s *InMemoryDocumentSessionOperations) RemoveBeforeDeleteListener(handlerIdx int) {
	s.onBeforeDelete[handlerIdx] = nil
}

func (s *InMemoryDocumentSessionOperations) AddBeforeQueryListener(handler func(interface{}, *BeforeQueryEventArgs)) int {
	s.onBeforeQuery = append(s.onBeforeQuery, handler)
	return len(s.onBeforeQuery) - 1
}

func (s *InMemoryDocumentSessionOperations) RemoveBeforeQueryListener(handlerIdx int) {
	s.onBeforeQuery[handlerIdx] = nil
}

func (s *InMemoryDocumentSessionOperations) GetgenerateEntityIDOnTheClient() *generateEntityIDOnTheClient {
	return s.generateEntityIDOnTheClient
}

func (s *InMemoryDocumentSessionOperations) GetEntityToJSON() *entityToJSON {
	return s.entityToJSON
}

// GetNumberOfEntitiesInUnitOfWork returns number of entinties
func (s *InMemoryDocumentSessionOperations) GetNumberOfEntitiesInUnitOfWork() int {
	return len(s.documents)
}

func (s *InMemoryDocumentSessionOperations) GetConventions() *DocumentConventions {
	return s._requestExecutor.conventions
}

func (s *InMemoryDocumentSessionOperations) GetDatabaseName() string {
	return s.databaseName
}

func (s *InMemoryDocumentSessionOperations) GenerateId(entity interface{}) string {
	return s.GetConventions().GenerateDocumentID(s.GetDatabaseName(), entity)
}

func (s *InMemoryDocumentSessionOperations) GetDocumentStore() *IDocumentStore {
	return s._documentStore
}

func (s *InMemoryDocumentSessionOperations) GetRequestExecutor() *RequestExecutor {
	return s._requestExecutor
}

func (s *InMemoryDocumentSessionOperations) GetOperations() *OperationExecutor {
	if s._operationExecutor == nil {
		dbName := s.GetDatabaseName()
		s._operationExecutor = s.GetDocumentStore().Operations().ForDatabase(dbName)
	}
	return s._operationExecutor
}

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
	if instance == nil {
		return nil, newIllegalArgumentError("Instance cannot be null")
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
	if instance == nil {
		return nil, newIllegalArgumentError("Instance cannot be null")
	}

	documentInfo, err := s.getDocumentInfo(instance)
	if err != nil {
		return nil, err
	}
	changeVector := jsonGetAsTextPointer(documentInfo.metadata, MetadataChangeVector)
	return changeVector, nil
}

// GetLastModifiedFor retursn last modified time for a given instance
func (s *InMemoryDocumentSessionOperations) GetLastModifiedFor(instance interface{}) (time.Time, bool) {
	panic("NYI")

	var res time.Time
	return res, false
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
			// optimized deletion: replace deleted element with last element
			// and shrink slice by 1
			n := len(docs)
			docs[i] = docs[n-1]
			docs[n-1] = nil // to help garbage collector
			*docsRef = docs[:n-1]
			return doc
		}
	}
	return nil
}

// getDocumentInfo returns documentInfo for a given instance
// Returns nil if not found
func (s *InMemoryDocumentSessionOperations) getDocumentInfo(instance interface{}) (*documentInfo, error) {
	documentInfo := getDocumentInfoByEntity(s.documents, instance)
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
	return stringArrayContainsNoCase(s._knownMissingIds, id)
}

// GetDocumentID returns id of a given instance
func (s *InMemoryDocumentSessionOperations) GetDocumentID(instance interface{}) string {
	if instance == nil {
		return ""
	}
	value := getDocumentInfoByEntity(s.documents, instance)
	if value == nil {
		return ""
	}
	return value.id
}

// IncrementRequestCount increments requests count
func (s *InMemoryDocumentSessionOperations) IncrementRequestCount() error {
	s.numberOfRequests++
	if s.numberOfRequests > s.maxNumberOfRequestsPerSession {
		return newIllegalStateError("exceeded max number of reqeusts per session of %d", s.maxNumberOfRequestsPerSession)
	}
	return nil
}

// TrackEntityInDocumentInfoOld tracks entity in documentInfo
func (s *InMemoryDocumentSessionOperations) TrackEntityInDocumentInfoOld(clazz reflect.Type, documentFound *documentInfo) (interface{}, error) {
	return s.TrackEntityOld(clazz, documentFound.id, documentFound.document, documentFound.metadata, false)
}

func (s *InMemoryDocumentSessionOperations) TrackEntityInDocumentInfo(result interface{}, documentFound *documentInfo) error {
	return s.TrackEntity(result, documentFound.id, documentFound.document, documentFound.metadata, false)
}

// TrackEntity tracks a given object
func (s *InMemoryDocumentSessionOperations) TrackEntity(result interface{}, id string, document map[string]interface{}, metadata map[string]interface{}, noTracking bool) error {
	if id == "" {
		s.DeserializeFromTransformer2(result, "", document)
		return nil
	}

	docInfo := s.documentsByID.getValue(id)
	if docInfo != nil {
		// the local instance may have been changed, we adhere to the current Unit of Work
		// instance, and return that, ignoring anything new.

		if docInfo.entity == nil {
			s.entityToJSON.ConvertToEntity2(result, id, document)
			docInfo.setEntity(result)
		} else {
			setInterfaceToValue(result, docInfo.entity)
		}

		if !noTracking {
			delete(s.includedDocumentsByID, id)
			setDocumentInfo(&s.documents, docInfo)
		}
		return nil
	}

	docInfo = s.includedDocumentsByID[id]
	if docInfo != nil {
		noSet := true
		if docInfo.entity == nil {
			s.entityToJSON.ConvertToEntity2(result, id, document)
			docInfo.setEntity(result)
			noSet = false
		}

		if !noTracking {
			delete(s.includedDocumentsByID, id)
			s.documentsByID.add(docInfo)
			setDocumentInfo(&s.documents, docInfo)
		}
		if noSet {
			setInterfaceToValue(result, docInfo.entity)
		}
		return nil
	}

	s.entityToJSON.ConvertToEntity2(result, id, document)

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
		setDocumentInfo(&s.documents, newDocumentInfo)
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

// TrackEntityOld tracks entity
func (s *InMemoryDocumentSessionOperations) TrackEntityOld(entityType reflect.Type, id string, document map[string]interface{}, metadata map[string]interface{}, noTracking bool) (interface{}, error) {
	var err error
	if id == "" {
		return s.DeserializeFromTransformer(entityType, "", document)
	}

	docInfo := s.documentsByID.getValue(id)
	if docInfo != nil {
		// the local instance may have been changed, we adhere to the current Unit of Work
		// instance, and return that, ignoring anything new.

		needsToMatchType := true
		if docInfo.entity == nil {
			needsToMatchType = false
			e, err := s.entityToJSON.ConvertToEntity(entityType, id, document)
			if err != nil {
				return nil, err
			}
			docInfo.setEntity(e)
		}

		if !noTracking {
			delete(s.includedDocumentsByID, id)
			setDocumentInfo(&s.documents, docInfo)
		}
		if needsToMatchType {
			// TODO: probably there's a better way. Figure out why docInfo.entity is **Foo in the first place
			// Test case: TestCachingOfDocumentInclude.cofi_can_avoid_using_server_for_multiload_with_include_if_everything_is_in_session_cache
			return matchValueToType(docInfo.entity, entityType), nil
		}
		return docInfo.entity, nil
	}

	docInfo = s.includedDocumentsByID[id]
	if docInfo != nil {
		if docInfo.entity == nil {
			e, err := s.entityToJSON.ConvertToEntity(entityType, id, document)
			if err != nil {
				return nil, err
			}
			docInfo.setEntity(e)
		}

		if !noTracking {
			delete(s.includedDocumentsByID, id)
			s.documentsByID.add(docInfo)
			setDocumentInfo(&s.documents, docInfo)
		}

		return docInfo.entity, nil
	}

	entity, err := s.entityToJSON.ConvertToEntity(entityType, id, document)
	if err != nil {
		return nil, err
	}

	changeVector := jsonGetAsTextPointer(metadata, MetadataChangeVector)
	if changeVector == nil {
		return nil, newIllegalStateError("Document %s must have Change Vector", id)
	}

	if !noTracking {
		newDocumentInfo := &documentInfo{}
		newDocumentInfo.id = id
		newDocumentInfo.document = document
		newDocumentInfo.metadata = metadata
		newDocumentInfo.setEntity(entity)
		newDocumentInfo.changeVector = changeVector

		s.documentsByID.add(newDocumentInfo)
		setDocumentInfo(&s.documents, newDocumentInfo)
	}

	return entity, nil
}

// return an error if entity cannot be deleted e.g. because it's a struct
func checkIsDeletable(entity interface{}) error {
	if entity == nil {
		return errors.New("can't delete nil values")
	}
	tp := reflect.TypeOf(entity)
	if tp.Kind() == reflect.Struct {
		return errors.New("can't delete struct values, must pass a pointer to struct")
	}
	if tp.Kind() == reflect.Ptr {
		if reflect.ValueOf(entity).IsNil() {
			return errors.New("can't delete nil values")
		}
		if tp.Elem() != nil && tp.Elem().Kind() == reflect.Ptr {
			return fmt.Errorf("can't delete values of type %T (double pointer)", entity)
		}
	}
	return nil
}

// DeleteEntity marks the specified entity for deletion. The entity will be deleted when SaveChanges is called.
func (s *InMemoryDocumentSessionOperations) DeleteEntity(entity interface{}) error {
	err := checkIsDeletable(entity)
	if err != nil {
		return err
	}

	value := getDocumentInfoByEntity(s.documents, entity)
	if value == nil {
		return newIllegalStateError("%#v is not associated with the session, cannot delete unknown entity instance", entity)
	}

	s.deletedEntities.add(entity)
	delete(s.includedDocumentsByID, value.id)
	s._knownMissingIds = append(s._knownMissingIds, value.id)
	return nil
}

// Delete marks the specified entity for deletion. The entity will be deleted when IDocumentSession.SaveChanges is called.
// WARNING: This method will not call beforeDelete listener!
func (s *InMemoryDocumentSessionOperations) Delete(id string) error {
	return s.DeleteWithChangeVector(id, nil)
}

func (s *InMemoryDocumentSessionOperations) DeleteWithChangeVector(id string, expectedChangeVector *string) error {
	if id == "" {
		return newIllegalArgumentError("Id cannot be empty")
	}

	var changeVector *string
	documentInfo := s.documentsByID.getValue(id)
	if documentInfo != nil {
		newObj := convertEntityToJSON(documentInfo.entity, documentInfo)
		if documentInfo.entity != nil && s.EntityChanged(newObj, documentInfo, nil) {
			return newIllegalStateError("Can't delete changed entity using identifier. Use delete(Class clazz, T entity) instead.")
		}

		if documentInfo.entity != nil {
			deleteDocumentInfoByEntity(&s.documents, documentInfo.entity)
		}

		s.documentsByID.remove(id)
		changeVector = documentInfo.changeVector
	}

	s._knownMissingIds = append(s._knownMissingIds, id)
	if !s.useOptimisticConcurrency {
		changeVector = nil
	}
	cmdData := NewDeleteCommandData(id, firstNonNilString(expectedChangeVector, changeVector))
	s.Defer(cmdData)
	return nil
}

// return an error if entity cannot be stored e.g. because it's a struct
func checkIsStorable(entity interface{}) error {
	if entity == nil {
		return errors.New("can't store nil values")
	}
	tp := reflect.TypeOf(entity)
	if tp.Kind() == reflect.Struct {
		return errors.New("can't store struct values, must pass a pointer to struct")
	}
	if tp.Kind() == reflect.Ptr {
		if reflect.ValueOf(entity).IsNil() {
			return errors.New("can't store nil values")
		}
		if tp.Elem() != nil && tp.Elem().Kind() == reflect.Ptr {
			return fmt.Errorf("can't store values of type %T (double pointer)", entity)
		}
	}
	return nil
}

// Store stores entity in the session. The entity will be saved when SaveChanges is called.
func (s *InMemoryDocumentSessionOperations) Store(entity interface{}) error {
	err := checkIsStorable(entity)
	if err != nil {
		return err
	}

	_, hasID := s.generateEntityIDOnTheClient.tryGetIDFromInstance(entity)
	concu := ConcurrencyCheckAuto
	if !hasID {
		concu = ConcurrencyCheckForced
	}
	return s.storeInternal(entity, nil, "", concu)
}

// StoreWithID stores  entity in the session, explicitly specifying its Id. The entity will be saved when SaveChanges is called.
func (s *InMemoryDocumentSessionOperations) StoreWithID(entity interface{}, id string) error {
	return s.storeInternal(entity, nil, id, ConcurrencyCheckAuto)
}

// StoreWithChangeVectorAndID stores entity in the session, explicitly specifying its id and change vector. The entity will be saved when SaveChanges is called.
func (s *InMemoryDocumentSessionOperations) StoreWithChangeVectorAndID(entity interface{}, changeVector *string, id string) error {
	concurr := ConcurrencyCheckDisabled
	if changeVector != nil {
		concurr = ConcurrencyCheckForced
	}

	return s.storeInternal(entity, changeVector, id, concurr)
}

// TODO: should this return an error?
func (s *InMemoryDocumentSessionOperations) RememberEntityForDocumentIdGeneration(entity interface{}) {
	err := newNotImplementedError("You cannot set GenerateDocumentIDsOnStore to false without implementing RememberEntityForDocumentIdGeneration")
	must(err)
}

func (s *InMemoryDocumentSessionOperations) storeInternal(entity interface{}, changeVector *string, id string, forceConcurrencyCheck ConcurrencyCheckMode) error {
	err := checkIsStorable(entity)
	if err != nil {
		return err
	}

	value := getDocumentInfoByEntity(s.documents, entity)
	if value != nil {
		value.changeVector = firstNonNilString(changeVector, value.changeVector)
		value.concurrencyCheckMode = forceConcurrencyCheck
		return nil
	}

	if id == "" {
		if s.generateDocumentKeysOnStore {
			id = s.generateEntityIDOnTheClient.generateDocumentKeyForStorage(entity)
		} else {
			s.RememberEntityForDocumentIdGeneration(entity)
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

	collectionName := s._requestExecutor.GetConventions().GetCollectionName(entity)
	metadata := map[string]interface{}{}
	if collectionName != "" {
		metadata[MetadataCollection] = collectionName
	}
	goType := s._requestExecutor.GetConventions().getGoTypeName(entity)
	if goType != "" {
		metadata[MetadataRavenGoType] = goType
	}
	if id != "" {
		s._knownMissingIds = stringArrayRemoveNoCase(s._knownMissingIds, id)
	}

	s.storeEntityInUnitOfWork(id, entity, changeVector, metadata, forceConcurrencyCheck)
	return nil
}

func (s *InMemoryDocumentSessionOperations) storeEntityInUnitOfWork(id string, entity interface{}, changeVector *string, metadata map[string]interface{}, forceConcurrencyCheck ConcurrencyCheckMode) {
	s.deletedEntities.remove(entity)
	if id != "" {
		s._knownMissingIds = stringArrayRemoveNoCase(s._knownMissingIds, id)
	}
	documentInfo := &documentInfo{}
	documentInfo.id = id
	documentInfo.metadata = metadata
	documentInfo.changeVector = changeVector
	documentInfo.concurrencyCheckMode = forceConcurrencyCheck
	documentInfo.setEntity(entity)
	documentInfo.newDocument = true
	documentInfo.document = nil

	setDocumentInfo(&s.documents, documentInfo)
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

func (s *InMemoryDocumentSessionOperations) PrepareForSaveChanges() (*SaveChangesData, error) {
	result := NewSaveChangesData(s)

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

func (s *InMemoryDocumentSessionOperations) prepareForEntitiesDeletion(result *SaveChangesData, changes map[string][]*DocumentsChanges) error {
	for deletedEntity := range s.deletedEntities.items {
		documentInfo := getDocumentInfoByEntity(s.documents, deletedEntity)
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
			command := result.GetDeferredCommandsMap()[idType]
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
					deleteDocumentInfoByEntity(&s.documents, documentInfo.entity)
					result.AddEntity(documentInfo.entity)
				}

				s.documentsByID.remove(documentInfo.id)
			}

			if !s.useOptimisticConcurrency {
				changeVector = nil
			}

			beforeDeleteEventArgs := newBeforeDeleteEventArgs(s, documentInfo.id, documentInfo.entity)
			for _, handler := range s.onBeforeDelete {
				if handler != nil {
					handler(s, beforeDeleteEventArgs)
				}
			}

			cmdData := NewDeleteCommandData(documentInfo.id, changeVector)
			result.AddSessionCommandData(cmdData)
		}

		if len(changes) == 0 {
			s.deletedEntities.clear()
		}
	}
	return nil
}

func (s *InMemoryDocumentSessionOperations) prepareForEntitiesPuts(result *SaveChangesData) error {
	for _, entityValue := range s.documents {
		if entityValue.ignoreChanges {
			continue
		}
		entityKey := entityValue.entity

		dirtyMetadata := s.UpdateMetadataModifications(entityValue)

		document := convertEntityToJSON(entityKey, entityValue)

		if !s.EntityChanged(document, entityValue, nil) && !dirtyMetadata {
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
			beforeStoreEventArgs := NewBeforeStoreEventArgs(s, entityValue.id, entityKey)
			for _, handler := range s.onBeforeStore {
				if handler != nil {
					handler(s, beforeStoreEventArgs)
				}
			}
			if beforeStoreEventArgs.isMetadataAccessed() {
				s.UpdateMetadataModifications(entityValue)
			}
			if beforeStoreEventArgs.isMetadataAccessed() || s.EntityChanged(document, entityValue, nil) {
				document = convertEntityToJSON(entityKey, entityValue)
			}
		}

		entityValue.newDocument = false
		result.AddEntity(entityKey)

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
		cmdData := NewPutCommandDataWithJSON(entityValue.id, changeVector, document)
		result.AddSessionCommandData(cmdData)
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

func (s *InMemoryDocumentSessionOperations) EntityChanged(newObj map[string]interface{}, documentInfo *documentInfo, changes map[string][]*DocumentsChanges) bool {
	return jsonOperationEntityChanged(newObj, documentInfo, changes)
}

func (s *InMemoryDocumentSessionOperations) WhatChanged() (map[string][]*DocumentsChanges, error) {
	changes := map[string][]*DocumentsChanges{}
	err := s.prepareForEntitiesDeletion(nil, changes)
	if err != nil {
		return nil, err
	}
	s.GetAllEntitiesChanges(changes)
	return changes, nil
}

// Gets a value indicating whether any of the entities tracked by the session has changes.
func (s *InMemoryDocumentSessionOperations) HasChanges() bool {
	panic("NYI")
	/*
		for (Map.Entry<Object, documentInfo> entity : documentsByEntity.entrySet()) {
			ObjectNode document = entityToJSON.convertEntityToJSON(entity.getKey(), entity.getValue());
			if (entityChanged(document, entity.getValue(), null)) {
				return true;
			}
		}

		return !deletedEntities.isEmpty();
	*/
	return false
}

// Determines whether the specified entity has changed.
func (s *InMemoryDocumentSessionOperations) HasChanged(entity interface{}) bool {
	documentInfo := getDocumentInfoByEntity(s.documents, entity)

	if documentInfo == nil {
		return false
	}

	document := convertEntityToJSON(entity, documentInfo)
	return s.EntityChanged(document, documentInfo, nil)
}

func (s *InMemoryDocumentSessionOperations) GetAllEntitiesChanges(changes map[string][]*DocumentsChanges) {
	for _, docInfo := range s.documentsByID.inner {
		s.UpdateMetadataModifications(docInfo)
		entity := docInfo.entity
		newObj := convertEntityToJSON(entity, docInfo)
		s.EntityChanged(newObj, docInfo, changes)
	}
}

// IgnoreChangesFor marks the entity as one that should be ignore for change tracking purposes,
// it still takes part in the session, but is ignored for SaveChanges.
func (s *InMemoryDocumentSessionOperations) IgnoreChangesFor(entity interface{}) {
	docInfo, _ := s.getDocumentInfo(entity)
	docInfo.ignoreChanges = true
}

// Evict evicts the specified entity from the session.
// Remove the entity from the delete queue and stops tracking changes for this entity.
func (s *InMemoryDocumentSessionOperations) Evict(entity interface{}) {
	deleted := deleteDocumentInfoByEntity(&s.documents, entity)
	if deleted != nil {
		s.documentsByID.remove(deleted.id)
	}

	s.deletedEntities.remove(entity)
}

// Clear clears the session
func (s *InMemoryDocumentSessionOperations) Clear() {
	s.documents = nil
	s.deletedEntities.clear()
	s.documentsByID = nil
	s._knownMissingIds = nil
	s.includedDocumentsByID = nil
}

// Defer defers a command to be executed on saveChanges()
func (s *InMemoryDocumentSessionOperations) Defer(command ICommandData) {
	a := []ICommandData{command}
	s.DeferMany(a)
}

// DeferMany defers commands to be executed on saveChanges()
func (s *InMemoryDocumentSessionOperations) DeferMany(commands []ICommandData) {
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
	if s._isDisposed {
		return
	}

	s._isDisposed = true

	// nothing more to do for now
}

// Close performs application-defined tasks associated with freeing, releasing, or resetting unmanaged resources.
func (s *InMemoryDocumentSessionOperations) Close() {
	s._close(true)
}

// RegisterMissing registers missing value id
func (s *InMemoryDocumentSessionOperations) RegisterMissing(id string) {
	s._knownMissingIds = append(s._knownMissingIds, id)
}

// UnregisterMissing unregisters missing value id
func (s *InMemoryDocumentSessionOperations) UnregisterMissing(id string) {
	s._knownMissingIds = stringArrayRemoveNoCase(s._knownMissingIds, id)
}

// RegisterIncludes registers includes object
func (s *InMemoryDocumentSessionOperations) RegisterIncludes(includes map[string]interface{}) {
	if includes == nil {
		return
	}

	for _, fieldValue := range includes {
		// TODO: this needs to check if value inside is nil
		if fieldValue == nil {
			continue
		}
		json, ok := fieldValue.(map[string]interface{})
		panicIf(!ok, "fieldValue of unsupported type %T", fieldValue)
		newDocumentInfo := getNewDocumentInfo(json)
		if jsonExtensionsTryGetConflict(newDocumentInfo.metadata) {
			continue
		}

		s.includedDocumentsByID[newDocumentInfo.id] = newDocumentInfo
	}
}

func (s *InMemoryDocumentSessionOperations) RegisterMissingIncludes(results ArrayNode, includes map[string]interface{}, includePaths []string) {
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

func (s *InMemoryDocumentSessionOperations) DeserializeFromTransformer2(result interface{}, id string, document map[string]interface{}) {
	s.entityToJSON.ConvertToEntity2(result, id, document)
}

func (s *InMemoryDocumentSessionOperations) DeserializeFromTransformer(clazz reflect.Type, id string, document map[string]interface{}) (interface{}, error) {
	return s.entityToJSON.ConvertToEntity(clazz, id, document)
}

func (s *InMemoryDocumentSessionOperations) checkIfIdAlreadyIncluded(ids []string, includes []string) bool {
	for _, id := range ids {
		if stringArrayContainsNoCase(s._knownMissingIds, id) {
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

			IncludesUtil_include(documentInfo.getDocument(), include, s -> {
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
	e, err := s.entityToJSON.ConvertToEntity(reflect.TypeOf(entity), documentInfo.id, document)
	if err != nil {
		return err
	}
	documentInfo.setEntity(e)

	err = copyValueProperties(entity, documentInfo.entity)
	if err != nil {
		return newRuntimeError("Unable to refresh entity: %s", err)
	}
	return nil
}

func (s *InMemoryDocumentSessionOperations) getOperationResult(results interface{}, result interface{}) error {
	// TODO: is this a no-op?
	//fmt.Printf("InMemoryDocumentSessionOperations.getOperationResult: trying to set results (%T) to result (%T)\n", results, result)
	return errors.New("NYI")
}

func (s *InMemoryDocumentSessionOperations) OnAfterSaveChangesInvoke(afterSaveChangesEventArgs *AfterSaveChangesEventArgs) {
	for _, handler := range s.onAfterSaveChanges {
		if handler != nil {
			handler(s, afterSaveChangesEventArgs)
		}
	}
}

func (s *InMemoryDocumentSessionOperations) OnBeforeQueryInvoke(beforeQueryEventArgs *BeforeQueryEventArgs) {
	for _, handler := range s.onBeforeQuery {
		if handler != nil {
			handler(s, beforeQueryEventArgs)
		}
	}
}

func (s *InMemoryDocumentSessionOperations) processQueryParameters(clazz reflect.Type, indexName string, collectionName string, conventions *DocumentConventions) (string, string) {
	isIndex := stringIsNotBlank(indexName)
	isCollection := stringIsNotEmpty(collectionName)

	if isIndex && isCollection {
		//throw new IllegalStateError("Parameters indexName and collectionName are mutually exclusive. Please specify only one of them.");
		panic("Parameters indexName and collectionName are mutually exclusive. Please specify only one of them.")
	}

	if !isIndex && !isCollection {
		collectionName = conventions.GetCollectionName(clazz)
	}

	return indexName, collectionName
}

type SaveChangesData struct {
	deferredCommands    []ICommandData
	deferredCommandsMap map[idTypeAndName]ICommandData
	sessionCommands     []ICommandData
	entities            []interface{}
	options             *BatchOptions
}

func NewSaveChangesData(session *InMemoryDocumentSessionOperations) *SaveChangesData {
	return &SaveChangesData{
		deferredCommands:    copyDeferredCommands(session.deferredCommands),
		deferredCommandsMap: copyDeferredCommandsMap(session.deferredCommandsMap),
		options:             session._saveChangesOptions,
	}
}

func (d *SaveChangesData) GetDeferredCommands() []ICommandData {
	return d.deferredCommands
}

func (d *SaveChangesData) GetSessionCommands() []ICommandData {
	return d.sessionCommands
}

func (d *SaveChangesData) GetEntities() []interface{} {
	return d.entities
}

func (d *SaveChangesData) GetOptions() *BatchOptions {
	return d.options
}

func (d *SaveChangesData) GetDeferredCommandsMap() map[idTypeAndName]ICommandData {
	return d.deferredCommandsMap
}

func (d *SaveChangesData) AddSessionCommandData(cmd ICommandData) {
	d.sessionCommands = append(d.sessionCommands, cmd)
}

func (d *SaveChangesData) AddEntity(entity interface{}) {
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
