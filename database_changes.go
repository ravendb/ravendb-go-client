package ravendb

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
)

// In Java it's hidden behind IDatabaseChanges which also contains IConnectableChanges

// DatabaseChanges notifies about changes to a database
type DatabaseChanges struct {
	commandID int32 // atomic

	requestExecutor *RequestExecutor
	conventions     *DocumentConventions
	database        string

	onDispose func()

	client   *websocket.Conn
	muClient sync.Mutex

	task         *completableFuture
	doWorkCancel context.CancelFunc
	tcs          *completableFuture

	mu            sync.Mutex // protects confirmations and counters maps
	confirmations map[int]*completableFuture
	counters      map[string]*databaseConnectionState

	immediateConnection int32 // atomic bool

	connectionStatusChanged []func()
	onError                 []func(error)
}

func (c *DatabaseChanges) nextCommandID() int {
	v := atomic.AddInt32(&c.commandID, 1)
	return int(v)
}

func newDatabaseChanges(requestExecutor *RequestExecutor, databaseName string, onDispose func()) *DatabaseChanges {
	res := &DatabaseChanges{
		requestExecutor: requestExecutor,
		conventions:     requestExecutor.GetConventions(),
		database:        databaseName,
		tcs:             newCompletableFuture(),
		onDispose:       onDispose,
		confirmations:   map[int]*completableFuture{},
		counters:        map[string]*databaseConnectionState{},
	}

	res.task = newCompletableFuture()
	var ctx context.Context
	ctx, res.doWorkCancel = context.WithCancel(context.Background())
	go func() {
		err := res.doWork(ctx)
		if err != nil {
			res.task.completeWithError(err)
		} else {
			res.task.complete(nil)
		}
	}()

	return res
}

func (c *DatabaseChanges) getWsClient() *websocket.Conn {
	c.muClient.Lock()
	res := c.client
	c.muClient.Unlock()
	return res
}

func (c *DatabaseChanges) setWsClient(client *websocket.Conn) {
	c.muClient.Lock()
	c.client = client
	c.muClient.Unlock()
}

func (c *DatabaseChanges) IsConnected() bool {
	client := c.getWsClient()
	return client != nil
}

func (c *DatabaseChanges) EnsureConnectedNow() error {
	_, err := c.tcs.Get()
	return err
}

func (c *DatabaseChanges) AddConnectionStatusChanged(handler func()) int {
	c.mu.Lock()
	idx := len(c.connectionStatusChanged)
	c.connectionStatusChanged = append(c.connectionStatusChanged, handler)
	c.mu.Unlock()
	return idx
}

func (c *DatabaseChanges) RemoveConnectionStatusChanged(handlerIdx int) {
	if handlerIdx != -1 {
		c.connectionStatusChanged[handlerIdx] = nil
	}
}

type CloseFunc func()

func (c *DatabaseChanges) ForIndex(indexName string, cb func(*IndexChange)) (CloseFunc, error) {
	counter, err := c.getOrAddConnectionState("indexes/"+indexName, "watch-index", "unwatch-index", indexName)
	if err != nil {
		return nil, err
	}

	filtered := func(change *IndexChange) {
		if strings.EqualFold(change.Name, indexName) {
			cb(change)
		}
	}
	idx := counter.addOnIndexChangeNotification(filtered)
	cancel := func() {
		counter.removeOnIndexChangeNotification(idx)
	}

	return cancel, nil
}

func (c *DatabaseChanges) getLastConnectionStateError() error {
	for _, counter := range c.counters {
		valueLastError := counter.lastError
		if valueLastError != nil {
			return valueLastError
		}
	}
	return nil
}

func (c *DatabaseChanges) ForDocument(docID string, cb func(*DocumentChange)) (CloseFunc, error) {
	counter, err := c.getOrAddConnectionState("docs/"+docID, "watch-doc", "unwatch-doc", docID)
	if err != nil {
		return nil, err
	}

	filtered := func(v *DocumentChange) {
		if strings.EqualFold(v.ID, docID) {
			cb(v)
		}
	}
	idx := counter.addOnDocumentChangeNotification(filtered)
	cancel := func() {
		counter.removeOnDocumentChangeNotification(idx)
	}
	return cancel, nil
}

func (c *DatabaseChanges) ForAllDocuments(cb func(*DocumentChange)) (CloseFunc, error) {
	counter, err := c.getOrAddConnectionState("all-docs", "watch-docs", "unwatch-docs", "")
	if err != nil {
		return nil, err
	}
	idx := counter.addOnDocumentChangeNotification(cb)
	cancel := func() {
		counter.removeOnDocumentChangeNotification(idx)
	}
	return cancel, nil
}

func (c *DatabaseChanges) ForOperationID(operationID int64, cb func(*OperationStatusChange)) (CloseFunc, error) {
	opIDStr := i64toa(operationID)
	counter, err := c.getOrAddConnectionState("operations/"+opIDStr, "watch-operation", "unwatch-operation", opIDStr)
	if err != nil {
		return nil, err
	}

	filtered := func(v *OperationStatusChange) {
		if v.OperationID == operationID {
			cb(v)
		}
	}

	idx := counter.addOnOperationChangeNotification(filtered)
	cancel := func() {
		counter.removeOnOperationChangeNotification(idx)
	}
	return cancel, nil
}

func (c *DatabaseChanges) ForAllOperations(cb func(*OperationStatusChange)) (CloseFunc, error) {
	counter, err := c.getOrAddConnectionState("all-operations", "watch-operations", "unwatch-operations", "")
	if err != nil {
		return nil, err
	}
	idx := counter.addOnOperationChangeNotification(cb)
	cancel := func() {
		counter.removeOnOperationChangeNotification(idx)
	}
	return cancel, nil
}

func (c *DatabaseChanges) ForAllIndexes(cb func(*IndexChange)) (CloseFunc, error) {
	counter, err := c.getOrAddConnectionState("all-indexes", "watch-indexes", "unwatch-indexes", "")
	if err != nil {
		return nil, err
	}

	idx := counter.addOnIndexChangeNotification(cb)
	cancel := func() {
		counter.removeOnIndexChangeNotification(idx)
	}
	return cancel, nil
}

func (c *DatabaseChanges) ForDocumentsStartingWith(docIDPrefix string, cb func(*DocumentChange)) (CloseFunc, error) {
	counter, err := c.getOrAddConnectionState("prefixes/"+docIDPrefix, "watch-prefix", "unwatch-prefix", docIDPrefix)
	if err != nil {
		return nil, err
	}
	filtered := func(v *DocumentChange) {
		n := len(docIDPrefix)
		if n > len(v.ID) {
			return
		}
		prefix := v.ID[:n]
		if strings.EqualFold(prefix, docIDPrefix) {
			cb(v)
		}
	}

	idx := counter.addOnDocumentChangeNotification(filtered)
	cancel := func() {
		counter.removeOnDocumentChangeNotification(idx)
	}
	return cancel, nil
}

func (c *DatabaseChanges) ForDocumentsInCollection(collectionName string, cb func(*DocumentChange)) (CloseFunc, error) {
	if collectionName == "" {
		return nil, newIllegalArgumentError("CollectionName cannot be empty")
	}

	counter, err := c.getOrAddConnectionState("collections/"+collectionName, "watch-collection", "unwatch-collection", collectionName)
	if err != nil {
		return nil, err
	}

	filtered := func(v *DocumentChange) {
		if strings.EqualFold(collectionName, v.CollectionName) {
			cb(v)
		}
	}
	idx := counter.addOnDocumentChangeNotification(filtered)
	cancel := func() {
		counter.removeOnDocumentChangeNotification(idx)
	}
	return cancel, nil
}

func (c *DatabaseChanges) ForDocumentsInCollectionOfType(clazz reflect.Type, cb func(*DocumentChange)) (CloseFunc, error) {
	collectionName := c.conventions.GetCollectionName(clazz)
	return c.ForDocumentsInCollection(collectionName, cb)
}

func (c *DatabaseChanges) invokeConnectionStatusChanged() {
	{
		// our internal processing
		if c.IsConnected() {
			c.tcs.complete(c)
		} else if c.tcs.IsDone() {
			c.tcs = newCompletableFuture()
		}
	}

	// call externally registered handlers outside of a lock
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

func (c *DatabaseChanges) AddOnError(handler func(error)) int {
	c.mu.Lock()
	defer c.mu.Unlock()

	idx := len(c.onError)
	c.onError = append(c.onError, handler)
	return idx
}

func (c *DatabaseChanges) RemoveOnError(handlerIdx int) {
	c.onError[handlerIdx] = nil
}

func (c *DatabaseChanges) invokeOnError(err error) {
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

func (c *DatabaseChanges) Close() {
	c.mu.Lock()
	for _, confirmation := range c.confirmations {
		confirmation.cancel(false)
	}
	c.mu.Unlock()

	c.doWorkCancel()

	client := c.getWsClient()
	if client != nil {
		//fmt.Printf("DatabaseChanges.Close(): before client.Close()\n")
		err := client.Close()
		if err != nil {
			dbg("DatabaseChanges.Close(): client.Close() failed with %s\n", err)
		}
		c.setWsClient(nil)
	}

	c.mu.Lock()
	c.counters = nil
	c.mu.Unlock()

	c.task.Get()
	c.invokeConnectionStatusChanged()
	if c.onDispose != nil {
		c.onDispose()
	}
}

func (c *DatabaseChanges) getOrAddConnectionState(name string, watchCommand string, unwatchCommand string, value string) (*databaseConnectionState, error) {
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

	counter = newDatabaseConnectionState(onConnect, onDisconnect)
	c.mu.Lock()
	c.counters[name] = counter
	c.mu.Unlock()

	if atomic.LoadInt32(&c.immediateConnection) != 0 {
		counter.onConnect()
	}
	return counter, nil
}

func (c *DatabaseChanges) send(command, value string) error {
	taskCompletionSource := newCompletableFuture()

	currentCommandID := c.nextCommandID()

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
	if client == nil {
		return errors.New("connection is closed")
	}
	err := client.WriteJSON(o)
	c.mu.Lock()
	c.confirmations[currentCommandID] = taskCompletionSource
	c.mu.Unlock()

	if err != nil {
		dbg("DatabaseChanges.send: WriteJSON() failed with %s\n", err)
		return err
	}

	_, err = taskCompletionSource.GetWithTimeout(time.Second * 15)
	return err
}

func toWebSocketPath(path string) string {
	path = strings.Replace(path, "http://", "ws://", -1)
	return strings.Replace(path, "https://", "wss://", -1)
}

func isCtxCancelled(ctx context.Context) bool {
	select {
	case <-ctx.Done():
		return true
	default:
		return false
	}
}

func (c *DatabaseChanges) doWork(ctx context.Context) error {
	_, err := c.requestExecutor.getPreferredNode()
	if err != nil {
		c.invokeConnectionStatusChanged()
		c.notifyAboutError(err)
		c.tcs.completeWithError(err)
		return err
	}

	urlString := c.requestExecutor.GetURL() + "/databases/" + c.database + "/changes"
	urlString = toWebSocketPath(urlString)

	for {

		if isCtxCancelled(ctx) {
			return nil
		}

		var err error
		panicIf(c.IsConnected(), "impoosible: cannot be connected")

		dialer := *websocket.DefaultDialer
		dialer.HandshakeTimeout = time.Second * 2
		re := c.requestExecutor
		if re.Certificate != nil || re.TrustStore != nil {
			dialer.TLSClientConfig, err = newTLSConfig(re.Certificate, re.TrustStore)
			if err != nil {
				return err
			}
		}

		ctx := context.Background()
		ctx, _ = context.WithTimeout(ctx, time.Second*5)
		var client *websocket.Conn
		client, _, err = dialer.DialContext(ctx, urlString, nil)
		c.setWsClient(client)

		// recheck cancellation because it might have been cancelled
		// since DialContext()
		if isCtxCancelled(ctx) {
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

		atomic.StoreInt32(&c.immediateConnection, 1)

		var counters []*databaseConnectionState
		c.mu.Lock()
		for _, counter := range c.counters {
			counters = append(counters, counter)
		}
		c.mu.Unlock()

		for _, counter := range counters {
			counter.onConnect()
		}

		c.invokeConnectionStatusChanged()
		c.processMessages(ctx)
		c.invokeConnectionStatusChanged()
		shouldReconnect := c.reconnectClient(ctx)

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

func (c *DatabaseChanges) reconnectClient(ctx context.Context) bool {
	if isCtxCancelled(ctx) {
		return false
	}

	atomic.StoreInt32(&c.immediateConnection, 0)

	c.invokeConnectionStatusChanged()
	return true
}

func (c *DatabaseChanges) notifySubscribers(typ string, value interface{}, states []*databaseConnectionState) error {
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

func (c *DatabaseChanges) notifyAboutError(e error) {
	c.invokeOnError(e)

	c.mu.Lock()
	defer c.mu.Unlock()
	for _, state := range c.counters {
		state.error(e)
	}
}

func (c *DatabaseChanges) processMessages(ctx context.Context) {
	var err error
	for {
		var msgArray []interface{} // an array of objects
		client := c.getWsClient()
		if client == nil {
			break
		}
		err = client.ReadJSON(&msgArray)
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
				if !isCtxCancelled(ctx) {
					c.notifyAboutError(newRuntimeError("%s", errStr))
				}
			case "Confirm":
				commandID, ok := jsonGetAsInt(msgNode, "CommandId")
				if ok {
					c.mu.Lock()
					future := c.confirmations[commandID]
					c.mu.Unlock()
					if future != nil {
						future.complete(nil)
					}
				}
			default:
				val := msgNode["Value"]
				var states []*databaseConnectionState
				for _, state := range c.counters {
					states = append(states, state)
				}
				c.notifySubscribers(typ, val, states)
			}
		}
	}
	// TODO: check for io.EOF for clean connection close?
	if !isCtxCancelled(ctx) {
		c.notifyAboutError(err)
	}
}
