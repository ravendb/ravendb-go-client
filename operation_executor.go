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
		res.databaseName = store.GetDatabase()
	}
	panicIf(res.databaseName == "", "databaseName is empty")
	res.requestExecutor = store.GetRequestExecutorWithDatabase(res.databaseName)
	return res
}

func (e *OperationExecutor) ForDatabase(databaseName string) *OperationExecutor {
	if strings.EqualFold(e.databaseName, databaseName) {
		return e
	}

	return NewOperationExecutorWithDatabaseName(e.store, databaseName)
}

func (e *OperationExecutor) Send(operation IOperation) error {
	return e.SendWithSessionInfo(operation, nil)
}

// Note: we don't return a result because we could only return interface{}
// The caller has access to operation and can access strongly typed
// command and its result
func (e *OperationExecutor) SendWithSessionInfo(operation IOperation, sessionInfo *SessionInfo) error {
	command := operation.GetCommand(e.store, e.requestExecutor.GetConventions(), e.requestExecutor.Cache)
	return e.requestExecutor.ExecuteCommandWithSessionInfo(command, sessionInfo)
}

func (e *OperationExecutor) SendAsync(operation IOperation) (*Operation, error) {
	return e.SendAsyncWithSessionInfo(operation, nil)
}

func (e *OperationExecutor) SendAsyncWithSessionInfo(operation IOperation, sessionInfo *SessionInfo) (*Operation, error) {
	command := operation.GetCommand(e.store, e.requestExecutor.GetConventions(), e.requestExecutor.Cache)

	err := e.requestExecutor.ExecuteCommandWithSessionInfo(command, sessionInfo)
	if err != nil {
		return nil, err
	}

	changes := func() *DatabaseChanges {
		return e.store.Changes()
	}
	result := getCommandOperationIDResult(command)

	return NewOperation(e.requestExecutor, changes, e.requestExecutor.GetConventions(), result.getOperationId()), nil

}

//     public PatchStatus send(PatchOperation operation, SessionInfo sessionInfo) {
//    public <TEntity> PatchOperation.Result<TEntity> send(Class<TEntity> entityClass, PatchOperation operation, SessionInfo sessionInfo) {
