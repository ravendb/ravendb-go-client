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
	_clazz            reflect.Type
	_revisions        bool
	_logger           *log.Logger
	_store            *DocumentStore
	_dbName           string
	_processingCts    *cancellationTokenSource
	_options          *SubscriptionWorkerOptions
	_subscriber       func(*SubscriptionBatch) error
	_tcpClient        net.Conn
	_parser           *json.Decoder
	_disposed         int32 // atomic
	_subscriptionTask *completableFuture

	afterAcknowledgment           []func(*SubscriptionBatch)
	onSubscriptionConnectionRetry []func(error)

	_redirectNode                     *ServerNode
	_subscriptionLocalRequestExecutor *RequestExecutor

	lastConnectionFailure time.Time
	onClosed              func(*SubscriptionWorker)

	mu sync.Mutex
}

func (w *SubscriptionWorker) getTcpClient() net.Conn {
	w.mu.Lock()
	res := w._tcpClient
	w.mu.Unlock()
	return res
}

func (w *SubscriptionWorker) setTcpClient(c net.Conn) {
	w.mu.Lock()
	w._tcpClient = c
	w.mu.Unlock()
}

func (w *SubscriptionWorker) isDisposed() bool {
	v := atomic.LoadInt32(&w._disposed)
	return v != 0
}

func (w *SubscriptionWorker) markDisposed() {
	atomic.StoreInt32(&w._disposed, 1)
}

// AddAfterAcknowledgmentListener adds callback function that will be called after
// listener has been acknowledged.
// Returns id that can be used in RemoveAfterAcknowledgmentListener
func (w *SubscriptionWorker) AddAfterAcknowledgmentListener(handler func(*SubscriptionBatch)) int {
	w.afterAcknowledgment = append(w.afterAcknowledgment, handler)
	return len(w.afterAcknowledgment) - 1
}

// RemoveAfterAcknowledgmentListener removes a callback added with AddAfterAcknowledgmentListener
func (w *SubscriptionWorker) RemoveAfterAcknowledgmentListener(id int) {
	w.afterAcknowledgment[id] = nil
}

// AddOnSubscriptionConnectionRetry adds a callback funtion that will be called when subscription
// conection is retried.
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
		_clazz:         clazz,
		_options:       options,
		_revisions:     withRevisions,
		_store:         documentStore,
		_dbName:        dbName,
		_processingCts: &cancellationTokenSource{},
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
	w._processingCts.cancel()
	w.closeTcpClient() // we disconnect immediately

	if w._subscriptionTask != nil && waitForSubscriptionTask {
		// just need to wait for it to end
		w._subscriptionTask.Get()
	}
	return nil
}

func (w *SubscriptionWorker) Run(processDocuments func(*SubscriptionBatch) error) (*completableFuture, error) {
	if w._subscriptionTask != nil {
		return nil, newIllegalStateError("The subscription is already running")
	}

	w._subscriber = processDocuments

	w._subscriptionTask = w.runSubscriptionAsync()
	return w._subscriptionTask, nil
}

func (w *SubscriptionWorker) getCurrentNodeTag() string {
	if w._redirectNode != nil {
		return w._redirectNode.ClusterTag
	}
	return ""
}

func (w *SubscriptionWorker) getSubscriptionName() string {
	if w._options != nil {
		return w._options.SubscriptionName
	}
	return ""
}

func (w *SubscriptionWorker) connectToServer() (net.Conn, error) {
	command := NewGetTcpInfoCommand("Subscription/"+w._dbName, w._dbName)
	requestExecutor := w._store.GetRequestExecutor(w._dbName)

	var err error
	if w._redirectNode != nil {
		err = requestExecutor.Execute(w._redirectNode, -1, command, false, nil)
		if err != nil {
			w._redirectNode = nil
			// if we failed to talk to a node, we'll forget about it and let the topology to
			// redirect us to the current node
			return nil, newRuntimeError(err.Error())
		}
	} else {
		if err = requestExecutor.ExecuteCommand(command); err != nil {
			return nil, err
		}
	}

	uri := command.Result.URL
	var serverCert []byte
	if command.Result.Certificate != nil {
		serverCert = []byte(*command.Result.Certificate)
	}
	cert := w._store.GetCertificate()
	tcpClient, err := tcpConnect(uri, serverCert, cert)
	if err != nil {
		return nil, err
	}
	w.setTcpClient(tcpClient)
	databaseName := w._dbName
	if databaseName == "" {
		databaseName = w._store.GetDatabase()
	}

	tcpHeaderMsg := &TcpConnectionHeaderMessage{
		Operation:        OperationSubscription,
		DatabaseName:     databaseName,
		OperationVersion: SUBSCRIPTION_TCP_VERSION,
	}

	header, err := jsonMarshal(tcpHeaderMsg)
	if err != nil {
		return nil, err
	}

	options, err := jsonMarshal(w._options)
	if err != nil {
		return nil, err
	}

	_, err = w.getTcpClient().Write(header)
	if err != nil {
		return nil, err
	}
	w._parser = json.NewDecoder(w.getTcpClient())

	var reply *TcpConnectionHeaderResponse
	//Reading reply from server
	err = w._parser.Decode(&reply)

	switch reply.Status {
	case TcpConnectionStatusOk:
		// no-op
	case TcpConnectionStatusAuthorizationFailed:
		return nil, newAuthorizationError("Cannot access database " + databaseName + " because " + reply.Message)
	case TcpConnectionStatusTcpVersionMismatch:
		//Kindly request the server to drop the connection
		dropMsg := &TcpConnectionHeaderMessage{
			Operation:        OperationDrop,
			DatabaseName:     databaseName,
			OperationVersion: SUBSCRIPTION_TCP_VERSION,
			Info:             "Couldn't agree on subscription tcp version ours: " + strconv.Itoa(SUBSCRIPTION_TCP_VERSION) + " theirs: " + strconv.Itoa(reply.Version),
		}
		header, _ = jsonMarshal(dropMsg)
		tcpClient.Write(header)
		return nil, newIllegalStateError("Can't connect to database " + databaseName + " because: " + reply.Message)
	}

	_, err = tcpClient.Write(options)
	if err != nil {
		return nil, err
	}
	if w._subscriptionLocalRequestExecutor != nil {
		w._subscriptionLocalRequestExecutor.Close()
	}
	conv := w._store.GetConventions()
	cert = requestExecutor.GetCertificate()
	uri = command.requestedNode.URL
	w._subscriptionLocalRequestExecutor = RequestExecutorCreateForSingleNodeWithoutConfigurationUpdates(uri, w._dbName, cert, conv)
	return tcpClient, nil
}

func (w *SubscriptionWorker) assertConnectionState(connectionStatus *SubscriptionConnectionServerMessage) error {
	if connectionStatus.Type == SubscriptionServerMessageError {
		if strings.Contains(connectionStatus.Exception, "DatabaseDoesNotExistException") {
			return newDatabaseDoesNotExistError(w._dbName + " does not exists. " + connectionStatus.Message)
		}
	}

	if connectionStatus.Type != SubscriptionServerMessageConnectionStatus {
		return newIllegalStateError("Server returned illegal type message when expecting connection status, was:" + connectionStatus.Type)
	}

	switch connectionStatus.Status {
	case SubscriptionConnectionStatusAccepted:
	case SubscriptionConnectionStatusInUse:
		return newSubscriptionInUseError("Subscription with id " + w._options.SubscriptionName + " cannot be opened, because it's in use and the connection strategy is " + w._options.Strategy)
	case SubscriptionConnectionStatusClosed:
		return newSubscriptionClosedError("Subscription with id " + w._options.SubscriptionName + " was closed. " + connectionStatus.Exception)
	case SubscriptionConnectionStatusInvalid:
		return newSubscriptionInvalidStateError("Subscription with id " + w._options.SubscriptionName + " cannot be opened, because it is in invalid state. " + connectionStatus.Exception)
	case SubscriptionConnectionStatusNotFound:
		return newSubscriptionDoesNotExistError("Subscription with id " + w._options.SubscriptionName + " cannot be opened, because it does not exist. " + connectionStatus.Exception)
	case SubscriptionConnectionStatusRedirect:
		// data := connectionStatus.Data
		/// TODO
		//appropriateNode := data[nameof(SubscriptionConnectionServerMessage.SubscriptionRedirectData.RedirectedTag)]?.ToString();
		appropriateNode := ""
		err := newSubscriptionDoesNotBelongToNodeError("Subscription With Id '%s' cannot be processed by current node, it will be redirected to %s", w._options.SubscriptionName, appropriateNode)
		err.appropriateNode = appropriateNode
		return err
	case SubscriptionConnectionStatusConcurrencyReconnect:
		return newSubscriptionChangeVectorUpdateConcurrencyError(connectionStatus.Message)
	default:
		return newIllegalStateError("Subscription " + w._options.SubscriptionName + " could not be opened, reason: " + connectionStatus.Status)
	}
	return nil
}

func (w *SubscriptionWorker) processSubscriptionInner() error {
	if err := w._processingCts.getToken().throwIfCancellationRequested(); err != nil {
		return err
	}

	socket, err := w.connectToServer()
	if err != nil {
		return err
	}

	defer socket.Close()
	if err := w._processingCts.getToken().throwIfCancellationRequested(); err != nil {
		return err
	}

	tcpClientCopy := w.getTcpClient()

	connectionStatus, err := w.readNextObject(tcpClientCopy)
	if err != nil {
		return err
	}

	if w._processingCts.getToken().isCancellationRequested() {
		return nil
	}

	if (connectionStatus.Type != SubscriptionServerMessageConnectionStatus) || (connectionStatus.Status != SubscriptionConnectionStatusAccepted) {
		if err = w.assertConnectionState(connectionStatus); err != nil {
			return err
		}
	}

	w.lastConnectionFailure = time.Time{}
	if w._processingCts.getToken().isCancellationRequested() {
		return nil
	}

	notifiedSubscriber := newCompletableFutureAlreadyCompleted(nil)
	batch := NewSubscriptionBatch(w._clazz, w._revisions, w._subscriptionLocalRequestExecutor, w._store, w._dbName, w._logger)

	for !w._processingCts.getToken().isCancellationRequested() {
		// start the read from the server

		readFromServer := newCompletableFuture()
		go func() {
			res, err := w.readSingleSubscriptionBatchFromServer(tcpClientCopy, batch)
			// TODO: wrap IOException errors in RuntimError
			if err != nil {
				readFromServer.completeWithError(err)
			} else {
				readFromServer.complete(res)
			}
		}()

		_, err := notifiedSubscriber.Get()
		if err != nil {
			// if the subscriber errored, we shut down
			w.closeTcpClient()
			return err
		}

		incomingBatchI, err := readFromServer.Get()
		if err != nil {
			return err
		}
		incomingBatch := incomingBatchI.([]*SubscriptionConnectionServerMessage)
		if err = w._processingCts.getToken().throwIfCancellationRequested(); err != nil {
			return err
		}
		lastReceivedChangeVector, err := batch.initialize(incomingBatch)
		if err != nil {
			return err
		}

		notifiedSubscriber = newCompletableFuture()
		go func() {
			err := w._subscriber(batch)
			if err != nil {
				if !w._options.IgnoreSubscriberErrors {
					/*TODO:
					if (_logger.isDebugEnabled()) {
						_logger.debug("Subscription " + _options.getSubscriptionName() + ". Subscriber threw an exception on document batch", ex);
					}*/
					// TODO: wrap original error
					err = newSubscriberErrorError("Subscriber threw an exception in subscription " + w._options.SubscriptionName)
					notifiedSubscriber.completeWithError(err)
					return
				}
			}
			if tcpClientCopy != nil {
				err = w.sendAck(lastReceivedChangeVector, tcpClientCopy)
				if err != nil {
					// TODO: wrap in RuntimeError
					notifiedSubscriber.completeWithError(err)
					return
				}
			}
			notifiedSubscriber.complete(nil)
		}()
	}
	return nil
}

func (w *SubscriptionWorker) processSubscription() error {
	err := w.processSubscriptionInner()
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

func (w *SubscriptionWorker) readSingleSubscriptionBatchFromServer(socket net.Conn, batch *SubscriptionBatch) ([]*SubscriptionConnectionServerMessage, error) {
	var incomingBatch []*SubscriptionConnectionServerMessage
	endOfBatch := false

	for !endOfBatch && !w._processingCts.getToken().isCancellationRequested() {
		receivedMessage, err := w.readNextObject(socket)
		if err != nil {
			return nil, err
		}

		if receivedMessage == nil || w._processingCts.getToken().isCancellationRequested() {
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
			batch.Items = nil
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

// TODO: no need to pass socket
func (w *SubscriptionWorker) readNextObject(socket net.Conn) (*SubscriptionConnectionServerMessage, error) {
	if w._processingCts.getToken().isCancellationRequested() {
		return nil, nil
	}

	if w.isDisposed() { //if we are disposed, nothing to do...
		return nil, nil
	}

	var res *SubscriptionConnectionServerMessage
	err := w._parser.Decode(&res)
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

func (w *SubscriptionWorker) runSubscriptionAsync() *completableFuture {
	future := newCompletableFuture()
	go func() {
		for !w._processingCts.getToken().isCancellationRequested() {
			w.closeTcpClient()
			if w._logger != nil {
				w._logger.Print("Subscription " + w._options.SubscriptionName + ". Connecting to server...")
			}

			ex := w.processSubscription()
			if ex == nil {
				continue
			}

			if w._processingCts.getToken().isCancellationRequested() {
				if !w.isDisposed() {
					future.completeWithError(ex)
					return
				}
			}
			/* TODO:
			if (_logger.isInfoEnabled()) {
				_logger.info("Subscription " + _options.getSubscriptionName() + ". Pulling task threw the following exception", ex);
			}
			*/
			shouldReconnect, err := w.shouldTryToReconnect(ex)
			if err != nil || !shouldReconnect {
				/*
					if (_logger.isErrorEnabled()) {
						_logger.error("Connection to subscription " + _options.getSubscriptionName() + " have been shut down because of an error", ex);
					}
				*/
				future.completeWithError(ex)
				return
			}
			time.Sleep(time.Duration(w._options.TimeToWaitBeforeConnectionRetry))
			for _, cb := range w.onSubscriptionConnectionRetry {
				cb(ex)
			}
		}
		future.complete(nil)
	}()
	return future
}

func (w *SubscriptionWorker) assertLastConnectionFailure() error {
	if w.lastConnectionFailure.IsZero() {
		w.lastConnectionFailure = time.Now()
		return nil
	}

	dur := time.Now().Sub(w.lastConnectionFailure)

	if dur > time.Duration(w._options.MaxErroneousPeriod) {
		return newSubscriptionInvalidStateError("Subscription connection was in invalid state for more than %s and therefore will be terminated", time.Duration(w._options.MaxErroneousPeriod))
	}
	return nil
}

func (w *SubscriptionWorker) shouldTryToReconnect(ex error) (bool, error) {
	//ex = ExceptionsUtils.unwrapException(ex);
	if se, ok := ex.(*SubscriptionDoesNotBelongToNodeError); ok {
		if err := w.assertLastConnectionFailure(); err != nil {
			return false, err
		}

		requestExecutor := w._store.GetRequestExecutor(w._dbName)
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

		w._redirectNode = nodeToRedirectTo
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
		w._processingCts.cancel()
		return false, nil
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
		tcpClient.Close()
		w.setTcpClient(nil)
	}
}
