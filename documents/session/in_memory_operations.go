package session

import (
	"errors"
	"../identity"
	"../../data"
	documents "../../documents"
	ravenHttp "../../http"
	"fmt"
	"net/http"
)

type ConcurrencyCheckMode uint8

const (
	AUTO ConcurrencyCheckMode = iota
	FORCED
	DISABLED
)

type InMemoryDocumentSessionOperator struct{
	generateDocumentIdsOnStore bool
	documentInfoCache map[string]DocumentInfo
	IdGenerator identity.OnClientIdGenerator

	database string
	store documents.DocumentStore
	requestExecutor ravenHttp.RequestExecutor
}

type DocumentInfo struct{
	Id string
	Etag int64
	Document map[string]map[string]interface{}
	Metadata map[string]interface{}
	ConcurrencyCheckMode ConcurrencyCheckMode
	IgnoreChanges, IsNewDocument bool
	Entity interface{}
}

func NewInMemoryDocumentSessionOperator(dbName string, documentStore documents.DocumentStore, requestExecutor ravenHttp.RequestExecutor) (*InMemoryDocumentSessionOperator, error){
	idGenerator, _ := identity.NewOnClientIdGenerator()
	return &InMemoryDocumentSessionOperator{true, make(map[string]DocumentInfo),*idGenerator, dbName, documentStore, requestExecutor}, nil
}

func NewDocumentInfo(document map[string]map[string]interface{}) (*DocumentInfo, error){
	var metadata map[string]interface{}
	metadata, ok := document[data.METADATA_KEY]
	if !ok {
		return nil, InvalidOperationError{map[string]interface{}(document), "metadata"}
	}
	id, ok := metadata[data.METADATA_ID]
	if !ok {
		return nil, InvalidOperationError{metadata, "id"}
	}
	etag, ok := metadata[data.METADATA_ID]
	if !ok {
		return nil, InvalidOperationError{metadata, "etag"}
	}
	return &DocumentInfo{string(id), int64(etag), document, metadata, FORCED, false, true,nil}, nil
}

func (sessionOperator InMemoryDocumentSessionOperator) storeInternal(entity interface{}, etag int64, id string, forceConcurrencyCheck session.ConcurrencyCheckMode) error{
	var documentInfo DocumentInfo
	if id == ""{
		if sessionOperator.generateDocumentIdsOnStore{
			id = sessionOperator.IdGenerator.GenerateDocumentIdForStorage(entity)
		}else{
			sessionOperator.rememberEntityForDocumentIdGeneration(entity);//todo should be overrided
		}
	}else{
		// Store it back into the Id field so the client has access to it
		sessionOperator.IdGenerator.TrySetIdentity(entity, id);//todo
	}
	documentInfo, ok := sessionOperator.documentInfoCache[id]
	if !ok{
		if etag != 0{
			documentInfo.Etag = etag
		}
		documentInfo.ConcurrencyCheckMode = FORCED
		return nil
	}



	return nil
}

//Stores the specified dynamic entity in the session. The entity will be saved when SaveChanges is called.
func (sessionOperator InMemoryDocumentSessionOperator) Store(entity interface{}, etag int64, id string) error{
	if entity == nil{
		return errors.New("documents: store empty object")
	}
	var concurrencyCheckMode ConcurrencyCheckMode
	switch{
	case etag == 0 && id == "":
		possibleId, ok := identity.LookupIdFromInstance(entity)
		if ok{
			id = possibleId
			concurrencyCheckMode = AUTO
		}else{
			concurrencyCheckMode = FORCED
		}
	case etag != 0 && id == "":
		concurrencyCheckMode = FORCED
	case etag == 0 && id != "":
		concurrencyCheckMode = AUTO
	case etag != 0 && id != "":
		concurrencyCheckMode = FORCED
	default:
		concurrencyCheckMode = DISABLED
	}
	return sessionOperator.storeInternal(entity, etag, id, concurrencyCheckMode)
}

//Marks the specified entity for deletion. The entity will be deleted when SaveChanges is called.
func (sessionOperator InMemoryDocumentSessionOperator) Delete(arg interface{}) error{
	return nil
}

func (sessionOperator InMemoryDocumentSessionOperator) GetDatabase() string{
	return sessionOperator.database
}

func (sessionOperator InMemoryDocumentSessionOperator) GetRequestExecutor() ravenHttp.RequestExecutor{
	return sessionOperator.requestExecutor
}

//errors

type InvalidOperationError struct{
	document map[string]interface{}
	field string
}

func (e InvalidOperationError) Error() string{
	return fmt.Sprintf("session: Document must have a %s", e.field)
}