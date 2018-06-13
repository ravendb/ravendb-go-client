package ravendb

import (
	"fmt"
	"strings"
	"sync"
	"time"
)

// DocumentStore represents a database
type DocumentStore struct {
	// from DocumentStoreBase
	//List<EventHandler<BeforeStoreEventArgs>> onBeforeStore = new ArrayList<>();
	//List<EventHandler<AfterSaveChangesEventArgs>> onAfterSaveChanges = new ArrayList<>();
	//List<EventHandler<BeforeDeleteEventArgs>> onBeforeDelete = new ArrayList<>();
	//List<EventHandler<BeforeQueryEventArgs>> onBeforeQuery = new ArrayList<>();
	//List<EventHandler<SessionCreatedEventArgs>> onSessionCreated = new ArrayList<>();
	disposed     bool
	conventions  *DocumentConventions
	urls         []string // urls for HTTP endopoints of server nodes
	initialized  bool
	_certificate *KeyStore
	database     string // name of the database

	// TODO: _databaseChanges
	// TODO: _aggressiveCacheChanges
	// maps database name to its RequestsExecutor
	requestsExecutors            map[string]*RequestExecutor
	_multiDbHiLo                 *MultiDatabaseHiLoIDGenerator
	maintenanceOperationExecutor *MaintenanceOperationExecutor
	operationExecutor            *OperationExecutor
	identifier                   string
	_aggressiveCachingUsed       bool

	afterClose  []func(*DocumentStore)
	beforeClose []func(*DocumentStore)

	// old
	mu sync.Mutex
}

// from DocumentStoreBase
func (s *DocumentStore) getConventions() *DocumentConventions {
	if s.conventions == nil {
		s.conventions = NewDocumentConventions()
	}
	return s.conventions
}

func (s *DocumentStore) setConventions(conventions *DocumentConventions) {
	s.conventions = conventions
}

func (s *DocumentStore) getUrls() []string {
	return s.urls
}

func (s *DocumentStore) setUrls(value []string) {
	panicIf(len(value) == 0, "value is empty")
	for i, s := range value {
		value[i] = strings.TrimSuffix(s, "/")
	}
	s.urls = value
}

func (s *DocumentStore) ensureNotClosed() {
	// TODO: implement me
}

/*
public void addBeforeStoreListener(EventHandler<BeforeStoreEventArgs> handler) {
	this.onBeforeStore.add(handler);

}
public void removeBeforeStoreListener(EventHandler<BeforeStoreEventArgs> handler) {
	this.onBeforeStore.remove(handler);
}

public void addAfterSaveChangesListener(EventHandler<AfterSaveChangesEventArgs> handler) {
	this.onAfterSaveChanges.add(handler);
}

public void removeAfterSaveChangesListener(EventHandler<AfterSaveChangesEventArgs> handler) {
	this.onAfterSaveChanges.remove(handler);
}

public void addBeforeDeleteListener(EventHandler<BeforeDeleteEventArgs> handler) {
	this.onBeforeDelete.add(handler);
}
public void removeBeforeDeleteListener(EventHandler<BeforeDeleteEventArgs> handler) {
	this.onBeforeDelete.remove(handler);
}

public void addBeforeQueryListener(EventHandler<BeforeQueryEventArgs> handler) {
	this.onBeforeQuery.add(handler);
}
public void removeBeforeQueryListener(EventHandler<BeforeQueryEventArgs> handler) {
	this.onBeforeQuery.remove(handler);
}
*/

func (s *DocumentStore) assertInitialized() {
	panicIf(!s.initialized, "DocumentStore must be initialized")
}

func (s *DocumentStore) getDatabase() string {
	return s.database
}

func (s *DocumentStore) setDatabase(database string) {
	panicIf(s.initialized, "is already initialized")
	s.database = database
}

func (s *DocumentStore) getCertificate() *KeyStore {
	return s._certificate
}

func (s *DocumentStore) setCertificate(certificate *KeyStore) {
	panicIf(s.initialized, "is already initialized")
	s._certificate = certificate
}

func (s *DocumentStore) aggressivelyCache() {
	s.aggressivelyCacheWithDatabase("")
}

func (s *DocumentStore) aggressivelyCacheWithDatabase(database string) {
	s.aggressivelyCacheForDatabase(time.Hour*24, database)
}

//    protected void registerEvents(InMemoryDocumentSessionOperations session) {

// NewDocumentStore creates a DocumentStore
func NewDocumentStore() *DocumentStore {
	s := &DocumentStore{
		requestsExecutors: map[string]*RequestExecutor{},
		conventions:       NewDocumentConventions(),
	}
	return s
}

func NewDocumentStoreWithUrlAndDatabase(url string, database string) *DocumentStore {
	res := NewDocumentStore()
	res.setUrls([]string{url})
	res.setDatabase(database)
	return res
}

func NewDocumentStoreWithUrlsAndDatabase(urls []string, database string) *DocumentStore {
	res := NewDocumentStore()
	res.setUrls(urls)
	res.setDatabase(database)
	return res
}

func (s *DocumentStore) getIdentifier() string {
	if s.identifier != "" {
		return s.identifier
	}

	if len(s.urls) == 0 {
		return ""
	}

	if s.database != "" {
		return strings.Join(s.urls, ",") + " (DB: " + s.database + ")"
	}

	return strings.Join(s.urls, ",")
}

func (s *DocumentStore) setIdentifier(identifier string) {
	s.identifier = identifier
}

// Close closes the store
func (s *DocumentStore) Close() {
	for _, fn := range s.beforeClose {
		fn(s)
	}
	s.beforeClose = nil

	// TODO: evict _aggressiveCacheChanges

	// TODO: close _databaseChanges

	if s._multiDbHiLo != nil {
		s._multiDbHiLo.ReturnUnusedRange()
	}

	s.disposed = true

	for _, fn := range s.afterClose {
		fn(s)
	}
	s.afterClose = nil

	for _, re := range s.requestsExecutors {
		re.close()
	}
}

// OpenSession opens a new session to document store.
func (s *DocumentStore) OpenSession() (*DocumentSession, error) {
	return s.OpenSessionWithOptions(NewSessionOptions())
}

func (s *DocumentStore) OpenSessionWithDatabase(database string) (*DocumentSession, error) {
	sessionOptions := NewSessionOptions()
	sessionOptions.setDatabase(database)
	return s.OpenSessionWithOptions(sessionOptions)
}

func (s *DocumentStore) OpenSessionWithOptions(options *SessionOptions) (*DocumentSession, error) {
	s.assertInitialized()
	s.ensureNotClosed()

	sessionID := NewUUID().String()
	databaseName := firstNonEmptyString(options.getDatabase(), s.getDatabase())
	requestExecutor := options.getRequestExecutor()
	if requestExecutor == nil {
		requestExecutor = s.GetRequestExecutorWithDatabase(databaseName)
	}
	session := NewDocumentSession(databaseName, s, sessionID, requestExecutor)
	//s.registerEvents(session);
	//s.afterSessionCreated(session);
	return session, nil
}

func (s *DocumentStore) GetRequestExecutor() *RequestExecutor {
	return s.GetRequestExecutorWithDatabase("")
}

// GetRequestExecutorWithDatabase gets a request executor for a given database
func (s *DocumentStore) GetRequestExecutorWithDatabase(database string) *RequestExecutor {
	s.assertInitialized()
	if database == "" {
		database = s.getDatabase()
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	executor, ok := s.requestsExecutors[database]
	if ok {
		return executor
	}

	if !s.getConventions().isDisableTopologyUpdates() {
		executor = RequestExecutor_create(s.getUrls(), s.getDatabase(), s.getCertificate(), s.getConventions())
	} else {
		executor = RequestExecutor_createForSingleNodeWithConfigurationUpdates(s.getUrls()[0], s.getDatabase(), s.getCertificate(), s.getConventions())
	}
	s.requestsExecutors[database] = executor
	return executor
}

// Initialize initializes document store,
// Must be called before executing any operation.
func (s *DocumentStore) Initialize() (*DocumentStore, error) {
	if s.initialized {
		return s, nil
	}
	err := s.assertValidConfiguration()
	if err != nil {
		return nil, err
	}

	conventions := s.conventions
	if conventions.getDocumentIdGenerator() == nil {
		generator := NewMultiDatabaseHiLoIdGenerator(s, s.getConventions())
		s._multiDbHiLo = generator
		genID := func(dbName string, entity Object) string {
			return generator.GenerateDocumentID(dbName, entity)
		}
		conventions.setDocumentIdGenerator(genID)
	}
	s.initialized = true
	return s, nil
}

func (s *DocumentStore) assertValidConfiguration() error {
	if len(s.urls) == 0 {
		return fmt.Errorf("Must provide urls to NewDocumentStore")
	}
	return nil
}

func (s *DocumentStore) disableAggressiveCaching() {
	s.disableAggressiveCachingWithDatabase("")
}

func (s *DocumentStore) disableAggressiveCachingWithDatabase(databaseName string) {
	// TODO: implement me
}

//    public IDatabaseChanges changes() {
//    public IDatabaseChanges changes(string database) {
//    protected IDatabaseChanges createDatabaseChanges(string database) {
//     public Exception getLastDatabaseChangesStateException() {
//    public Exception getLastDatabaseChangesStateException(string database) {

func (s *DocumentStore) aggressivelyCacheFor(cacheDuration time.Duration) {
	// TODO: implement me
}

func (s *DocumentStore) aggressivelyCacheForDatabase(cacheDuration time.Duration, database string) {
	// TODO: implement me
}

//    private void listenToChangesAndUpdateTheCache(string database) {

func (s *DocumentStore) addBeforeCloseListener(fn func(*DocumentStore)) {
	s.beforeClose = append(s.beforeClose, fn)
}

//   public void removeBeforeCloseListener(EventHandler<VoidArgs> event) {

func (s *DocumentStore) addAfterCloseListener(fn func(*DocumentStore)) {
	s.afterClose = append(s.afterClose, fn)
}

//    public void removeAfterCloseListener(EventHandler<VoidArgs> event) {

func (s *DocumentStore) maintenance() *MaintenanceOperationExecutor {
	s.assertInitialized()

	if s.maintenanceOperationExecutor == nil {
		s.maintenanceOperationExecutor = NewMaintenanceOperationExecutor(s)
	}

	return s.maintenanceOperationExecutor
}

func (s *DocumentStore) operations() *OperationExecutor {
	if s.operationExecutor == nil {
		s.operationExecutor = NewOperationExecutor(s)
	}

	return s.operationExecutor
}

//    public BulkInsertOperation bulkInsert() {
//    public BulkInsertOperation bulkInsert(string database) {
