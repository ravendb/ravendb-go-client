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

// Note: in Java IChangesConnectionState hides changeSubscribers

type changeSubscribers struct {
	name           string // is key of DatabaseChanges.subscribers
	watchCommand   string
	unwatchCommand string
	commandValue   string

	onDocumentChange        []func(*DocumentChange)
	onIndexChange           []func(*IndexChange)
	onOperationStatusChange []func(*OperationStatusChange)
}

func (s *changeSubscribers) hasRegisteredHandlers() bool {
	// s.mu must be locked here
	for _, cb := range s.onDocumentChange {
		if cb != nil {
			return true
		}
	}
	for _, cb := range s.onIndexChange {
		if cb != nil {
			return true
		}
	}
	for _, cb := range s.onOperationStatusChange {
		if cb != nil {
			return true
		}
	}
	return false
}

// DatabaseChanges notifies about changes to a database
type DatabaseChanges struct {
	commandID int32 // atomic

	requestExecutor *RequestExecutor
	conventions     *DocumentConventions
	database        string

	onDispose func()

	client   *websocket.Conn
	muClient sync.Mutex

	task         chan error
	doWorkCancel context.CancelFunc
	tcs          *completableFuture

	mu          sync.Mutex // protects subscribers maps
	subscribers map[string]*changeSubscribers

	muConfirmations sync.Mutex
	confirmations   map[int]*completableFuture

	immediateConnection atomicBool

	connectionStatusChanged []func()
	onError                 []func(error)

	lastError error
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
		subscribers:     map[string]*changeSubscribers{},
	}

	res.task = make(chan error, 1)
	var ctx context.Context
	ctx, res.doWorkCancel = context.WithCancel(context.Background())
	go func() {
		err := res.doWork(ctx)
		res.task <- err
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

type CancelFunc func()

func (c *DatabaseChanges) ForIndex(indexName string, cb func(*IndexChange)) (CancelFunc, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	subscribers, err := c.getOrAddSubscribers("indexes/"+indexName, "watch-index", "unwatch-index", indexName)
	if err != nil {
		return nil, err
	}

	filtered := func(change *IndexChange) {
		if strings.EqualFold(change.Name, indexName) {
			cb(change)
		}
	}
	idx := len(subscribers.onIndexChange)
	subscribers.onIndexChange = append(subscribers.onIndexChange, filtered)
	cancel := func() {
		c.mu.Lock()
		defer c.mu.Unlock()
		subscribers.onIndexChange[idx] = nil
		c.maybeDisconnectSubscribers(subscribers)
	}

	return cancel, nil
}

func (c *DatabaseChanges) getLastConnectionStateError() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.lastError
}

func (c *DatabaseChanges) ForDocument(docID string, cb func(*DocumentChange)) (CancelFunc, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	subscribers, err := c.getOrAddSubscribers("docs/"+docID, "watch-doc", "unwatch-doc", docID)
	if err != nil {
		return nil, err
	}

	filtered := func(change *DocumentChange) {
		panicIf(change.ID != docID, "v.ID (%s) != docID (%s)", change.ID, docID)
		cb(change)
	}
	idx := len(subscribers.onDocumentChange)
	subscribers.onDocumentChange = append(subscribers.onDocumentChange, filtered)
	cancel := func() {
		c.mu.Lock()
		defer c.mu.Unlock()
		subscribers.onDocumentChange[idx] = nil
		c.maybeDisconnectSubscribers(subscribers)
	}
	return cancel, nil
}

func (c *DatabaseChanges) ForAllDocuments(cb func(*DocumentChange)) (CancelFunc, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	subscribers, err := c.getOrAddSubscribers("all-docs", "watch-docs", "unwatch-docs", "")
	if err != nil {
		return nil, err
	}
	idx := len(subscribers.onDocumentChange)
	subscribers.onDocumentChange = append(subscribers.onDocumentChange, cb)
	cancel := func() {
		c.mu.Lock()
		defer c.mu.Unlock()
		subscribers.onDocumentChange[idx] = nil
		c.maybeDisconnectSubscribers(subscribers)
	}
	return cancel, nil
}

func (c *DatabaseChanges) ForOperationID(operationID int64, cb func(*OperationStatusChange)) (CancelFunc, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	opIDStr := i64toa(operationID)
	subscribers, err := c.getOrAddSubscribers("operations/"+opIDStr, "watch-operation", "unwatch-operation", opIDStr)
	if err != nil {
		return nil, err
	}

	filtered := func(v *OperationStatusChange) {
		if v.OperationID == operationID {
			cb(v)
		}
	}
	idx := len(subscribers.onOperationStatusChange)
	subscribers.onOperationStatusChange = append(subscribers.onOperationStatusChange, filtered)

	cancel := func() {
		c.mu.Lock()
		defer c.mu.Unlock()
		subscribers.onOperationStatusChange[idx] = nil
		c.maybeDisconnectSubscribers(subscribers)
	}
	return cancel, nil
}

func (c *DatabaseChanges) ForAllOperations(cb func(*OperationStatusChange)) (CancelFunc, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	subscribers, err := c.getOrAddSubscribers("all-operations", "watch-operations", "unwatch-operations", "")
	if err != nil {
		return nil, err
	}
	idx := len(subscribers.onOperationStatusChange)
	subscribers.onOperationStatusChange = append(subscribers.onOperationStatusChange, cb)

	cancel := func() {
		c.mu.Lock()
		defer c.mu.Unlock()
		subscribers.onOperationStatusChange[idx] = nil
		c.maybeDisconnectSubscribers(subscribers)
	}
	return cancel, nil
}

func (c *DatabaseChanges) removeSubscriber(subscribers *changeSubscribers, callbacks []func(), idx int) {

}

func (c *DatabaseChanges) ForAllIndexes(cb func(*IndexChange)) (CancelFunc, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	subscribers, err := c.getOrAddSubscribers("all-indexes", "watch-indexes", "unwatch-indexes", "")
	if err != nil {
		return nil, err
	}
	idx := len(subscribers.onIndexChange)
	subscribers.onIndexChange = append(subscribers.onIndexChange, cb)
	cancel := func() {
		c.mu.Lock()
		defer c.mu.Unlock()
		subscribers.onIndexChange[idx] = nil
		c.maybeDisconnectSubscribers(subscribers)
	}
	return cancel, nil
}

func (c *DatabaseChanges) ForDocumentsStartingWith(docIDPrefix string, cb func(*DocumentChange)) (CancelFunc, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	subscribers, err := c.getOrAddSubscribers("prefixes/"+docIDPrefix, "watch-prefix", "unwatch-prefix", docIDPrefix)
	if err != nil {
		return nil, err
	}
	filtered := func(change *DocumentChange) {
		n := len(docIDPrefix)
		if n > len(change.ID) {
			return
		}
		prefix := change.ID[:n]
		if strings.EqualFold(prefix, docIDPrefix) {
			cb(change)
		}
	}
	idx := len(subscribers.onDocumentChange)
	subscribers.onDocumentChange = append(subscribers.onDocumentChange, filtered)

	cancel := func() {
		c.mu.Lock()
		defer c.mu.Unlock()
		subscribers.onDocumentChange[idx] = nil
		c.maybeDisconnectSubscribers(subscribers)
	}
	return cancel, nil
}

func (c *DatabaseChanges) ForDocumentsInCollection(collectionName string, cb func(*DocumentChange)) (CancelFunc, error) {
	if collectionName == "" {
		return nil, newIllegalArgumentError("CollectionName cannot be empty")
	}

	c.mu.Lock()
	defer c.mu.Unlock()
	subscribers, err := c.getOrAddSubscribers("collections/"+collectionName, "watch-collection", "unwatch-collection", collectionName)
	if err != nil {
		return nil, err
	}

	filtered := func(v *DocumentChange) {
		if strings.EqualFold(collectionName, v.CollectionName) {
			cb(v)
		}
	}

	idx := len(subscribers.onDocumentChange)
	subscribers.onDocumentChange = append(subscribers.onDocumentChange, filtered)

	cancel := func() {
		c.mu.Lock()
		defer c.mu.Unlock()
		subscribers.onDocumentChange[idx] = nil
		c.maybeDisconnectSubscribers(subscribers)
	}
	return cancel, nil
}

func (c *DatabaseChanges) ForDocumentsInCollectionOfType(clazz reflect.Type, cb func(*DocumentChange)) (CancelFunc, error) {
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

func (c *DatabaseChanges) Close() {
	c.muConfirmations.Lock()
	for _, confirmation := range c.confirmations {
		confirmation.cancel(false)
	}
	c.muConfirmations.Unlock()

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
	c.subscribers = nil
	c.mu.Unlock()

	<-c.task

	c.invokeConnectionStatusChanged()
	if c.onDispose != nil {
		c.onDispose()
	}
}

func (c *DatabaseChanges) getOrAddSubscribers(name string, watchCommand string, unwatchCommand string, value string) (*changeSubscribers, error) {
	// must be called while holding mu lock
	subscribers, ok := c.subscribers[name]

	if ok {
		return subscribers, nil
	}

	subscribers = &changeSubscribers{
		name:           name,
		watchCommand:   watchCommand,
		unwatchCommand: unwatchCommand,
		commandValue:   value,
	}
	c.subscribers[name] = subscribers

	if c.immediateConnection.isTrue() {
		if err := c.connectSubscribers(subscribers); err != nil {
			return nil, err
		}

	}
	return subscribers, nil
}

func (c *DatabaseChanges) maybeDisconnectSubscribers(subscribers *changeSubscribers) {
	if !subscribers.hasRegisteredHandlers() {
		c.disconnectSubscribers(subscribers)
	}
}

func (c *DatabaseChanges) disconnectSubscribers(subscribers *changeSubscribers) {
	// called while holding a mu lock
	name := subscribers.name
	if c.IsConnected() {
		go c.send(subscribers.unwatchCommand, subscribers.commandValue)
		// ignoring error: if we are not connected then we unsubscribed
		// already because connections drops with all subscriptions
	}
	delete(c.subscribers, name)
}

func (c *DatabaseChanges) connectSubscribers(subscribers *changeSubscribers) error {
	// called inside the lock, so unlock while sending blocking call
	c.mu.Unlock()
	defer c.mu.Lock()
	return c.send(subscribers.watchCommand, subscribers.commandValue)
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
	if err != nil {
		dbg("DatabaseChanges.send: WriteJSON() failed with %s\n", err)
		return err
	}

	c.muConfirmations.Lock()
	c.confirmations[currentCommandID] = taskCompletionSource
	c.muConfirmations.Unlock()

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

		ctxDial, _ := context.WithTimeout(ctx, time.Second*5)
		var client *websocket.Conn
		client, _, err = dialer.DialContext(ctxDial, urlString, nil)
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

		(&c.immediateConnection).set(true)

		c.mu.Lock()
		for _, subscribers := range c.subscribers {
			c.connectSubscribers(subscribers)
		}
		c.mu.Unlock()

		c.invokeConnectionStatusChanged()
		c.processMessages(ctx)
		c.invokeConnectionStatusChanged()

		shouldReconnect := false
		{
			if !isCtxCancelled(ctx) {
				(&c.immediateConnection).set(false)
				shouldReconnect = true
			}
		}

		c.muConfirmations.Lock()
		for _, confirmation := range c.confirmations {
			confirmation.cancel(false)
		}

		c.confirmations = map[int]*completableFuture{}
		c.muConfirmations.Unlock()

		if !shouldReconnect {
			return nil
		}

		// wait before next retry
		time.Sleep(time.Second)
	}
}

func (s *changeSubscribers) sendDocumentChange(documentChange *DocumentChange) {
	for _, f := range s.onDocumentChange {
		if f != nil {
			f(documentChange)
		}
	}
}

func (s *changeSubscribers) sendIndexChange(indexChange *IndexChange) {
	for _, f := range s.onIndexChange {
		if f != nil {
			f(indexChange)
		}
	}
}

func (s *changeSubscribers) sendOperationStatusChange(operationStatusChange *OperationStatusChange) {
	for _, f := range s.onOperationStatusChange {
		if f != nil {
			f(operationStatusChange)
		}
	}
}

func (c *DatabaseChanges) notifySubscribers(typ string, value interface{}) error {
	switch typ {
	case "DocumentChange":
		var documentChange *DocumentChange
		err := decodeJSONAsStruct(value, &documentChange)
		if err != nil {
			dbg("notifySubscribers: '%s' decodeJSONAsStruct failed with %s\n", typ, err)
			return err
		}
		c.mu.Lock()
		defer c.mu.Unlock()
		for _, s := range c.subscribers {
			s.sendDocumentChange(documentChange)
		}
	case "IndexChange":
		var indexChange *IndexChange
		err := decodeJSONAsStruct(value, &indexChange)
		if err != nil {
			dbg("notifySubscribers: '%s' decodeJSONAsStruct failed with %s\n", typ, err)
			return err
		}
		c.mu.Lock()
		defer c.mu.Unlock()
		for _, s := range c.subscribers {
			s.sendIndexChange(indexChange)
		}
	case "OperationStatusChange":
		var operationStatusChange *OperationStatusChange
		err := decodeJSONAsStruct(value, &operationStatusChange)
		if err != nil {
			dbg("notifySubscribers: '%s' decodeJSONAsStruct failed with %s\n", typ, err)
			return err
		}
		c.mu.Lock()
		defer c.mu.Unlock()
		for _, s := range c.subscribers {
			s.sendOperationStatusChange(operationStatusChange)
		}
	default:
		return fmt.Errorf("notifySubscribers: unsupported type '%s'", typ)
	}
	return nil
}

func (c *DatabaseChanges) notifyAboutError(err error) {
	// call onError handlers outside of a lock
	var handlers []func(error)
	c.mu.Lock()
	c.lastError = err
	if len(c.onError) > 0 {
		handlers = append(handlers, c.onError...)
	}
	c.mu.Unlock()

	for _, fn := range handlers {
		if fn != nil {
			fn(err)
		}
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
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				dbg("webSocketChangesProcessor.processMessages() ReadJSON() failed with %s\n", err)
			} else {
				err = nil
			}
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
					c.muConfirmations.Lock()
					future := c.confirmations[commandID]
					c.muConfirmations.Unlock()
					if future != nil {
						future.complete(nil)
					}
				}
			default:
				val := msgNode["Value"]
				c.notifySubscribers(typ, val)
			}
		}
	}
	// TODO: check for io.EOF for clean connection close?
	if !isCtxCancelled(ctx) {
		dbg("Not cancelled so calling notifyAboutError(), err = %v\n", err)
		c.notifyAboutError(err)
	}
}
