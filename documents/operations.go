package documents

import (
	"../http"
)

type IOperation interface{
	GetCommand(store IDocumentStore) http.RavenRequestable
}

type IOperationExecutor interface{
	Send(operation IOperation) error
	GetStore() IDocumentStore
}

type OperationExecutor struct{
	store IDocumentStore
	databaseName string
	requestExecutor http.RequestExecutor
}

type AdminOperationExecutor struct{
	operationExecutor, serverOperationExecutor IOperationExecutor
}

type ServerOperationExecutor struct{

}

func NewOperationExecutor(store IDocumentStore, databaseName string) *OperationExecutor{
	return &OperationExecutor{store:store, databaseName:databaseName, requestExecutor: store.GetRequestExecutor(databaseName)}
}

func NewAdminOperationExecutor(store IDocumentStore, databaseName string) *AdminOperationExecutor{
	operationExecutor := NewOperationExecutor(store, databaseName)
	return &AdminOperationExecutor{operationExecutor, nil}
}

func NewServerOperationExecutor(store IDocumentStore, databaseName string) *OperationExecutor{
	return &OperationExecutor{store:store, databaseName:databaseName, requestExecutor: store.GetRequestExecutor(databaseName)}
}

func (executor OperationExecutor) Send(operation IOperation) error{
	command := operation.GetCommand(executor.store)
	return executor.requestExecutor.ExecuteOnCurrentNode(command)
}

func (executor OperationExecutor) GetStore() IDocumentStore{
	return executor.store
}

func (executor AdminOperationExecutor) GetStore() IDocumentStore{
	return executor.operationExecutor.GetStore()
}

func (executor AdminOperationExecutor) Send(operation IOperation) error{
	return executor.operationExecutor.Send(operation)
}

func (executor AdminOperationExecutor) Server(operation IOperation) IOperationExecutor{
	if &executor.serverOperationExecutor == nil{
		executor.serverOperationExecutor = NewServerOperationExecutor(executor.GetStore(), "")
	}
	return executor.serverOperationExecutor
}