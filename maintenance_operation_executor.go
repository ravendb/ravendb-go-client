package ravendb

import "strings"

type MaintenanceOperationExecutor struct {
	store                   *DocumentStore
	databaseName            String
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
		res.requestExecutor = store.GetRequestExecutor(res.databaseName)
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

// NOTE: send() can't be done without generics. Instead use:
// Execute*Command(server().GetRequestExecutor())
