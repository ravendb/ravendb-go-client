package ravendb

import (
	"fmt"
)

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
	metadata             map[string]interface{}
	document             map[string]interface{}
	metadataInstance     map[string]interface{}
	entity               interface{}
	newDocuemnt          bool
	collection           string
}

// TODO: rename to saveChangesData
type _SaveChangesData struct {
	commands             []*CommandData
	entities             []interface{}
	deferredCommandCount int
}

// DocumentSession is a Unit of Work for accessing RavenDB server
// https://sourcegraph.com/github.com/ravendb/RavenDB-Python-Client/-/blob/pyravendb/store/document_session.py#L18
// https://sourcegraph.com/github.com/ravendb/RavenDB-Python-Client/-/blob/pyravendb/store/document_session.py#L18
type DocumentSession struct {
	SessionID                 string
	Database                  string
	documentStore             *DocumentStore
	RequestsExecutor          *RequestsExecutor
	NumberOfRequestsInSession int
	Conventions               *DocumentConventions
	// in-flight objects scheduled to Store, before calling SaveChanges
	// key is name of the type
	// TODO: this uses value semantics, so it works as expected for
	// pointers to structs, but 2 different structs with the same content
	// will match the same object. Should I disallow storing non-pointer structs?
	// convert non-pointer structs to structs?
	documentsByEntity map[interface{}]*DocumentInfo
	deletedEntities   map[interface{}]struct{}
	// ids of entities that were deleted
	knownMissingIDs  []string
	documentsByID    map[string]interface{}
	deferredCommands []*CommandData
}

// NewDocumentSession creates a new DocumentSession
func NewDocumentSession(dbName string, documentStore *DocumentStore, id string, re *RequestsExecutor) *DocumentSession {
	res := &DocumentSession{
		SessionID:         id,
		Database:          dbName,
		documentStore:     documentStore,
		RequestsExecutor:  re,
		Conventions:       documentStore.Conventions,
		documentsByEntity: map[interface{}]*DocumentInfo{},
		documentsByID:     map[string]interface{}{},
		deletedEntities:   map[interface{}]struct{}{},
	}
	return res
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

	// TODO:
	// s.assertNoNonUniqueInstance(entity, entityID)
	if entityID == "" {
		entityID = s.documentStore.generateID(s.Database, entity)
		trySetIDOnEntity(entity, entityID)
	}

	// TODO: implement this
	/*
		for command in self._defer_commands:
			if command.key == entity_id:
				raise exceptions.InvalidOperationException(
					"Can't store document, there is a deferred command registered for this document in the session. "
					"Document id: " + entity_id)
	*/
	if _, ok := s.deletedEntities[entity]; ok {
		return fmt.Errorf("Can't store object, it was already deleted in this session.  Document id: %s", entityID)
	}
	metadata := buildDefaultMetadata(entity)
	s.saveEntity(entityID, entity, nil, metadata, nil, forceConcurrencyCheck)
	return nil
}

// https://sourcegraph.com/github.com/ravendb/ravendb-jvm-client@v4.0/-/blob/src/main/java/net/ravendb/client/documents/session/InMemoryDocumentSessionOperations.java#L665
// https://sourcegraph.com/github.com/ravendb/RavenDB-Python-Client@v4.0/-/blob/pyravendb/store/document_session.py#L101
func (s *DocumentSession) saveEntity(key string, entity interface{}, originalMetadata map[string]interface{}, metadata map[string]interface{}, document *DocumentInfo, concurrencyCheckMode ConcurrencyCheckMode) {
	// TODO: can key here be ever empty?
	delete(s.deletedEntities, entity)
	if key != "" {
		removeStringFromArray(&s.knownMissingIDs, key)
		if _, ok := s.documentsByID[key]; ok {
			return
		}
	}
	if key != "" {
		s.documentsByID[key] = entity
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

// TODO: move to DocumentConventions
func buildDefaultMetadata(entity interface{}) map[string]interface{} {
	res := map[string]interface{}{}
	if entity == nil {
		return res
	}
	fullTypeName := getTypeName(entity)
	typeName := getShortTypeName(entity)
	collectionName := pluralize(typeName)
	res["@collection"] = collectionName
	res["Raven-Go-Type"] = fullTypeName
	return res
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
	// TODO: implement me
	panicIf(true, "NYI")
}

// SaveChanges saves documents to database queued with Store
func (s *DocumentSession) SaveChanges() error {
	data := &_SaveChangesData{
		commands:             s.deferredCommands,
		deferredCommandCount: len(s.deferredCommands),
	}
	s.deferredCommands = nil
	s.prepareForDeleteCommands(data)
	s.prepareForPutsCommands(data)
	if len(data.commands) == 0 {
		return nil
	}

	err := s.incrementRequetsCount()
	if err != nil {
		return err
	}

	panicIf(true, "NYI")
	return nil
}

const (
	// TODO: this should be in DocumentConventiosn
	maxNumberOfRequestPerSession = 32
)

// https://sourcegraph.com/github.com/ravendb/RavenDB-Python-Client@v4.0/-/blob/pyravendb/store/document_session.py#L380
func (s *DocumentSession) incrementRequetsCount() error {
	s.NumberOfRequestsInSession++
	if s.NumberOfRequestsInSession > maxNumberOfRequestPerSession {
		return fmt.Errorf("exceeded max number of reqeusts per session of %d", maxNumberOfRequestPerSession)
	}
	return nil
}

// https://sourcegraph.com/github.com/ravendb/RavenDB-Python-Client@v4.0/-/blob/pyravendb/store/document_session.py#L294
func (s *DocumentSession) getConcurrencyCheckMode(entity interface{}, key string, changeVector string) ConcurrencyCheckMode {
	// TODO: port the logic from python
	return ConcurrencyCheckDisabled
}
