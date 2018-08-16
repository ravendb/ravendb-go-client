package ravendb

import "sync"

var (
	_ IDatabaseChanges = &DatabaseChanges{}
)

type DatabaseChanges struct {
	_commandId int

	//Semaphore _semaphore = new Semaphore(1);

	_requestExecutor *RequestExecutor
	_conventions     *DocumentConventions
	_database        string

	_onDispose Runnable

	//_client *WebSocketClient
	//_clientSession *Session
	//_processor *WebSocketChangesProcessor

	_task *CompletableFuture
	_cts  *CancellationTokenSource
	_tcs  *CompletableFuture

	_confirmations sync.Map // int => *CompletableFuture
	_counters      sync.Map // toLower(string) -> *DatabaseConnectionState

	_immediateConnection AtomicInteger
}

func NewDatabaseChanges(requestExecutor *RequestExecutor, databaseName string, onDispose Runnable) *DatabaseChanges {
	res := &DatabaseChanges{
		_requestExecutor: requestExecutor,
		_conventions:     requestExecutor.GetConventions(),
		_database:        databaseName,
		_tcs:             NewCompletableFuture(),
		_cts:             NewCancellationTokenSource(),
		_onDispose:       onDispose,
	}

	//res._client = res.createWebSocketClient(_requestExecutor),
	//res._task = CompletableFuture.runAsync(() -> doWork());
	//res.addConnectionStatusChanged(res._connectionStatusEventHandler)
	return res
}

func (c *DatabaseChanges) isConnected() bool {
	panic("NYI")
	return false
}

func (c *DatabaseChanges) ensureConnectedNow() {
	panic("NYI")

}

func (c *DatabaseChanges) addConnectionStatusChanged(handler EventHandler) {
	panic("NYI")

}

func (c *DatabaseChanges) removeConnectionStatusChanged(handler EventHandler) {
	panic("NYI")

}

func (c *DatabaseChanges) addOnError(handler func(error)) {
	panic("NYI")

}

func (c *DatabaseChanges) removeOnError(handler func(error)) {
	panic("NYI")
}

func (c *DatabaseChanges) forAllOperations() IChangesObservable_OperationStatusChange {
	panic("NYI")
	return nil
}
