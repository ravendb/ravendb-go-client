package ravendb

import (
	"crypto/tls"
	"crypto/x509"
	"strings"
	"sync"
	"time"
)

// Note: Java's IDocumentStore is DocumentStore
// Note: Java's DocumentStoreBase is folded into DocumentStore

// DocumentStore represents a database
type DocumentStore struct {
	// from DocumentStoreBase
	onBeforeStore      []func(*BeforeStoreEventArgs)
	onAfterSaveChanges []func(*AfterSaveChangesEventArgs)

	onBeforeDelete []func(*BeforeDeleteEventArgs)
	onBeforeQuery  []func(*BeforeQueryEventArgs)
	// TODO: there's no way to register for this event
	onSessionCreated []func(*SessionCreatedEventArgs)
	subscriptions    *DocumentSubscriptions

	disposed    bool
	conventions *DocumentConventions
	urls        []string // urls for HTTP endopoints of server nodes
	initialized bool
	Certificate *tls.Certificate
	TrustStore  *x509.Certificate
	database    string // name of the database

	// maps database name to DatabaseChanges. Must be protected with mutex
	databaseChanges map[string]*DatabaseChanges

	// Note: access must be protected with mu
	// Lazy.Value is **EvictItemsFromCacheBasedOnChanges
	aggressiveCacheChanges map[string]*evictItemsFromCacheBasedOnChanges

	// maps database name to its RequestsExecutor
	// access must be protected with mu
	// TODO: in Java is ConcurrentMap<String, RequestExecutor> requestExecutors
	// so must protect access with mutex and use case-insensitive lookup
	requestsExecutors map[string]*RequestExecutor

	multiDbHiLo                  *MultiDatabaseHiLoIDGenerator
	maintenanceOperationExecutor *MaintenanceOperationExecutor
	operationExecutor            *OperationExecutor
	identifier                   string
	aggressiveCachingUsed        bool

	afterClose  []func(*DocumentStore)
	beforeClose []func(*DocumentStore)

	mu sync.Mutex
}

// methods from DocumentStoreBase

// GetConventions returns DocumentConventions
func (s *DocumentStore) GetConventions() *DocumentConventions {
	if s.conventions == nil {
		s.conventions = NewDocumentConventions()
	}
	return s.conventions
}

// SetConventions sets DocumentConventions
func (s *DocumentStore) SetConventions(conventions *DocumentConventions) {
	s.assertNotInitialized("conventions")
	s.conventions = conventions
}

// Subscriptions returns DocumentSubscriptions which allows subscribing to changes in store
func (s *DocumentStore) Subscriptions() *DocumentSubscriptions {
	return s.subscriptions
}

// GetUrls returns urls of all RavenDB nodes
func (s *DocumentStore) GetUrls() []string {
	return s.urls
}

// SetUrls sets initial urls of RavenDB nodes
func (s *DocumentStore) SetUrls(urls []string) {
	panicIf(len(urls) == 0, "urls is empty")
	s.assertNotInitialized("urls")
	for i, s := range urls {
		urls[i] = strings.TrimSuffix(s, "/")
	}
	s.urls = urls
}

func (s *DocumentStore) ensureNotClosed() error {
	if s.disposed {
		return newIllegalStateError("The document store has already been disposed and cannot be used")
	}
	return nil
}

// AddBeforeStoreStoreListener registers a function that will be called before storing ab entity.
// It'll be registered with every new session.
// Returns listener id that can be passed to RemoveBeforeStoreListener to unregister
// the listener.
func (s *DocumentStore) AddBeforeStoreListener(handler func(*BeforeStoreEventArgs)) int {
	id := len(s.onBeforeStore)
	s.onBeforeStore = append(s.onBeforeStore, handler)
	return id

}

// RemoveBeforeStoreListener removes a listener given id returned by AddBeforeStoreListener
func (s *DocumentStore) RemoveBeforeStoreListener(handlerID int) {
	s.onBeforeStore[handlerID] = nil
}

// AddAfterSaveChangesListener registers a function that will be called before saving changes.
// It'll be registered with every new session.
// Returns listener id that can be passed to RemoveAfterSaveChangesListener to unregister
// the listener.
func (s *DocumentStore) AddAfterSaveChangesListener(handler func(*AfterSaveChangesEventArgs)) int {
	s.onAfterSaveChanges = append(s.onAfterSaveChanges, handler)
	return len(s.onAfterSaveChanges) - 1
}

// RemoveAfterSaveChangesListener removes a listener given id returned by AddAfterSaveChangesListener
func (s *DocumentStore) RemoveAfterSaveChangesListener(handlerID int) {
	s.onAfterSaveChanges[handlerID] = nil
}

// AddBeforeDeleteListener registers a function that will be called before deleting an entity.
// It'll be registered with every new session.
// Returns listener id that can be passed to RemoveBeforeDeleteListener to unregister
// the listener.
func (s *DocumentStore) AddBeforeDeleteListener(handler func(*BeforeDeleteEventArgs)) int {
	s.onBeforeDelete = append(s.onBeforeDelete, handler)
	return len(s.onBeforeDelete) - 1
}

// RemoveBeforeDeleteListener removes a listener given id returned by AddBeforeDeleteListener
func (s *DocumentStore) RemoveBeforeDeleteListener(handlerID int) {
	s.onBeforeDelete[handlerID] = nil
}

// AddBeforeQueryListener registers a function that will be called before running a query.
// It allows customizing query via DocumentQueryCustomization.
// It'll be registered with every new session.
// Returns listener id that can be passed to RemoveBeforeQueryListener to unregister
// the listener.
func (s *DocumentStore) AddBeforeQueryListener(handler func(*BeforeQueryEventArgs)) int {
	s.onBeforeQuery = append(s.onBeforeQuery, handler)
	return len(s.onBeforeQuery) - 1
}

// RemoveBeforeQueryListener removes a listener given id returned by AddBeforeQueryListener
func (s *DocumentStore) RemoveBeforeQueryListener(handlerID int) {
	s.onBeforeQuery[handlerID] = nil
}

func (s *DocumentStore) registerEvents(session *InMemoryDocumentSessionOperations) {
	// TODO: unregister those events?
	for _, handler := range s.onBeforeStore {
		if handler != nil {
			session.AddBeforeStoreListener(handler)
		}
	}

	for _, handler := range s.onAfterSaveChanges {
		if handler != nil {
			session.AddAfterSaveChangesListener(handler)
		}
	}

	for _, handler := range s.onBeforeDelete {
		if handler != nil {
			session.AddBeforeDeleteListener(handler)
		}
	}

	for _, handler := range s.onBeforeQuery {
		if handler != nil {
			session.AddBeforeQueryListener(handler)
		}
	}
}

func (s *DocumentStore) afterSessionCreated(session *InMemoryDocumentSessionOperations) {
	for _, handler := range s.onSessionCreated {
		if handler != nil {
			args := &SessionCreatedEventArgs{
				Session: session,
			}
			handler(args)
		}
	}
}

func (s *DocumentStore) assertInitialized() error {
	if !s.initialized {
		return newIllegalStateError("DocumentStore must be initialized")
	}
	return nil
}

func (s *DocumentStore) assertNotInitialized(property string) {
	panicIf(s.initialized, "You cannot set '%s' after the document store has been initialized.", property)
}

func (s *DocumentStore) GetDatabase() string {
	return s.database
}

func (s *DocumentStore) SetDatabase(database string) {
	s.assertNotInitialized("database")
	s.database = database
}

func (s *DocumentStore) AggressivelyCache(database string) (CancelFunc, error) {
	return s.AggressivelyCacheForDatabase(time.Hour*24, database)
}

func newDocumentStore() *DocumentStore {
	s := &DocumentStore{
		requestsExecutors:      map[string]*RequestExecutor{},
		conventions:            NewDocumentConventions(),
		databaseChanges:        map[string]*DatabaseChanges{},
		aggressiveCacheChanges: map[string]*evictItemsFromCacheBasedOnChanges{},
	}
	s.subscriptions = newDocumentSubscriptions(s)
	return s
}

func NewDocumentStore(urls []string, database string) *DocumentStore {
	res := newDocumentStore()
	if len(urls) > 0 {
		res.SetUrls(urls)
	}
	if database != "" {
		res.SetDatabase(database)
	}
	return res
}

// Get an identifier of the store. For debugging / testing.
func (s *DocumentStore) GetIdentifier() string {
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

func (s *DocumentStore) SetIdentifier(identifier string) {
	s.identifier = identifier
}

// Close closes the Store
func (s *DocumentStore) Close() {
	if s.disposed {
		redbg("DocumentStore.Close: already disposed\n")
		return
	}
	redbg("DocumentStore.Close\n")

	for _, fn := range s.beforeClose {
		fn(s)
	}
	s.beforeClose = nil

	for _, evict := range s.aggressiveCacheChanges {
		evict.Close()
	}

	for _, changes := range s.databaseChanges {
		changes.Close()
	}

	if s.multiDbHiLo != nil {
		s.multiDbHiLo.ReturnUnusedRange()
	}

	if s.Subscriptions() != nil {
		_ = s.Subscriptions().Close()
	}

	s.disposed = true

	for _, fn := range s.afterClose {
		fn(s)
	}
	s.afterClose = nil

	for _, re := range s.requestsExecutors {
		re.Close()
	}
}

// OpenSession opens a new session to document Store.
// If database is not given, we'll use store's database name
func (s *DocumentStore) OpenSession(database string) (*DocumentSession, error) {
	sessionOptions := &SessionOptions{
		Database: database,
	}
	return s.OpenSessionWithOptions(sessionOptions)
}

func (s *DocumentStore) OpenSessionWithOptions(options *SessionOptions) (*DocumentSession, error) {
	if err := s.assertInitialized(); err != nil {
		return nil, err
	}
	if err := s.ensureNotClosed(); err != nil {
		return nil, err
	}

	sessionID := NewUUID().String()
	databaseName := options.Database
	if databaseName == "" {
		databaseName = s.GetDatabase()
	}
	requestExecutor := options.RequestExecutor
	if requestExecutor == nil {
		requestExecutor = s.GetRequestExecutor(databaseName)
	}
	session := NewDocumentSession(databaseName, s, sessionID, requestExecutor)
	s.registerEvents(session.InMemoryDocumentSessionOperations)
	s.afterSessionCreated(session.InMemoryDocumentSessionOperations)
	return session, nil
}

func (s *DocumentStore) ExecuteIndex(task *IndexCreationTask, database string) error {
	if err := s.assertInitialized(); err != nil {
		return err
	}
	return task.Execute(s, s.conventions, database)
}

func (s *DocumentStore) ExecuteIndexes(tasks []*IndexCreationTask, database string) error {
	if err := s.assertInitialized(); err != nil {
		return err
	}
	indexesToAdd := indexCreationCreateIndexesToAdd(tasks, s.conventions)

	op := NewPutIndexesOperation(indexesToAdd...)
	if database == "" {
		database = s.GetDatabase()
	}
	return s.Maintenance().ForDatabase(database).Send(op)
}

// GetRequestExecutor gets a request executor.
// database is optional
func (s *DocumentStore) GetRequestExecutor(database string) *RequestExecutor {
	must(s.assertInitialized())
	if database == "" {
		database = s.GetDatabase()
	}
	database = strings.ToLower(database)

	s.mu.Lock()
	executor, ok := s.requestsExecutors[database]
	s.mu.Unlock()

	if ok {
		return executor
	}

	if !s.GetConventions().IsDisableTopologyUpdates() {
		executor = RequestExecutorCreate(s.GetUrls(), database, s.Certificate, s.TrustStore, s.GetConventions())
	} else {
		executor = RequestExecutorCreateForSingleNodeWithConfigurationUpdates(s.GetUrls()[0], database, s.Certificate, s.TrustStore, s.GetConventions())
	}

	s.mu.Lock()
	s.requestsExecutors[database] = executor
	s.mu.Unlock()

	return executor
}

// Initialize initializes document Store,
// Must be called before executing any operation.
func (s *DocumentStore) Initialize() error {
	if s.initialized {
		return nil
	}
	err := s.assertValidConfiguration()
	if err != nil {
		return err
	}

	conventions := s.conventions
	if conventions.GetDocumentIDGenerator() == nil {
		generator := NewMultiDatabaseHiLoIDGenerator(s, s.GetConventions())
		s.multiDbHiLo = generator
		genID := func(dbName string, entity interface{}) (string, error) {
			return generator.GenerateDocumentID(dbName, entity)
		}
		conventions.SetDocumentIDGenerator(genID)
	}
	s.initialized = true
	return nil
}

func (s *DocumentStore) assertValidConfiguration() error {
	if len(s.urls) == 0 {
		return newIllegalArgumentError("Must provide urls to NewDocumentStore")
	}
	return nil
}

type RestoreCaching struct {
	re  *RequestExecutor
	old *AggressiveCacheOptions
}

func (r *RestoreCaching) Close() error {
	r.re.aggressiveCaching = r.old
	return nil
}

func (s *DocumentStore) DisableAggressiveCaching(databaseName string) *RestoreCaching {
	if databaseName == "" {
		databaseName = s.GetDatabase()
	}

	re := s.GetRequestExecutor(databaseName)
	old := re.aggressiveCaching
	re.aggressiveCaching = nil
	res := &RestoreCaching{
		re:  re,
		old: old,
	}
	return res
}

func (s *DocumentStore) Changes(database string) *DatabaseChanges {
	must(s.assertInitialized())

	if database == "" {
		database = s.GetDatabase()
	}

	s.mu.Lock()
	changes, ok := s.databaseChanges[database]
	s.mu.Unlock()

	if !ok {
		changes = s.createDatabaseChanges(database)

		s.mu.Lock()
		s.databaseChanges[database] = changes
		s.mu.Unlock()

	}
	return changes
}

func (s *DocumentStore) createDatabaseChanges(database string) *DatabaseChanges {
	panicIf(database == "", "database can't be empty string")
	onDispose := func() {
		s.mu.Lock()
		delete(s.databaseChanges, database)
		s.mu.Unlock()
	}
	re := s.GetRequestExecutor(database)
	return newDatabaseChanges(re, database, onDispose)
}

func (s *DocumentStore) GetLastDatabaseChangesStateError(database string) error {
	if database == "" {
		database = s.GetDatabase()
	}

	s.mu.Lock()
	databaseChanges, ok := s.databaseChanges[database]
	s.mu.Unlock()

	if !ok {
		return nil
	}
	ch := databaseChanges
	return ch.getLastConnectionStateError()
}

func (s *DocumentStore) AggressivelyCacheFor(cacheDuration time.Duration) (CancelFunc, error) {
	return s.AggressivelyCacheForDatabase(cacheDuration, "")
}

func (s *DocumentStore) AggressivelyCacheForDatabase(cacheDuration time.Duration, database string) (CancelFunc, error) {
	if database == "" {
		database = s.GetDatabase()
	}
	if database == "" {
		return nil, newIllegalArgumentError("must have database")
	}
	s.mu.Lock()
	cachingUsed := s.aggressiveCachingUsed
	s.mu.Unlock()
	if !cachingUsed {
		err := s.listenToChangesAndUpdateTheCache(database)
		if err != nil {
			return nil, err
		}
	}

	// TODO: protect access to aggressiveCaching
	opts := &AggressiveCacheOptions{
		Duration: cacheDuration,
	}
	re := s.GetRequestExecutor(database)
	oldOpts := re.aggressiveCaching
	re.aggressiveCaching = opts

	restorer := func() {
		re.aggressiveCaching = oldOpts
	}
	return restorer, nil
}

func (s *DocumentStore) listenToChangesAndUpdateTheCache(database string) error {
	s.mu.Lock()
	s.aggressiveCachingUsed = true
	evict := s.aggressiveCacheChanges[database]
	s.mu.Unlock()

	if evict != nil {
		return nil
	}

	evict, err := newEvictItemsFromCacheBasedOnChanges(s, database)
	if err != nil {
		return err
	}
	s.mu.Lock()
	s.aggressiveCacheChanges[database] = evict
	s.mu.Unlock()
	return nil
}

func (s *DocumentStore) AddBeforeCloseListener(fn func(*DocumentStore)) int {
	s.beforeClose = append(s.beforeClose, fn)
	return len(s.beforeClose) - 1
}

func (s *DocumentStore) RemoveBeforeCloseListener(idx int) {
	s.beforeClose[idx] = nil
}

func (s *DocumentStore) AddAfterCloseListener(fn func(*DocumentStore)) int {
	s.afterClose = append(s.afterClose, fn)
	return len(s.afterClose) - 1
}

func (s *DocumentStore) RemoveAfterCloseListener(idx int) {
	s.afterClose[idx] = nil
}

func (s *DocumentStore) Maintenance() *MaintenanceOperationExecutor {
	must(s.assertInitialized())

	if s.maintenanceOperationExecutor == nil {
		s.maintenanceOperationExecutor = NewMaintenanceOperationExecutor(s, "")
	}

	return s.maintenanceOperationExecutor
}

func (s *DocumentStore) Operations() *OperationExecutor {
	if s.operationExecutor == nil {
		s.operationExecutor = NewOperationExecutor(s, "")
	}

	return s.operationExecutor
}

func (s *DocumentStore) BulkInsert(database string) *BulkInsertOperation {
	if database == "" {
		database = s.GetDatabase()
	}
	return NewBulkInsertOperation(database, s)
}
