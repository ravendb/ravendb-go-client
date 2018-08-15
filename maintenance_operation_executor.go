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
		databaseName: firstNonEmptyString(databaseName, store.GetDatabase()),
	}
	if res.databaseName != "" {
		res.requestExecutor = store.GetRequestExecutorWithDatabase(res.databaseName)
	}
	return res
}

func (e *MaintenanceOperationExecutor) Server() *ServerOperationExecutor {
	if e.serverOperationExecutor == nil {
		e.serverOperationExecutor = NewServerOperationExecutor(e.store)
	}
	return e.serverOperationExecutor
}

func (e *MaintenanceOperationExecutor) ForDatabase(databaseName string) *MaintenanceOperationExecutor {
	if strings.EqualFold(e.databaseName, databaseName) {
		return e
	}
	return NewMaintenanceOperationExecutorWithDatabase(e.store, databaseName)
}

func (e *MaintenanceOperationExecutor) Send(operation IMaintenanceOperation) error {
	err := e.assertDatabaseNameSet()
	if err != nil {
		return err
	}
	command := operation.GetCommand(e.requestExecutor.getConventions())
	err = e.requestExecutor.ExecuteCommand(command)
	return err
}

func (e *MaintenanceOperationExecutor) assertDatabaseNameSet() error {
	if e.databaseName == "" {
		return NewIllegalStateException("Cannot use maintenance without a database defined, did you forget to call forDatabase?")
	}
	return nil
}
