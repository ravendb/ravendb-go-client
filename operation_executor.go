package ravendb

import "strings"

type OperationExecutor struct {
	store           *DocumentStore
	databaseName    string
	requestExecutor *RequestExecutor
}

func NewOperationExecutor(store *DocumentStore) *OperationExecutor {
	return NewOperationExecutorWithDatabaseName(store, "")
}

func NewOperationExecutorWithDatabaseName(store *DocumentStore, databaseName string) *OperationExecutor {
	res := &OperationExecutor{
		store:        store,
		databaseName: databaseName,
	}
	if res.databaseName == "" {
		res.databaseName = store.getDatabase()
	}
	panicIf(res.databaseName == "", "databaseName is empty")
	res.requestExecutor = store.GetRequestExecutorWithDatabase(res.databaseName)
	return res
}

func (e *OperationExecutor) forDatabase(databaseName string) *OperationExecutor {
	if strings.EqualFold(e.databaseName, databaseName) {
		return e
	}

	return NewOperationExecutorWithDatabaseName(e.store, databaseName)
}

func (e *OperationExecutor) send(operation IOperation) error {
	return e.sendWithSessionInfo(operation, nil)
}

// Note: we don't return a result because we could only return interface{}
// The caller has access to operation and can access strongly typed
// command and its result
func (e *OperationExecutor) sendWithSessionInfo(operation IOperation, sessionInfo *SessionInfo) error {
	command := operation.getCommand(e.store, e.requestExecutor.getConventions(), e.requestExecutor.getCache())
	return e.requestExecutor.executeCommandWithSessionInfo(command, sessionInfo)
}

//     public Operation sendAsync(IOperation<OperationIdResult> operation) {
//    public Operation sendAsync(IOperation<OperationIdResult> operation, SessionInfo sessionInfo) {
//     public PatchStatus send(PatchOperation operation, SessionInfo sessionInfo) {
//    public <TEntity> PatchOperation.Result<TEntity> send(Class<TEntity> entityClass, PatchOperation operation, SessionInfo sessionInfo) {
