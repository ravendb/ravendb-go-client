package ravendb

import (
	"net/http"
	"strings"
)

type OperationExecutor struct {
	store           *DocumentStore
	databaseName    string
	requestExecutor *RequestExecutor
}

func NewOperationExecutor(store *DocumentStore, databaseName string) *OperationExecutor {
	res := &OperationExecutor{
		store:        store,
		databaseName: databaseName,
	}
	if res.databaseName == "" {
		res.databaseName = store.GetDatabase()
	}
	panicIf(res.databaseName == "", "databaseName is empty")
	res.requestExecutor = store.GetRequestExecutor(res.databaseName)
	return res
}

func (e *OperationExecutor) ForDatabase(databaseName string) *OperationExecutor {
	if strings.EqualFold(e.databaseName, databaseName) {
		return e
	}

	return NewOperationExecutor(e.store, databaseName)
}

// Note: we don't return a result because we could only return interface{}
// The caller has access to operation and can access strongly typed
// command and its result
// sessionInfo can be nil
func (e *OperationExecutor) Send(operation IOperation, sessionInfo *SessionInfo) error {
	command, err := operation.GetCommand(e.store, e.requestExecutor.GetConventions(), e.requestExecutor.Cache)
	if err != nil {
		return err
	}
	return e.requestExecutor.ExecuteCommand(command, sessionInfo)
}

// sessionInfo can be nil
func (e *OperationExecutor) SendAsync(operation IOperation, sessionInfo *SessionInfo) (*Operation, error) {
	command, err := operation.GetCommand(e.store, e.requestExecutor.GetConventions(), e.requestExecutor.Cache)
	if err != nil {
		return nil, err
	}

	if err = e.requestExecutor.ExecuteCommand(command, sessionInfo); err != nil {
		return nil, err
	}

	changes := func() *DatabaseChanges {
		return e.store.Changes("")
	}
	result := getCommandOperationIDResult(command)

	return NewOperation(e.requestExecutor, changes, e.requestExecutor.GetConventions(), result.OperationID), nil
}

// Note: use SendPatchOperation() instead and check PatchOperationResult.Status
// public PatchStatus send(PatchOperation operation) {
// public PatchStatus send(PatchOperation operation, SessionInfo sessionInfo) {

func (e *OperationExecutor) SendPatchOperation(operation *PatchOperation, sessionInfo *SessionInfo) (*PatchOperationResult, error) {
	conventions := e.requestExecutor.GetConventions()
	cache := e.requestExecutor.Cache
	command, err := operation.GetCommand(e.store, conventions, cache)
	if err != nil {
		return nil, err
	}
	if err = e.requestExecutor.ExecuteCommand(command, sessionInfo); err != nil {
		return nil, err
	}

	cmdResult := operation.Command.Result
	result := &PatchOperationResult{
		Status:   cmdResult.Status,
		Document: cmdResult.ModifiedDocument,
	}
	switch operation.Command.StatusCode {
	case http.StatusNotModified:
		result.Status = PatchStatusNotModified
	case http.StatusNotFound:
		result.Status = PatchStatusDocumentDoesNotExist
	}
	return result, nil
}
