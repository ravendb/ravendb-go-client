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
	clientSessionIDCounter int32 = 1
)

func newClientSessionID() int {
	newID := atomic.AddInt32(&clientSessionIDCounter, 1)
	return int(newID)
}

// InMemoryDocumentSessionOperations represents database operations queued
// in memory
type InMemoryDocumentSessionOperations struct {
	clientSessionID  int
	deletedEntities  map[interface{}]struct{}
	RequestsExecutor *RequestsExecutor
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
	knownMissingIDs map[string]struct{}

	// Note: skipping unused externalState
	// Note: skipping unused getCurrentSessionNode

	documentsByID *DocumentsById

	// Translate between an ID and its associated entity
	// TODO: ignore case for keys
	includedDocumentsByID map[string]*DocumentInfo

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
}

// NewInMemoryDocumentSessionOperations creates new InMemoryDocumentSessionOperations
func NewInMemoryDocumentSessionOperations(dbName string, store *DocumentStore, re *RequestsExecutor) *InMemoryDocumentSessionOperations {
	clientSessionID := newClientSessionID()
	res := InMemoryDocumentSessionOperations{
		clientSessionID:               clientSessionID,
		deletedEntities:               map[interface{}]struct{}{},
		RequestsExecutor:              re,
		generateDocumentKeysOnStore:   true,
		sessionInfo:                   SessionInfo{SessionID: clientSessionID},
		documentsByID:                 NewDocumentsById(),
		includedDocumentsByID:         map[string]*DocumentInfo{},
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
	return s.RequestsExecutor.Conventions
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
	// TODO: id check, assertNoNonUniqueInstance()
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
	_, ok := s.knownMissingIDs[id]
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
			delete(s.includedDocumentsByID, id)
			s.documentsByEntity[docInfo.entity] = docInfo
		}
		return docInfo.entity, nil
	}

	docInfo = s.includedDocumentsByID[id]
	if docInfo != nil {
		if docInfo.entity == nil {
			docInfo.entity = convertToEntity(entityType, id, document)
		}

		if !noTracking {
			delete(s.includedDocumentsByID, id)
			s.documentsByID.add(docInfo)
			s.documentsByEntity[docInfo.entity] = docInfo
		}

		return docInfo.entity, nil
	}

	entity := convertToEntity(entityType, id, document)

	changeVector := jsonGetAsText(metadata, MetadataCHANGE_VECTOR)
	if changeVector == "" {
		return nil, NewIllegalStateError(fmt.Sprintf("Document %s must have Change Vector", id))
	}

	if !noTracking {
		newDocumentInfo := NewDocumentInfo()
		newDocumentInfo.id = id
		newDocumentInfo.setDocument(document)
		newDocumentInfo.setMetadata(metadata)
		newDocumentInfo.setEntity(entity)
		newDocumentInfo.setChangeVector(changeVector)

		s.documentsByID.add(newDocumentInfo)
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
		return NewIllegalStateError(fmt.Sprintf("%#v is not associated with the session, cannot delete unknown entity instance", entity))
	}

	s.deletedEntities[entity] = struct{}{}
	delete(s.includedDocumentsByID, value.getId())
	s.knownMissingIDs[value.getId()] = struct{}{}
	return nil
}

// Marks the specified entity for deletion. The entity will be deleted when IDocumentSession.SaveChanges is called.
// WARNING: This method will not call beforeDelete listener!
func (s *InMemoryDocumentSessionOperations) Delete(id string) error {
	return s.DeleteWithChangeVector(id, "")
}

func (s *InMemoryDocumentSessionOperations) DeleteWithChangeVector(id string, expectedChangeVector string) error {
	if id == "" {
		return NewIllegalArgumentError("Id cannot be null")
	}

	changeVector := ""
	documentInfo := s.documentsByID.getValue(id)
	if documentInfo != nil {
		newObj := convertEntityToJson(documentInfo.getEntity(), documentInfo)
		if documentInfo.getEntity() != nil && s.entityChanged(newObj, documentInfo, nil) {
			return NewIllegalStateError("Can't delete changed entity using identifier. Use delete(Class clazz, T entity) instead.")
		}

		if documentInfo.getEntity() != nil {
			delete(s.documentsByEntity, documentInfo.getEntity())
		}

		s.documentsByID.remove(id)
		changeVector = documentInfo.getChangeVector()
	}

	s.knownMissingIDs[id] = struct{}{}
	if !s.useOptimisticConcurrency {
		changeVector = ""
	}
	// TODO: remove
	fmt.Printf("%s\n", changeVector)
	//defer(new DeleteCommandData(id, ObjectUtils.firstNonNull(expectedChangeVector, changeVector)));
	return nil
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
