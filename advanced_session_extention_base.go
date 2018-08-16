package ravendb

type AdvancedSessionExtentionBase struct {
	session             *InMemoryDocumentSessionOperations
	documentsByEntity   map[interface{}]*DocumentInfo
	requestExecutor     *RequestExecutor
	sessionInfo         *SessionInfo
	documentStore       *IDocumentStore
	deferredCommandsMap map[IdTypeAndName]ICommandData

	deletedEntities *ObjectSet
	documentsById   *DocumentsById
}

func NewAdvancedSessionExtentionBase(session *InMemoryDocumentSessionOperations) *AdvancedSessionExtentionBase {
	return &AdvancedSessionExtentionBase{
		session:             session,
		documentsByEntity:   session.documentsByEntity,
		requestExecutor:     session.GetRequestExecutor(),
		sessionInfo:         session.sessionInfo,
		documentStore:       session.GetDocumentStore(),
		deferredCommandsMap: session.deferredCommandsMap,
		deletedEntities:     session.deletedEntities,
		documentsById:       session.documentsById,
	}
}

func (e *AdvancedSessionExtentionBase) DeferMany(commands []ICommandData) {
	e.session.DeferMany(commands)
}
