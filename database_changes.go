package ravendb

import (
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

var (
	_ IDatabaseChanges = &DatabaseChanges{}
)

type DatabaseChanges struct {
	_commandId int // TODO: make atomic?

	// TODO: why semaphore of size 1 and not a mutex?
	_semaphore chan bool

	_requestExecutor *RequestExecutor
	_conventions     *DocumentConventions
	_database        string

	_onDispose Runnable

	_client *websocket.Conn

	_task *CompletableFuture
	_cts  *CancellationTokenSource
	_tcs  *CompletableFuture

	mu             sync.Mutex // protects _confirmations and _counters maps
	_confirmations map[int]*CompletableFuture
	_counters      map[string]*DatabaseConnectionState

	_immediateConnection atomicInteger

	_connectionStatusEventHandlerIdx int
	_connectionStatusChanged         []func()
	onError                          []func(error)
}

func NewDatabaseChanges(requestExecutor *RequestExecutor, databaseName string, onDispose Runnable) *DatabaseChanges {
	res := &DatabaseChanges{
		_requestExecutor:                 requestExecutor,
		_conventions:                     requestExecutor.GetConventions(),
		_database:                        databaseName,
		_tcs:                             NewCompletableFuture(),
		_cts:                             NewCancellationTokenSource(),
		_onDispose:                       onDispose,
		_semaphore:                       make(chan bool, 1),
		_connectionStatusEventHandlerIdx: -1,
		_confirmations:                   map[int]*CompletableFuture{},
		_counters:                        map[string]*DatabaseConnectionState{},
	}

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
	res._connectionStatusEventHandlerIdx = res.AddConnectionStatusChanged(_connectionStatusEventHandler)
	return res
}

func (c *DatabaseChanges) onConnectionStatusChanged() {
	c.semAcquire()
	defer c.semRelease()

	if c.IsConnected() {
		c._tcs.Complete(c)
		return
	}

	if c._tcs.IsDone() {
		c._tcs = NewCompletableFuture()
	}
}

func (c *DatabaseChanges) IsConnected() bool {
	// TODO: should be protected aginst multi-threading
	return c._client != nil
}

func (c *DatabaseChanges) EnsureConnectedNow() error {
	_, err := c._tcs.Get()
	return err
}

func (c *DatabaseChanges) AddConnectionStatusChanged(handler func()) int {
	c.mu.Lock()
	idx := len(c._connectionStatusChanged)
	c._connectionStatusChanged = append(c._connectionStatusChanged, handler)
	c.mu.Unlock()
	return idx
}

func (c *DatabaseChanges) RemoveConnectionStatusChanged(handlerIdx int) {
	if handlerIdx != -1 {
		c._connectionStatusChanged[handlerIdx] = nil
	}
}

func (c *DatabaseChanges) ForIndex(indexName string) (IChangesObservable, error) {
	counter, err := c.getOrAddConnectionState("indexes/"+indexName, "watch-index", "unwatch-index", indexName)
	if err != nil {
		return nil, err
	}

	filter := func(notification interface{}) bool {
		v := notification.(*IndexChange)
		return strings.EqualFold(v.Name, indexName)
	}

	taskedObservable := NewChangesObservable(ChangesType_INDEX, counter, filter)
	return taskedObservable, nil
}

func (c *DatabaseChanges) getLastConnectionStateException() error {
	for _, counter := range c._counters {
		valueLastException := counter.lastException
		if valueLastException != nil {
			return valueLastException
		}
	}
	return nil
}

func (c *DatabaseChanges) ForDocument(docId string) (IChangesObservable, error) {
	counter, err := c.getOrAddConnectionState("docs/"+docId, "watch-doc", "unwatch-doc", docId)
	if err != nil {
		return nil, err
	}

	filter := func(notification interface{}) bool {
		v := notification.(*DocumentChange)
		return strings.EqualFold(v.ID, docId)
	}
	taskedObservable := NewChangesObservable(ChangesType_DOCUMENT, counter, filter)
	return taskedObservable, nil
}

func filterAlwaysTrue(notification interface{}) bool {
	dbg("filterAlwaysTrue: %T\n", notification)
	return true
}

func (c *DatabaseChanges) ForAllDocuments() (IChangesObservable, error) {
	counter, err := c.getOrAddConnectionState("all-docs", "watch-docs", "unwatch-docs", "")
	if err != nil {
		return nil, err
	}
	taskedObservable := NewChangesObservable(ChangesType_DOCUMENT, counter, filterAlwaysTrue)
	return taskedObservable, nil
}

func (c *DatabaseChanges) ForOperationId(operationId int) (IChangesObservable, error) {
	opIDStr := strconv.Itoa(operationId)
	counter, err := c.getOrAddConnectionState("operations/"+opIDStr, "watch-operation", "unwatch-operation", opIDStr)
	if err != nil {
		return nil, err
	}

	filter := func(notification interface{}) bool {
		v := notification.(*OperationStatusChange)
		return v.OperationID == operationId
	}
	taskedObservable := NewChangesObservable(ChangesType_OPERATION, counter, filter)
	return taskedObservable, nil
}

func (c *DatabaseChanges) ForAllOperations() (IChangesObservable, error) {
	counter, err := c.getOrAddConnectionState("all-operations", "watch-operations", "unwatch-operations", "")
	if err != nil {
		return nil, err
	}

	taskedObservable := NewChangesObservable(ChangesType_OPERATION, counter, filterAlwaysTrue)

	return taskedObservable, nil
}

func (c *DatabaseChanges) ForAllIndexes() (IChangesObservable, error) {
	counter, err := c.getOrAddConnectionState("all-indexes", "watch-indexes", "unwatch-indexes", "")
	if err != nil {
		return nil, err
	}

	taskedObservable := NewChangesObservable(ChangesType_INDEX, counter, filterAlwaysTrue)

	return taskedObservable, nil
}

func (c *DatabaseChanges) ForDocumentsStartingWith(docIdPrefix string) (IChangesObservable, error) {
	counter, err := c.getOrAddConnectionState("prefixes/"+docIdPrefix, "watch-prefix", "unwatch-prefix", docIdPrefix)
	if err != nil {
		return nil, err
	}
	filter := func(notification interface{}) bool {
		v := notification.(*DocumentChange)
		n := len(docIdPrefix)
		if n > len(v.ID) {
			return false
		}
		prefix := v.ID[:n]
		return strings.EqualFold(prefix, docIdPrefix)
	}

	taskedObservable := NewChangesObservable(ChangesType_DOCUMENT, counter, filter)

	return taskedObservable, nil
}

func (c *DatabaseChanges) ForDocumentsInCollection(collectionName string) (IChangesObservable, error) {
	if collectionName == "" {
		return nil, NewIllegalArgumentException("CollectionName cannot be empty")
	}

	counter, err := c.getOrAddConnectionState("collections/"+collectionName, "watch-collection", "unwatch-collection", collectionName)
	if err != nil {
		return nil, err
	}

	filter := func(notification interface{}) bool {
		v := notification.(*DocumentChange)
		return strings.EqualFold(collectionName, v.CollectionName)
	}

	taskedObservable := NewChangesObservable(ChangesType_DOCUMENT, counter, filter)

	return taskedObservable, nil
}

/*
func (c *DatabaseChanges) ForDocumentsInCollection(Class<?> clazz) (IChangesObservable, error) {
	String collectionName = _conventions.getCollectionName(clazz);
	return forDocumentsInCollection(collectionName);
}
*/

func (c *DatabaseChanges) ForDocumentsOfType(typeName string) (IChangesObservable, error) {
	if typeName == "" {
		return nil, NewIllegalArgumentException("TypeName cannot be empty")
	}

	encodedTypeName := UrlUtils_escapeDataString(typeName)

	counter, err := c.getOrAddConnectionState("types/"+typeName, "watch-type", "unwatch-type", encodedTypeName)
	if err != nil {
		return nil, err
	}

	filter := func(notification interface{}) bool {
		v := notification.(*DocumentChange)
		return strings.EqualFold(typeName,
			v.TypeName)
	}

	taskedObservable := NewChangesObservable(ChangesType_DOCUMENT, counter, filter)

	return taskedObservable, nil
}

/*
   public IChangesObservable<DocumentChange> ForDocumentsOfType(Class<?> clazz) {
       if (clazz == null) {
           throw new IllegalArgumentException("Clazz cannot be null");
       }

       String className = _conventions.getFindJavaClassName().apply(clazz);
       return forDocumentsOfType(className);
   }

*/

func (c *DatabaseChanges) invokeConnectionStatusChanged() {
	var dup []func()
	c.mu.Lock()
	for _, fn := range c._connectionStatusChanged {
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
	idx := len(c.onError)
	c.onError = append(c.onError, handler)
	return idx
}

func (c *DatabaseChanges) RemoveOnError(handlerIdx int) {
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
	c.mu.Lock()
	for _, confirmation := range c._confirmations {
		confirmation.Cancel(false)
	}
	c.semAcquire()
	c._client.Close()
	c._client = nil
	c.semRelease()

	c._cts.cancel()
	c._counters = nil
	c.mu.Unlock()
	c._task.Get()
	c.invokeConnectionStatusChanged()
	c.RemoveConnectionStatusChanged(c._connectionStatusEventHandlerIdx)
	if c._onDispose != nil {
		c._onDispose()
	}
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
		dbg("getOrAddConnectionState() onDisconnect()\n")
		if c.IsConnected() {
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
		dbg("getOrAddConnectionState() onConnect()\n")
		c.send(watchCommand, value)
	}

	counter = NewDatabaseConnectionState(onConnect, onDisconnect)
	c._counters[name] = counter

	if c._immediateConnection.get() != 0 {
		counter.onConnect()
	}
	return counter, nil
}

func (c *DatabaseChanges) semAcquire() {
	c._semaphore <- true
}

func (c *DatabaseChanges) semRelease() {
	<-c._semaphore
}

func (c *DatabaseChanges) send(command, value string) error {
	taskCompletionSource := NewCompletableFuture()

	c.semAcquire()

	c._commandId++
	currentCommandId := c._commandId

	o := struct {
		CommandID int    `json:"CommandId"`
		Command   string `json:"Command"`
		Param     string `json:"Param"`
	}{
		CommandID: currentCommandId,
		Command:   command,
		Param:     value,
	}

	err := c._client.WriteJSON(o)
	c._confirmations[currentCommandId] = taskCompletionSource

	c.semRelease()
	if err != nil {
		dbg("DatabaseChanges.send: WriteJSON() failed with %s\n", err)
		return err
	}

	dbg("DatabaseChanges.send: '%s' '%s', id: %d\n", command, value, currentCommandId)
	_, err = taskCompletionSource.GetWithTimeout(time.Second * 15)
	dbg("DatabaseChanges.send: got response for command %d\n", currentCommandId)
	return err
}

func toWebSocketPath(path string) string {
	path = strings.Replace(path, "http://", "ws://", -1)
	return strings.Replace(path, "https://", "wss://", -1)
}

func (c *DatabaseChanges) doWork() error {
	dbg("doWork()\n")
	_, err := c._requestExecutor.getPreferredNode()
	if err != nil {
		c.invokeConnectionStatusChanged()
		c.notifyAboutError(err)
		dbg("doWork(): err: %s\n", err)
		return err
	}

	urlString := c._requestExecutor.GetUrl() + "/databases/" + c._database + "/changes"
	urlString = toWebSocketPath(urlString)

	for {
		if c._cts.getToken().isCancellationRequested() {
			dbg("doWork(): isCancellationRequested()\n")
			return nil
		}

		var processor *WebSocketChangesProcessor
		var err error
		panicIf(c.IsConnected(), "impoosible: cannot be connected")

		dbg("doWork(): before dial %s\n", urlString)
		c._client, _, err = websocket.DefaultDialer.Dial(urlString, nil)
		dbg("doWork(): after dial\n")
		if err != nil {
			dbg("doWork(): websocket.Dial(%s) failed with %s()\n", urlString, err)
			time.Sleep(time.Second)
			continue
		}

		processor = NewWebSocketChangesProcessor(c._client)
		go processor.processMessages(c)
		c._immediateConnection.set(1)

		// TODO: make thread safe
		for _, counter := range c._counters {
			counter.onConnect()
		}
		c.invokeConnectionStatusChanged()
		_, err = processor.processing.Get()
		c.invokeConnectionStatusChanged()
		shouldReconnect := c.reconnectClient()

		for _, confirmation := range c._confirmations {
			confirmation.Cancel(false)
		}

		for k := range c._confirmations {
			delete(c._confirmations, k)
		}

		if !shouldReconnect {
			return nil
		}

		// wait before next retry
		time.Sleep(time.Second)
	}
}

func (c *DatabaseChanges) reconnectClient() bool {
	if c._cts.getToken().isCancellationRequested() {
		return false
	}

	c._immediateConnection.set(0)

	c.invokeConnectionStatusChanged()
	return true
}

func (c *DatabaseChanges) notifySubscribers(typ string, value interface{}, states []*DatabaseConnectionState) error {
	dbg("notifySubscribers: typ '%s', val: %v, len(states): %d\n", typ, value, len(states))
	switch typ {
	case "DocumentChange":
		var documentChange *DocumentChange
		err := decodeJSONAsStruct(value, &documentChange)
		if err != nil {
			dbg("notifySubscribers: decodeJSONAsStruct failed with %s\n", err)
			return err
		}
		for _, state := range states {
			state.sendDocumentChange(documentChange)
		}
	case "IndexChange":
		var indexChange *IndexChange
		err := decodeJSONAsStruct(value, &indexChange)
		if err != nil {
			return err
		}
		for _, state := range states {
			state.sendIndexChange(indexChange)
		}
	case "OperationStatusChange":
		var operationStatusChange *OperationStatusChange
		err := decodeJSONAsStruct(value, &operationStatusChange)
		if err != nil {
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
	if c._cts.getToken().isCancellationRequested() {
		return
	}

	c.invokeOnError(e)

	c.mu.Lock()
	defer c.mu.Unlock()
	for _, state := range c._counters {
		state.error(e)
	}
}

type WebSocketChangesProcessor struct {
	processing *CompletableFuture
	client     *websocket.Conn
}

func NewWebSocketChangesProcessor(client *websocket.Conn) *WebSocketChangesProcessor {
	return &WebSocketChangesProcessor{
		processing: NewCompletableFuture(),
		client:     client,
	}
}

func (p *WebSocketChangesProcessor) processMessages(changes *DatabaseChanges) {
	var err error
	dbg("WebSocketChangesProcessor.processMessages()\n")
	for {
		var msgArray []interface{} // an array of objects
		dbg("WebSocketChangesProcessor.processMessages() before ReadJSON()\n")
		err = p.client.ReadJSON(&msgArray)
		if err != nil {
			dbg("WebSocketChangesProcessor.processMessages() ReadJSON() failed with %s\n", err)
			break
		}
		dbg("WebSocketChangesProcessor.processMessages() msgArray: %T %v\n", msgArray, msgArray)
		for _, msgNodeV := range msgArray {
			msgNode := msgNodeV.(map[string]interface{})
			typ, _ := jsonGetAsString(msgNode, "Type")
			switch typ {
			case "Error":
				errStr, _ := jsonGetAsString(msgNode, "Error")
				changes.notifyAboutError(NewRuntimeException("%s", errStr))
			case "Confirm":
				commandID, ok := jsonGetAsInt(msgNode, "CommandId")
				if ok {
					changes.semAcquire()
					future := changes._confirmations[commandID]
					changes.semRelease()
					if future != nil {
						future.Complete(nil)
					}
				}
			default:
				val := msgNode["Value"]
				var states []*DatabaseConnectionState
				for _, state := range changes._counters {
					states = append(states, state)
				}
				changes.notifySubscribers(typ, val, states)
			}

		}
	}
	// TODO: check for io.EOF for clean connection close?
	changes.notifyAboutError(err)
	p.processing.CompleteExceptionally(err)
}
