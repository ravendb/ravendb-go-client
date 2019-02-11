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

	mu sync.Mutex
}

func (s *changeSubscribers) registerOnDocumentChange(fn func(*DocumentChange)) int {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.onDocumentChange = append(s.onDocumentChange, fn)
	return len(s.onDocumentChange) - 1
}

func (s *changeSubscribers) unregisterOnDocumentChange(idx int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.onDocumentChange[idx] = nil
}

func (s *changeSubscribers) registerOnIndexChange(fn func(*IndexChange)) int {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.onIndexChange = append(s.onIndexChange, fn)
	return len(s.onIndexChange) - 1
}

func (s *changeSubscribers) unregisterOnIndexChange(idx int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.onIndexChange[idx] = nil
}

func (s *changeSubscribers) registerOnOperationStatusChange(fn func(*OperationStatusChange)) int {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.onOperationStatusChange = append(s.onOperationStatusChange, fn)
	return len(s.onOperationStatusChange) - 1
}

func (s *changeSubscribers) unregisterOnOperationStatusChange(idx int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.onOperationStatusChange[idx] = nil
}

func (s *changeSubscribers) sendDocumentChange(change *DocumentChange) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, f := range s.onDocumentChange {
		if f != nil {
			f(change)
		}
	}
}

func (s *changeSubscribers) sendIndexChange(change *IndexChange) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, f := range s.onIndexChange {
		if f != nil {
			f(change)
		}
	}
}

func (s *changeSubscribers) sendOperationStatusChange(change *OperationStatusChange) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, f := range s.onOperationStatusChange {
		if f != nil {
			f(change)
		}
	}
}

func (s *changeSubscribers) hasRegisteredHandlers() bool {
	s.mu.Lock()
	defer s.mu.Unlock()

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

type commandConfirmation struct {
	timeStart    time.Time
	duration     time.Duration
	ch           chan bool
	completed    int32 // atomic
	wasCancelled bool
}

func newCommandConfirmation() *commandConfirmation {
	return &commandConfirmation{
		timeStart: time.Now(),
		ch:        make(chan bool, 1), // don't block the sender
	}
}

func (c *commandConfirmation) confirm(wasCancelled bool) {
	new := atomic.AddInt32(&c.completed, 1)
	if new > 1 {
		// was already completed
		return
	}
	c.duration = time.Since(c.timeStart)
	c.wasCancelled = wasCancelled
	c.ch <- true
}

func (c *commandConfirmation) waitForConfirmation(timeout time.Duration) bool {
	select {
	case <-c.ch:
		return true
	case <-time.After(timeout):
		return false
	}
}

type commandToSend struct {
	command string
	value   string
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

	doWorkCancel    context.CancelFunc
	cancelRequested atomicBool

	// will be notified if doWork goroutine finishes
	chanWorkCompleted chan error

	// will be notified if we connect or fail to connect
	// allows waiting for connection being established
	chanIsConnected chan error

	chanSend chan (*commandToSend)

	subscribers sync.Map // string => *changeSubscribers

	mu sync.Mutex

	confirmations sync.Map // int -> *commandConfirmation

	immediateConnection atomicBool

	connectionStatusChanged []func()
	onError                 []func(error)

	lastError error
}

func (c *DatabaseChanges) isCancelRequested() bool {
	return c.cancelRequested.isTrue()
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
		chanIsConnected: make(chan error, 1),
		onDispose:       onDispose,
		chanSend:        make(chan *commandToSend, 16),
	}

	res.chanWorkCompleted = make(chan error, 1)
	var ctx context.Context
	ctx, res.doWorkCancel = context.WithCancel(context.Background())
	go func() {
		err := res.doWork(ctx)
		res.chanWorkCompleted <- err
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
	err := <-c.chanIsConnected
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
	idx := subscribers.registerOnIndexChange(filtered)
	cancel := func() {
		subscribers.unregisterOnIndexChange(idx)
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
	idx := subscribers.registerOnDocumentChange(filtered)
	cancel := func() {
		subscribers.unregisterOnDocumentChange(idx)
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

	idx := subscribers.registerOnDocumentChange(cb)
	cancel := func() {
		subscribers.unregisterOnDocumentChange(idx)
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

	idx := subscribers.registerOnOperationStatusChange(filtered)
	cancel := func() {
		subscribers.unregisterOnOperationStatusChange(idx)
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

	idx := subscribers.registerOnOperationStatusChange(cb)
	cancel := func() {
		subscribers.unregisterOnOperationStatusChange(idx)
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

	idx := subscribers.registerOnIndexChange(cb)
	cancel := func() {
		subscribers.unregisterOnIndexChange(idx)
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

	idx := subscribers.registerOnDocumentChange(filtered)
	cancel := func() {
		subscribers.unregisterOnDocumentChange(idx)
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

	idx := subscribers.registerOnDocumentChange(filtered)
	cancel := func() {
		subscribers.unregisterOnDocumentChange(idx)
		c.maybeDisconnectSubscribers(subscribers)
	}
	return cancel, nil
}

func (c *DatabaseChanges) ForDocumentsInCollectionOfType(clazz reflect.Type, cb func(*DocumentChange)) (CancelFunc, error) {
	collectionName := c.conventions.GetCollectionName(clazz)
	return c.ForDocumentsInCollection(collectionName, cb)
}

func (c *DatabaseChanges) invokeConnectionStatusChanged() {
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

func (c *DatabaseChanges) cancelOutstandingRequests() {
	rangeFn := func(key, val interface{}) bool {
		confirmation := val.(*commandConfirmation)
		confirmation.confirm(true)
		return true
	}
	c.confirmations.Range(rangeFn)
}

func (c *DatabaseChanges) Close() {
	c.doWorkCancel()
	(&c.cancelRequested).set(true)

	c.cancelOutstandingRequests()

	client := c.getWsClient()
	if client != nil {
		//fmt.Printf("DatabaseChanges.Close(): before client.Close()\n")
		err := client.Close()
		if err != nil {
			dbg("DatabaseChanges.Close(): client.Close() failed with %s\n", err)
		}
		c.setWsClient(nil)
	}

	<-c.chanWorkCompleted

	c.invokeConnectionStatusChanged()
	if c.onDispose != nil {
		c.onDispose()
	}
}

func (c *DatabaseChanges) getOrAddSubscribers(name string, watchCommand string, unwatchCommand string, value string) (*changeSubscribers, error) {
	// must be called while holding mu lock
	subscribersI, ok := c.subscribers.Load(name)

	if ok {
		return subscribersI.(*changeSubscribers), nil
	}

	subscribers := &changeSubscribers{
		name:           name,
		watchCommand:   watchCommand,
		unwatchCommand: unwatchCommand,
		commandValue:   value,
	}
	c.subscribers.Store(name, subscribers)

	if c.IsConnected() {
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
	name := subscribers.name
	if c.IsConnected() {
		go c.send(subscribers.unwatchCommand, subscribers.commandValue)
		// ignoring error: if we are not connected then we unsubscribed
		// already because connections drops with all subscriptions
	}
	c.subscribers.Delete(name)
}

func (c *DatabaseChanges) connectSubscribers(subscribers *changeSubscribers) error {
	return c.send(subscribers.watchCommand, subscribers.commandValue)
}

func (c *DatabaseChanges) send(command, value string) error {
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

	confirmation := newCommandConfirmation()

	c.confirmations.Store(currentCommandID, confirmation)
	confirmation.waitForConfirmation(time.Second * 15)
	return nil
}

func toWebSocketPath(path string) string {
	path = strings.Replace(path, "http://", "ws://", -1)
	return strings.Replace(path, "https://", "wss://", -1)
}

func (c *DatabaseChanges) doWorkInner(ctx context.Context) error {
	var err error
	dialer := *websocket.DefaultDialer
	dialer.HandshakeTimeout = time.Second * 2
	re := c.requestExecutor
	if re.Certificate != nil || re.TrustStore != nil {
		dialer.TLSClientConfig, err = newTLSConfig(re.Certificate, re.TrustStore)
		if err != nil {
			return err
		}
	}

	urlString, err := c.requestExecutor.GetURL()
	if err != nil {
		return err
	}
	urlString += "/databases/" + c.database + "/changes"
	urlString = toWebSocketPath(urlString)

	ctxDial, _ := context.WithTimeout(ctx, time.Second*5)
	var client *websocket.Conn
	client, _, err = dialer.DialContext(ctxDial, urlString, nil)
	c.setWsClient(client)

	// recheck cancellation because it might have been cancelled
	// since DialContext()
	if c.isCancelRequested() {
		if client != nil {
			client.Close()
			c.setWsClient(nil)
		}
		return err
	}

	if err != nil {
		time.Sleep(time.Second)
	}
	c.chanIsConnected <- nil
	close(c.chanIsConnected)

	connectFn := func(key, value interface{}) bool {
		subscribers := value.(*changeSubscribers)
		c.connectSubscribers(subscribers)
		return true
	}
	c.subscribers.Range(connectFn)
	return nil
}

func (c *DatabaseChanges) doWork(ctx context.Context) error {
	_, err := c.requestExecutor.getPreferredNode()
	if err != nil {
		c.invokeConnectionStatusChanged()
		c.notifyAboutError(err)
		c.chanIsConnected <- err
		close(c.chanIsConnected)
		return err
	}

	for {

		if c.isCancelRequested() {
			return nil
		}

		panicIf(c.IsConnected(), "impoosible: cannot be connected")
		c.doWorkInner(ctx)

		c.invokeConnectionStatusChanged()
		c.processMessages(ctx)
		c.invokeConnectionStatusChanged()
		c.setWsClient(nil)

		shouldReconnect := !c.isCancelRequested()

		c.cancelOutstandingRequests()
		c.confirmations = sync.Map{}

		if !shouldReconnect {
			return nil
		}

		c.chanIsConnected = make(chan error, 1)

		// wait before next retry
		time.Sleep(time.Second)
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
		fn := func(key, value interface{}) bool {
			s := value.(*changeSubscribers)
			s.sendDocumentChange(documentChange)
			return true
		}
		c.subscribers.Range(fn)
	case "IndexChange":
		var indexChange *IndexChange
		err := decodeJSONAsStruct(value, &indexChange)
		if err != nil {
			dbg("notifySubscribers: '%s' decodeJSONAsStruct failed with %s\n", typ, err)
			return err
		}
		fn := func(key, value interface{}) bool {
			s := value.(*changeSubscribers)
			s.sendIndexChange(indexChange)
			return true
		}
		c.subscribers.Range(fn)
	case "OperationStatusChange":
		var operationStatusChange *OperationStatusChange
		err := decodeJSONAsStruct(value, &operationStatusChange)
		if err != nil {
			dbg("notifySubscribers: '%s' decodeJSONAsStruct failed with %s\n", typ, err)
			return err
		}
		fn := func(key, value interface{}) bool {
			s := value.(*changeSubscribers)
			s.sendOperationStatusChange(operationStatusChange)
			return true
		}
		c.subscribers.Range(fn)
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
				if !c.isCancelRequested() {
					c.notifyAboutError(newRuntimeError("%s", errStr))
				}
			case "Confirm":
				commandID, ok := jsonGetAsInt(msgNode, "CommandId")
				if ok {
					v, ok := c.confirmations.Load(commandID)
					if ok {
						confirmation := v.(*commandConfirmation)
						confirmation.confirm(false)
					}
				}
			default:
				val := msgNode["Value"]
				c.notifySubscribers(typ, val)
			}
		}
	}

	if err != nil && !c.isCancelRequested() {
		dbg("Not cancelled so calling notifyAboutError(), err = %v\n", err)
		c.notifyAboutError(err)
	}
}
