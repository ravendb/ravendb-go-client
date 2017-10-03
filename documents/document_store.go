package documents

import (
	"errors"
	ravenHttp "../http"
	"github.com/google/uuid"
)

type IDocumentStore interface{
	OpenSession() (*DocumentSession, error)
}

type DocumentStore struct{
	Urls []string
	Database string
	initialized, closed bool

	OperationsExecutor, AdminOperationsExecutor, ServerOperationsExecutor IOperationExecutor
}

func NewDocumentStore(DefaultDBName string) (*DocumentStore, error){

	return &DocumentStore{}, nil
}

func (store *DocumentStore) OpenSession() (*DocumentSession, error){
	options := NewSessionOptions(nil)
	return store.OpenSessionWithOptions(options)
}

func (store *DocumentStore) OpenSessionWithDatabase(database string) (*DocumentSession, error){
	options := NewSessionOptions(database)
	return store.OpenSessionWithOptions(options)
}

func (store *DocumentStore) OpenSessionWithOptions(options ISessionOptions) (*DocumentSession, error){
	store.AssertInitialized()
	store.EnsureNotClosed()

	sessionId := uuid.New()
	databaseName := options.GetDatabase()
	requestExecutor := options.GetRequestExecutor()
	return NewDocumentSession(databaseName, *store, sessionId, requestExecutor)
}

func (store *DocumentStore) AssertInitialized() error{
	if(store.initialized) {
		return errors.New("You cannot open a session or access the database commands before initializing the document store. Did you forget calling Initialize()?")
	}
	return nil
}

func (store *DocumentStore) EnsureNotClosed() error{
	if(store.closed) {
		return errors.New("The document store has already been disposed and cannot be used")
	}
	return nil
}

func (store *DocumentStore) Initialize() error{
	if store.initialized {
		return nil
	}
	if err := store.validateConfiguration(); err != nil{
		return err
	}

}

func (store *DocumentStore) validateConfiguration() error{
	if store.Urls == nil{
		return errors.New("Store Urls is empty")
	}
	return nil
}

func (store *DocumentStore) GetRequestExecutor() ravenHttp.RequestExecutor{

}

func (store *DocumentStore) Operations() IOperationExecutor{
	if &store.OperationsExecutor == nil{
		store.OperationsExecutor = NewOperationExecutor(store, "")
	}
	return store.OperationsExecutor
}

func (store *DocumentStore) Admin() IOperationExecutor{
	if &store.AdminOperationsExecutor == nil{
		store.AdminOperationsExecutor = NewAdminOperationExecutor(store, "")
	}
	return store.AdminOperationsExecutor
}