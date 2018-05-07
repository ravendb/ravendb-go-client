package ravendb

// DocumentSession is a Unit of Work for accessing RavenDB server
// https://sourcegraph.com/github.com/ravendb/RavenDB-Python-Client/-/blob/pyravendb/store/document_session.py#L18
// https://sourcegraph.com/github.com/ravendb/RavenDB-Python-Client/-/blob/pyravendb/store/document_session.py#L18
type DocumentSession struct {
	SessionID        string
	Database         string
	documentStore    *DocumentStore
	requestsExecutor *RequestsExecutor
	// TODO: move rields
	// documentsByID map[string] ??
}

// NewDocumentSession creates a new DocumentSession
func NewDocumentSession(dbName string, documentStore *DocumentStore, id string, re *RequestsExecutor) *DocumentSession {
	res := &DocumentSession{
		SessionID:        id,
		Database:         dbName,
		documentStore:    documentStore,
		requestsExecutor: re,
	}
	return res
}
