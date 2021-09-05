package ravendb

import (
	"context"
	"encoding/json"
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

// enableDatabaseChangesDebugOutput enables debug logging of
// database_changes.go code
var enableDatabaseChangesDebugOutput bool

// for debugging DatabaseChanges code
func dcdbg(format string, args ...interface{}) {
	// change to true to enable debug output
	if enableDatabaseChangesDebugOutput {
		fmt.Printf(format, args...)
	}
}

type changeSubscribers struct {
	name           string // is key of DatabaseChanges.subscribers
	watchCommand   string
	unwatchCommand string
	commandValue   string

	onDocumentChange        sync.Map // int -> func(*DocumentChange)
	onIndexChange           sync.Map // int -> func(*IndexChange)
	onOperationStatusChange sync.Map // int -> func(*OperationStatusChange)

	nextID int32 // atomic
}

func (s *changeSubscribers) getNextID() int {
	n := atomic.AddInt32(&s.nextID, 1)
	return int(n)
}

func (s *changeSubscribers) registerOnDocumentChange(fn func(*DocumentChange)) int {
	id := s.getNextID()
	s.onDocumentChange.Store(id, fn)
	return id
}

func (s *changeSubscribers) unregisterOnDocumentChange(id int) {
	s.onDocumentChange.Delete(id)
}

func (s *changeSubscribers) registerOnIndexChange(fn func(*IndexChange)) int {
	id := s.getNextID()
	s.onIndexChange.Store(id, fn)
	return id
}

func (s *changeSubscribers) unregisterOnIndexChange(id int) {
	s.onIndexChange.Delete(id)
}

func (s *changeSubscribers) registerOnOperationStatusChange(fn func(*OperationStatusChange)) int {
	id := s.getNextID()
	s.onOperationStatusChange.Store(id, fn)
	return id
}

func (s *changeSubscribers) unregisterOnOperationStatusChange(id int) {
	s.onOperationStatusChange.Delete(id)
}

func (s *changeSubscribers) sendDocumentChange(change *DocumentChange) {
	s.onDocumentChange.Range(func(k, v interface{}) bool {
		f := v.(func(documentChange *DocumentChange))
		f(change)
		return true
	})
}

func (s *changeSubscribers) sendIndexChange(change *IndexChange) {
	s.onIndexChange.Range(func(k, v interface{}) bool {
		f := v.(func(documentChange *IndexChange))
		f(change)
		return true
	})
}

func (s *changeSubscribers) sendOperationStatusChange(change *OperationStatusChange) {
	s.onOperationStatusChange.Range(func(k, v interface{}) bool {
		f := v.(func(documentChange *OperationStatusChange))
		f(change)
		return true
	})
}

func (s *changeSubscribers) hasRegisteredHandlers() bool {
	// there is no sync.Map.Count() so we have to enumerate to see
	// if there are any registered handlers
	hasHandlers := false
	fn := func(k, v interface{}) bool {
		hasHandlers = true
		// only care about the first one
		return false
	}
	s.onDocumentChange.Range(fn)
	s.onIndexChange.Range(fn)
	s.onOperationStatusChange.Range(fn)
	return hasHandlers
}

func newDatabaseChangesCommand(id int, command string, value string) *databaseChangesCommand {
	return &databaseChangesCommand{
		id:        id,
		command:   command,
		value:     value,
		timeStart: time.Now(),
		ch:        make(chan bool, 1), // don't block the sender
	}
}

type databaseChangesCommand struct {
	// the data we Send
	id      int
	command string
	value   string

	// used to wait for notifications
	timeStart    time.Time
	duration     time.Duration
	ch           chan bool
	completed    int32 // atomic
	wasCancelled bool
}

func (c *databaseChangesCommand) confirm(wasCancelled bool) {
	v := atomic.AddInt32(&c.completed, 1)
	if v > 1 {
		// was already completed
		return
	}
	c.duration = time.Since(c.timeStart)
	c.wasCancelled = wasCancelled
	c.ch <- true
}

func (c *databaseChangesCommand) waitForConfirmation(timeout time.Duration) bool {
	select {
	case <-c.ch:
		return true
	case <-time.After(timeout):
		return false
	}
}

// DatabaseChanges notifies about changes to a database
type DatabaseChanges struct {
	commandID           int32 // atomic
	connStatusChangedID int32 // atomic

	requestExecutor *RequestExecutor
	conventions     *DocumentConventions
	database        string

	onClose func()

	ctxCancel    context.Context
	doWorkCancel context.CancelFunc

	// will be notified if we connect or fail to connect
	// allows waiting for connection being established
	chIsConnected chan error

	chCommands      chan *databaseChangesCommand
	chWorkCompleted chan error

	subscribers sync.Map // string => *changeSubscribers

	mu sync.Mutex

	// commands that have been sent to the server but not confirmed
	outstandingCommands sync.Map // int -> *commandConfirmation

	// TODO: why is this not used?
	// immediateConnection int32 // atomic

	connectionStatusChanged []func()
	onError                 []func(error)

	lastError atomic.Value // error
}

func (c *DatabaseChanges) isClosed() bool {
	select {
	case <-c.ctxCancel.Done():
		return true
	default:
		return false
	}
}

func (c *DatabaseChanges) nextCommandID() int {
	v := atomic.AddInt32(&c.commandID, 1)
	return int(v)
}

func newDatabaseChanges(requestExecutor *RequestExecutor, databaseName string, onClose func()) *DatabaseChanges {
	res := &DatabaseChanges{
		requestExecutor: requestExecutor,
		conventions:     requestExecutor.GetConventions(),
		database:        databaseName,
		chIsConnected:   make(chan error, 1),
		onClose:         onClose,
		chWorkCompleted: make(chan error, 1),
		chCommands:      make(chan *databaseChangesCommand, 32),
	}

	res.ctxCancel, res.doWorkCancel = context.WithCancel(context.Background())

	go func() {
		_, err := requestExecutor.getPreferredNode()
		if err != nil {
			dcdbg("newDatabaseChanges: getPreferredNode() failed with %s\n", err)
			res.notifyAboutError(err)
			res.chWorkCompleted <- err
			close(res.chWorkCompleted)
			return
		}

		err = res.doWork(res.ctxCancel)
		res.chWorkCompleted <- err
		close(res.chWorkCompleted)
	}()

	return res
}

func (c *DatabaseChanges) EnsureConnectedNow() error {
	select {
	case <-c.ctxCancel.Done():
		dcdbg("DatabaseChanges(): EnsureConnectedNow(): is closed\n")
		return errors.New("DatabaseChanges.EnsureConnectedNow(): Close() has been called")
	case err := <-c.chWorkCompleted:
		dcdbg("DatabaseChanges(): EnsureConnectedNow(): chnWorkCompleted notified\n")
		return err
	case err := <-c.chIsConnected:
		dcdbg("DatabaseChanges(): EnsureConnectedNow(): chanIsConnected notified\n")
		return err
	case <-time.After(time.Second * 15):
		dcdbg("DatabaseChanges(): EnsureConnectedNow(): timed out waiting for connection\n")
		return errors.New("timed out waiting for connection")
	}
}

func (c *DatabaseChanges) AddConnectionStatusChanged(handler func()) int {
	c.mu.Lock()
	idx := len(c.connectionStatusChanged)
	c.connectionStatusChanged = append(c.connectionStatusChanged, handler)
	c.mu.Unlock()
	return idx
}

func (c *DatabaseChanges) RemoveConnectionStatusChanged(handlerID int) {
	if handlerID != -1 {
		c.mu.Lock()
		c.connectionStatusChanged[handlerID] = nil
		c.mu.Unlock()
	}
}

type CancelFunc func()

// ForIndex registers a callback that will be called for changes in an index with a given name.
// It returns a function to call to unregister the callback.
func (c *DatabaseChanges) ForIndex(indexName string, cb func(*IndexChange)) (CancelFunc, error) {
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
	if v := c.lastError.Load(); v == nil {
		return nil
	} else {
		return v.(error)
	}
}

// ForDocument registers a callback that will be called for changes on a ocument with a given id
// It returns a function to call to unregister the callback.
func (c *DatabaseChanges) ForDocument(docID string, cb func(*DocumentChange)) (CancelFunc, error) {
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

// ForAllDocuments registers a callback that will be called for changes on all documents.
// It returns a function to call to unregister the callback.
func (c *DatabaseChanges) ForAllDocuments(cb func(*DocumentChange)) (CancelFunc, error) {
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

// ForOperationID registers a callback that will be called when a change happens to operation with a given id.
// It returns a function to call to unregister the callback.
func (c *DatabaseChanges) ForOperationID(operationID int64, cb func(*OperationStatusChange)) (CancelFunc, error) {
	opIDStr := i64toa(operationID)
	subscribers, err := c.getOrAddSubscribers("operations/"+opIDStr, "watch-operation", "unwatch-operation", opIDStr)
	if err != nil {
		return nil, err
	}

	filtered := func(change *OperationStatusChange) {
		if change.OperationID == operationID {
			cb(change)
		}
	}

	idx := subscribers.registerOnOperationStatusChange(filtered)
	cancel := func() {
		subscribers.unregisterOnOperationStatusChange(idx)
		c.maybeDisconnectSubscribers(subscribers)
	}
	return cancel, nil
}

// ForAllOperations registers a callback that will be called when any operation changes status.
// It returns a function to call to unregister the callback.
func (c *DatabaseChanges) ForAllOperations(cb func(change *OperationStatusChange)) (CancelFunc, error) {
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

// ForAllIndexes registers a callback that will be called when a change on any index happens.
// It returns a function to call to unregister the callback.
func (c *DatabaseChanges) ForAllIndexes(cb func(*IndexChange)) (CancelFunc, error) {
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

// ForDocumentsStartingWith registers a callback that will be called for changes on documents whose id starts with
// a given prefix. It returns a function to call to unregister the callback.
func (c *DatabaseChanges) ForDocumentsStartingWith(docIDPrefix string, cb func(*DocumentChange)) (CancelFunc, error) {
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

// ForDocumentsInCollection registers a callback that will be called on changes for documents in a given collection.
// It returns a function to call to unregister the callback.
func (c *DatabaseChanges) ForDocumentsInCollection(collectionName string, cb func(*DocumentChange)) (CancelFunc, error) {
	if collectionName == "" {
		return nil, newIllegalArgumentError("CollectionName cannot be empty")
	}

	subscribers, err := c.getOrAddSubscribers("collections/"+collectionName, "watch-collection", "unwatch-collection", collectionName)
	if err != nil {
		return nil, err
	}

	filtered := func(change *DocumentChange) {
		if strings.EqualFold(collectionName, change.CollectionName) {
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

// ForDocumentsInCollectionOfType registers a callback that will be called on changes for documents of a given type.
// It returns a function to call to unregister the callback.
func (c *DatabaseChanges) ForDocumentsInCollectionOfType(clazz reflect.Type, cb func(*DocumentChange)) (CancelFunc, error) {
	collectionName := c.conventions.getCollectionName(clazz)
	return c.ForDocumentsInCollection(collectionName, cb)
}

func (c *DatabaseChanges) invokeConnectionStatusChanged() {
	// make a copy of callers so that we can call outside of a lock
	c.mu.Lock()
	dup := append([]func(){}, c.connectionStatusChanged...)
	c.mu.Unlock()

	for _, fn := range dup {
		if fn != nil {
			fn()
		}
	}
}

func (c *DatabaseChanges) AddOnError(handler func(error)) int {
	c.mu.Lock()
	defer c.mu.Unlock()
	idx := len(c.onError)
	c.onError = append(c.onError, handler)
	return idx
}

func (c *DatabaseChanges) RemoveOnError(handlerID int) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.onError[handlerID] = nil
}

// cancel outstanding commands to unblock those waiting for their completion
func (c *DatabaseChanges) cancelOutstandingCommands() {
	rangeFn := func(key, val interface{}) bool {
		cmd := val.(*databaseChangesCommand)
		cmd.confirm(true)
		dcdbg("DatabaseChanges: cancelled outstanding command %d '%s %s'\n", cmd.id, cmd.command, cmd.value)
		c.outstandingCommands.Delete(key)
		return true
	}
	c.outstandingCommands.Range(rangeFn)
}

// Close closes DatabaseChanges and release its resources
func (c *DatabaseChanges) Close() {
	dcdbg("DatabaseChanges: Close()\n")
	//debug.PrintStack()
	select {
	case <-c.chWorkCompleted:
		dcdbg("DatabaseChanges.Close(): has already been closed because chanWorkCompleted notified\n")
	default:
		// no-op
	}

	c.doWorkCancel()
	c.cancelOutstandingCommands()

	select {
	case <-c.chWorkCompleted:
	case <-time.After(time.Second * 5):
		dcdbg("DatabaseChanges.Close(): timed out waiting for chanWorkCompleted\n")
	}

	if c.onClose != nil {
		c.onClose()
	}
}

func fmtDCCommand(cmd, value string) string {
	if value == "" {
		return cmd
	}
	return cmd + " " + value
}

func (c *DatabaseChanges) getOrAddSubscribers(name string, watchCommand string, unwatchCommand string, value string) (*changeSubscribers, error) {
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
	if err := c.connectSubscribers(subscribers); err != nil {
		return nil, err
	}
	return subscribers, nil
}

func (c *DatabaseChanges) maybeDisconnectSubscribers(subscribers *changeSubscribers) {
	if !subscribers.hasRegisteredHandlers() {
		c.disconnectSubscribers(subscribers)
	}
}

func (c *DatabaseChanges) disconnectSubscribers(subscribers *changeSubscribers) {
	_ = c.send(subscribers.unwatchCommand, subscribers.commandValue, false)
	// ignoring error: if we are not connected then we unsubscribed
	// already because connections drops with all subscriptions
	c.subscribers.Delete(subscribers.name)
}

func (c *DatabaseChanges) connectSubscribers(subscribers *changeSubscribers) error {
	return c.send(subscribers.watchCommand, subscribers.commandValue, true)
}

func (c *DatabaseChanges) send(command, value string, waitForConfirmation bool) error {
	if c.isClosed() {
		return errors.New("Send() called after Close()")
	}

	id := c.nextCommandID()
	cmd := newDatabaseChangesCommand(id, command, value)
	dcdbg("DatabaseChanges: Send(): command id: %d, command: '%s', wait: %v\n", id, fmtDCCommand(command, value), waitForConfirmation)
	if waitForConfirmation {
		c.outstandingCommands.Store(id, cmd)
	}

	c.mu.Lock()
	chCommands := c.chCommands
	c.mu.Unlock()
	chCommands <- cmd

	if waitForConfirmation {
		cmd.waitForConfirmation(time.Second * 15)
	}
	return nil
}

func startSendWorker(conn *websocket.Conn, chCommands chan *databaseChangesCommand) chan error {
	chFailed := make(chan error, 1)
	go func() {
		dcdbg("starting a chCommands reading loop\n")
		for cmd := range chCommands {
			dcdbg("got command with id %d to Send. Command: %s, param: %s\n", cmd.id, cmd.command, cmd.value)
			o := struct {
				CommandID int    `json:"CommandId"`
				Command   string `json:"Command"`
				Param     string `json:"Param"`
			}{
				CommandID: cmd.id,
				Command:   cmd.command,
				Param:     cmd.value,
			}
			err := conn.SetWriteDeadline(time.Now().Add(time.Second * 3))
			if err != nil {
				dcdbg("DatabaseChanges: SetWriteDeadline() failed with %s\n", err)
				chFailed <- err
				return
			}
			err = conn.WriteJSON(o)
			if err != nil {
				dcdbg("DatabaseChanges: conn.WriteJSON() failed with %s\n", err)
				chFailed <- err
				return
			}
			dcdbg("wrote command with id %d to socket\n", cmd.id)
		}
		dcdbg("DatabaseChanges: Send worker finished\n")
	}()
	return chFailed
}

func toWebSocketPath(path string) string {
	path = strings.Replace(path, "http://", "ws://", -1)
	return strings.Replace(path, "https://", "wss://", -1)
}

// returns true if we should try to reconnect
func (c *DatabaseChanges) doWorkInner(ctx context.Context) (error, bool) {
	var err error
	dialer := *websocket.DefaultDialer
	dialer.HandshakeTimeout = time.Second * 2

	re := c.requestExecutor
	if re.Certificate != nil || re.TrustStore != nil {
		dialer.TLSClientConfig, err = newTLSConfig(re.Certificate, re.TrustStore)
		if err != nil {
			return err, false
		}
	}

	urlString, err := c.requestExecutor.GetURL()
	if err != nil {
		return err, false
	}
	urlString += "/databases/" + c.database + "/changes"
	urlString = toWebSocketPath(urlString)

	ctxDial, cancel := context.WithTimeout(ctx, time.Second*2)
	var client *websocket.Conn
	client, _, err = dialer.DialContext(ctxDial, urlString, nil)
	cancel()

	if err != nil {
		dcdbg("DatabaseChanges: dialer.DialContext failed with '%s'\n", err)
		return err, false
	}

	var chWriterFailed chan error
	chWriterFailed = startSendWorker(client, c.chCommands)
	var chReaderFailed chan error
	chReaderFailed = c.startProcessMessagesWorker(ctx, client)

	connectFn := func(key, value interface{}) bool {
		subscribers := value.(*changeSubscribers)
		_ = c.connectSubscribers(subscribers)
		return true
	}
	c.subscribers.Range(connectFn)

	c.invokeConnectionStatusChanged()

	c.chIsConnected <- nil
	// close so that subsequent channel reads also return immediately
	close(c.chIsConnected)

	shouldReconnect := true
	err = nil
	select {
	case err = <-chWriterFailed:
		dcdbg("DatabaseChanges: writer failed with '%s'\n", err)
	case err = <-chReaderFailed:
		if err != nil {
			dcdbg("DatabaseChanges: reader failed with '%s'\n", err)
		} else {
			dcdbg("DatabaseChanges: reader finished cleanly\n")
		}
	case <-ctx.Done():
		dcdbg("cancellation requested\n")
		shouldReconnect = false
	}

	c.mu.Lock()
	chCommands := c.chCommands
	c.chCommands = make(chan *databaseChangesCommand, 32)
	c.mu.Unlock()
	close(chCommands)
	_ = client.Close()

	c.invokeConnectionStatusChanged()
	return err, shouldReconnect
}

func (c *DatabaseChanges) doWork(ctx context.Context) error {
	for {
		err, shouldReconnect := c.doWorkInner(ctx)
		if err != nil {
			dcdbg("DatabaseChanges: doWorkInner() failed with '%s'\n", err)
		}
		c.cancelOutstandingCommands()
		if !shouldReconnect {
			return err
		}
		// wait before next retry
		time.Sleep(time.Second)
	}
}

func (c *DatabaseChanges) notifySubscribers(typ string, value interface{}) error {
	dcdbg("DatabnaseChanges: notifySubscribers(): %s, %v\n", typ, value)
	switch typ {
	case "DocumentChange":
		var documentChange *DocumentChange
		err := decodeJSONAsStruct(value, &documentChange)
		if err != nil {
			dcdbg("notifySubscribers: '%s' decodeJSONAsStruct failed with %s\n", typ, err)
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
			dcdbg("notifySubscribers: '%s' decodeJSONAsStruct failed with %s\n", typ, err)
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
			dcdbg("notifySubscribers: '%s' decodeJSONAsStruct failed with %s\n", typ, err)
			return err
		}
		fn := func(key, value interface{}) bool {
			s := value.(*changeSubscribers)
			s.sendOperationStatusChange(operationStatusChange)
			return true
		}
		c.subscribers.Range(fn)
	default:
		dcdbg("DatabnaseChanges: notifySubscribers(): unsupported type '%s'\n", typ)
		return fmt.Errorf("notifySubscribers: unsupported type '%s'", typ)
	}
	return nil
}

func (c *DatabaseChanges) notifyAboutError(err error) {
	if c.isClosed() {
		return
	}
	panicIf(err == nil, "err is nil")
	c.lastError.Store(err)

	// make a copy so that we can call outside of a lock
	c.mu.Lock()
	handlers := append([]func(error){}, c.onError...)
	c.mu.Unlock()

	for _, fn := range handlers {
		if fn != nil {
			fn(err)
		}
	}
}

func (c *DatabaseChanges) startProcessMessagesWorker(ctx context.Context, conn *websocket.Conn) chan error {
	chFailed := make(chan error, 1)
	go func() {
		var err error
		for {
			var msgArray []interface{} // an array of objects
			err = conn.ReadJSON(&msgArray)
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					dcdbg("DatabaseChanges: ReadJSON() failed with %s\n", err)
				} else {
					dcdbg("DatabaseChanges: ReadJSON() failed with %s, turning into no error\n", err)
					err = nil
				}
				break
			}
			if len(msgArray) == 0 {
				continue
			}

			if enableDatabaseChangesDebugOutput {
				s, _ := json.Marshal(msgArray)
				dcdbg("DatatabaseChange: received messages:\n%s\n", s)
			}

			for _, msgNodeV := range msgArray {
				msgNode := msgNodeV.(map[string]interface{})
				typ, ok := jsonGetAsText(msgNode, "Type")
				if !ok {
					continue
				}
				switch typ {
				case "Error":
					errStr, _ := jsonGetAsText(msgNode, "Error")
					c.notifyAboutError(newRuntimeError("%s", errStr))
				case "Confirm":
					commandID, ok := jsonGetAsInt(msgNode, "CommandId")
					if ok {
						v, ok := c.outstandingCommands.Load(commandID)
						if ok {
							cmd := v.(*databaseChangesCommand)
							cmd.confirm(false)
							dcdbg("DatabaseChanges: confirmed command id %d, command '%s'\n", cmd.id, fmtDCCommand(cmd.command, cmd.value))
						}
					}
				default:
					if val, ok := msgNode["Value"]; ok {
						// sometimes a message is {"TopologyChange":true}
						_ = c.notifySubscribers(typ, val)
					}
				}
			}
		}
		if err != nil {
			dcdbg("Not cancelled so calling notifyAboutError(), err = %v\n", err)
			c.notifyAboutError(err)
		}
		chFailed <- err
	}()
	return chFailed
}
