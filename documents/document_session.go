package documents

import (
	"github.com/google/uuid"
	"github.com/ravendb/ravendb-go-client/http"
)

type IDocumentSession interface {
	SaveChanges() error
	Store(interface{}, int64, string) error
	Delete(interface{}) error
	Load(string) (interface{}, bool)
}

type ISessionOptions interface {
	GetDatabase() string
	GetRequestExecutor() http.RequestExecutor
}

type DocumentSession struct {
	inMemoryDocumentSessionOperator InMemoryDocumentSessionOperator
}

type SessionOptions struct {
	database        string
	requestExecutor http.RequestExecutor
}

func NewDocumentSession(dbName string, store DocumentStore, id uuid.UUID, requestExecutor http.RequestExecutor) (*DocumentSession, error) {
	inMemoryDocumentSessionOperator, err := NewInMemoryDocumentSessionOperator(dbName, requestExecutor)
	return &DocumentSession{*inMemoryDocumentSessionOperator}, err
}

func NewSessionOptions(database string, requestExecutor http.RequestExecutor) *SessionOptions {
	return &SessionOptions{database, requestExecutor}
}

func (sessionOperator SessionOptions) GetDatabase() string {
	return sessionOperator.database
}

func (sessionOperator SessionOptions) GetRequestExecutor() http.RequestExecutor {
	return sessionOperator.requestExecutor
}

//Saves all the pending changes to the server.
func (documentSession DocumentSession) SaveChanges() {
	//todo
}

//Stores the specified dynamic entity in the session. The entity will be saved when SaveChanges is called.
func (documentSession DocumentSession) Store(object interface{}, etag int64, id string) error {
	return documentSession.inMemoryDocumentSessionOperator.Store(object, etag, id)
}

//Marks the specified entity for deletion. The entity will be deleted when SaveChanges is called.
func (documentSession DocumentSession) Delete(arg interface{}) error {
	return documentSession.inMemoryDocumentSessionOperator.Delete(arg)
}

func (documentSession DocumentSession) DeleteById(id string, expectedChangeVector string) error {
	return documentSession.inMemoryDocumentSessionOperator.Delete(id)
}

func (documentSession DocumentSession) Load(id string) (interface{}, bool) {
	return nil, false
}
