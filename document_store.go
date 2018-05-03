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
	requestsExecutors map[string]*RequestsExecutor
}

// NewDocumentStore creates a DocumentStore
// https://sourcegraph.com/github.com/ravendb/RavenDB-Python-Client@v4.0/-/blob/pyravendb/store/document_store.py#L13
func NewDocumentStore(urls []string, db string) *DocumentStore {
	res := &DocumentStore{
		urls:     urls,
		database: db,
	}
	return res
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
	if s.database == "" {
		return fmt.Errorf("Must provide database name to NewDocumentStore")
	}
	return nil
}

func (s *DocumentStore) assertInitialized() {
	panicIf(!s.isInitialized, "DocumentStore must be initialized")
}

// GetRequestExecutor gets a request executor for a given database
// https://sourcegraph.com/github.com/ravendb/RavenDB-Python-Client@v4.0/-/blob/pyravendb/store/document_store.py#L84
// https://sourcegraph.com/github.com/ravendb/ravendb-jvm-client@v4.0/-/blob/src/main/java/net/ravendb/client/documents/DocumentStore.java#L159
func (s *DocumentStore) GetRequestExecutor(dbName string) *RequestsExecutor {
	s.assertInitialized()
	if dbName == "" {
		dbName = s.database
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if re, ok := s.requestsExecutors[dbName]; ok {
		return re
	}
	// TODO: certificate, conventions
	re := CreateRequestsExecutor(s.urls, dbName)
	s.requestsExecutors[dbName] = re
	return re
}

// OpenSession opens a new session to document store.
// https://sourcegraph.com/github.com/ravendb/RavenDB-Python-Client@v4.0/-/blob/pyravendb/store/document_store.py#L112
func (s *DocumentStore) OpenSession() (*DocumentSession, error) {
	s.assertInitialized()
	res := &DocumentSession{}

	//session_id = uuid.uuid4()
	//requests_executor = self.get_request_executor(database) if request_executor is None else request_executor
	//return DocumentSession(database, self, requests_executor, session_id)

	return res, nil
}
