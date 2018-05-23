package ravendb

import (
	"reflect"
)

// ObjectNode is an alias for a json document represented as a map
// Name comes from Java implementation
type ObjectNode = map[string]interface{}

type JsonNodeType = interface{}

// TODO: remove it, it only exists to make initial porting faster
type Object = interface{}

// TODO: remove it, it only exists to make initial porting faster
type String = string

// DocumentSession is a Unit of Work for accessing RavenDB server
// https://sourcegraph.com/github.com/ravendb/RavenDB-Python-Client/-/blob/pyravendb/store/document_session.py#L18
// https://sourcegraph.com/github.com/ravendb/RavenDB-Python-Client/-/blob/pyravendb/store/document_session.py#L18
type DocumentSession struct {
	*InMemoryDocumentSessionOperations

	// _attachments *IAttachmentsSessionOperations
	// _revisions *IRevisionsSessionOperations
}

//    public IAdvancedSessionOperations advanced() {
//    public ILazySessionOperations lazily() {
//    public IEagerSessionOperations eagerly() {
//    public IAttachmentsSessionOperations attachments() {
//    public IRevisionsSessionOperations revisions() {

// NewDocumentSession creates a new DocumentSession
func NewDocumentSession(dbName string, store *DocumentStore, id string, re *RequestExecutor) *DocumentSession {
	res := &DocumentSession{
		InMemoryDocumentSessionOperations: NewInMemoryDocumentSessionOperations(dbName, store, re, id),
	}

	//res._attachments: NewDocumentSessionAttachments(res)
	//res._revisions = NewDocumentSessionRevisions(res)

	return res
}

func (s *DocumentSession) SaveChanges() error {
	saveChangeOperation := NewBatchOperation(s.InMemoryDocumentSessionOperations)

	command := saveChangeOperation.createRequest()
	if command == nil {
		return nil
	}
	exec := s.RequestExecutor.GetCommandExecutor(false)
	result, err := ExecuteBatchCommand(exec, command)
	if err != nil {
		return err
	}
	saveChangeOperation.setResult(result)
	return nil
}

// https://sourcegraph.com/github.com/ravendb/ravendb-jvm-client@v4.0/-/blob/src/main/java/net/ravendb/client/documents/session/InMemoryDocumentSessionOperations.java#L665
// https://sourcegraph.com/github.com/ravendb/RavenDB-Python-Client@v4.0/-/blob/pyravendb/store/document_session.py#L101
func (s *DocumentSession) saveEntity(key string, entity interface{}, originalMetadata map[string]interface{}, metadata map[string]interface{}, document *DocumentInfo, concurrencyCheckMode ConcurrencyCheckMode) {
	// Note: key can be empty
	delete(s.deletedEntities, entity)
	if key != "" {
		delete(s._knownMissingIds, key)
		if v := s.documentsById.getValue(key); v == nil {
			return
		}
	}
	if key != "" {
		s.documentsById.add(document)
	}
	if document == nil {
		document = &DocumentInfo{}
	}
	document.id = key
	document.metadata = metadata
	document.changeVector = ""
	document.concurrencyCheckMode = concurrencyCheckMode
	document.entity = entity
	document.newDocument = true
	s.documentsByEntity[entity] = document
}

func (s *DocumentSession) convertAndSaveEntity(key string, document ObjectNode, objectType reflect.Value) {
	if v := s.documentsById.getValue(key); v == nil {
		return
	}
	// TODO: convert_to_entity
	panicIf(true, "NYI")
}
