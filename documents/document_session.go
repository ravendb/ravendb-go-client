package documents

import (
	"../http"
	"./session"
	"github.com/google/uuid"
)

type IDocumentSession interface{
	SaveChanges() error
	Store(interface{}, int64, string) error
	Delete(interface{}) error
}

type DocumentSession struct{
	inMemoryDocumentSessionOperator session.InMemoryDocumentSessionOperator
}

func NewDocumentSession(dbName string, store DocumentStore, id uuid.UUID, requestExecutor http.RequestExecutor) (*DocumentSession, error){
	return &DocumentSession{}, nil
}

//Saves all the pending changes to the server.
func (documentSession DocumentSession) SaveChanges(){
//todo
}

//Stores the specified dynamic entity in the session. The entity will be saved when SaveChanges is called.
func (documentSession DocumentSession) Store(object interface{}, etag int64, id string) error{
	return documentSession.inMemoryDocumentSessionOperator.Store(object, etag, id)
}

//Marks the specified entity for deletion. The entity will be deleted when SaveChanges is called.
func (documentSession DocumentSession) Delete(arg interface{}) error{
	return documentSession.inMemoryDocumentSessionOperator.Delete(arg)
}

