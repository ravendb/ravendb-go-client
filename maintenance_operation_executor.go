package ravendb

import "strings"

type MaintenanceOperationExecutor struct {
	store                   *DocumentStore
	databaseName            string
	requestExecutor         *RequestExecutor
	serverOperationExecutor *ServerOperationExecutor
}

func NewMaintenanceOperationExecutor(store *DocumentStore) *MaintenanceOperationExecutor {
	return NewMaintenanceOperationExecutorWithDatabase(store, "")
}

func NewMaintenanceOperationExecutorWithDatabase(store *DocumentStore, databaseName string) *MaintenanceOperationExecutor {

	res := &MaintenanceOperationExecutor{
		store:        store,
		databaseName: firstNonEmptyString(databaseName, store.getDatabase()),
	}
	if res.databaseName != "" {
		res.requestExecutor = store.GetRequestExecutorWithDatabase(res.databaseName)
	}
	return res
}

func (e *MaintenanceOperationExecutor) server() *ServerOperationExecutor {
	if e.serverOperationExecutor == nil {
		e.serverOperationExecutor = NewServerOperationExecutor(e.store)
	}
	return e.serverOperationExecutor
}

func (e *MaintenanceOperationExecutor) forDatabase(databaseName string) *MaintenanceOperationExecutor {
	if strings.EqualFold(e.databaseName, databaseName) {
		return e
	}
	return NewMaintenanceOperationExecutorWithDatabase(e.store, databaseName)
}

func (e *MaintenanceOperationExecutor) send(operation IMaintenanceOperation) error {
	err := e.assertDatabaseNameSet()
	if err != nil {
		return err
	}
	command := operation.getCommand(e.requestExecutor.getConventions())
	err = e.requestExecutor.executeCommand(command)
	return err
}

func (e *MaintenanceOperationExecutor) assertDatabaseNameSet() error {
	if e.databaseName == "" {
		return NewIllegalStateException("Cannot use maintenance without a database defined, did you forget to call forDatabase?")
	}
	return nil
}
