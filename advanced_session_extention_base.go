package ravendb

// AdvancedSessionExtensionBase implements common advanced session operations
type AdvancedSessionExtensionBase struct {
	session             *InMemoryDocumentSessionOperations
	documents           []*documentInfo
	requestExecutor     *RequestExecutor
	sessionInfo         *SessionInfo
	documentStore       *DocumentStore
	deferredCommandsMap map[idTypeAndName]ICommandData

	deletedEntities *objectSet
	documentsByID   *documentsByID
}

func newAdvancedSessionExtensionBase(session *InMemoryDocumentSessionOperations) *AdvancedSessionExtensionBase {
	return &AdvancedSessionExtensionBase{
		session:             session,
		documents:           session.documents,
		requestExecutor:     session.GetRequestExecutor(),
		sessionInfo:         session.sessionInfo,
		documentStore:       session.GetDocumentStore(),
		deferredCommandsMap: session.deferredCommandsMap,
		deletedEntities:     session.deletedEntities,
		documentsByID:       session.documentsByID,
	}
}

// Defer defers multiple commands to be executed on SaveChnages
func (e *AdvancedSessionExtensionBase) Defer(commands ...ICommandData) {
	e.session.Defer(commands...)
}
