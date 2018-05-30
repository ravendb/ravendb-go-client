package ravendb

import (
	"fmt"
	"sync"
)

// DocumentStore represents a database
type DocumentStore struct {
	urls          []string // urls for HTTP endopoints of server nodes
	database      string   // name of the database
	isInitialized bool
	mu            sync.Mutex
	// maps database name to its RequestsExecutor
	requestsExecutors            map[string]*RequestExecutor
	Conventions                  *DocumentConventions
	_multiDbHiLo                 *MultiDatabaseHiLoIDGenerator
	maintenanceOperationExecutor *MaintenanceOperationExecutor
}

func (s *DocumentStore) getConventions() *DocumentConventions {
	return s.Conventions
}

func (s *DocumentStore) getURLS() []string {
	return s.urls
}

// NewDocumentStore creates a DocumentStore
// https://sourcegraph.com/github.com/ravendb/RavenDB-Python-Client@v4.0/-/blob/pyravendb/store/document_store.py#L13
func NewDocumentStore(urls []string, db string) *DocumentStore {
	s := &DocumentStore{
		urls:              urls,
		database:          db,
		requestsExecutors: map[string]*RequestExecutor{},
		Conventions:       NewDocumentConventions(),
	}

	// TODO: this belongs also to NewDocumentStore
	if len(s.urls) == 0 {
		err := fmt.Errorf("Must provide urls to NewDocumentStore")
		must(err)
	}
	// TODO: for some operations (like listing databases) you don't need database name
	if s.database == "" {
		err := fmt.Errorf("Must provide database name to NewDocumentStore")
		must(err)
	}
	return s
}

// Initialize initializes document store,
// Must be called before executing any operation.
// https://sourcegraph.com/github.com/ravendb/RavenDB-Python-Client@v4.0/-/blob/pyravendb/store/document_store.py#L96
func (s *DocumentStore) Initialize() error {
	if s.isInitialized {
		return nil
	}
	// TODO: this belongs also to NewDocumentStore
	if len(s.urls) == 0 {
		return fmt.Errorf("Must provide urls to NewDocumentStore")
	}
	// TODO: for some operations (like listing databases) you don't need database name
	if s.database == "" {
		return fmt.Errorf("Must provide database name to NewDocumentStore")
	}
	conventions := s.Conventions
	if conventions.getDocumentIdGenerator() == nil {
		generator := NewMultiDatabaseHiLoIDGenerator(s)
		s._multiDbHiLo = generator
		genID := func(dbName string, entity Object) string {
			return generator.GenerateDocumentID(dbName, entity)
		}
		conventions.setDocumentIdGenerator(genID)
	}

	s.isInitialized = true
	return nil
}

func (s *DocumentStore) assertInitialized() {
	panicIf(!s.isInitialized, "DocumentStore must be initialized")
}

// GetRequestExecutor gets a request executor for a given database
// https://sourcegraph.com/github.com/ravendb/RavenDB-Python-Client@v4.0/-/blob/pyravendb/store/document_store.py#L84
// https://sourcegraph.com/github.com/ravendb/ravendb-jvm-client@v4.0/-/blob/src/main/java/net/ravendb/client/documents/DocumentStore.java#L159
func (s *DocumentStore) GetRequestExecutor(dbName string) *RequestExecutor {
	s.assertInitialized()
	if dbName == "" {
		dbName = s.database
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if re, ok := s.requestsExecutors[dbName]; ok {
		return re
	}
	// TODO: certificate
	re := RequestExecutor_create(s.urls, dbName, nil, s.Conventions)
	s.requestsExecutors[dbName] = re
	return re
}

// TODO: this is temporary, should be on RequestsExecutor
func (s *DocumentStore) getSimpleExecutor() CommandExecutorFunc {
	node := &ServerNode{
		URL:        s.urls[0],
		Database:   s.database,
		ClusterTag: "0",
	}
	return MakeSimpleExecutor(node)
}

// OpenSession opens a new session to document store.
// https://sourcegraph.com/github.com/ravendb/RavenDB-Python-Client@v4.0/-/blob/pyravendb/store/document_store.py#L112
func (s *DocumentStore) OpenSession() (*DocumentSession, error) {
	s.assertInitialized()

	sessionID := NewUUID().String()
	re := s.GetRequestExecutor(s.database)
	return NewDocumentSession(s.database, s, sessionID, re), nil
}

// Close closes the store
func (s *DocumentStore) Close() {
	if s._multiDbHiLo != nil {
		s._multiDbHiLo.ReturnUnusedRange()
	}
	// TODO: more
}

func (s *DocumentStore) getDatabase() string {
	return s.database
}

func (s *DocumentStore) maintenance() *MaintenanceOperationExecutor {
	s.assertInitialized()

	if s.maintenanceOperationExecutor == nil {
		s.maintenanceOperationExecutor = NewMaintenanceOperationExecutor(s)
	}

	return s.maintenanceOperationExecutor
}
