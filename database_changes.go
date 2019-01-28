package ravendb

import (
	"context"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// In Java it's hidden behind IDatabaseChanges which also contains IConnectableChanges

// databaseChanges notifies about changes to a database
type databaseChanges struct {
	commandID atomicInteger

	semaphore sync.Mutex

	requestExecutor *RequestExecutor
	conventions     *DocumentConventions
	database        string

	onDispose Runnable

	client   *websocket.Conn
	muClient sync.Mutex

	task *completableFuture
	_cts *cancellationTokenSource
	tcs  *completableFuture

	mu            sync.Mutex // protects confirmations and counters maps
	confirmations map[int]*completableFuture
	counters      map[string]*DatabaseConnectionState

	immediateConnection atomicInteger

	connectionStatusEventHandlerIdx int
	connectionStatusChanged         []func()
	onError                         []func(error)
}

func newDatabaseChanges(requestExecutor *RequestExecutor, databaseName string, onDispose Runnable) *databaseChanges {
	res := &databaseChanges{
		requestExecutor:                 requestExecutor,
		conventions:                     requestExecutor.GetConventions(),
		database:                        databaseName,
		tcs:                             newCompletableFuture(),
		_cts:                            &cancellationTokenSource{},
		onDispose:                       onDispose,
		connectionStatusEventHandlerIdx: -1,
		confirmations:                   map[int]*completableFuture{},
		counters:                        map[string]*DatabaseConnectionState{},
	}

	res.task = newCompletableFuture()
	go func() {
		err := res.doWork()
		if err != nil {
			res.task.completeWithError(err)
		} else {
			res.task.complete(nil)
		}
	}()

	_connectionStatusEventHandler := func() {
		res.onConnectionStatusChanged()
	}
	res.connectionStatusEventHandlerIdx = res.AddConnectionStatusChanged(_connectionStatusEventHandler)
	return res
}

func (c *databaseChanges) onConnectionStatusChanged() {
	c.semAcquire()
	defer c.semRelease()

	if c.IsConnected() {
		c.tcs.complete(c)
		return
	}

	if c.tcs.IsDone() {
		c.tcs = newCompletableFuture()
	}
}

func (c *databaseChanges) getWsClient() *websocket.Conn {
	c.muClient.Lock()
	res := c.client
	c.muClient.Unlock()
	return res
}

func (c *databaseChanges) setWsClient(client *websocket.Conn) {
	c.muClient.Lock()
	c.client = client
	c.muClient.Unlock()
}

func (c *databaseChanges) IsConnected() bool {
	client := c.getWsClient()
	return client != nil
}

func (c *databaseChanges) EnsureConnectedNow() error {
	_, err := c.tcs.Get()
	return err
}

func (c *databaseChanges) AddConnectionStatusChanged(handler func()) int {
	c.mu.Lock()
	idx := len(c.connectionStatusChanged)
	c.connectionStatusChanged = append(c.connectionStatusChanged, handler)
	c.mu.Unlock()
	return idx
}

func (c *databaseChanges) RemoveConnectionStatusChanged(handlerIdx int) {
	if handlerIdx != -1 {
		c.connectionStatusChanged[handlerIdx] = nil
	}
}

func (c *databaseChanges) ForIndex(indexName string) (IChangesObservable, error) {
	counter, err := c.getOrAddConnectionState("indexes/"+indexName, "watch-index", "unwatch-index", indexName)
	if err != nil {
		return nil, err
	}

	filter := func(notification interface{}) bool {
		v := notification.(*IndexChange)
		return strings.EqualFold(v.Name, indexName)
	}

	taskedObservable := NewChangesObservable(ChangeIndex, counter, filter)
	return taskedObservable, nil
}

func (c *databaseChanges) getLastConnectionStateError() error {
	for _, counter := range c.counters {
		valueLastError := counter.lastError
		if valueLastError != nil {
			return valueLastError
		}
	}
	return nil
}

func (c *databaseChanges) ForDocument(docID string) (IChangesObservable, error) {
	counter, err := c.getOrAddConnectionState("docs/"+docID, "watch-doc", "unwatch-doc", docID)
	if err != nil {
		return nil, err
	}

	filter := func(notification interface{}) bool {
		v := notification.(*DocumentChange)
		return strings.EqualFold(v.ID, docID)
	}
	taskedObservable := NewChangesObservable(ChangeDocument, counter, filter)
	return taskedObservable, nil
}

func filterAlwaysTrue(notification interface{}) bool {
	return true
}

func (c *databaseChanges) ForAllDocuments() (IChangesObservable, error) {
	counter, err := c.getOrAddConnectionState("all-docs", "watch-docs", "unwatch-docs", "")
	if err != nil {
		return nil, err
	}
	taskedObservable := NewChangesObservable(ChangeDocument, counter, filterAlwaysTrue)
	return taskedObservable, nil
}

func (c *databaseChanges) ForOperationID(operationID int) (IChangesObservable, error) {
	opIDStr := strconv.Itoa(operationID)
	counter, err := c.getOrAddConnectionState("operations/"+opIDStr, "watch-operation", "unwatch-operation", opIDStr)
	if err != nil {
		return nil, err
	}

	filter := func(notification interface{}) bool {
		v := notification.(*OperationStatusChange)
		return v.OperationID == operationID
	}
	taskedObservable := NewChangesObservable(ChangeOperation, counter, filter)
	return taskedObservable, nil
}

func (c *databaseChanges) ForAllOperations() (IChangesObservable, error) {
	counter, err := c.getOrAddConnectionState("all-operations", "watch-operations", "unwatch-operations", "")
	if err != nil {
		return nil, err
	}

	taskedObservable := NewChangesObservable(ChangeOperation, counter, filterAlwaysTrue)

	return taskedObservable, nil
}

func (c *databaseChanges) ForAllIndexes() (IChangesObservable, error) {
	counter, err := c.getOrAddConnectionState("all-indexes", "watch-indexes", "unwatch-indexes", "")
	if err != nil {
		return nil, err
	}

	taskedObservable := NewChangesObservable(ChangeIndex, counter, filterAlwaysTrue)

	return taskedObservable, nil
}

func (c *databaseChanges) ForDocumentsStartingWith(docIDPrefix string) (IChangesObservable, error) {
	counter, err := c.getOrAddConnectionState("prefixes/"+docIDPrefix, "watch-prefix", "unwatch-prefix", docIDPrefix)
	if err != nil {
		return nil, err
	}
	filter := func(notification interface{}) bool {
		v := notification.(*DocumentChange)
		n := len(docIDPrefix)
		if n > len(v.ID) {
			return false
		}
		prefix := v.ID[:n]
		return strings.EqualFold(prefix, docIDPrefix)
	}

	taskedObservable := NewChangesObservable(ChangeDocument, counter, filter)

	return taskedObservable, nil
}

func (c *databaseChanges) ForDocumentsInCollection(collectionName string) (IChangesObservable, error) {
	if collectionName == "" {
		return nil, newIllegalArgumentError("CollectionName cannot be empty")
	}

	counter, err := c.getOrAddConnectionState("collections/"+collectionName, "watch-collection", "unwatch-collection", collectionName)
	if err != nil {
		return nil, err
	}

	filter := func(notification interface{}) bool {
		v := notification.(*DocumentChange)
		return strings.EqualFold(collectionName, v.CollectionName)
	}

	taskedObservable := NewChangesObservable(ChangeDocument, counter, filter)

	return taskedObservable, nil
}

func (c *databaseChanges) ForDocumentsInCollectionOfType(clazz reflect.Type) (IChangesObservable, error) {
	collectionName := c.conventions.GetCollectionName(clazz)
	return c.ForDocumentsInCollection(collectionName)
}

func (c *databaseChanges) invokeConnectionStatusChanged() {
	var dup []func()
	c.mu.Lock()
	for _, fn := range c.connectionStatusChanged {
		if fn != nil {
			dup = append(dup, fn)
		}
	}
	c.mu.Unlock()

	for _, fn := range dup {
		fn()
	}
}

func (c *databaseChanges) AddOnError(handler func(error)) int {
	c.mu.Lock()
	defer c.mu.Unlock()

	idx := len(c.onError)
	c.onError = append(c.onError, handler)
	return idx
}

func (c *databaseChanges) RemoveOnError(handlerIdx int) {
	c.onError[handlerIdx] = nil
}

func (c *databaseChanges) invokeOnError(err error) {
	// make a copy so that we can safely access outside of a lock
	c.mu.Lock()
	if len(c.onError) == 0 {
		c.mu.Unlock()
		return
	}
	handlers := append([]func(error){}, c.onError...)
	c.mu.Unlock()

	for _, fn := range handlers {
		if fn != nil {
			fn(err)
		}
	}
}

func (c *databaseChanges) Close() {
	//fmt.Printf("databaseChanges.Close()\n")
	c.mu.Lock()
	for _, confirmation := range c.confirmations {
		confirmation.cancel(false)
	}
	c.mu.Unlock()

	//fmt.Printf("databaseChanges.Close() before _cts.cancel\n")
	c.mu.Lock()
	c._cts.cancel()
	c.mu.Unlock()
	//fmt.Printf("databaseChanges.Close() after _cts.cancel\n")

	client := c.getWsClient()
	if client != nil {
		//fmt.Printf("databaseChanges.Close(): before client.Close()\n")
		err := client.Close()
		if err != nil {
			dbg("databaseChanges.Close(): client.Close() failed with %s\n", err)
		}
		c.setWsClient(nil)
	}

	c.mu.Lock()
	c.counters = nil
	c.mu.Unlock()

	c.task.Get()
	c.invokeConnectionStatusChanged()
	c.RemoveConnectionStatusChanged(c.connectionStatusEventHandlerIdx)
	if c.onDispose != nil {
		c.onDispose()
	}
}

func (c *databaseChanges) getOrAddConnectionState(name string, watchCommand string, unwatchCommand string, value string) (*DatabaseConnectionState, error) {
	c.mu.Lock()
	counter, ok := c.counters[name]
	c.mu.Unlock()

	if ok {
		return counter, nil
	}

	s := name
	onDisconnect := func() {
		if c.IsConnected() {
			c.send(unwatchCommand, value)
			// ignoring error: if we are not connected then we unsubscribed
			// already because connections drops with all subscriptions
		}

		c.mu.Lock()
		state := c.counters[s]
		delete(c.counters, s)
		c.mu.Unlock()
		state.Close()
	}

	onConnect := func() {
		c.send(watchCommand, value)
	}

	counter = NewDatabaseConnectionState(onConnect, onDisconnect)
	c.mu.Lock()
	c.counters[name] = counter
	c.mu.Unlock()

	if c.immediateConnection.get() != 0 {
		counter.onConnect()
	}
	return counter, nil
}

func (c *databaseChanges) semAcquire() {
	c.semaphore.Lock()
}

func (c *databaseChanges) semRelease() {
	c.semaphore.Unlock()
}

func (c *databaseChanges) send(command, value string) error {
	taskCompletionSource := newCompletableFuture()

	c.semAcquire()

	currentCommandID := c.commandID.incrementAndGet()

	o := struct {
		CommandID int    `json:"CommandId"`
		Command   string `json:"Command"`
		Param     string `json:"Param"`
	}{
		CommandID: currentCommandID,
		Command:   command,
		Param:     value,
	}

	client := c.getWsClient()
	err := client.WriteJSON(o)
	c.confirmations[currentCommandID] = taskCompletionSource

	c.semRelease()
	if err != nil {
		dbg("databaseChanges.send: WriteJSON() failed with %s\n", err)
		return err
	}

	_, err = taskCompletionSource.GetWithTimeout(time.Second * 15)
	return err
}

func toWebSocketPath(path string) string {
	path = strings.Replace(path, "http://", "ws://", -1)
	return strings.Replace(path, "https://", "wss://", -1)
}

func (c *databaseChanges) doWork() error {
	_, err := c.requestExecutor.getPreferredNode()
	if err != nil {
		c.invokeConnectionStatusChanged()
		c.notifyAboutError(err)
		return err
	}

	urlString := c.requestExecutor.GetURL() + "/databases/" + c.database + "/changes"
	urlString = toWebSocketPath(urlString)

	for {
		if c._cts.getToken().isCancellationRequested() {
			return nil
		}

		var processor *webSocketChangesProcessor
		var err error
		panicIf(c.IsConnected(), "impoosible: cannot be connected")

		dialer := *websocket.DefaultDialer
		dialer.HandshakeTimeout = time.Second * 2
		re := c.requestExecutor
		if re.Certificate != nil || re.TrustStore != nil {
			tlsConfig, err := newTLSConfig(re.Certificate, re.TrustStore)
			if err != nil {
				return err
			}
			dialer.TLSClientConfig = tlsConfig
		}

		ctx := context.Background()
		var client *websocket.Conn
		client, _, err = dialer.DialContext(ctx, urlString, nil)
		c.setWsClient(client)

		// recheck cancellation because it might have been cancelled
		// since DialContext()
		if c._cts.getToken().isCancellationRequested() {
			if client != nil {
				client.Close()
				c.setWsClient(nil)
			}
			return err
		}

		if err != nil {
			time.Sleep(time.Second)
			continue
		}

		processor = newWebSocketChangesProcessor(client)
		go processor.processMessages(c)
		c.immediateConnection.set(1)

		c.mu.Lock()
		for _, counter := range c.counters {
			counter.onConnect()
		}
		c.mu.Unlock()

		c.invokeConnectionStatusChanged()
		processor.processing.Get()
		c.invokeConnectionStatusChanged()
		shouldReconnect := c.reconnectClient()

		for _, confirmation := range c.confirmations {
			confirmation.cancel(false)
		}

		for k := range c.confirmations {
			delete(c.confirmations, k)
		}

		if !shouldReconnect {
			return nil
		}

		// wait before next retry
		time.Sleep(time.Second)
	}
}

func (c *databaseChanges) reconnectClient() bool {
	if c._cts.getToken().isCancellationRequested() {
		return false
	}

	c.immediateConnection.set(0)

	c.invokeConnectionStatusChanged()
	return true
}

func (c *databaseChanges) notifySubscribers(typ string, value interface{}, states []*DatabaseConnectionState) error {
	switch typ {
	case "DocumentChange":
		var documentChange *DocumentChange
		err := decodeJSONAsStruct(value, &documentChange)
		if err != nil {
			dbg("notifySubscribers: '%s' decodeJSONAsStruct failed with %s\n", typ, err)
			return err
		}
		for _, state := range states {
			state.sendDocumentChange(documentChange)
		}
	case "IndexChange":
		var indexChange *IndexChange
		err := decodeJSONAsStruct(value, &indexChange)
		if err != nil {
			dbg("notifySubscribers: '%s' decodeJSONAsStruct failed with %s\n", typ, err)
			return err
		}
		for _, state := range states {
			state.sendIndexChange(indexChange)
		}
	case "OperationStatusChange":
		var operationStatusChange *OperationStatusChange
		err := decodeJSONAsStruct(value, &operationStatusChange)
		if err != nil {
			dbg("notifySubscribers: '%s' decodeJSONAsStruct failed with %s\n", typ, err)
			return err
		}
		for _, state := range states {
			state.sendOperationStatusChange(operationStatusChange)
		}
	default:
		return fmt.Errorf("notifySubscribers: unsupported type '%s'", typ)
	}
	return nil
}

func (c *databaseChanges) notifyAboutError(e error) {
	if c._cts.getToken().isCancellationRequested() {
		return
	}

	c.invokeOnError(e)

	c.mu.Lock()
	defer c.mu.Unlock()
	for _, state := range c.counters {
		state.error(e)
	}
}

type webSocketChangesProcessor struct {
	processing *completableFuture
	client     *websocket.Conn
}

func newWebSocketChangesProcessor(client *websocket.Conn) *webSocketChangesProcessor {
	return &webSocketChangesProcessor{
		processing: newCompletableFuture(),
		client:     client,
	}
}

func (p *webSocketChangesProcessor) processMessages(changes *databaseChanges) {
	var err error
	for {
		var msgArray []interface{} // an array of objects
		err = p.client.ReadJSON(&msgArray)
		if err != nil {
			dbg("webSocketChangesProcessor.processMessages() ReadJSON() failed with %s\n", err)
			break
		}

		for _, msgNodeV := range msgArray {
			msgNode := msgNodeV.(map[string]interface{})
			typ, _ := jsonGetAsString(msgNode, "Type")
			switch typ {
			case "Error":
				errStr, _ := jsonGetAsString(msgNode, "Error")
				changes.notifyAboutError(newRuntimeError("%s", errStr))
			case "Confirm":
				commandID, ok := jsonGetAsInt(msgNode, "CommandId")
				if ok {
					changes.semAcquire()
					future := changes.confirmations[commandID]
					changes.semRelease()
					if future != nil {
						future.complete(nil)
					}
				}
			default:
				val := msgNode["Value"]
				var states []*DatabaseConnectionState
				for _, state := range changes.counters {
					states = append(states, state)
				}
				changes.notifySubscribers(typ, val, states)
			}
		}
	}
	// TODO: check for io.EOF for clean connection close?
	changes.notifyAboutError(err)
	p.processing.completeWithError(err)
}
