package ravendb

import "sync"

var (
	_ IDatabaseChanges = &DatabaseChanges{}
)

type DatabaseChanges struct {
	_commandId int

	// TODO: why semaphore of size 1 and not a mutex?
	_semaphore chan bool

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

	_immediateConnection atomicInteger

	_connectionStatusChanged []func()
}

func NewDatabaseChanges(requestExecutor *RequestExecutor, databaseName string, onDispose Runnable) *DatabaseChanges {
	res := &DatabaseChanges{
		_requestExecutor: requestExecutor,
		_conventions:     requestExecutor.GetConventions(),
		_database:        databaseName,
		_tcs:             NewCompletableFuture(),
		_cts:             NewCancellationTokenSource(),
		_onDispose:       onDispose,
		_semaphore:       make(chan bool, 1),
	}

	//res._client = res.createWebSocketClient(_requestExecutor),
	//res._task = CompletableFuture.runAsync(() -> doWork());

	_connectionStatusEventHandler := func() {
		res.onConnectionStatusChanged()
	}
	res.addConnectionStatusChanged(_connectionStatusEventHandler)
	return res
}

func (c *DatabaseChanges) onConnectionStatusChanged() {
	c._semaphore <- true // acquire
	defer func() {
		<-c._semaphore // release
	}()

	if c.isConnected() {
		c._tcs.Complete(c)
		return
	}

	if c._tcs.IsDone() {
		c._tcs = NewCompletableFuture()
	}
}

func (c *DatabaseChanges) isConnected() bool {
	panic("NYI")
	return false
}

func (c *DatabaseChanges) ensureConnectedNow() {
	panic("NYI")

}

func (c *DatabaseChanges) addConnectionStatusChanged(handler func()) int {
	idx := len(c._connectionStatusChanged)
	c._connectionStatusChanged = append(c._connectionStatusChanged, handler)
	return idx

}

func (c *DatabaseChanges) removeConnectionStatusChanged(handlerIdx int) {
	c._connectionStatusChanged[handlerIdx] = nil
}

func (c *DatabaseChanges) invokeConnectionStatusChanged() {
	for _, fn := range c._connectionStatusChanged {
		if fn != nil {
			fn()
		}
	}
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
