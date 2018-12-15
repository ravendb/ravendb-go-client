package ravendb

import (
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
	deletedEntities             *ObjectSet
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
	includedDocumentsByID map[string]*DocumentInfo

	// hold the data required to manage the data for RavenDB's Unit of Work
	// TODO: this uses value semantics, so it works as expected for
	// pointers to structs, but 2 different structs with the same content
	// will match the same object. Should I disallow storing non-pointer structs?
	// convert non-pointer structs to structs?
	documents []*DocumentInfo

	_documentStore *DocumentStore

	databaseName string

	numberOfRequests int

	Conventions *DocumentConventions

	maxNumberOfRequestsPerSession int
	useOptimisticConcurrency      bool

	deferredCommands []ICommandData

	// Note: using value type so that lookups are based on value
	deferredCommandsMap map[IdTypeAndName]ICommandData

	generateEntityIdOnTheClient *GenerateEntityIdOnTheClient
	entityToJson                *EntityToJson

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
		deletedEntities:               NewObjectSet(),
		_requestExecutor:              re,
		generateDocumentKeysOnStore:   true,
		sessionInfo:                   &SessionInfo{SessionID: clientSessionID},
		documentsByID:                 newDocumentsByID(),
		includedDocumentsByID:         map[string]*DocumentInfo{},
		documents:                     []*DocumentInfo{},
		_documentStore:                store,
		databaseName:                  dbName,
		maxNumberOfRequestsPerSession: re.conventions._maxNumberOfRequestsPerSession,
		useOptimisticConcurrency:      re.conventions.UseOptimisticConcurrency,
		deferredCommandsMap:           make(map[IdTypeAndName]ICommandData),
	}

	genIDFunc := func(entity interface{}) string {
		return res.GenerateId(entity)
	}
	res.generateEntityIdOnTheClient = NewGenerateEntityIdOnTheClient(re.conventions, genIDFunc)
	res.entityToJson = NewEntityToJson(res)
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

func (s *InMemoryDocumentSessionOperations) GetGenerateEntityIdOnTheClient() *GenerateEntityIdOnTheClient {
	return s.generateEntityIdOnTheClient
}

func (s *InMemoryDocumentSessionOperations) GetEntityToJson() *EntityToJson {
	return s.entityToJson
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
	return s.GetConventions().GenerateDocumentId(s.GetDatabaseName(), entity)
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
func (s *InMemoryDocumentSessionOperations) GetMetadataFor(instance interface{}) (*IMetadataDictionary, error) {
	if instance == nil {
		return nil, NewIllegalArgumentException("Instance cannot be null")
	}

	documentInfo, err := s.GetDocumentInfo(instance)
	if err != nil {
		return nil, err
	}
	if documentInfo.metadataInstance != nil {
		return documentInfo.metadataInstance, nil
	}

	metadataAsJson := documentInfo.metadata
	metadata := NewMetadataAsDictionaryWithSource(metadataAsJson)
	documentInfo.metadataInstance = metadata
	return metadata, nil
}

// GetChangeVectorFor returns metadata for a given instance
// empty string means there is not change vector
func (s *InMemoryDocumentSessionOperations) GetChangeVectorFor(instance interface{}) (*string, error) {
	if instance == nil {
		return nil, NewIllegalArgumentException("Instance cannot be null")
	}

	documentInfo, err := s.GetDocumentInfo(instance)
	if err != nil {
		return nil, err
	}
	changeVector := jsonGetAsTextPointer(documentInfo.metadata, Constants_Documents_Metadata_CHANGE_VECTOR)
	return changeVector, nil
}

// GetLastModifiedFor retursn last modified time for a given instance
func (s *InMemoryDocumentSessionOperations) GetLastModifiedFor(instance interface{}) (time.Time, bool) {
	panic("NYI")

	var res time.Time
	return res, false
}

func getDocumentInfoByEntity(docs []*DocumentInfo, entity interface{}) *DocumentInfo {
	for _, doc := range docs {
		if doc.entity == entity {
			return doc
		}
	}
	return nil
}

// adds or replaces DocumentInfo in a list by entity
func setDocumentInfo(docsRef *[]*DocumentInfo, toAdd *DocumentInfo) {
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

// returns deleted DocumentInfo
func deleteDocumentInfoByEntity(docsRef *[]*DocumentInfo, entity interface{}) *DocumentInfo {
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

// GetDocumentInfo returns DocumentInfo for a given instance
// Returns nil if not found
func (s *InMemoryDocumentSessionOperations) GetDocumentInfo(instance interface{}) (*DocumentInfo, error) {
	documentInfo := getDocumentInfoByEntity(s.documents, instance)
	if documentInfo != nil {
		return documentInfo, nil
	}

	id, ok := s.generateEntityIdOnTheClient.tryGetIdFromInstance(instance)
	if !ok {
		return nil, NewIllegalStateException("Could not find the document id for %s", instance)
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
	return StringArrayContainsNoCase(s._knownMissingIds, id)
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
		return NewIllegalStateException("exceeded max number of reqeusts per session of %d", s.maxNumberOfRequestsPerSession)
	}
	return nil
}

// TrackEntityInDocumentInfoOld tracks entity in DocumentInfo
func (s *InMemoryDocumentSessionOperations) TrackEntityInDocumentInfoOld(clazz reflect.Type, documentFound *DocumentInfo) (interface{}, error) {
	return s.TrackEntityOld(clazz, documentFound.id, documentFound.document, documentFound.metadata, false)
}

func (s *InMemoryDocumentSessionOperations) TrackEntityInDocumentInfo(result interface{}, documentFound *DocumentInfo) error {
	return s.TrackEntity(result, documentFound.id, documentFound.document, documentFound.metadata, false)
}

// TrackEntity tracks a given object
func (s *InMemoryDocumentSessionOperations) TrackEntity(result interface{}, id string, document ObjectNode, metadata ObjectNode, noTracking bool) error {
	if id == "" {
		s.DeserializeFromTransformer2(result, "", document)
		return nil
	}

	docInfo := s.documentsByID.getValue(id)
	// TODO: this used to always be false. After fixing the logic it now crashes in ConvertToEntity2
	// Temporarily disable this code path (doesn't affect tests although possibly some currently failing
	// tests are due to this.
	// Re-enable this code path and fix crashes.
	if docInfo != nil {
		// the local instance may have been changed, we adhere to the current Unit of Work
		// instance, and return that, ignoring anything new.

		if docInfo.entity == nil {
			s.entityToJson.ConvertToEntity2(result, id, document)
			docInfo.entity = result
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
			s.entityToJson.ConvertToEntity2(result, id, document)
			docInfo.entity = result
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

	s.entityToJson.ConvertToEntity2(result, id, document)

	changeVector := jsonGetAsTextPointer(metadata, Constants_Documents_Metadata_CHANGE_VECTOR)
	if changeVector == nil {
		return NewIllegalStateException("Document %s must have Change Vector", id)
	}

	if !noTracking {
		newDocumentInfo := NewDocumentInfo()
		newDocumentInfo.id = id
		newDocumentInfo.document = document
		newDocumentInfo.metadata = metadata
		newDocumentInfo.entity = result
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
func (s *InMemoryDocumentSessionOperations) TrackEntityOld(entityType reflect.Type, id string, document ObjectNode, metadata ObjectNode, noTracking bool) (interface{}, error) {
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
			docInfo.entity, err = s.entityToJson.ConvertToEntity(entityType, id, document)
			if err != nil {
				return nil, err
			}
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
			docInfo.entity, err = s.entityToJson.ConvertToEntity(entityType, id, document)
			if err != nil {
				return nil, err
			}
		}

		if !noTracking {
			delete(s.includedDocumentsByID, id)
			s.documentsByID.add(docInfo)
			setDocumentInfo(&s.documents, docInfo)
		}

		return docInfo.entity, nil
	}

	entity, err := s.entityToJson.ConvertToEntity(entityType, id, document)
	if err != nil {
		return nil, err
	}

	changeVector := jsonGetAsTextPointer(metadata, Constants_Documents_Metadata_CHANGE_VECTOR)
	if changeVector == nil {
		return nil, NewIllegalStateException("Document %s must have Change Vector", id)
	}

	if !noTracking {
		newDocumentInfo := NewDocumentInfo()
		newDocumentInfo.id = id
		newDocumentInfo.document = document
		newDocumentInfo.metadata = metadata
		newDocumentInfo.entity = entity
		newDocumentInfo.changeVector = changeVector

		s.documentsByID.add(newDocumentInfo)
		setDocumentInfo(&s.documents, newDocumentInfo)
	}

	return entity, nil
}

// DeleteEntity marks the specified entity for deletion. The entity will be deleted when SaveChanges is called.
func (s *InMemoryDocumentSessionOperations) DeleteEntity(entity interface{}) error {
	if entity == nil {
		return NewIllegalArgumentException("Entity cannot be null")
	}

	value := getDocumentInfoByEntity(s.documents, entity)
	if value == nil {
		return NewIllegalStateException("%#v is not associated with the session, cannot delete unknown entity instance", entity)
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
		return NewIllegalArgumentException("Id cannot be empty")
	}

	var changeVector *string
	documentInfo := s.documentsByID.getValue(id)
	if documentInfo != nil {
		newObj := EntityToJson_convertEntityToJson(documentInfo.entity, documentInfo)
		if documentInfo.entity != nil && s.EntityChanged(newObj, documentInfo, nil) {
			return NewIllegalStateException("Can't delete changed entity using identifier. Use delete(Class clazz, T entity) instead.")
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

// Store stores entity in the session. The entity will be saved when SaveChanges is called.
func (s *InMemoryDocumentSessionOperations) Store(entity interface{}) error {
	_, hasID := s.generateEntityIdOnTheClient.tryGetIdFromInstance(entity)
	concu := ConcurrencyCheck_AUTO
	if !hasID {
		concu = ConcurrencyCheck_FORCED
	}
	return s.storeInternal(entity, nil, "", concu)
}

// StoreWithID stores  entity in the session, explicitly specifying its Id. The entity will be saved when SaveChanges is called.
func (s *InMemoryDocumentSessionOperations) StoreWithID(entity interface{}, id string) error {
	return s.storeInternal(entity, nil, id, ConcurrencyCheck_AUTO)
}

// StoreWithChangeVectorAndID stores entity in the session, explicitly specifying its id and change vector. The entity will be saved when SaveChanges is called.
func (s *InMemoryDocumentSessionOperations) StoreWithChangeVectorAndID(entity interface{}, changeVector *string, id string) error {
	concurr := ConcurrencyCheck_DISABLED
	if changeVector != nil {
		concurr = ConcurrencyCheck_FORCED
	}

	return s.storeInternal(entity, changeVector, id, concurr)
}

// TODO: should this return an error?
func (s *InMemoryDocumentSessionOperations) RememberEntityForDocumentIdGeneration(entity interface{}) {
	err := NewNotImplementedException("You cannot set GenerateDocumentIdsOnStore to false without implementing RememberEntityForDocumentIdGeneration")
	must(err)
}

func (s *InMemoryDocumentSessionOperations) storeInternal(entity interface{}, changeVector *string, id string, forceConcurrencyCheck ConcurrencyCheckMode) error {
	if nil == entity {
		return NewIllegalArgumentException("Entity cannot be null")
	}

	value := getDocumentInfoByEntity(s.documents, entity)
	if value != nil {
		value.changeVector = firstNonNilString(changeVector, value.changeVector)
		value.concurrencyCheckMode = forceConcurrencyCheck
		return nil
	}

	if id == "" {
		if s.generateDocumentKeysOnStore {
			id = s.generateEntityIdOnTheClient.generateDocumentKeyForStorage(entity)
		} else {
			s.RememberEntityForDocumentIdGeneration(entity)
		}
	} else {
		// Store it back into the Id field so the client has access to it
		s.generateEntityIdOnTheClient.trySetIdentity(entity, id)
	}

	tmp := NewIdTypeAndName(id, CommandType_CLIENT_ANY_COMMAND, "")
	if _, ok := s.deferredCommandsMap[tmp]; ok {
		return NewIllegalStateException("Can't Store document, there is a deferred command registered for this document in the session. Document id: %s", id)
	}

	if s.deletedEntities.contains(entity) {
		return NewIllegalStateException("Can't Store object, it was already deleted in this session.  Document id: %s", id)
	}

	// we make the check here even if we just generated the ID
	// users can override the ID generation behavior, and we need
	// to detect if they generate duplicates.

	if err := s.assertNoNonUniqueInstance(entity, id); err != nil {
		return err
	}

	collectionName := s._requestExecutor.GetConventions().GetCollectionName(entity)
	metadata := ObjectNode{}
	if collectionName != "" {
		metadata[Constants_Documents_Metadata_COLLECTION] = collectionName
	}
	goType := s._requestExecutor.GetConventions().GetGoTypeName(entity)
	if goType != "" {
		metadata[Constants_Documents_Metadata_RAVEN_GO_TYPE] = goType
	}
	if id != "" {
		s._knownMissingIds = StringArrayRemoveNoCase(s._knownMissingIds, id)
	}

	s.StoreEntityInUnitOfWork(id, entity, changeVector, metadata, forceConcurrencyCheck)
	return nil
}

func (s *InMemoryDocumentSessionOperations) StoreEntityInUnitOfWork(id string, entity interface{}, changeVector *string, metadata ObjectNode, forceConcurrencyCheck ConcurrencyCheckMode) {
	s.deletedEntities.remove(entity)
	if id != "" {
		s._knownMissingIds = StringArrayRemoveNoCase(s._knownMissingIds, id)
	}
	documentInfo := NewDocumentInfo()
	documentInfo.id = id
	documentInfo.metadata = metadata
	documentInfo.changeVector = changeVector
	documentInfo.concurrencyCheckMode = forceConcurrencyCheck
	documentInfo.entity = entity
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

	return NewNonUniqueObjectException("Attempted to associate a different object with id '" + id + "'.")
}

func (s *InMemoryDocumentSessionOperations) PrepareForSaveChanges() (*SaveChangesData, error) {
	result := NewSaveChangesData(s)

	s.deferredCommands = nil
	s.deferredCommandsMap = make(map[IdTypeAndName]ICommandData)

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

func (s *InMemoryDocumentSessionOperations) UpdateMetadataModifications(documentInfo *DocumentInfo) bool {
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
			change := NewDocumentsChanges()
			change.setFieldNewValue("")
			change.setFieldOldValue("")
			change.setChange(DocumentsChanges_ChangeType_DOCUMENT_DELETED)

			docChanges = append(docChanges, change)
			changes[documentInfo.id] = docChanges
		} else {
			idType := NewIdTypeAndName(documentInfo.id, CommandType_CLIENT_ANY_COMMAND, "")
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

			beforeDeleteEventArgs := NewBeforeDeleteEventArgs(s, documentInfo.id, documentInfo.entity)
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

		document := EntityToJson_convertEntityToJson(entityKey, entityValue)

		if !s.EntityChanged(document, entityValue, nil) && !dirtyMetadata {
			continue
		}

		idType := NewIdTypeAndName(entityValue.id, CommandType_CLIENT_NOT_ATTACHMENT, "")
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
				document = EntityToJson_convertEntityToJson(entityKey, entityValue)
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
			if entityValue.concurrencyCheckMode != ConcurrencyCheck_DISABLED {
				// if the user didn't provide a change vector, we'll test for an empty one
				tmp := ""
				changeVector = firstNonNilString(entityValue.changeVector, &tmp)
			} else {
				changeVector = nil // TODO: redundant
			}
		} else if entityValue.concurrencyCheckMode == ConcurrencyCheck_FORCED {
			changeVector = entityValue.changeVector
		} else {
			changeVector = nil // TODO: redundant
		}
		cmdData := NewPutCommandDataWithJson(entityValue.id, changeVector, document)
		result.AddSessionCommandData(cmdData)
	}
	return nil
}

func (s *InMemoryDocumentSessionOperations) throwInvalidModifiedDocumentWithDeferredCommand(resultCommand ICommandData) error {
	err := NewIllegalStateException("Cannot perform save because document " + resultCommand.getId() + " has been modified by the session and is also taking part in deferred " + resultCommand.getType() + " command")
	return err
}

func (s *InMemoryDocumentSessionOperations) throwInvalidDeletedDocumentWithDeferredCommand(resultCommand ICommandData) error {
	err := NewIllegalStateException("Cannot perform save because document " + resultCommand.getId() + " has been deleted by the session and is also taking part in deferred " + resultCommand.getType() + " command")
	return err
}

func (s *InMemoryDocumentSessionOperations) EntityChanged(newObj ObjectNode, documentInfo *DocumentInfo, changes map[string][]*DocumentsChanges) bool {
	return JsonOperation_entityChanged(newObj, documentInfo, changes)
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
		for (Map.Entry<Object, DocumentInfo> entity : documentsByEntity.entrySet()) {
			ObjectNode document = entityToJson.convertEntityToJson(entity.getKey(), entity.getValue());
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

	document := EntityToJson_convertEntityToJson(entity, documentInfo)
	return s.EntityChanged(document, documentInfo, nil)
}

func (s *InMemoryDocumentSessionOperations) GetAllEntitiesChanges(changes map[string][]*DocumentsChanges) {
	for _, docInfo := range s.documentsByID.inner {
		s.UpdateMetadataModifications(docInfo)
		entity := docInfo.entity
		newObj := EntityToJson_convertEntityToJson(entity, docInfo)
		s.EntityChanged(newObj, docInfo, changes)
	}
}

// IgnoreChangesFor marks the entity as one that should be ignore for change tracking purposes,
// it still takes part in the session, but is ignored for SaveChanges.
func (s *InMemoryDocumentSessionOperations) IgnoreChangesFor(entity interface{}) {
	docInfo, _ := s.GetDocumentInfo(entity)
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
	idType := NewIdTypeAndName(command.getId(), command.getType(), command.GetName())
	s.deferredCommandsMap[idType] = command
	idType = NewIdTypeAndName(command.getId(), CommandType_CLIENT_ANY_COMMAND, "")
	s.deferredCommandsMap[idType] = command

	cmdType := command.getType()
	isAttachmentCmd := (cmdType == CommandType_ATTACHMENT_PUT) || (cmdType == CommandType_ATTACHMENT_DELETE)
	if !isAttachmentCmd {
		idType = NewIdTypeAndName(command.getId(), CommandType_CLIENT_NOT_ATTACHMENT, "")
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
	s._knownMissingIds = StringArrayRemoveNoCase(s._knownMissingIds, id)
}

// RegisterIncludes registers includes object
func (s *InMemoryDocumentSessionOperations) RegisterIncludes(includes ObjectNode) {
	if includes == nil {
		return
	}

	for _, fieldValue := range includes {
		// TODO: this needs to check if value inside is nil
		if fieldValue == nil {
			continue
		}
		json, ok := fieldValue.(ObjectNode)
		panicIf(!ok, "fieldValue of unsupported type %T", fieldValue)
		newDocumentInfo := DocumentInfo_getNewDocumentInfo(json)
		if JsonExtensions_tryGetConflict(newDocumentInfo.metadata) {
			continue
		}

		s.includedDocumentsByID[newDocumentInfo.id] = newDocumentInfo
	}
}

func (s *InMemoryDocumentSessionOperations) RegisterMissingIncludes(results ArrayNode, includes ObjectNode, includePaths []string) {
	if len(includePaths) == 0 {
		return
	}
	// TODO: ?? This is a no-op in Java
	/*
		for _, result := range results {
			for _, include := range includePaths {
				if include == Constants_Documents_Indexing_Fields_DOCUMENT_ID_FIELD_NAME {
					continue
				}
				// TODO: IncludesUtil.include() but it's a no-op in Java code
			}
		}
	*/
}

func (s *InMemoryDocumentSessionOperations) DeserializeFromTransformer2(result interface{}, id string, document ObjectNode) {
	s.entityToJson.ConvertToEntity2(result, id, document)
}

func (s *InMemoryDocumentSessionOperations) DeserializeFromTransformer(clazz reflect.Type, id string, document ObjectNode) (interface{}, error) {
	return s.entityToJson.ConvertToEntity(clazz, id, document)
}

func (s *InMemoryDocumentSessionOperations) checkIfIdAlreadyIncluded(ids []string, includes []string) bool {
	for _, id := range ids {
		if StringArrayContainsNoCase(s._knownMissingIds, id) {
			continue
		}

		// Check if document was already loaded, the check if we've received it through include
		documentInfo := s.documentsByID.getValue(id)
		if documentInfo == nil {
			documentInfo, _ = s.includedDocumentsByID[id]
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

func (s *InMemoryDocumentSessionOperations) refreshInternal(entity interface{}, cmd *GetDocumentsCommand, documentInfo *DocumentInfo) error {
	document := cmd.Result.Results[0]
	if document == nil {
		return NewIllegalStateException("Document '%s' no longer exists and was probably deleted", documentInfo.id)
	}

	value := document[Constants_Documents_Metadata_KEY]
	meta := value.(ObjectNode)
	documentInfo.metadata = meta

	if documentInfo.metadata != nil {
		changeVector := jsonGetAsTextPointer(meta, Constants_Documents_Metadata_CHANGE_VECTOR)
		documentInfo.changeVector = changeVector
	}
	documentInfo.document = document
	var err error
	documentInfo.entity, err = s.entityToJson.ConvertToEntity(reflect.TypeOf(entity), documentInfo.id, document)
	if err != nil {
		return err
	}

	err = BeanUtils_copyProperties(entity, documentInfo.entity)
	if err != nil {
		return NewRuntimeException("Unable to refresh entity: %s", err)
	}
	return nil
}

func isPtrStruct(t reflect.Type) bool {
	if t.Kind() != reflect.Ptr {
		return false
	}
	return t.Elem() != nil && t.Elem().Kind() == reflect.Struct
}

func isMapStringToPtrStruct(t reflect.Type) bool {
	if t.Kind() != reflect.Map {
		return false
	}

	if t.Key().Kind() != reflect.String {
		return false
	}

	return isPtrStruct(t.Elem())
}

func (s *InMemoryDocumentSessionOperations) getOperationResult(clazz reflect.Type, result interface{}) (interface{}, error) {
	// result is map[string]interface{}, where interface{} is ptr-to-struct
	// clazz describes what type the caller wanted ([]*struct, map[string]*struct or *struct)
	// this converts result to type of clazz

	//fmt.Printf("getOperationResult(): clazz is '%s', result is '%v' of type '%T'\n", clazz, result, result)

	if result == nil {
		res := Defaults_defaultValue(clazz)
		//fmt.Printf("getOperationResult(): returning result '%#v' of type 'T%s'\n", res, res)
		return res, nil
	}

	resultType := reflect.ValueOf(result).Type()
	//fmt.Printf("getOperationResult: result type: %T, resultType: %s, clazz: %s, result: %v\n", result, resultType, clazz, result)
	if resultType == clazz {
		return result, nil
	}

	if clazz.Kind() == reflect.Slice {
		res := reflect.MakeSlice(clazz, 0, 0)
		arr, ok := result.([]interface{})
		if !ok {
			return nil, fmt.Errorf("result is '%T' and not []interface{}", result)
		}
		// Note: this might be different than Java due to randomized map tranversal in Go
		for _, el := range arr {
			// TODO: don't panic if el type != clazz.Elem() type
			v := reflect.ValueOf(el)
			res = reflect.Append(res, v)
		}
		return res.Interface(), nil
	}

	resultMap, ok := result.(map[string]interface{})
	if !ok {
		return nil, NewIllegalStateException("result must be of type map[string]interface{}, is: %T", result)
	}

	if isMapStringToPtrStruct(clazz) {
		mapValueType := clazz.Elem()
		mapType := reflect.MapOf(stringType, mapValueType)
		m := reflect.MakeMap(mapType)

		if len(resultMap) == 0 {
			res := m.Interface()
			//fmt.Printf("getOperationResult(): returning result '%#v' of type 'T%s'\n", res, res)
			return res, nil
		}

		// Note: if the value in the database for a give id doesn't exist,
		// result has the key with nil value and we preserve that to match Java
		for k, v := range resultMap {
			//fmt.Printf("k: '%s', vt: '%T', v: '%v'\n", k, v, v)
			key := reflect.ValueOf(k)
			res := reflect.ValueOf(v)
			m.SetMapIndex(key, res)
		}

		return m.Interface(), nil
	}

	if !isPtrStruct(clazz) {
		return nil, NewIllegalStateException("expected clazz to be of type ptr-to-struct, is: %T", clazz)
	}

	if len(resultMap) == 0 {
		return nil, nil
	}

	for _, v := range resultMap {
		//fmt.Printf("getOperationResult: v: '%v' v type: '%T', clazz: %s\n", v, v, clazz)
		// TODO: assert that type of v is the same as clazz?
		return v, nil
	}

	//fmt.Printf("getOperationResult(): returning nil, clazz is '%s', result is '%v' of type '%T'\n", clazz, result, result)
	return nil, nil
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
		//throw new IllegalStateException("Parameters indexName and collectionName are mutually exclusive. Please specify only one of them.");
		panic("Parameters indexName and collectionName are mutually exclusive. Please specify only one of them.")
	}

	if !isIndex && !isCollection {
		collectionName = conventions.GetCollectionName(clazz)
	}

	return indexName, collectionName
}

type SaveChangesData struct {
	deferredCommands    []ICommandData
	deferredCommandsMap map[IdTypeAndName]ICommandData
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

func (d *SaveChangesData) GetDeferredCommandsMap() map[IdTypeAndName]ICommandData {
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

func copyDeferredCommandsMap(in map[IdTypeAndName]ICommandData) map[IdTypeAndName]ICommandData {
	res := map[IdTypeAndName]ICommandData{}
	for k, v := range in {
		res[k] = v
	}
	return res
}
