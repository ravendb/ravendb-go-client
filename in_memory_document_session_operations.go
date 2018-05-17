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

// SessionInfo describes a session
type SessionInfo struct {
	SessionID int
}

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
	// Note: skipping unused saveChangeOptions
	// Note: skipping unused isDisposed
	ID string

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

	deferredCommands []*CommandData

	// Note: using value type so that lookups are based on value
	deferredCommandsMap map[IdTypeAndName]*CommandData
}

// NewInMemoryDocumentSessionOperations creates new InMemoryDocumentSessionOperations
func NewInMemoryDocumentSessionOperations(dbName string, store *DocumentStore, re *RequestExecutor) *InMemoryDocumentSessionOperations {
	clientSessionID := newClientSessionID()
	res := InMemoryDocumentSessionOperations{
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
	return &res
}

// GetNumberOfEntitiesInUnitOfWork returns number of entinties
func (s *InMemoryDocumentSessionOperations) GetNumberOfEntitiesInUnitOfWork() int {
	return len(s.documentsByEntity)
}

func (s *InMemoryDocumentSessionOperations) getConventions() *DocumentConventions {
	return s.RequestExecutor.Conventions
}

// GetMetadataFor returns metadata for a given instance
func (s *InMemoryDocumentSessionOperations) GetMetadataFor(instance interface{}) IMetadataDictionary {
	panicIf(true, "NYI")
	return nil
}

// GetChangeVectorFor returns metadata for a given instance
// empty string means there is not change vector
func (s *InMemoryDocumentSessionOperations) GetChangeVectorFor(instance interface{}) string {
	panicIf(true, "NYI")
	return ""
}

// GetLastModifiedFor retursn last modified time for a given instance
func (s *InMemoryDocumentSessionOperations) GetLastModifiedFor(instance interface{}) (time.Time, bool) {
	panicIf(true, "NYI")
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

	/* TODO: port this
	Reference<String> idRef = new Reference<>();
	if (!generateEntityIdOnTheClient.tryGetIdFromInstance(instance, idRef)) {
		throw new IllegalStateException("Could not find the document id for " + instance);
	}

	if err := s.assertNoNonUniqueInstance(instance, id); err != nil {
		return nil, err
	}
	*/

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
func (s *InMemoryDocumentSessionOperations) IncrementRequetsCount() error {
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
			docInfo.entity = convertToEntity(entityType, id, document)
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
			docInfo.entity = convertToEntity(entityType, id, document)
		}

		if !noTracking {
			delete(s.includedDocumentsById, id)
			s.documentsById.add(docInfo)
			s.documentsByEntity[docInfo.entity] = docInfo
		}

		return docInfo.entity, nil
	}

	entity := convertToEntity(entityType, id, document)

	changeVector := jsonGetAsText(metadata, MetadataCHANGE_VECTOR)
	if changeVector == "" {
		return nil, NewIllegalStateError("Document %s must have Change Vector", id)
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
		return NewIllegalArgumentError("Entity cannot be null")
	}

	value := s.documentsByEntity[entity]
	if value == nil {
		return NewIllegalStateError("%#v is not associated with the session, cannot delete unknown entity instance", entity)
	}

	s.deletedEntities[entity] = struct{}{}
	delete(s.includedDocumentsById, value.getId())
	s._knownMissingIds[value.getId()] = struct{}{}
	return nil
}

// Marks the specified entity for deletion. The entity will be deleted when IDocumentSession.SaveChanges is called.
// WARNING: This method will not call beforeDelete listener!
func (s *InMemoryDocumentSessionOperations) Delete(id string) error {
	return s.DeleteWithChangeVector(id, "")
}

func (s *InMemoryDocumentSessionOperations) DeleteWithChangeVector(id string, expectedChangeVector string) error {
	if id == "" {
		return NewIllegalArgumentError("Id cannot be empty")
	}

	changeVector := ""
	documentInfo := s.documentsById.getValue(id)
	if documentInfo != nil {
		newObj := convertEntityToJson(documentInfo.getEntity(), documentInfo)
		if documentInfo.getEntity() != nil && s.entityChanged(newObj, documentInfo, nil) {
			return NewIllegalStateError("Can't delete changed entity using identifier. Use delete(Class clazz, T entity) instead.")
		}

		if documentInfo.getEntity() != nil {
			delete(s.documentsByEntity, documentInfo.getEntity())
		}

		s.documentsById.remove(id)
		changeVector = documentInfo.getChangeVector()
	}

	s._knownMissingIds[id] = struct{}{}
	if !s.useOptimisticConcurrency {
		changeVector = ""
	}
	// TODO: remove
	fmt.Printf("%s\n", changeVector)
	//defer(new DeleteCommandData(id, ObjectUtils.firstNonNull(expectedChangeVector, changeVector)));
	return nil
}

// Stores the specified entity in the session. The entity will be saved when SaveChanges is called.
func (s *InMemoryDocumentSessionOperations) StoreEntity(entity Object) error {
	panicIf(true, "NYI")
	return nil
	// TODO: implememnt
	//_, hasId := tryGetIdFromInstance(entity);
	//s.storeInternal(entity, null, null, !hasId ? ConcurrencyCheckMode.FORCED : ConcurrencyCheckMode.AUTO);
}

/// Stores the specified entity in the session, explicitly specifying its Id. The entity will be saved when SaveChanges is called.
func (s *InMemoryDocumentSessionOperations) StoreEntityWithID(entity Object, id String) error {
	return s.storeInternal(entity, "", id, ConcurrencyCheckAuto)
}

// Stores the specified entity in the session, explicitly specifying its Id. The entity will be saved when SaveChanges is called.
func (s *InMemoryDocumentSessionOperations) Store(entity Object, changeVector String, id String) error {
	concurr := ConcurrencyCheckDisabled
	if changeVector != "" {
		concurr = ConcurrencyCheckForced
	}

	return s.storeInternal(entity, changeVector, id, concurr)
}

func (s *InMemoryDocumentSessionOperations) storeInternal(entity Object, changeVector String, id String, forceConcurrencyCheck ConcurrencyCheckMode) error {
	if nil == entity {
		return NewIllegalArgumentError("Entity cannot be null")
	}

	value := s.documentsByEntity[entity]
	if value != nil {
		value.setChangeVector(firstNonEmptyString(changeVector, value.getChangeVector()))
		value.setConcurrencyCheckMode(forceConcurrencyCheck)
		return nil
	}

	if id == "" {
		if s.generateDocumentKeysOnStore {
			// TODO:: fix me
			//id = generateEntityIdOnTheClient.generateDocumentKeyForStorage(entity);
		} else {
			//TODO: fix me
			//rememberEntityForDocumentIdGeneration(entity);
		}
	} else {
		// Store it back into the Id field so the client has access to it
		// TODO: fix me
		//generateEntityIdOnTheClient.trySetIdentity(entity, id);
	}

	tmp := NewIdTypeAndName(id, CommandType_CLIENT_ANY_COMMAND, "")
	if _, ok := s.deferredCommandsMap[tmp]; ok {
		return NewIllegalStateError("Can't store document, there is a deferred command registered for this document in the session. Document id: %s", id)
	}

	if _, ok := s.deletedEntities[entity]; ok {
		return NewIllegalStateError("Can't store object, it was already deleted in this session.  Document id: %s", id)
	}

	// we make the check here even if we just generated the ID
	// users can override the ID generation behavior, and we need
	// to detect if they generate duplicates.

	if err := s.assertNoNonUniqueInstance(entity, id); err != nil {
		return err
	}

	// collectionName := s.RequestExecutor.getConventions().getCollectionName(entity)

	/*
		ObjectMapper mapper = JsonExtensions.getDefaultMapper();
		ObjectNode metadata = mapper.createObjectNode();

		if (collectionName != null) {
			metadata.set(Constants.Documents.Metadata.COLLECTION, mapper.convertValue(collectionName, JsonNode.class));
		}

		String javaType = _requestExecutor.getConventions().getJavaClassName(entity.getClass());
		if (javaType != null) {
			metadata.set(Constants.Documents.Metadata.RAVEN_JAVA_TYPE, mapper.convertValue(javaType, TextNode.class));
		}

		if (id != null) {
			_knownMissingIds.remove(id);
		}

		storeEntityInUnitOfWork(id, entity, changeVector, metadata, forceConcurrencyCheck);
	*/
	return nil
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

	return NewNonUniqueObjectError("Attempted to associate a different object with id '" + id + "'.")
}

func (s *InMemoryDocumentSessionOperations) entityChanged(newObj ObjectNode, documentInfo *DocumentInfo, changes map[string][]*DocumentsChanges) bool {
	//return JsonOperation.entityChanged(newObj, documentInfo, changes);
	return false
}

func (s *InMemoryDocumentSessionOperations) deserializeFromTransformer(clazz reflect.Type, id string, document ObjectNode) interface{} {
	panicIf(true, "NYI")
	//return entityToJson.convertToEntity(clazz, id, document);
	return nil
}
