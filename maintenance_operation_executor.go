package ravendb

import "strings"

type MaintenanceOperationExecutor struct {
	store                   *DocumentStore
	databaseName            string
	requestExecutor         *RequestExecutor
	serverOperationExecutor *ServerOperationExecutor
}

func NewMaintenanceOperationExecutor(store *DocumentStore, databaseName string) *MaintenanceOperationExecutor {

	res := &MaintenanceOperationExecutor{
		store:        store,
		databaseName: firstNonEmptyString(databaseName, store.GetDatabase()),
	}
	if res.databaseName != "" {
		res.requestExecutor = store.GetRequestExecutor(res.databaseName)
	}
	return res
}

func (e *MaintenanceOperationExecutor) GetRequestExecutor() *RequestExecutor {
	if e.requestExecutor != nil {
		return e.requestExecutor
	}
	if e.databaseName != "" {
		e.requestExecutor = e.store.GetRequestExecutor(e.databaseName)
	}
	return e.requestExecutor
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
	return NewMaintenanceOperationExecutor(e.store, databaseName)
}

func (e *MaintenanceOperationExecutor) Send(operation IMaintenanceOperation) error {
	if err := e.assertDatabaseNameSet(); err != nil {
		return err
	}
	command, err := operation.GetCommand(e.GetRequestExecutor().GetConventions())
	if err != nil {
		return err
	}
	return e.GetRequestExecutor().ExecuteCommand(command, nil)
}

func (e *MaintenanceOperationExecutor) SendAsync(operation IMaintenanceOperation) (*Operation, error) {
	if err := e.assertDatabaseNameSet(); err != nil {
		return nil, err
	}
	command, err := operation.GetCommand(e.GetRequestExecutor().GetConventions())
	if err != nil {
		return nil, err
	}
	if err = e.GetRequestExecutor().ExecuteCommand(command, nil); err != nil {
		return nil, err
	}
	fn := func() *DatabaseChanges {
		return e.store.Changes("")
	}
	re := e.GetRequestExecutor()
	id := getCommandOperationIDResult(command)
	return NewOperation(re, fn, re.GetConventions(), id.OperationID), nil
}

func (e *MaintenanceOperationExecutor) assertDatabaseNameSet() error {
	if e.databaseName == "" {
		return newIllegalStateError("Cannot use maintenance without a database defined, did you forget to call forDatabase?")
	}
	return nil
}
