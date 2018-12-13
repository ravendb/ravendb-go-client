package ravendb

type AdvancedSessionExtentionBase struct {
	session             *InMemoryDocumentSessionOperations
	documents           []*DocumentInfo
	requestExecutor     *RequestExecutor
	sessionInfo         *SessionInfo
	documentStore       *IDocumentStore
	deferredCommandsMap map[IdTypeAndName]ICommandData

	deletedEntities *ObjectSet
	documentsByID   *documentsByID
}

func NewAdvancedSessionExtentionBase(session *InMemoryDocumentSessionOperations) *AdvancedSessionExtentionBase {
	return &AdvancedSessionExtentionBase{
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

func (e *AdvancedSessionExtentionBase) DeferMany(commands []ICommandData) {
	e.session.DeferMany(commands)
}
