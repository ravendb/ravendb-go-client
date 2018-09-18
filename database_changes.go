package ravendb

import (
	"sync"

	"github.com/gorilla/websocket"
)

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

	_client *websocket.Conn
	_task   *CompletableFuture
	_cts    *CancellationTokenSource
	_tcs    *CompletableFuture

	mu             sync.Mutex // protects _confirmations and _counters maps
	_confirmations map[int]*CompletableFuture
	_counters      map[string]*DatabaseConnectionState

	_immediateConnection atomicInteger

	_connectionStatusChanged []func()
	onError                  []func(error)
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
	res._task = NewCompletableFuture()
	go func() {
		err := res.doWork()
		if err != nil {
			res._task.CompleteExceptionally(err)
		} else {
			res._task.Complete(nil)
		}
	}()

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
	return c._client != nil
}

func (c *DatabaseChanges) ensureConnectedNow() error {
	_, err := c._tcs.Get()
	return err
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

func (c *DatabaseChanges) addOnError(handler func(error)) int {
	idx := len(c.onError)
	c.onError = append(c.onError, handler)
	return idx
}

func (c *DatabaseChanges) removeOnError(handlerIdx int) {
	c.onError[handlerIdx] = nil
}

func (c *DatabaseChanges) invokeOnError(err error) {
	for _, fn := range c.onError {
		if fn != nil {
			fn(err)
		}
	}
}

func (c *DatabaseChanges) Close() {
	panic("NYI")
}

func (c *DatabaseChanges) getOrAddConnectionState(name string, watchCommand string, unwatchCommand string, value string) (*DatabaseConnectionState, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	counter, ok := c._counters[name]
	if ok {
		return counter, nil
	}

	s := name
	onDisconnect := func() {
		if c.isConnected() {
			err := c.send(unwatchCommand, value)
			if err != nil {
				// if we are not connected then we unsubscribed already
				// because connections drops with all subscriptions
			}
		}

		c.mu.Lock()
		state := c._counters[s]
		delete(c._counters, s)
		c.mu.Unlock()
		state.Close()
	}

	onConnect := func() {
		c.send(watchCommand, value)
	}

	counter = NewDatabaseConnectionState(onConnect, onDisconnect)
	c._counters[name] = counter

	if c._immediateConnection.get() == 0 {
		counter.onConnect()
	}
	return counter, nil
}

func (c *DatabaseChanges) send(command, value string) error {
	panic("NYI")
	return nil
}

func (c *DatabaseChanges) doWork() error {
	_, err := c._requestExecutor.getPreferredNode()
	if err != nil {
		c.invokeConnectionStatusChanged()
		c.notifyAboutError(err)
		return err
	}
	panic("NYI")
	return nil
}

func (c *DatabaseChanges) reconnectClient() bool {
	panic("NYI")
	return false
}

func (c *DatabaseChanges) forAllOperations() IChangesObservable_OperationStatusChange {
	panic("NYI")
	return nil
}

func (c *DatabaseChanges) notifyAboutError(e error) {
	panic("NYI")
	// TODO: implement this
	/*
		if (_cts.getToken().isCancellationRequested()) {
			return;
		}
	*/

	c.invokeOnError(e)

	c.mu.Lock()
	defer c.mu.Unlock()
	for _, state := range c._counters {
		state.error(e)
	}
}
