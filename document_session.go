package ravendb

import (
	"errors"
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
	knownMissingIDs []string
	// TODO: move rields
	// documentsByID map[string] ??
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

	return errors.New("FYI")
}

// SaveChanges saves documents to database queued with Store
func (s *DocumentSession) SaveChanges() error {
	panicIf(true, "NYI")
	return nil
}

// https://sourcegraph.com/github.com/ravendb/RavenDB-Python-Client@v4.0/-/blob/pyravendb/store/document_session.py#L294
func (s *DocumentSession) getConcurrencyCheckMode(entity interface{}, key string, changeVector string) ConcurrencyCheckMode {
	// TODO: port the logic from python
	return ConcurrencyCheckDisabled
}
