package ravendb

import "strings"

type OperationExecutor struct {
	store           *DocumentStore
	databaseName    String
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

func (e *OperationExecutor) forDatabase(databaseName String) *OperationExecutor {
	if strings.EqualFold(e.databaseName, databaseName) {
		return e
	}

	return NewOperationExecutorWithDatabaseName(e.store, databaseName)
}

// TODO: make arg IOperation
func (e *OperationExecutor) send(command RavenCommand) error {
	return e.sendWithSessionInfo(command, nil)
}

// TODO: make arg IOperation
// TODO: java returns a result
func (e *OperationExecutor) sendWithSessionInfo(command RavenCommand, sessionInfo *SessionInfo) error {
	return e.requestExecutor.executeCommandWithSessionInfo(command, sessionInfo)
}

//     public Operation sendAsync(IOperation<OperationIdResult> operation) {
//    public Operation sendAsync(IOperation<OperationIdResult> operation, SessionInfo sessionInfo) {
//     public PatchStatus send(PatchOperation operation, SessionInfo sessionInfo) {
//    public <TEntity> PatchOperation.Result<TEntity> send(Class<TEntity> entityClass, PatchOperation operation, SessionInfo sessionInfo) {
