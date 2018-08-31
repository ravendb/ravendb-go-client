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
	_clientSessionID   int
	deletedEntities    *ObjectSet
	_requestExecutor   *RequestExecutor
	_operationExecutor *OperationExecutor
	// Note: pendingLazyOperations and onEvaluateLazy not used
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

	documentsById *DocumentsById

	// Translate between an ID and its associated entity
	// TODO: ignore case for keys
	includedDocumentsById map[string]*DocumentInfo

	// hold the data required to manage the data for RavenDB's Unit of Work
	// TODO: this uses value semantics, so it works as expected for
	// pointers to structs, but 2 different structs with the same content
	// will match the same object. Should I disallow storing non-pointer structs?
	// convert non-pointer structs to structs?
	documentsByEntity map[interface{}]*DocumentInfo

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
		documentsById:                 NewDocumentsById(),
		includedDocumentsById:         map[string]*DocumentInfo{},
		documentsByEntity:             map[interface{}]*DocumentInfo{},
		_documentStore:                store,
		databaseName:                  dbName,
		maxNumberOfRequestsPerSession: re.conventions._maxNumberOfRequestsPerSession,
		useOptimisticConcurrency:      re.conventions.UseOptimisticConcurrency,
		deferredCommandsMap:           make(map[IdTypeAndName]ICommandData),
	}

	genIDFunc := func(entity Object) string {
		return res.GenerateId(entity)
	}
	res.generateEntityIdOnTheClient = NewGenerateEntityIdOnTheClient(re.conventions, genIDFunc)
	res.entityToJson = NewEntityToJson(res)
	return res
}

func (s *InMemoryDocumentSessionOperations) GetDeferredCommandsCount() int {
	return len(s.deferredCommands)
}

func (s *InMemoryDocumentSessionOperations) AddBeforeStoreListener(handler func(interface{}, *BeforeStoreEventArgs)) {
	s.onBeforeStore = append(s.onBeforeStore, handler)

}
func (s *InMemoryDocumentSessionOperations) RemoveBeforeStoreListener(handler func(interface{}, *BeforeStoreEventArgs)) {
	panic("NYI")
	//this.onBeforeStore.remove(handler);
}

func (s *InMemoryDocumentSessionOperations) AddAfterSaveChangesListener(handler func(interface{}, *AfterSaveChangesEventArgs)) {
	s.onAfterSaveChanges = append(s.onAfterSaveChanges, handler)
}

func (s *InMemoryDocumentSessionOperations) RemoveAfterSaveChangesListener(handler func(interface{}, *AfterSaveChangesEventArgs)) {
	panic("NYI")
	//this.onAfterSaveChanges.remove(handler);
}

func (s *InMemoryDocumentSessionOperations) AddBeforeDeleteListener(handler func(interface{}, *BeforeDeleteEventArgs)) {
	s.onBeforeDelete = append(s.onBeforeDelete, handler)
}

func (s *InMemoryDocumentSessionOperations) RemoveBeforeDeleteListener(handler func(interface{}, *BeforeDeleteEventArgs)) {
	panic("NYI")
	//this.onBeforeDelete.remove(handler);
}

func (s *InMemoryDocumentSessionOperations) AddBeforeQueryListener(handler func(interface{}, *BeforeQueryEventArgs)) {
	s.onBeforeQuery = append(s.onBeforeQuery, handler)
}

func (s *InMemoryDocumentSessionOperations) RemoveBeforeQueryListener(handler func(interface{}, *BeforeQueryEventArgs)) {
	panic("NYI")
	//this.onBeforeQuery.remove(handler);
}

func (s *InMemoryDocumentSessionOperations) GetGenerateEntityIdOnTheClient() *GenerateEntityIdOnTheClient {
	return s.generateEntityIdOnTheClient
}

func (s *InMemoryDocumentSessionOperations) GetEntityToJson() *EntityToJson {
	return s.entityToJson
}

// GetNumberOfEntitiesInUnitOfWork returns number of entinties
func (s *InMemoryDocumentSessionOperations) GetNumberOfEntitiesInUnitOfWork() int {
	return len(s.documentsByEntity)
}

func (s *InMemoryDocumentSessionOperations) GetConventions() *DocumentConventions {
	return s._requestExecutor.conventions
}

func (s *InMemoryDocumentSessionOperations) GetDatabaseName() string {
	return s.databaseName
}

func (s *InMemoryDocumentSessionOperations) GenerateId(entity Object) string {
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
	documentInfo.setMetadataInstance(metadata)
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

// GetDocumentInfo returns DocumentInfo for a given instance
// Returns nil if not found
func (s *InMemoryDocumentSessionOperations) GetDocumentInfo(instance interface{}) (*DocumentInfo, error) {
	documentInfo := s.documentsByEntity[instance]
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
	documentInfo := s.documentsById.getValue(id)
	if documentInfo != nil && documentInfo.document != nil {
		// is loaded
		return true
	}
	if s.IsDeleted(id) {
		return true
	}
	_, found := s.includedDocumentsById[id]
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
	value := s.documentsByEntity[instance]
	if value == nil {
		return ""
	}
	return value.id
}

// IncrementRequetsCount increments requests count
func (s *InMemoryDocumentSessionOperations) IncrementRequestCount() error {
	s.numberOfRequests++
	if s.numberOfRequests > s.maxNumberOfRequestsPerSession {
		return NewIllegalStateException("exceeded max number of reqeusts per session of %d", s.maxNumberOfRequestsPerSession)
	}
	return nil
}

// TrackEntityInDocumentInfo tracks entity in DocumentInfo
func (s *InMemoryDocumentSessionOperations) TrackEntityInDocumentInfoOld(clazz reflect.Type, documentFound *DocumentInfo) (interface{}, error) {
	return s.TrackEntityOld(clazz, documentFound.id, documentFound.document, documentFound.metadata, false)
}

func (s *InMemoryDocumentSessionOperations) TrackEntityInDocumentInfo(result interface{}, documentFound *DocumentInfo) error {
	return s.TrackEntity(result, documentFound.id, documentFound.document, documentFound.metadata, false)
}

func (s *InMemoryDocumentSessionOperations) TrackEntity(result interface{}, id string, document ObjectNode, metadata ObjectNode, noTracking bool) error {
	if id == "" {
		s.DeserializeFromTransformer2(result, "", document)
		return nil
	}

	docInfo := s.documentsByEntity[id]
	if docInfo != nil {
		// the local instance may have been changed, we adhere to the current Unit of Work
		// instance, and return that, ignoring anything new.

		if docInfo.entity == nil {
			s.entityToJson.ConvertToEntity2(docInfo.entity, id, document)
		}

		if !noTracking {
			delete(s.includedDocumentsById, id)
			s.documentsByEntity[docInfo.entity] = docInfo
		}
		setInterfaceToValue(result, docInfo.entity)
		return nil
	}

	docInfo = s.includedDocumentsById[id]
	if docInfo != nil {
		noSet := true
		if docInfo.entity == nil {
			s.entityToJson.ConvertToEntity2(result, id, document)
			docInfo.setEntity(result)
			noSet = false
		}

		if !noTracking {
			delete(s.includedDocumentsById, id)
			s.documentsById.add(docInfo)
			s.documentsByEntity[docInfo.entity] = docInfo
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
		newDocumentInfo.setDocument(document)
		newDocumentInfo.setMetadata(metadata)
		newDocumentInfo.setEntity(result)
		newDocumentInfo.setChangeVector(changeVector)

		s.documentsById.add(newDocumentInfo)
		s.documentsByEntity[result] = newDocumentInfo
	}

	return nil
}

// TrackEntity tracks entity
func (s *InMemoryDocumentSessionOperations) TrackEntityOld(entityType reflect.Type, id string, document ObjectNode, metadata ObjectNode, noTracking bool) (interface{}, error) {
	if id == "" {
		return s.DeserializeFromTransformer(entityType, "", document), nil
	}

	docInfo := s.documentsByEntity[id]
	if docInfo != nil {
		// the local instance may have been changed, we adhere to the current Unit of Work
		// instance, and return that, ignoring anything new.

		if docInfo.entity == nil {
			docInfo.entity = s.entityToJson.ConvertToEntity(entityType, id, document)
		}

		if !noTracking {
			delete(s.includedDocumentsById, id)
			s.documentsByEntity[docInfo.entity] = docInfo
		}
		return docInfo.entity, nil
	}

	docInfo = s.includedDocumentsById[id]
	if docInfo != nil {
		if docInfo.entity == nil {
			docInfo.entity = s.entityToJson.ConvertToEntity(entityType, id, document)
		}

		if !noTracking {
			delete(s.includedDocumentsById, id)
			s.documentsById.add(docInfo)
			s.documentsByEntity[docInfo.entity] = docInfo
		}

		return docInfo.entity, nil
	}

	entity := s.entityToJson.ConvertToEntity(entityType, id, document)

	changeVector := jsonGetAsTextPointer(metadata, Constants_Documents_Metadata_CHANGE_VECTOR)
	if changeVector == nil {
		return nil, NewIllegalStateException("Document %s must have Change Vector", id)
	}

	if !noTracking {
		newDocumentInfo := NewDocumentInfo()
		newDocumentInfo.id = id
		newDocumentInfo.setDocument(document)
		newDocumentInfo.setMetadata(metadata)
		newDocumentInfo.setEntity(entity)
		newDocumentInfo.setChangeVector(changeVector)

		s.documentsById.add(newDocumentInfo)
		s.documentsByEntity[entity] = newDocumentInfo
	}

	return entity, nil
}

// Marks the specified entity for deletion. The entity will be deleted when SaveChanges is called.
func (s *InMemoryDocumentSessionOperations) DeleteEntity(entity interface{}) error {
	if entity == nil {
		return NewIllegalArgumentException("Entity cannot be null")
	}

	value := s.documentsByEntity[entity]
	if value == nil {
		return NewIllegalStateException("%#v is not associated with the session, cannot delete unknown entity instance", entity)
	}

	s.deletedEntities.add(entity)
	delete(s.includedDocumentsById, value.id)
	s._knownMissingIds = append(s._knownMissingIds, value.id)
	return nil
}

// Marks the specified entity for deletion. The entity will be deleted when IDocumentSession.SaveChanges is called.
// WARNING: This method will not call beforeDelete listener!
func (s *InMemoryDocumentSessionOperations) Delete(id string) error {
	return s.DeleteWithChangeVector(id, nil)
}

func (s *InMemoryDocumentSessionOperations) DeleteWithChangeVector(id string, expectedChangeVector *string) error {
	if id == "" {
		return NewIllegalArgumentException("Id cannot be empty")
	}

	var changeVector *string
	documentInfo := s.documentsById.getValue(id)
	if documentInfo != nil {
		newObj := EntityToJson_convertEntityToJson(documentInfo.entity, documentInfo)
		if documentInfo.entity != nil && s.EntityChanged(newObj, documentInfo, nil) {
			return NewIllegalStateException("Can't delete changed entity using identifier. Use delete(Class clazz, T entity) instead.")
		}

		if documentInfo.entity != nil {
			delete(s.documentsByEntity, documentInfo.entity)
		}

		s.documentsById.remove(id)
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
func (s *InMemoryDocumentSessionOperations) Store(entity Object) error {
	_, hasID := s.generateEntityIdOnTheClient.tryGetIdFromInstance(entity)
	concu := ConcurrencyCheck_AUTO
	if !hasID {
		concu = ConcurrencyCheck_FORCED
	}
	return s.storeInternal(entity, nil, "", concu)
}

// StoreWithID stores  entity in the session, explicitly specifying its Id. The entity will be saved when SaveChanges is called.
func (s *InMemoryDocumentSessionOperations) StoreWithID(entity Object, id string) error {
	return s.storeInternal(entity, nil, id, ConcurrencyCheck_AUTO)
}

// StoreWithChangeVectorAndID stores entity in the session, explicitly specifying its id and change vector. The entity will be saved when SaveChanges is called.
func (s *InMemoryDocumentSessionOperations) StoreWithChangeVectorAndID(entity Object, changeVector *string, id string) error {
	concurr := ConcurrencyCheck_DISABLED
	if changeVector != nil {
		concurr = ConcurrencyCheck_FORCED
	}

	return s.storeInternal(entity, changeVector, id, concurr)
}

// TODO: should this return an error?
func (s *InMemoryDocumentSessionOperations) RememberEntityForDocumentIdGeneration(entity Object) {
	err := NewNotImplementedException("You cannot set GenerateDocumentIdsOnStore to false without implementing RememberEntityForDocumentIdGeneration")
	must(err)
}

func (s *InMemoryDocumentSessionOperations) storeInternal(entity Object, changeVector *string, id string, forceConcurrencyCheck ConcurrencyCheckMode) error {
	if nil == entity {
		return NewIllegalArgumentException("Entity cannot be null")
	}

	value := s.documentsByEntity[entity]
	if value != nil {
		value.setChangeVector(firstNonNilString(changeVector, value.changeVector))
		value.setConcurrencyCheckMode(forceConcurrencyCheck)
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

func (s *InMemoryDocumentSessionOperations) StoreEntityInUnitOfWork(id string, entity Object, changeVector *string, metadata ObjectNode, forceConcurrencyCheck ConcurrencyCheckMode) {
	s.deletedEntities.remove(entity)
	if id != "" {
		s._knownMissingIds = StringArrayRemoveNoCase(s._knownMissingIds, id)
	}
	documentInfo := NewDocumentInfo()
	documentInfo.setId(id)
	documentInfo.setMetadata(metadata)
	documentInfo.setChangeVector(changeVector)
	documentInfo.setConcurrencyCheckMode(forceConcurrencyCheck)
	documentInfo.setEntity(entity)
	documentInfo.setNewDocument(true)
	documentInfo.setDocument(nil)

	s.documentsByEntity[entity] = documentInfo
	if id != "" {
		s.documentsById.add(documentInfo)
	}
}

func (s *InMemoryDocumentSessionOperations) assertNoNonUniqueInstance(entity Object, id string) error {
	nLastChar := len(id) - 1
	if len(id) == 0 || id[nLastChar] == '|' || id[nLastChar] == '/' {
		return nil
	}
	info := s.documentsById.getValue(id)
	if info == nil || info.entity == entity {
		return nil
	}

	return NewNonUniqueObjectException("Attempted to associate a different object with id '" + id + "'.")
}

func (s *InMemoryDocumentSessionOperations) PrepareForSaveChanges() (*SaveChangesData, error) {
	result := NewSaveChangesData(s)

	s.deferredCommands = nil
	s.deferredCommandsMap = make(map[IdTypeAndName]ICommandData)

	err := s.PrepareForEntitiesDeletion(result, nil)
	if err != nil {
		return nil, err
	}
	err = s.PrepareForEntitiesPuts(result)
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

func (s *InMemoryDocumentSessionOperations) PrepareForEntitiesDeletion(result *SaveChangesData, changes map[string][]*DocumentsChanges) error {
	for deletedEntity := range s.deletedEntities.items {
		documentInfo := s.documentsByEntity[deletedEntity]
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
			documentInfo = s.documentsById.getValue(documentInfo.id)

			if documentInfo != nil {
				changeVector = documentInfo.changeVector

				if documentInfo.entity != nil {
					delete(s.documentsByEntity, documentInfo.entity)
					result.AddEntity(documentInfo.entity)
				}

				s.documentsById.remove(documentInfo.id)
			}

			if !s.useOptimisticConcurrency {
				changeVector = nil
			}

			beforeDeleteEventArgs := NewBeforeDeleteEventArgs(s, documentInfo.id, documentInfo.entity)
			for _, handler := range s.onBeforeDelete {
				handler(s, beforeDeleteEventArgs)
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

func (s *InMemoryDocumentSessionOperations) PrepareForEntitiesPuts(result *SaveChangesData) error {
	for entityKey, entityValue := range s.documentsByEntity {
		if entityValue.ignoreChanges {
			continue
		}

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
				handler(s, beforeStoreEventArgs)
			}
			if beforeStoreEventArgs.isMetadataAccessed() {
				s.UpdateMetadataModifications(entityValue)
			}
			if beforeStoreEventArgs.isMetadataAccessed() || s.EntityChanged(document, entityValue, nil) {
				document = EntityToJson_convertEntityToJson(entityKey, entityValue)
			}
		}

		entityValue.setNewDocument(false)
		result.AddEntity(entityKey)

		if entityValue.id != "" {
			s.documentsById.remove(entityValue.id)
		}

		entityValue.setDocument(document)

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
	err := s.PrepareForEntitiesDeletion(nil, changes)
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
func (s *InMemoryDocumentSessionOperations) HasChanged(entity Object) bool {
	documentInfo := s.documentsByEntity[entity]

	if documentInfo == nil {
		return false
	}

	document := EntityToJson_convertEntityToJson(entity, documentInfo)
	return s.EntityChanged(document, documentInfo, nil)
}

func (s *InMemoryDocumentSessionOperations) GetAllEntitiesChanges(changes map[string][]*DocumentsChanges) {
	for _, docInfo := range s.documentsById.inner {
		s.UpdateMetadataModifications(docInfo)
		entity := docInfo.entity
		newObj := EntityToJson_convertEntityToJson(entity, docInfo)
		s.EntityChanged(newObj, docInfo, changes)
	}
}

// Mark the entity as one that should be ignore for change tracking purposes,
// it still takes part in the session, but is ignored for SaveChanges.
func (s *InMemoryDocumentSessionOperations) IgnoreChangesFor(entity Object) {
	docInfo, _ := s.GetDocumentInfo(entity)
	docInfo.setIgnoreChanges(true)
}

// Evicts the specified entity from the session.
// Remove the entity from the delete queue and stops tracking changes for this entity.
func (s *InMemoryDocumentSessionOperations) Evict(entity Object) {
	documentInfo := s.documentsByEntity[entity]
	if documentInfo != nil {
		delete(s.documentsByEntity, entity)
		s.documentsById.remove(documentInfo.id)
	}

	s.deletedEntities.remove(entity)
}

func (s *InMemoryDocumentSessionOperations) Clear() {
	s.documentsByEntity = nil
	s.deletedEntities.clear()
	s.documentsById = nil
	s._knownMissingIds = nil
	s.includedDocumentsById = nil
}

// Defer commands to be executed on saveChanges()
func (s *InMemoryDocumentSessionOperations) Defer(command ICommandData) {
	a := []ICommandData{command}
	s.DeferMany(a)
}

// Defer commands to be executed on saveChanges()
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

/**
 * Performs application-defined tasks associated with freeing, releasing, or resetting unmanaged resources.
 */
func (s *InMemoryDocumentSessionOperations) Close() {
	s._close(true)
}

func (s *InMemoryDocumentSessionOperations) RegisterMissing(id string) {
	s._knownMissingIds = append(s._knownMissingIds, id)
}

func (s *InMemoryDocumentSessionOperations) UnregisterMissing(id string) {
	s._knownMissingIds = StringArrayRemoveNoCase(s._knownMissingIds, id)
}

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

		s.includedDocumentsById[newDocumentInfo.id] = newDocumentInfo
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

func (s *InMemoryDocumentSessionOperations) DeserializeFromTransformer(clazz reflect.Type, id string, document ObjectNode) interface{} {
	return s.entityToJson.ConvertToEntity(clazz, id, document)
}

func (s *InMemoryDocumentSessionOperations) checkIfIdAlreadyIncluded(ids []string, includes []string) bool {
	for _, id := range ids {
		if StringArrayContainsNoCase(s._knownMissingIds, id) {
			continue
		}

		// Check if document was already loaded, the check if we've received it through include
		documentInfo := s.documentsById.getValue(id)
		if documentInfo == nil {
			documentInfo, _ = s.includedDocumentsById[id]
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

func (s *InMemoryDocumentSessionOperations) refreshInternal(entity Object, cmd *GetDocumentsCommand, documentInfo *DocumentInfo) error {
	document := cmd.Result.GetResults()[0]
	if document == nil {
		return NewIllegalStateException("Document '%s' no longer exists and was probably deleted", documentInfo.id)
	}

	value := document[Constants_Documents_Metadata_KEY]
	meta := value.(ObjectNode)
	documentInfo.setMetadata(meta)

	if documentInfo.metadata != nil {
		changeVector := jsonGetAsTextPointer(meta, Constants_Documents_Metadata_CHANGE_VECTOR)
		documentInfo.setChangeVector(changeVector)
	}
	documentInfo.setDocument(document)
	documentInfo.setEntity(s.entityToJson.ConvertToEntity(GetTypeOf(entity), documentInfo.id, document))

	err := BeanUtils_copyProperties(entity, documentInfo.entity)
	if err != nil {
		return NewRuntimeException("Unable to refresh entity: %s", err)
	}
	return nil
}

//TODO: protected static <T> T getOperationResult(Class<T> clazz, Object result) {

func (s *InMemoryDocumentSessionOperations) OnAfterSaveChangesInvoke(afterSaveChangesEventArgs *AfterSaveChangesEventArgs) {
	for _, handler := range s.onAfterSaveChanges {
		handler(s, afterSaveChangesEventArgs)
	}
}

func (s *InMemoryDocumentSessionOperations) OnBeforeQueryInvoke(beforeQueryEventArgs *BeforeQueryEventArgs) {
	for _, handler := range s.onBeforeQuery {
		handler(s, beforeQueryEventArgs)
	}
}

func (s *InMemoryDocumentSessionOperations) processQueryParameters(clazz reflect.Type, indexName string, collectionName string, conventions *DocumentConventions) (string, string) {
	isIndex := StringUtils_isNotBlank(indexName)
	isCollection := StringUtils_isNotEmpty(collectionName)

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
	entities            []Object
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

func (d *SaveChangesData) GetEntities() []Object {
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

func (d *SaveChangesData) AddEntity(entity Object) {
	d.entities = append(d.entities, entity)
}

// TODO: make faster
func copyDeferredCommands(in []ICommandData) []ICommandData {
	res := []ICommandData{}
	for _, d := range in {
		res = append(res, d)
	}
	return res
}

func copyDeferredCommandsMap(in map[IdTypeAndName]ICommandData) map[IdTypeAndName]ICommandData {
	res := map[IdTypeAndName]ICommandData{}
	for k, v := range in {
		res[k] = v
	}
	return res
}
