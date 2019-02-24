package ravendb

import (
	"encoding/json"
	"log"
	"net"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// SubscriptionWorker describes subscription worker
type SubscriptionWorker struct {
	clazz     reflect.Type
	revisions bool
	logger    *log.Logger
	store     *DocumentStore
	dbName    string

	cancellationRequested int32 // atomic, > 0 means cancellation was requested
	options               *SubscriptionWorkerOptions
	chResults 			chan []*SubscriptionBatchItem
	tcpClient             atomic.Value // net.Conn
	parser                *json.Decoder
	disposed              int32 // atomic
	// this channel is closed when worker
	chDone chan struct{}


	afterAcknowledgment           []func(*subscriptionBatch)
	onSubscriptionConnectionRetry []func(error)

	redirectNode                     *ServerNode
	subscriptionLocalRequestExecutor *RequestExecutor

	lastConnectionFailure time.Time
	supportedFeatures     *supportedFeatures
	onClosed              func(*SubscriptionWorker)

	err atomic.Value // error
	mu  sync.Mutex
}

// Err returns a potential error, available after worker finished
func (w *SubscriptionWorker) Err() error {
	if v := w.err.Load(); v == nil {
		return nil
	} else {
		return v.(error)
	}
}

func (w *SubscriptionWorker) isCancellationRequested() bool {
	v := atomic.LoadInt32(&w.cancellationRequested)
	return v > 0
}

// Cancel requests the worker to finish. It doesn't happen immediately.
// To check if the worker finished, use HasFinished
// To wait
func (w *SubscriptionWorker) Cancel() {
	atomic.AddInt32(&w.cancellationRequested, 1)
	// we might be reading from a connection, so break that loop
	// by closing the connection
	w.closeTcpClient()
}

// IsDone returns true if the worker has finished
func (w *SubscriptionWorker) IsDone() bool {
	if w.chDone == nil {
		// not started yet
		return true
	}
	select {
	case <-w.chDone:
		return true
	default:
		return false
	}
}

// WaitUntilFinished waits until worker finishes for up to a timeout and
// returns an error.
// If timeout is 0, it waits indefinitely.
func (w *SubscriptionWorker) WaitUntilFinished(timeout time.Duration) error {
	if w.chDone == nil {
		// not started yet
		return newSubscriptionInvalidStateError("SubscriptionWorker has not yet been started with Run()")
	}

	if timeout == 0 {
		<-w.chDone
		return w.Err()
	}

	select {
	case <-w.chDone:
	// no-op, we're here if already finished (channel closed)
	case <-time.After(timeout):
		return  NewTimeoutError("timed out waiting for subscription worker to finish")
	}
	return w.Err()
}

func (w *SubscriptionWorker) getTcpClient() net.Conn {
	if conn := w.tcpClient.Load(); conn == nil {
		return nil
	} else {
		return conn.(net.Conn)
	}
}

func (w *SubscriptionWorker) isDisposed() bool {
	v := atomic.LoadInt32(&w.disposed)
	return v != 0
}

func (w *SubscriptionWorker) markDisposed() {
	atomic.StoreInt32(&w.disposed, 1)
}

// AddAfterAcknowledgmentListener adds callback function that will be called after
// listener has been acknowledged.
// Returns id that can be used in RemoveAfterAcknowledgmentListener
func (w *SubscriptionWorker) AddAfterAcknowledgmentListener(handler func(*subscriptionBatch)) int {
	w.afterAcknowledgment = append(w.afterAcknowledgment, handler)
	return len(w.afterAcknowledgment) - 1
}

// RemoveAfterAcknowledgmentListener removes a callback added with AddAfterAcknowledgmentListener
func (w *SubscriptionWorker) RemoveAfterAcknowledgmentListener(id int) {
	w.afterAcknowledgment[id] = nil
}

// AddOnSubscriptionConnectionRetry adds a callback function that will be called
// when subscription  connection is retried.
// Returns id that can be used in RemoveOnSubscriptionConnectionRetry
func (w *SubscriptionWorker) AddOnSubscriptionConnectionRetry(handler func(error)) int {
	w.onSubscriptionConnectionRetry = append(w.onSubscriptionConnectionRetry, handler)
	return len(w.onSubscriptionConnectionRetry) - 1
}

// RemoveOnSubscriptionConnectionRetry removes a callback added with AddOnSubscriptionConnectionRetry
func (w *SubscriptionWorker) RemoveOnSubscriptionConnectionRetry(id int) {
	w.onSubscriptionConnectionRetry[id] = nil
}

// NewSubscriptionWorker returns new SubscriptionWorker
func NewSubscriptionWorker(clazz reflect.Type, options *SubscriptionWorkerOptions, withRevisions bool, documentStore *DocumentStore, dbName string) (*SubscriptionWorker, error) {

	if options.SubscriptionName == "" {
		return nil, newIllegalArgumentError("SubscriptionConnectionOptions must specify the subscriptionName")
	}

	if dbName == "" {
		dbName = documentStore.GetDatabase()
	}

	res := &SubscriptionWorker{
		clazz:     clazz,
		options:   options,
		revisions: withRevisions,
		store:     documentStore,
		dbName:    dbName,
	}

	return res, nil
}

// Close closes a subscription
func (w *SubscriptionWorker) Close() error {
	return w.close(true)
}

func (w *SubscriptionWorker) close(waitForSubscriptionTask bool) error {
	if w.isDisposed() {
		return nil
	}
	defer func() {
		if w.onClosed != nil {
			w.onClosed(w)
		}
	}()
	w.markDisposed()
	w.Cancel()

	if waitForSubscriptionTask {
		_ = w.WaitUntilFinished(0)
	}

	if w.subscriptionLocalRequestExecutor != nil {
		w.subscriptionLocalRequestExecutor.Close()
	}
	return nil
}

func (w *SubscriptionWorker) Run() (chan []*SubscriptionBatchItem, error) {
	if w.chDone != nil {
		return nil, newIllegalStateError("The subscription is already running")
	}

	// unbuffered so that we can ack to the server that the user processed
	// a batch
	w.chResults = make(chan []*SubscriptionBatchItem)
	w.chDone = make(chan struct{})
	go func() {
		w.runSubscriptionAsync(w.chResults)
	}()
	return w.chResults, nil
}

func (w *SubscriptionWorker) getCurrentNodeTag() string {
	if w.redirectNode != nil {
		return w.redirectNode.ClusterTag
	}
	return ""
}

func (w *SubscriptionWorker) getSubscriptionName() string {
	if w.options != nil {
		return w.options.SubscriptionName
	}
	return ""
}

func (w *SubscriptionWorker) connectToServer() (net.Conn, error) {
	command := NewGetTcpInfoCommand("Subscription/"+w.dbName, w.dbName)
	requestExecutor := w.store.GetRequestExecutor(w.dbName)

	var err error
	if w.redirectNode != nil {
		err = requestExecutor.Execute(w.redirectNode, -1, command, false, nil)
		if err != nil {
			w.redirectNode = nil
			// if we failed to talk to a node, we'll forget about it and let the topology to
			// redirect us to the current node
			return nil, newRuntimeError(err.Error())
		}
	} else {
		if err = requestExecutor.ExecuteCommand(command, nil); err != nil {
			return nil, err
		}
	}

	uri := command.Result.URL
	var serverCert []byte
	if command.Result.Certificate != nil {
		serverCert = []byte(*command.Result.Certificate)
	}
	cert := w.store.Certificate
	tcpClient, err := tcpConnect(uri, serverCert, cert)
	if err != nil {
		return nil, err
	}
	w.tcpClient.Store(tcpClient)
	databaseName := w.dbName
	if databaseName == "" {
		databaseName = w.store.GetDatabase()
	}

	parameters := &tcpNegotiateParameters{}
	parameters.database = databaseName
	parameters.operation = operationSubscription
	parameters.version = subscriptionTCPVersion
	fn := func(s string) int {
		n, _ := w.readServerResponseAndGetVersion(s)
		return n
	}
	parameters.readResponseAndGetVersionCallback = fn
	parameters.destinationNodeTag = w.getCurrentNodeTag()
	parameters.destinationUrl = command.Result.URL

	w.supportedFeatures, err = negotiateProtocolVersion(tcpClient, parameters)
	if err != nil {
		return nil, err
	}

	if w.supportedFeatures.protocolVersion <= 0 {
		return nil, newIllegalStateError(w.options.SubscriptionName + " : TCP negotiation resulted with an invalid protocol version: " + strconv.Itoa(w.supportedFeatures.protocolVersion))
	}

	options, err := jsonMarshal(w.options)
	if err != nil {
		return nil, err
	}

	_, err = tcpClient.Write(options)
	if err != nil {
		return nil, err
	}
	if w.subscriptionLocalRequestExecutor != nil {
		w.subscriptionLocalRequestExecutor.Close()
	}
	conv := w.store.GetConventions()
	cert = requestExecutor.Certificate
	trustStore := requestExecutor.TrustStore
	uri = command.requestedNode.URL
	w.subscriptionLocalRequestExecutor = RequestExecutorCreateForSingleNodeWithoutConfigurationUpdates(uri, w.dbName, cert, trustStore, conv)
	return tcpClient, nil
}

func (w *SubscriptionWorker) ensureParser() {
	if w.parser == nil {
		w.parser = json.NewDecoder(w.getTcpClient())
	}
}

func (w *SubscriptionWorker) readServerResponseAndGetVersion(url string) (int, error) {
	//Reading reply from server
	w.ensureParser()
	var reply *tcpConnectionHeaderResponse
	err := w.parser.Decode(&reply)
	if err != nil {
		return 0, err
	}

	switch reply.Status {
	case tcpConnectionStatusOk:
		return reply.Version, nil
	case tcpConnectionStatusAuthorizationFailed:
		return 0, newAuthorizationError("Cannot access database " + w.dbName + " because " + reply.Message)
	case tcpConnectionStatusTcpVersionMismatch:
		if reply.Version != outOfRangeStatus {
			return reply.Version, nil
		}
		// Kindly request the server to drop the connection
		_ = w.sendDropMessage(reply)
		return 0, newIllegalStateError("Can't connect to database " + w.dbName + " because: " + reply.Message)
	}

	return 0, newIllegalStateError("Unknown status '%s'", reply.Status)
}

func (w *SubscriptionWorker) sendDropMessage(reply *tcpConnectionHeaderResponse) error {
	dropMsg := &tcpConnectionHeaderMessage{}
	dropMsg.Operation = operationDrop
	dropMsg.DatabaseName = w.dbName
	dropMsg.OperationVersion = subscriptionTCPVersion
	dropMsg.Info = "Couldn't agree on subscription tcp version ours: " + strconv.Itoa(subscriptionTCPVersion) + " theirs: " + strconv.Itoa(reply.Version)
	header, err := jsonMarshal(dropMsg)
	if err != nil {
		return err
	}
	tcpClient := w.getTcpClient()
	if _, err = tcpClient.Write(header); err != nil {
		return err
	}
	return nil
}

func (w *SubscriptionWorker) assertConnectionState(connectionStatus *SubscriptionConnectionServerMessage) error {
	//fmt.Printf("assertConnectionStatus: %v\n", connectionStatus)
	if connectionStatus.Type == SubscriptionServerMessageError {
		if strings.Contains(connectionStatus.Exception, "DatabaseDoesNotExistException") {
			return newDatabaseDoesNotExistError(w.dbName + " does not exists. " + connectionStatus.Message)
		}
	}

	if connectionStatus.Type != SubscriptionServerMessageConnectionStatus {
		return newIllegalStateError("Server returned illegal type message when expecting connection status, was:" + connectionStatus.Type)
	}

	switch connectionStatus.Status {
	case SubscriptionConnectionStatusAccepted:
	case SubscriptionConnectionStatusInUse:
		return newSubscriptionInUseError("Subscription with id " + w.options.SubscriptionName + " cannot be opened, because it's in use and the connection strategy is " + w.options.Strategy)
	case SubscriptionConnectionStatusClosed:
		return newSubscriptionClosedError("Subscription with id " + w.options.SubscriptionName + " was closed. " + connectionStatus.Exception)
	case SubscriptionConnectionStatusInvalid:
		return newSubscriptionInvalidStateError("Subscription with id " + w.options.SubscriptionName + " cannot be opened, because it is in invalid state. " + connectionStatus.Exception)
	case SubscriptionConnectionStatusNotFound:
		return newSubscriptionDoesNotExistError("Subscription with id " + w.options.SubscriptionName + " cannot be opened, because it does not exist. " + connectionStatus.Exception)
	case SubscriptionConnectionStatusRedirect:
		data := connectionStatus.Data
		appropriateNode, _ := jsonGetAsText(data, "RedirectedTag")
		err := newSubscriptionDoesNotBelongToNodeError("Subscription With id %s cannot be processed by current node, it will be redirected to %s", w.options.SubscriptionName, appropriateNode)
		err.appropriateNode = appropriateNode
		return err
	case SubscriptionConnectionStatusConcurrencyReconnect:
		return newSubscriptionChangeVectorUpdateConcurrencyError(connectionStatus.Message)
	default:
		return newIllegalStateError("Subscription " + w.options.SubscriptionName + " could not be opened, reason: " + connectionStatus.Status)
	}
	return nil
}

func (w *SubscriptionWorker) processSubscriptionInner(chResults chan []*SubscriptionBatchItem) error {
	if w.isCancellationRequested() {
		return throwCancellationRequested()
	}

	socket, err := w.connectToServer()
	if err != nil {
		return err
	}

	defer func() {
		_ = socket.Close()
	}()
	if w.isCancellationRequested() {
		return throwCancellationRequested()
	}

	tcpClientCopy := w.getTcpClient()

	connectionStatus, err := w.readNextObject()
	if err != nil {
		return err
	}

	if w.isCancellationRequested() {
		return nil
	}

	if (connectionStatus.Type != SubscriptionServerMessageConnectionStatus) || (connectionStatus.Status != SubscriptionConnectionStatusAccepted) {
		if err = w.assertConnectionState(connectionStatus); err != nil {
			return err
		}
	}

	w.lastConnectionFailure = time.Time{}
	if w.isCancellationRequested() {
		return nil
	}

	batch := newSubscriptionBatch(w.clazz, w.revisions, w.subscriptionLocalRequestExecutor, w.store, w.dbName, w.logger)

	for !w.isCancellationRequested() {
		incomingBatch, err := w.readSingleSubscriptionBatchFromServer(batch)
		if err != nil {
			return err
		}
		if w.isCancellationRequested() {
			return throwCancellationRequested()
		}
		lastReceivedChangeVector, err := batch.initialize(incomingBatch)
		if err != nil {
			return err
		}

		// send on a goroutine so that we can continue reading from
		// the server
		go func(items []*SubscriptionBatchItem, lastChangeVector string, tcpClient net.Conn) {
			// protect against panics because a client might never
			// read the results and when we close chResults
			// channel and we have pending sends, chResults <- batch
			// will panic
			defer func() {
				recover()
			}()

			chResults <- items

			if tcpClient != nil {
				// TODO: what to do in case of an error?
				_ = w.sendAck(lastChangeVector, tcpClient)
				// if !w.options.IgnoreSubscriberErrors {
			}
		}(batch.items, lastReceivedChangeVector, tcpClientCopy)
	}
	return nil
}

func (w *SubscriptionWorker) processSubscription(chResults chan []*SubscriptionBatchItem) error {
	err := w.processSubscriptionInner(chResults)
	if err == nil {
		return nil
	}
	if _, ok := err.(*OperationCancelledError); ok {
		if !w.isDisposed() {
			return err
		}
		// otherwise this is thrown when shutting down, it
		// isn't an error, so we don't need to treat
		// it as such
		return nil
	}

	return err
}

func (w *SubscriptionWorker) readSingleSubscriptionBatchFromServer(batch *subscriptionBatch) ([]*SubscriptionConnectionServerMessage, error) {
	var incomingBatch []*SubscriptionConnectionServerMessage
	endOfBatch := false

	for !endOfBatch && !w.isCancellationRequested() {
		receivedMessage, err := w.readNextObject()
		if err != nil {
			return nil, err
		}

		if receivedMessage == nil || w.isCancellationRequested() {
			break
		}

		switch receivedMessage.Type {
		case SubscriptionServerMessageData:
			incomingBatch = append(incomingBatch, receivedMessage)
		case SubscriptionServerMessageEndOfBatch:
			endOfBatch = true
		case SubscriptionServerMessageConfirm:
			for _, cb := range w.afterAcknowledgment {
				cb(batch)
			}
			incomingBatch = nil
			//batch.Items = nil
		case SubscriptionServerMessageConnectionStatus:
			if err = w.assertConnectionState(receivedMessage); err != nil {
				return nil, err
			}
		case SubscriptionServerMessageError:
			return nil, throwSubscriptionError(receivedMessage)
		default:
			return nil, throwInvalidServerResponse(receivedMessage)
		}
	}

	return incomingBatch, nil
}

func throwInvalidServerResponse(receivedMessage *SubscriptionConnectionServerMessage) error {
	return newIllegalArgumentError("Unrecognized message " + receivedMessage.Type + " type received from server")
}

func throwSubscriptionError(receivedMessage *SubscriptionConnectionServerMessage) error {
	exc := receivedMessage.Exception
	if exc == "" {
		exc = "None"
	}
	return newIllegalStateError("Connected terminated by server. Exception: " + exc)
}

func (w *SubscriptionWorker) readNextObject() (*SubscriptionConnectionServerMessage, error) {
	if w.isCancellationRequested() || w.isDisposed() {
		return nil, nil
	}

	var res *SubscriptionConnectionServerMessage
	err := w.parser.Decode(&res)
	return res, err
}

func (w *SubscriptionWorker) sendAck(lastReceivedChangeVector string, networkStream net.Conn) error {
	msg := &SubscriptionConnectionClientMessage{
		ChangeVector: &lastReceivedChangeVector,
		Type:         SubscriptionClientMessageAcknowledge,
	}
	ack, err := jsonMarshal(msg)
	if err != nil {
		return err
	}
	_, err = networkStream.Write(ack)
	return err
}

func (w *SubscriptionWorker) runSubscriptionAsync(chResults chan []*SubscriptionBatchItem) {
	//fmt.Printf("runSubscription(): %p started\n", w)
	defer func() {
		//fmt.Printf("runSubscriptionLoop() %p finished\n", w)

		// TODO: this reports data race on chResult but shouldn't
		close(chResults)
		close(w.chDone)
	}()

	for !w.isCancellationRequested() {
		w.closeTcpClient()
		if w.logger != nil {
			w.logger.Print("Subscription " + w.options.SubscriptionName + ". Connecting to server...")
		}

		//fmt.Printf("before w.processSubscription\n")
		ex := w.processSubscription(chResults)
		//fmt.Printf("after w.processSubscription, ex: %v\n", ex)
		if ex == nil {
			continue
		}

		if w.isCancellationRequested() {
			if !w.isDisposed() {
				w.err.Store(ex)
				return
			}
		}
		/* TODO:
		if (_logger.isInfoEnabled()) {
			_logger.info("Subscription " + _options.getSubscriptionName() + ". Pulling task threw the following exception", ex);
		}
		*/
		shouldReconnect, err := w.shouldTryToReconnect(ex)
		//fmt.Printf("shouldTryReconnect() returned err='%s'\n", err)
		if err != nil || !shouldReconnect {
			/*
				if (_logger.isErrorEnabled()) {
					_logger.error("Connection to subscription " + _options.getSubscriptionName() + " have been shut down because of an error", ex);
				}
			*/
			if err != nil {
				w.err.Store(err)
			}
			return
		}
		time.Sleep(time.Duration(w.options.TimeToWaitBeforeConnectionRetry))
		for _, cb := range w.onSubscriptionConnectionRetry {
			cb(ex)
		}
	}
}

func (w *SubscriptionWorker) assertLastConnectionFailure() error {
	if w.lastConnectionFailure.IsZero() {
		w.lastConnectionFailure = time.Now()
		return nil
	}

	dur := time.Since(w.lastConnectionFailure)

	if dur > time.Duration(w.options.MaxErroneousPeriod) {
		return newSubscriptionInvalidStateError("Subscription connection was in invalid state for more than %s and therefore will be terminated", time.Duration(w.options.MaxErroneousPeriod))
	}
	return nil
}

func (w *SubscriptionWorker) shouldTryToReconnect(ex error) (bool, error) {
	//fmt.Printf("shouldTryToReconnect, ex type: %T, ex v: %v, ex str: %s\n", ex, ex, ex)
	//ex = ExceptionsUtils.unwrapException(ex);
	if w.isCancellationRequested() {
		return false, nil
	}
	if se, ok := ex.(*SubscriptionDoesNotBelongToNodeError); ok {
		if err := w.assertLastConnectionFailure(); err != nil {
			return false, err
		}

		requestExecutor := w.store.GetRequestExecutor(w.dbName)
		if se.appropriateNode == "" {
			return true, nil
		}

		var nodeToRedirectTo *ServerNode
		for _, x := range requestExecutor.GetTopologyNodes() {
			if x.ClusterTag == se.appropriateNode {
				nodeToRedirectTo = x
				break
			}
		}

		if nodeToRedirectTo == nil {
			return false, newIllegalStateError("Could not redirect to " + se.appropriateNode + ", because it was not found in local topology, even after retrying")
		}

		w.redirectNode = nodeToRedirectTo
		return true, nil
	}

	if _, ok := ex.(*SubscriptionChangeVectorUpdateConcurrencyError); ok {
		return true, nil
	}

	_, ok1 := ex.(*SubscriptionInUseError)
	_, ok2 := ex.(*SubscriptionDoesNotExistError)
	_, ok3 := ex.(*SubscriptionClosedError)
	_, ok4 := ex.(*SubscriptionInvalidStateError)
	_, ok5 := ex.(*DatabaseDoesNotExistError)
	_, ok6 := ex.(*AuthorizationError)
	_, ok7 := ex.(*AllTopologyNodesDownError)
	_, ok8 := ex.(*SubscriberErrorError)
	if ok1 || ok2 || ok3 || ok4 || ok5 || ok6 || ok7 || ok8 {
		w.Cancel()
		return false, ex
	}

	if err := w.assertLastConnectionFailure(); err != nil {
		return false, err
	}
	return true, nil
}

func (w *SubscriptionWorker) closeTcpClient() {
	//w._parser = nil // Note: not necessary and causes data race

	tcpClient := w.getTcpClient()
	if tcpClient != nil {
		_ = tcpClient.Close()
	}
}
