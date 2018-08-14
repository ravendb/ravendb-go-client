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
	onBeforeStore      []func(interface{}, *BeforeStoreEventArgs)
	onAfterSaveChanges []func(interface{}, *AfterSaveChangesEventArgs)

	onBeforeDelete   []func(interface{}, *BeforeDeleteEventArgs)
	onBeforeQuery    []func(interface{}, *BeforeQueryEventArgs)
	onSessionCreated []func(interface{}, *SessionCreatedEventArgs)

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
func (s *DocumentStore) GetConventions() *DocumentConventions {
	if s.conventions == nil {
		s.conventions = NewDocumentConventions()
	}
	return s.conventions
}

func (s *DocumentStore) SetConventions(conventions *DocumentConventions) {
	s.conventions = conventions
}

func (s *DocumentStore) GetUrls() []string {
	return s.urls
}

func (s *DocumentStore) SetUrls(value []string) {
	panicIf(len(value) == 0, "value is empty")
	for i, s := range value {
		value[i] = strings.TrimSuffix(s, "/")
	}
	s.urls = value
}

func (s *DocumentStore) ensureNotClosed() {
	// TODO: implement me
}

func (s *DocumentStore) AddBeforeStoreListener(handler func(interface{}, *BeforeStoreEventArgs)) {
	s.onBeforeStore = append(s.onBeforeStore, handler)

}
func (s *DocumentStore) RemoveBeforeStoreListener(handler func(interface{}, *BeforeStoreEventArgs)) {
	panic("NYI")
	//this.onBeforeStore.remove(handler);
}

func (s *DocumentStore) AddAfterSaveChangesListener(handler func(interface{}, *AfterSaveChangesEventArgs)) {
	s.onAfterSaveChanges = append(s.onAfterSaveChanges, handler)
}

func (s *DocumentStore) RemoveAfterSaveChangesListener(handler func(interface{}, *AfterSaveChangesEventArgs)) {
	panic("NYI")
	//this.onAfterSaveChanges.remove(handler);
}

func (s *DocumentStore) AddBeforeDeleteListener(handler func(interface{}, *BeforeDeleteEventArgs)) {
	s.onBeforeDelete = append(s.onBeforeDelete, handler)
}

func (s *DocumentStore) RemoveBeforeDeleteListener(handler func(interface{}, *BeforeDeleteEventArgs)) {
	panic("NYI")
	//this.onBeforeDelete.remove(handler);
}

func (s *DocumentStore) AddBeforeQueryListener(handler func(interface{}, *BeforeQueryEventArgs)) {
	s.onBeforeQuery = append(s.onBeforeQuery, handler)
}

func (s *DocumentStore) RemoveBeforeQueryListener(handler func(interface{}, *BeforeQueryEventArgs)) {
	panic("NYI")
	//this.onBeforeQuery.remove(handler);
}

func (s *DocumentStore) RegisterEvents(session *InMemoryDocumentSessionOperations) {
	for _, handler := range s.onBeforeStore {
		session.addBeforeStoreListener(handler)
	}

	for _, handler := range s.onAfterSaveChanges {
		session.addAfterSaveChangesListener(handler)
	}

	for _, handler := range s.onBeforeDelete {
		session.addBeforeDeleteListener(handler)
	}

	for _, handler := range s.onBeforeQuery {
		session.addBeforeQueryListener(handler)
	}
}

func (s *DocumentStore) afterSessionCreated(session *InMemoryDocumentSessionOperations) {
	for _, handler := range s.onSessionCreated {
		handler(s, NewSessionCreatedEventArgs(session))
	}
}

func (s *DocumentStore) assertInitialized() {
	panicIf(!s.initialized, "DocumentStore must be initialized")
}

func (s *DocumentStore) GetDatabase() string {
	return s.database
}

func (s *DocumentStore) SetDatabase(database string) {
	panicIf(s.initialized, "is already initialized")
	s.database = database
}

func (s *DocumentStore) GetCertificate() *KeyStore {
	return s._certificate
}

func (s *DocumentStore) SetCertificate(certificate *KeyStore) {
	panicIf(s.initialized, "is already initialized")
	s._certificate = certificate
}

func (s *DocumentStore) AggressivelyCache() {
	s.AggressivelyCacheWithDatabase("")
}

func (s *DocumentStore) AggressivelyCacheWithDatabase(database string) {
	s.AggressivelyCacheForDatabase(time.Hour*24, database)
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
	res.SetUrls([]string{url})
	res.SetDatabase(database)
	return res
}

func NewDocumentStoreWithUrlsAndDatabase(urls []string, database string) *DocumentStore {
	res := NewDocumentStore()
	res.SetUrls(urls)
	res.SetDatabase(database)
	return res
}

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

// Close closes the store
func (s *DocumentStore) Close() {
	if s.disposed {
		return
	}

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
		re.Close()
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
	databaseName := firstNonEmptyString(options.getDatabase(), s.GetDatabase())
	requestExecutor := options.getRequestExecutor()
	if requestExecutor == nil {
		requestExecutor = s.GetRequestExecutorWithDatabase(databaseName)
	}
	session := NewDocumentSession(databaseName, s, sessionID, requestExecutor)
	s.RegisterEvents(session.InMemoryDocumentSessionOperations)
	s.afterSessionCreated(session.InMemoryDocumentSessionOperations)
	return session, nil
}

func (s *DocumentStore) ExecuteIndex(task *AbstractIndexCreationTask) error {
	return s.ExecuteIndexWithDatabase(task, "")
}

func (s *DocumentStore) ExecuteIndexWithDatabase(task *AbstractIndexCreationTask, database string) error {
	s.assertInitialized()
	return task.execute2(s, s.conventions, database)
}

func (s *DocumentStore) ExecuteIndexes(tasks []*AbstractIndexCreationTask) error {
	return s.ExecuteIndexesWithDatabase(tasks, "")
}

func (s *DocumentStore) ExecuteIndexesWithDatabase(tasks []*AbstractIndexCreationTask, database string) error {
	s.assertInitialized()
	indexesToAdd := IndexCreation_createIndexesToAdd(tasks, s.conventions)

	op := NewPutIndexesOperation(indexesToAdd...)
	if database == "" {
		database = s.GetDatabase()
	}
	return s.Maintenance().forDatabase(database).send(op)
}

func (s *DocumentStore) GetRequestExecutor() *RequestExecutor {
	return s.GetRequestExecutorWithDatabase("")
}

// GetRequestExecutorWithDatabase gets a request executor for a given database
func (s *DocumentStore) GetRequestExecutorWithDatabase(database string) *RequestExecutor {
	s.assertInitialized()
	if database == "" {
		database = s.GetDatabase()
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	executor, ok := s.requestsExecutors[database]
	if ok {
		return executor
	}

	if !s.GetConventions().isDisableTopologyUpdates() {
		executor = RequestExecutor_create(s.GetUrls(), s.GetDatabase(), s.GetCertificate(), s.GetConventions())
	} else {
		executor = RequestExecutor_createForSingleNodeWithConfigurationUpdates(s.GetUrls()[0], s.GetDatabase(), s.GetCertificate(), s.GetConventions())
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
		generator := NewMultiDatabaseHiLoIdGenerator(s, s.GetConventions())
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

type RestoreCaching struct {
	re  *RequestExecutor
	old *AggressiveCacheOptions
}

func (r *RestoreCaching) Close() {
	// TODO: in Java it's thread local
	r.re.aggressiveCaching = r.old
}

func (s *DocumentStore) DisableAggressiveCaching() *RestoreCaching {
	return s.DisableAggressiveCachingWithDatabase("")
}

func (s *DocumentStore) DisableAggressiveCachingWithDatabase(databaseName string) *RestoreCaching {
	if databaseName == "" {
		databaseName = s.GetDatabase()
	}

	re := s.GetRequestExecutorWithDatabase(databaseName)
	old := re.aggressiveCaching // TODO: is thread local
	re.aggressiveCaching = nil  // TODO: is thread local
	res := &RestoreCaching{
		re:  re,
		old: old,
	}
	return res
}

func (s *DocumentStore) Changes() *IDatabaseChanges {
	// TODO: implement me
	return nil
}

//    public IDatabaseChanges changes(string database) {
//    protected IDatabaseChanges createDatabaseChanges(string database) {
//     public Exception getLastDatabaseChangesStateException() {
//    public Exception getLastDatabaseChangesStateException(string database) {

func (s *DocumentStore) AggressivelyCacheFor(cacheDuration time.Duration) {
	// TODO: implement me
}

func (s *DocumentStore) AggressivelyCacheForDatabase(cacheDuration time.Duration, database string) {
	// TODO: implement me
}

//    private void listenToChangesAndUpdateTheCache(string database) {

func (s *DocumentStore) AddBeforeCloseListener(fn func(*DocumentStore)) {
	s.beforeClose = append(s.beforeClose, fn)
}

//   public void removeBeforeCloseListener(EventHandler<VoidArgs> event) {

func (s *DocumentStore) AddAfterCloseListener(fn func(*DocumentStore)) {
	s.afterClose = append(s.afterClose, fn)
}

//    public void removeAfterCloseListener(EventHandler<VoidArgs> event) {

func (s *DocumentStore) Maintenance() *MaintenanceOperationExecutor {
	s.assertInitialized()

	if s.maintenanceOperationExecutor == nil {
		s.maintenanceOperationExecutor = NewMaintenanceOperationExecutor(s)
	}

	return s.maintenanceOperationExecutor
}

func (s *DocumentStore) Operations() *OperationExecutor {
	if s.operationExecutor == nil {
		s.operationExecutor = NewOperationExecutor(s)
	}

	return s.operationExecutor
}

func (s *DocumentStore) BulkInsert() *BulkInsertOperation {
	return s.BulkInsertWithDatabase("")
}

func (s *DocumentStore) BulkInsertWithDatabase(database string) *BulkInsertOperation {
	if database == "" {
		database = s.GetDatabase()
	}
	return NewBulkInsertOperation(database, s)
}
