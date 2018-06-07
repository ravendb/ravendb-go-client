package ravendb

import (
	"fmt"
	"reflect"
	"sync/atomic"
	"time"
)

const (
	// TODO: fix names
	MetadataCOLLECTION    = "@collection"
	MetadataPROJECTION    = "@projection"
	MetadataKEY           = "@metadata"
	MetadataID            = "@id"
	MetadataCONFLICT      = "@conflict"
	MetadataID_PROPERTY   = "Id"
	MetadataFLAGS         = "@flags"
	MetadataATTACHMENTS   = "@attachments"
	MetadataINDEX_SCORE   = "@index-score"
	MetadataLAST_MODIFIED = "@last-modified"
	MetadataRAVEN_GO_TYPE = "Raven-Go-Type"
	MetadataCHANGE_VECTOR = "@change-vector"
	MetadataEXPIRES       = "@expires"
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
	_clientSessionID int
	deletedEntities  map[interface{}]struct{}
	RequestExecutor  *RequestExecutor
	// TODO: OperationExecutor
	// Note: pendingLazyOperations and onEvaluateLazy not used
	generateDocumentKeysOnStore bool
	sessionInfo                 SessionInfo
	_saveChangesOptions         *BatchOptions

	// Note: skipping unused isDisposed
	id string

	/* TODO:
	   private final List<EventHandler<BeforeStoreEventArgs>> onBeforeStore = new ArrayList<>();
	   private final List<EventHandler<AfterSaveChangesEventArgs>> onAfterSaveChanges = new ArrayList<>();
	   private final List<EventHandler<BeforeDeleteEventArgs>> onBeforeDelete = new ArrayList<>();
	   private final List<EventHandler<BeforeQueryEventArgs>> onBeforeQuery = new ArrayList<>();
	*/

	// ids of entities that were deleted
	_knownMissingIds map[string]struct{}

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

	documentStore *DocumentStore

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
		deletedEntities:               map[interface{}]struct{}{},
		RequestExecutor:               re,
		generateDocumentKeysOnStore:   true,
		sessionInfo:                   SessionInfo{SessionID: clientSessionID},
		documentsById:                 NewDocumentsById(),
		includedDocumentsById:         map[string]*DocumentInfo{},
		documentsByEntity:             map[interface{}]*DocumentInfo{},
		documentStore:                 store,
		databaseName:                  dbName,
		maxNumberOfRequestsPerSession: re.Conventions.MaxNumberOfRequestsPerSession,
		useOptimisticConcurrency:      re.Conventions.UseOptimisticConcurrency,
	}
	genIDFunc := func(entity Object) String {
		return res.generateId(entity)
	}
	res.generateEntityIdOnTheClient = NewGenerateEntityIdOnTheClient(re.Conventions, genIDFunc)
	res.entityToJson = NewEntityToJson(res)
	return res
}

func (s *InMemoryDocumentSessionOperations) getGenerateEntityIdOnTheClient() *GenerateEntityIdOnTheClient {
	return s.generateEntityIdOnTheClient
}

func (s *InMemoryDocumentSessionOperations) getEntityToJson() *EntityToJson {
	return s.entityToJson
}

// GetNumberOfEntitiesInUnitOfWork returns number of entinties
func (s *InMemoryDocumentSessionOperations) GetNumberOfEntitiesInUnitOfWork() int {
	return len(s.documentsByEntity)
}

func (s *InMemoryDocumentSessionOperations) getConventions() *DocumentConventions {
	return s.RequestExecutor.Conventions
}

func (s *InMemoryDocumentSessionOperations) getDatabaseName() string {
	return s.databaseName
}

func (s *InMemoryDocumentSessionOperations) generateId(entity Object) String {
	return s.getConventions().generateDocumentId(s.getDatabaseName(), entity)
}

// GetMetadataFor gets the metadata for the specified entity.
func (s *InMemoryDocumentSessionOperations) GetMetadataFor(instance interface{}) (*IMetadataDictionary, error) {
	if instance == nil {
		return nil, NewIllegalArgumentException("Instance cannot be null")
	}

	documentInfo, err := s.getDocumentInfo(instance)
	if err != nil {
		return nil, err
	}
	if documentInfo.getMetadataInstance() != nil {
		return documentInfo.getMetadataInstance(), nil
	}

	metadataAsJson := documentInfo.getMetadata()
	metadata := NewMetadataAsDictionaryWithSource(metadataAsJson)
	documentInfo.setMetadataInstance(metadata)
	return metadata, nil
}

// GetChangeVectorFor returns metadata for a given instance
// empty string means there is not change vector
func (s *InMemoryDocumentSessionOperations) GetChangeVectorFor(instance interface{}) (string, error) {
	if instance == nil {
		return "", NewIllegalArgumentException("Instance cannot be null")
	}

	documentInfo, err := s.getDocumentInfo(instance)
	if err != nil {
		return "", err
	}
	changeVector := jsonGetAsText(documentInfo.getMetadata(), Constants_Documents_Metadata_CHANGE_VECTOR)
	return changeVector, nil
}

// GetLastModifiedFor retursn last modified time for a given instance
func (s *InMemoryDocumentSessionOperations) GetLastModifiedFor(instance interface{}) (time.Time, bool) {
	panicIf(true, "NYI")

	var res time.Time
	return res, false
}

// GetDocumentInfo returns DocumentInfo for a given instance
// Returns nil if not found
func (s *InMemoryDocumentSessionOperations) getDocumentInfo(instance interface{}) (*DocumentInfo, error) {
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
	_, ok := s._knownMissingIds[id]
	return ok
}

// GetDocumentID returns id of a given instance
func (s *InMemoryDocumentSessionOperations) GetDocumentID(instance interface{}) string {
	panicIf(true, "NYI")
	return ""
}

// IncrementRequetsCount increments requests count
func (s *InMemoryDocumentSessionOperations) IncrementRequestCount() error {
	s.numberOfRequests++
	if s.numberOfRequests > s.maxNumberOfRequestsPerSession {
		return fmt.Errorf("exceeded max number of reqeusts per session of %d", s.maxNumberOfRequestsPerSession)
	}
	return nil
}

// TrackEntityInDocumentInfo tracks entity in DocumentInfo
func (s *InMemoryDocumentSessionOperations) TrackEntityInDocumentInfo(clazz reflect.Type, documentFound *DocumentInfo) (interface{}, error) {
	return s.TrackEntity(clazz, documentFound.id, documentFound.document, documentFound.metadata, false)
}

// TrackEntity tracks entity
func (s *InMemoryDocumentSessionOperations) TrackEntity(entityType reflect.Type, id string, document ObjectNode, metadata ObjectNode, noTracking bool) (interface{}, error) {
	if id == "" {
		return s.deserializeFromTransformer(entityType, "", document), nil
	}

	docInfo := s.documentsByEntity[id]
	if docInfo != nil {
		// the local instance may have been changed, we adhere to the current Unit of Work
		// instance, and return that, ignoring anything new.

		if docInfo.entity == nil {
			docInfo.entity = s.entityToJson.convertToEntity(entityType, id, document)
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
			docInfo.entity = s.entityToJson.convertToEntity(entityType, id, document)
		}

		if !noTracking {
			delete(s.includedDocumentsById, id)
			s.documentsById.add(docInfo)
			s.documentsByEntity[docInfo.entity] = docInfo
		}

		return docInfo.entity, nil
	}

	entity := s.entityToJson.convertToEntity(entityType, id, document)

	changeVector := jsonGetAsTextPointer(metadata, MetadataCHANGE_VECTOR)
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

	s.deletedEntities[entity] = struct{}{}
	delete(s.includedDocumentsById, value.getId())
	s._knownMissingIds[value.getId()] = struct{}{}
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
		newObj := EntityToJson_convertEntityToJson(documentInfo.getEntity(), documentInfo)
		if documentInfo.getEntity() != nil && s.entityChanged(newObj, documentInfo, nil) {
			return NewIllegalStateException("Can't delete changed entity using identifier. Use delete(Class clazz, T entity) instead.")
		}

		if documentInfo.getEntity() != nil {
			delete(s.documentsByEntity, documentInfo.getEntity())
		}

		s.documentsById.remove(id)
		changeVector = documentInfo.getChangeVector()
	}

	s._knownMissingIds[id] = struct{}{}
	if !s.useOptimisticConcurrency {
		changeVector = nil
	}
	cmdData := NewDeleteCommandData(id, firstNonNilString(expectedChangeVector, changeVector))
	s.Defer(cmdData)
	return nil
}

// Stores the specified entity in the session. The entity will be saved when SaveChanges is called.
func (s *InMemoryDocumentSessionOperations) StoreEntity(entity Object) error {
	_, hasID := s.generateEntityIdOnTheClient.tryGetIdFromInstance(entity)
	concu := ConcurrencyCheck_AUTO
	if !hasID {
		concu = ConcurrencyCheck_FORCED
	}
	return s.storeInternal(entity, nil, "", concu)
}

/// Stores the specified entity in the session, explicitly specifying its Id. The entity will be saved when SaveChanges is called.
func (s *InMemoryDocumentSessionOperations) StoreEntityWithID(entity Object, id String) error {
	return s.storeInternal(entity, nil, id, ConcurrencyCheck_AUTO)
}

// Stores the specified entity in the session, explicitly specifying its Id. The entity will be saved when SaveChanges is called.
func (s *InMemoryDocumentSessionOperations) Store(entity Object, changeVector *string, id String) error {
	concurr := ConcurrencyCheck_DISABLED
	if changeVector != nil {
		concurr = ConcurrencyCheck_FORCED
	}

	return s.storeInternal(entity, changeVector, id, concurr)
}

// TODO: should this return an error?
func (s *InMemoryDocumentSessionOperations) rememberEntityForDocumentIdGeneration(entity Object) {
	err := NewNotImplementedException("You cannot set GenerateDocumentIdsOnStore to false without implementing RememberEntityForDocumentIdGeneration")
	must(err)
}

func (s *InMemoryDocumentSessionOperations) storeInternal(entity Object, changeVector *string, id String, forceConcurrencyCheck ConcurrencyCheckMode) error {
	if nil == entity {
		return NewIllegalArgumentException("Entity cannot be null")
	}

	value := s.documentsByEntity[entity]
	if value != nil {
		value.setChangeVector(firstNonNilString(changeVector, value.getChangeVector()))
		value.setConcurrencyCheckMode(forceConcurrencyCheck)
		return nil
	}

	if id == "" {
		if s.generateDocumentKeysOnStore {
			id = s.generateEntityIdOnTheClient.generateDocumentKeyForStorage(entity)
		} else {
			s.rememberEntityForDocumentIdGeneration(entity)
		}
	} else {
		// Store it back into the Id field so the client has access to it
		s.generateEntityIdOnTheClient.trySetIdentity(entity, id)
	}

	tmp := NewIdTypeAndName(id, CommandType_CLIENT_ANY_COMMAND, "")
	if _, ok := s.deferredCommandsMap[tmp]; ok {
		return NewIllegalStateException("Can't store document, there is a deferred command registered for this document in the session. Document id: %s", id)
	}

	if _, ok := s.deletedEntities[entity]; ok {
		return NewIllegalStateException("Can't store object, it was already deleted in this session.  Document id: %s", id)
	}

	// we make the check here even if we just generated the ID
	// users can override the ID generation behavior, and we need
	// to detect if they generate duplicates.

	if err := s.assertNoNonUniqueInstance(entity, id); err != nil {
		return err
	}

	collectionName := s.RequestExecutor.getConventions().getCollectionName(entity)
	metadata := ObjectNode{}
	if collectionName != "" {
		metadata[Constants_Documents_Metadata_COLLECTION] = collectionName
	}
	goType := s.RequestExecutor.getConventions().getGoTypeName(entity)
	if goType != "" {
		metadata[Constants_Documents_Metadata_RAVEN_GO_TYPE] = goType
	}
	if id != "" {
		delete(s._knownMissingIds, id)
	}

	s.storeEntityInUnitOfWork(id, entity, changeVector, metadata, forceConcurrencyCheck)
	return nil
}

func (s *InMemoryDocumentSessionOperations) storeEntityInUnitOfWork(id String, entity Object, changeVector *string, metadata ObjectNode, forceConcurrencyCheck ConcurrencyCheckMode) {
	delete(s.deletedEntities, entity)
	if id != "" {
		delete(s._knownMissingIds, id)
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

func (s *InMemoryDocumentSessionOperations) assertNoNonUniqueInstance(entity Object, id String) error {
	nLastChar := len(id) - 1
	if len(id) == 0 || id[nLastChar] == '|' || id[nLastChar] == '/' {
		return nil
	}
	info := s.documentsById.getValue(id)
	if info == nil || info.getEntity() == entity {
		return nil
	}

	return NewNonUniqueObjectException("Attempted to associate a different object with id '" + id + "'.")
}

func (s *InMemoryDocumentSessionOperations) prepareForSaveChanges() *SaveChangesData {
	result := NewSaveChangesData(s)

	s.deferredCommands = nil
	s.deferredCommandsMap = nil

	s.prepareForEntitiesDeletion(result, nil)
	s.prepareForEntitiesPuts(result)

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
	return result
}

func (s *InMemoryDocumentSessionOperations) updateMetadataModifications(documentInfo *DocumentInfo) bool {
	dirty := false
	metadataInstance := documentInfo.getMetadataInstance()
	metadata := documentInfo.getMetadata()
	if metadataInstance != nil {
		if metadataInstance.isDirty() {
			dirty = true
		}
		props := metadataInstance.keySet()
		for _, prop := range props {
			propValue, ok := metadataInstance.get(prop)
			if !ok {
				dirty = true
				continue
			}
			if d, ok := propValue.(*MetadataAsDictionary); ok {
				if d.isDirty() {
					dirty = true
				}
			}
			metadata[prop] = propValue
		}
	}
	return dirty
}

// TODO: return an error
func (s *InMemoryDocumentSessionOperations) prepareForEntitiesDeletion(result *SaveChangesData, changes map[string][]*DocumentsChanges) {
	for deletedEntity := range s.deletedEntities {
		documentInfo := s.documentsByEntity[deletedEntity]
		if documentInfo == nil {
			continue
		}
		if len(changes) > 0 {
			docChanges := []*DocumentsChanges{}
			change := NewDocumentsChanges()
			change.setFieldNewValue("")
			change.setFieldOldValue("")
			change.setChange(DocumentsChanges_ChangeType_DOCUMENT_DELETED)

			docChanges = append(docChanges, change)
			changes[documentInfo.getId()] = docChanges
		} else {
			idType := NewIdTypeAndName(documentInfo.getId(), CommandType_CLIENT_ANY_COMMAND, "")
			command := result.getDeferredCommandsMap()[idType]
			if command != nil {
				s.throwInvalidDeletedDocumentWithDeferredCommand(command)
			}

			var changeVector *string
			documentInfo = s.documentsById.getValue(documentInfo.getId())

			if documentInfo != nil {
				changeVector = documentInfo.getChangeVector()

				if documentInfo.getEntity() != nil {
					delete(s.documentsByEntity, documentInfo.getEntity())
					result.addEntity(documentInfo.getEntity())
				}

				s.documentsById.remove(documentInfo.getId())
			}

			if !s.useOptimisticConcurrency {
				changeVector = nil
			}

			// TODO:
			//BeforeDeleteEventArgs beforeDeleteEventArgs = new BeforeDeleteEventArgs(this, documentInfo.getId(), documentInfo.getEntity());
			//EventHelper.invoke(onBeforeDelete, this, beforeDeleteEventArgs);

			cmdData := NewDeleteCommandData(documentInfo.getId(), changeVector)
			result.addSessionCommandData(cmdData)
		}

		if len(changes) == 0 {
			s.deletedEntities = nil
		}
	}
}

// TODO: return an error
func (s *InMemoryDocumentSessionOperations) prepareForEntitiesPuts(result *SaveChangesData) {
	for entityKey, entityValue := range s.documentsByEntity {
		if entityValue.isIgnoreChanges() {
			continue
		}

		dirtyMetadata := s.updateMetadataModifications(entityValue)

		document := EntityToJson_convertEntityToJson(entityKey, entityValue)

		if !s.entityChanged(document, entityValue, nil) && !dirtyMetadata {
			continue
		}

		idType := NewIdTypeAndName(entityValue.getId(), CommandType_CLIENT_NOT_ATTACHMENT, "")
		command := result.deferredCommandsMap[idType]
		if command != nil {
			s.throwInvalidModifiedDocumentWithDeferredCommand(command)
		}

		/* TODO:
		List<EventHandler<BeforeStoreEventArgs>> onBeforeStore = this.onBeforeStore;
		if (onBeforeStore != null && !onBeforeStore.isEmpty()) {
			BeforeStoreEventArgs beforeStoreEventArgs = new BeforeStoreEventArgs(this, entity.getValue().getId(), entity.getKey());
			EventHelper.invoke(onBeforeStore, this, beforeStoreEventArgs);

			if (beforeStoreEventArgs.isMetadataAccessed()) {
				updateMetadataModifications(entity.getValue());
			}

			if (beforeStoreEventArgs.isMetadataAccessed() || entityChanged(document, entity.getValue(), null)) {
				document = entityToJson.convertEntityToJson(entity.getKey(), entity.getValue());
			}
		}
		*/

		entityValue.setNewDocument(false)
		result.addEntity(entityKey)

		if entityValue.getId() != "" {
			s.documentsById.remove(entityValue.getId())
		}

		entityValue.setDocument(document)

		var changeVector *string
		if s.useOptimisticConcurrency {
			if entityValue.getConcurrencyCheckMode() != ConcurrencyCheck_DISABLED {
				// if the user didn't provide a change vector, we'll test for an empty one
				tmp := ""
				changeVector = firstNonNilString(entityValue.getChangeVector(), &tmp)
			} else {
				changeVector = nil // TODO: redundant
			}
		} else if entityValue.getConcurrencyCheckMode() == ConcurrencyCheck_FORCED {
			changeVector = entityValue.getChangeVector()
		} else {
			changeVector = nil // TODO: redundant
		}
		cmdData := NewPutCommandDataWithJson(entityValue.getId(), changeVector, document)
		result.addSessionCommandData(cmdData)
	}
}

// TODO: should return an error and be propagated
func (s *InMemoryDocumentSessionOperations) throwInvalidModifiedDocumentWithDeferredCommand(resultCommand ICommandData) {
	err := fmt.Errorf("Cannot perform save because document " + resultCommand.getId() + " has been modified by the session and is also taking part in deferred " + resultCommand.getType() + " command")
	must(err)
}

// TODO: should return an error and be propagated
func (s *InMemoryDocumentSessionOperations) throwInvalidDeletedDocumentWithDeferredCommand(resultCommand ICommandData) {
	err := fmt.Errorf("Cannot perform save because document " + resultCommand.getId() + " has been deleted by the session and is also taking part in deferred " + resultCommand.getType() + " command")
	must(err)
}

func (s *InMemoryDocumentSessionOperations) entityChanged(newObj ObjectNode, documentInfo *DocumentInfo, changes map[string][]*DocumentsChanges) bool {
	return JsonOperation_entityChanged(newObj, documentInfo, changes)
}

func (s *InMemoryDocumentSessionOperations) deserializeFromTransformer(clazz reflect.Type, id string, document ObjectNode) interface{} {
	panicIf(true, "NYI")
	//return entityToJson.convertToEntity(clazz, id, document);
	return nil
}

func (s *InMemoryDocumentSessionOperations) WhatChanged() map[string][]*DocumentsChanges {
	changes := map[string][]*DocumentsChanges{}
	s.prepareForEntitiesDeletion(nil, changes)
	panicIf(true, "NYI")
	/*
		getAllEntitiesChanges(changes);
	*/
	return changes
}

// Gets a value indicating whether any of the entities tracked by the session has changes.
func (s *InMemoryDocumentSessionOperations) hasChanges() bool {
	panicIf(true, "NYI")
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
func (s *InMemoryDocumentSessionOperations) hasChanged(entity Object) bool {
	documentInfo := s.documentsByEntity[entity]

	if documentInfo == nil {
		return false
	}

	document := EntityToJson_convertEntityToJson(entity, documentInfo)
	return s.entityChanged(document, documentInfo, nil)
}

func (s *InMemoryDocumentSessionOperations) getAllEntitiesChanges(changes map[string][]*DocumentsChanges) {
	for _, pairValue := range s.documentsById.inner {
		s.updateMetadataModifications(pairValue)
		newObj := EntityToJson_convertEntityToJson(pairValue.getEntity(), pairValue)
		s.entityChanged(newObj, pairValue, changes)
	}
}

// Mark the entity as one that should be ignore for change tracking purposes,
// it still takes part in the session, but is ignored for SaveChanges.
func (s *InMemoryDocumentSessionOperations) ignoreChangesFor(entity Object) {
	docInfo, _ := s.getDocumentInfo(entity)
	docInfo.setIgnoreChanges(true)
}

// Evicts the specified entity from the session.
// Remove the entity from the delete queue and stops tracking changes for this entity.
func (s *InMemoryDocumentSessionOperations) evict(entity Object) {
	documentInfo := s.documentsByEntity[entity]
	if documentInfo != nil {
		delete(s.documentsByEntity, entity)
		s.documentsById.remove(documentInfo.getId())
	}

	delete(s.deletedEntities, entity)
}

func (s *InMemoryDocumentSessionOperations) Clear() {
	s.documentsByEntity = nil
	s.deletedEntities = nil
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
	idType := NewIdTypeAndName(command.getId(), command.getType(), command.getName())
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

func (s *InMemoryDocumentSessionOperations) RegisterMissing(id String) {
	s._knownMissingIds[id] = struct{}{}
}

func (s *InMemoryDocumentSessionOperations) UnregisterMissing(id String) {
	delete(s._knownMissingIds, id)
}

func (s *InMemoryDocumentSessionOperations) registerIncludes(includes ObjectNode) {
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
		if JsonExtensions_tryGetConflict(newDocumentInfo.getMetadata()) {
			continue
		}

		s.includedDocumentsById[newDocumentInfo.getId()] = newDocumentInfo
	}
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

func (d *SaveChangesData) getDeferredCommands() []ICommandData {
	return d.deferredCommands
}

func (d *SaveChangesData) getSessionCommands() []ICommandData {
	return d.sessionCommands
}

func (d *SaveChangesData) getEntities() []Object {
	return d.entities
}

func (d *SaveChangesData) getOptions() *BatchOptions {
	return d.options
}

func (d *SaveChangesData) getDeferredCommandsMap() map[IdTypeAndName]ICommandData {
	return d.deferredCommandsMap
}

func (d *SaveChangesData) addSessionCommandData(cmd ICommandData) {
	d.sessionCommands = append(d.sessionCommands, cmd)
}

func (d *SaveChangesData) addEntity(entity Object) {
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
