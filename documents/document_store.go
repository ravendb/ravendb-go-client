package documents

import (
	"errors"

	"github.com/google/uuid"
	"github.com/ravendb/ravendb-go-client/data"
	ravenHttp "github.com/ravendb/ravendb-go-client/http"
)

type IDocumentStore interface {
	OpenSession() (*DocumentSession, error)
}

type DocumentStore struct {
	Urls                []string
	Database, ApiKey    string
	initialized, closed bool
	requestExecutors    map[string]ravenHttp.RequestExecutor
	DocumentConventions data.DocumentConvention

	OperationsExecutor, AdminOperationsExecutor, ServerOperationsExecutor IOperationExecutor
}

func NewDocumentStore(DefaultDBName string) (*DocumentStore, error) {

	return &DocumentStore{}, nil
}

func (store *DocumentStore) OpenSession() (*DocumentSession, error) {
	options := NewSessionOptions(nil)
	return store.OpenSessionWithOptions(options)
}

func (store *DocumentStore) OpenSessionWithDatabase(database string) (*DocumentSession, error) {
	options := NewSessionOptions(database)
	return store.OpenSessionWithOptions(options)
}

func (store *DocumentStore) OpenSessionWithOptions(options ISessionOptions) (*DocumentSession, error) {
	store.AssertInitialized()
	store.EnsureNotClosed()

	sessionId := uuid.New()
	databaseName := options.GetDatabase()
	requestExecutor := options.GetRequestExecutor()
	return NewDocumentSession(databaseName, *store, sessionId, requestExecutor)
}

func (store *DocumentStore) AssertInitialized() error {
	if store.initialized {
		return errors.New("You cannot open a session or access the database commands before initializing the document store. Did you forget calling Initialize()?")
	}
	return nil
}

func (store *DocumentStore) EnsureNotClosed() error {
	if store.closed {
		return errors.New("The document store has already been disposed and cannot be used")
	}
	return nil
}

func (store *DocumentStore) Initialize() error {
	if store.initialized {
		return nil
	}
	if err := store.validateConfiguration(); err != nil {
		return err
	}

}

func (store *DocumentStore) validateConfiguration() error {
	if store.Urls == nil {
		return errors.New("Store Urls is empty")
	}
	return nil
}

func (store *DocumentStore) GetRequestExecutor(database string) ravenHttp.RequestExecutor {
	if database == "" {
		database = store.Database
	}
	_, ok := store.requestExecutors[database]
	if !ok {
		executor, err := ravenHttp.NewRequestExecutor(database, store.ApiKey)
		if store.DocumentConventions.DisableTopologyUpdates == false {
			executor.Create(store.Urls, database, store.ApiKey)
		} else {
			executor.CreateForSingleNode(store.Urls[0], database, store.ApiKey)
		}
		store.requestExecutors[database] = *executor
	}
	return store.requestExecutors[database]
}

func (store *DocumentStore) Operations() IOperationExecutor {
	if &store.OperationsExecutor == nil {
		store.OperationsExecutor = NewOperationExecutor(store, "")
	}
	return store.OperationsExecutor
}

func (store *DocumentStore) Admin() IOperationExecutor {
	if &store.AdminOperationsExecutor == nil {
		store.AdminOperationsExecutor = NewAdminOperationExecutor(store, "")
	}
	return store.AdminOperationsExecutor
}
