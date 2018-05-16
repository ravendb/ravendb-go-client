package ravendb

import (
	"errors"
	"fmt"
	"reflect"
	"sync/atomic"
	"time"
)

// ObjectNode is an alias for a json document represented as a map
// Name comes from Java implementation
type ObjectNode = map[string]interface{}

// ConcurrencyCheckMode describes concurrency check
type ConcurrencyCheckMode int

const (
	// ConcurrencyCheckAuto is automatic optimistic concurrency check depending on UseOptimisticConcurrency setting or provided Change Vector
	ConcurrencyCheckAuto ConcurrencyCheckMode = iota
	// ConcurrencyCheckForced forces optimistic concurrency check even if UseOptimisticConcurrency is not set
	ConcurrencyCheckForced
	// ConcurrencyCheckDisabled disables optimistic concurrency check even if UseOptimisticConcurrency is set
	ConcurrencyCheckDisabled
)

// DocumentInfo stores information about entity in a session
type DocumentInfo struct {
	id                   string
	changeVector         string
	concurrencyCheckMode ConcurrencyCheckMode
	ignoreChanges        bool
	originalMetadata     map[string]interface{}
	metadata             ObjectNode
	document             ObjectNode
	originalValue        map[string]interface{}
	metadataInstance     map[string]interface{}
	entity               interface{}
	newDocuemnt          bool
	collection           string
}

// TODO: rename to saveChangesData, maybe
type _SaveChangesData struct {
	commands             []*CommandData
	entities             []interface{}
	deferredCommandCount int
}

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

/*
// BatchOptions describes options for batch operations
type BatchOptions struct {
	waitForReplicas                 bool
	numberOfReplicasToWaitFor       int
	waitForReplicasTimeout          time.Duration
	majority                        bool
	throwOnTimeoutInWaitForReplicas bool

	waitForIndexes                 bool
	waitForIndexesTimeout          time.Duration
	throwOnTimeoutInWaitForIndexes bool
	waitForSpecificIndexes         []string
}
*/

// InMemoryDocumentSessionOperations represents database operations queued
// in memory
// TODO: move to own file
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

	// TODO: ignore case for keys
	documentsByID map[string]*DocumentInfo

	// Translate between an ID and its associated entity
	// TODO: ignore case for keys
	// TODO: value is *DocumentInfo
	includedDocumentsByID map[string]JSONAsMap

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

// IMetadataDictionary describes metadata for a document
type IMetadataDictionary = map[string]interface{}

// NewInMemoryDocumentSessionOperations creates new InMemoryDocumentSessionOperations
func NewInMemoryDocumentSessionOperations(dbName string, store *DocumentStore, re *RequestsExecutor) *InMemoryDocumentSessionOperations {
	clientSessionID := newClientSessionID()
	res := InMemoryDocumentSessionOperations{
		clientSessionID:               clientSessionID,
		deletedEntities:               map[interface{}]struct{}{},
		RequestsExecutor:              re,
		generateDocumentKeysOnStore:   true,
		sessionInfo:                   SessionInfo{SessionID: clientSessionID},
		documentsByID:                 map[string]*DocumentInfo{},
		includedDocumentsByID:         map[string]JSONAsMap{},
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
	documentInfo := s.documentsByID[id]
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

// DocumentSession is a Unit of Work for accessing RavenDB server
// https://sourcegraph.com/github.com/ravendb/RavenDB-Python-Client/-/blob/pyravendb/store/document_session.py#L18
// https://sourcegraph.com/github.com/ravendb/RavenDB-Python-Client/-/blob/pyravendb/store/document_session.py#L18
type DocumentSession struct {
	*InMemoryDocumentSessionOperations

	SessionID string

	deferCommands []*CommandData
}

// NewDocumentSession creates a new DocumentSession
func NewDocumentSession(dbName string, store *DocumentStore, id string, re *RequestsExecutor) *DocumentSession {
	res := &DocumentSession{
		InMemoryDocumentSessionOperations: NewInMemoryDocumentSessionOperations(dbName, store, re),
		SessionID:                         id,
	}
	return res
}

//
func (s *DocumentSession) deferCmd(cmd *CommandData, rest ...*CommandData) {
	s.deferCommands = append(s.deferCommands, cmd)
	for _, cmd := range rest {
		s.deferCommands = append(s.deferCommands, cmd)
	}
}

func (s *DocumentSession) saveIncludes(includes map[string]JSONAsMap) {
	for key, value := range includes {
		if _, ok := s.documentsByID[key]; !ok {
			s.includedDocumentsByID[key] = value
		}
	}
}

// https://sourcegraph.com/github.com/ravendb/ravendb-jvm-client@v4.0/-/blob/src/main/java/net/ravendb/client/documents/session/InMemoryDocumentSessionOperations.java#L665
// https://sourcegraph.com/github.com/ravendb/RavenDB-Python-Client@v4.0/-/blob/pyravendb/store/document_session.py#L101
func (s *DocumentSession) saveEntity(key string, entity interface{}, originalMetadata map[string]interface{}, metadata map[string]interface{}, document *DocumentInfo, concurrencyCheckMode ConcurrencyCheckMode) {
	// TODO: can key here be ever empty?
	delete(s.deletedEntities, entity)
	if key != "" {
		delete(s.knownMissingIDs, key)
		if _, ok := s.documentsByID[key]; ok {
			return
		}
	}
	if key != "" {
		s.documentsByID[key] = document
	}
	if document == nil {
		document = &DocumentInfo{}
	}
	document.id = key
	document.metadata = metadata
	//document.changeVector = ""
	document.concurrencyCheckMode = concurrencyCheckMode
	document.entity = entity
	document.newDocuemnt = true
	s.documentsByEntity[entity] = document
}

func (s *DocumentSession) convertAndSaveEntity(key string, document JSONAsMap, objectType reflect.Value) {
	if _, ok := s.documentsByID[key]; ok {
		return
	}
	// TODO: convert_to_entity
	panicIf(true, "NYI")
}

func (s *DocumentSession) multiLoad(keys []string, res interface{}, includes []string) error {
	// TODO: get objectType which is reflect.Value of res where res is []*struct
	idsOfNotExistingObject := stringArrayCopy(keys)
	if len(includes) == 0 {
		var idsInIncludes []string
		for _, key := range idsOfNotExistingObject {
			if _, ok := s.includedDocumentsByID[key]; ok {
				idsInIncludes = append(idsInIncludes, key)
			}
		}
		for _, include := range idsInIncludes {
			panicIf(true, "NYI")
			//self._convert_and_save_entity(include, self._included_documents_by_id[include], object_type, nested_object_types)
			delete(s.includedDocumentsByID, include)
		}
		var a []string
		for _, key := range idsOfNotExistingObject {
			if _, ok := s.documentsByID[key]; !ok {
				a = append(a, key)
			}
		}
		idsOfNotExistingObject = a
	}

	var a []string
	for _, key := range idsOfNotExistingObject {
		if _, ok := s.knownMissingIDs[key]; !ok {
			a = append(a, key)
		}
	}
	idsOfNotExistingObject = a

	if len(idsOfNotExistingObject) > 0 {
		// TODO: propagate error
		s.IncrementRequetsCount()
		cmd := NewGetDocumentCommand(idsOfNotExistingObject, includes, false)
		exec := s.documentStore.GetRequestExecutor("").GetCommandExecutor(false)
		res, err := ExecuteGetDocumentCommand(exec, cmd)
		if err != nil {
			return err
		}
		results := res.Results
		for i := 0; i < len(results); i++ {
			key := idsOfNotExistingObject[i]
			jsonEntity := results[i]
			if len(jsonEntity) == 0 {
				s.knownMissingIDs[key] = struct{}{}
				continue
			}
			var objectType reflect.Value
			s.convertAndSaveEntity(key, jsonEntity, objectType)
		}
		s.saveIncludes(res.Includes)
	}
	return nil
}

// Load loads documents from a database based
func (s *DocumentSession) Load(keys []string, res interface{}, includes []string) error {
	if len(keys) == 0 {
		return errors.New("must provide keys")
	}
	return s.multiLoad(keys, res, includes)
}

// TODO: delete_by_entity
// TODO: delete

func (s *DocumentSession) assertNoNonUniqueInstance(entity interface{}, key string) {
	// TODO: implement me
}

// Store schedules entity for storing in the database. To actually save the
// data, call SaveSession.
// key and changeVector can be ""
// https://sourcegraph.com/github.com/ravendb/RavenDB-Python-Client@v4.0/-/blob/pyravendb/store/document_session.py#L248
// https://sourcegraph.com/github.com/ravendb/ravendb-jvm-client@v4.0/-/blob/src/main/java/net/ravendb/client/documents/session/InMemoryDocumentSessionOperations.java#L601
func (s *DocumentSession) Store(entity interface{}, key string, changeVector string) error {
	panicIf(entity == nil, "entity cannot be nil") // TODO: return as error?
	// TODO: check entity is a struct
	forceConcurrencyCheck := s.getConcurrencyCheckMode(entity, key, changeVector)
	if docInfo, ok := s.documentsByEntity[entity]; ok {
		if changeVector != "" {
			docInfo.changeVector = changeVector
		}
		docInfo.concurrencyCheckMode = forceConcurrencyCheck
		return nil
	}
	entityID := ""
	if key == "" {
		entityID, _ = tryGetIDFromInstance(entity)
	} else {
		trySetIDOnEntity(entity, key)
		entityID = key
	}

	s.assertNoNonUniqueInstance(entity, entityID)
	if entityID == "" {
		entityID = s.documentStore.generateID(s.databaseName, entity)
		trySetIDOnEntity(entity, entityID)
	}

	for _, command := range s.deferCommands {
		if command.key == entityID {
			return fmt.Errorf("Can't store document, there is a deferred command registered for this document in the session. Document id: %s", entityID)
		}
	}

	if _, ok := s.deletedEntities[entity]; ok {
		err := fmt.Errorf("Can't store object, it was already deleted in this session.  Document id: %s", entityID)
		return err
	}
	metadata := buildDefaultMetadata(entity)
	s.saveEntity(entityID, entity, nil, metadata, nil, forceConcurrencyCheck)
	return nil
}

// https://sourcegraph.com/github.com/ravendb/RavenDB-Python-Client@v4.0/-/blob/pyravendb/store/document_session.py#L294
func (s *DocumentSession) getConcurrencyCheckMode(entity interface{}, key string, changeVector string) ConcurrencyCheckMode {
	// TODO: port the logic from python
	return ConcurrencyCheckDisabled
}

// SaveChanges saves documents to database queued with Store
func (s *DocumentSession) SaveChanges() error {
	data := &_SaveChangesData{
		commands:             s.deferCommands,
		deferredCommandCount: len(s.deferCommands),
	}
	s.deferCommands = nil
	s.prepareForDeleteCommands(data)
	s.prepareForPutsCommands(data)
	if len(data.commands) == 0 {
		return nil
	}

	err := s.IncrementRequetsCount()
	if err != nil {
		return err
	}

	batchCommand := NewBatchCommand(data.commands)
	exec := s.RequestsExecutor.GetCommandExecutor(false)
	batchResult, err := ExecuteBatchCommand(exec, batchCommand)
	if err != nil {
		return err
	}
	// TODO: batch_result != None
	s.updateBatchResult(batchResult, data)
	return nil
}

func (s *DocumentSession) updateBatchResult(batchResult JSONArrayResult, data *_SaveChangesData) {
	batchResultLength := len(batchResult)
	for i := data.deferredCommandCount; i < batchResultLength; i++ {
		item := batchResult[i]
		typ := item["Type"]
		typStr := typ.(string)
		if typStr != "PUT" {
			continue
		}
		entity := data.entities[i-data.deferredCommandCount]
		documentInfo, ok := s.documentsByEntity[entity]
		if !ok {
			continue
		}
		// TODO: add helper getAsString()
		key := item["@id"]
		keyStr := key.(string)
		s.documentsByID[keyStr] = documentInfo
		// TODO: python code is document_info["change_vector"] = ["change_vector"]
		// which seems wrong
		delete(item, "Type")
		documentInfo.originalMetadata = copyJSONMap(item)
		documentInfo.metadata = item
		documentInfo.originalValue = structToJSONMap(entity)
	}
}

// https://sourcegraph.com/github.com/ravendb/RavenDB-Python-Client@v4.0/-/blob/pyravendb/store/document_session.py#L338
// https://sourcegraph.com/github.com/ravendb/ravendb-jvm-client@v4.0/-/blob/src/main/java/net/ravendb/client/documents/session/InMemoryDocumentSessionOperations.java#L742
func (s *DocumentSession) prepareForDeleteCommands(data *_SaveChangesData) {
	// TODO: we don't need to gather keysToDelete
	if len(s.deletedEntities) == 0 {
		return
	}
	var keysToDelete []string
	for _, entity := range s.deletedEntities {
		docInfo := s.documentsByEntity[entity]
		key := docInfo.id
		keysToDelete = append(keysToDelete, key)
	}

	for _, key := range keysToDelete {
		existingEntity, ok := s.documentsByID[key]
		if !ok {
			continue
		}
		var changeVector string
		if docInfo := s.documentsByEntity[existingEntity]; docInfo != nil {
			meta := docInfo.metadata
			// TODO: take optimistic concurrency setting into account
			if v, ok := meta["@change-vector"]; ok {
				changeVector = v.(string)
			}
			delete(s.documentsByEntity, existingEntity)
			delete(s.documentsByID, key)
			data.entities = append(data.entities, existingEntity)
			cmdData := NewDeleteCommandData(key, changeVector)
			data.commands = append(data.commands, cmdData)
		}
	}
	s.deletedEntities = nil
}

func (s *DocumentSession) prepareForPutsCommands(data *_SaveChangesData) {
	for entity, docInfo := range s.documentsByEntity {
		if !s.hasChange(entity) {
			continue
		}
		key := docInfo.id
		metadata := docInfo.metadata
		changeVector := ""
		// TODO: logic for changeVector
		document := structToJSONMap(entity)
		data.entities = append(data.entities, entity)
		if key != "" {
			delete(s.documentsByID, key)
			deleteID(document)
		}
		cmd := NewPutCommandData(key, changeVector, document, metadata)
		data.commands = append(data.commands, cmd)
	}
}

// hasChange returns true if entity has changes
func (s *DocumentSession) hasChange(entity interface{}) bool {
	// TODO: implement me
	return true
}
