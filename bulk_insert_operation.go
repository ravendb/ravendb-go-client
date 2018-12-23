package ravendb

import (
	"bytes"
	"io"
	"net/http"
	"strconv"
	"strings"
)

// Note: the implementation details are different from Java
// We take advantage of a pipe: a read end is passed as io.Reader
// to the request. A write end is what we use to write to the request.

var _ RavenCommand = &BulkInsertCommand{}

// BulkInsertCommand describes build insert command
type BulkInsertCommand struct {
	RavenCommandBase

	_stream io.Reader

	_id int

	useCompression bool

	Result *http.Response
}

// NewBulkInsertCommand returns new BulkInsertCommand
func NewBulkInsertCommand(id int, stream io.Reader, useCompression bool) *BulkInsertCommand {
	cmd := &BulkInsertCommand{
		RavenCommandBase: NewRavenCommandBase(),

		_stream:        stream,
		_id:            id,
		useCompression: useCompression,
	}
	return cmd
}

// CreateRequest creates a request
func (c *BulkInsertCommand) CreateRequest(node *ServerNode) (*http.Request, error) {
	url := node.GetUrl() + "/databases/" + node.GetDatabase() + "/bulk_insert?id=" + strconv.Itoa(c._id)
	// TODO: implement compression. It must be attached to the writer
	//message.setEntity(useCompression ? new GzipCompressingEntity(_stream) : _stream)
	return NewHttpPostReader(url, c._stream)
}

// SetResponse sets response
func (c *BulkInsertCommand) SetResponse(response []byte, fromCache bool) error {
	return newNotImplementedError("Not implemented")
}

// TODO: port this. Currently send can't be over-written
/*
 CloseableHttpResponse send(CloseableHttpClient client, HttpRequestBase request) throws IOException {
	try {
		return super.send(client, request)
	} catch (Exception e) {
		_stream.errorOnRequestStart(e)
		throw e
	}
}
*/

// BulkInsertOperation represents bulk insert operation
type BulkInsertOperation struct {
	_generateEntityIDOnTheClient *generateEntityIDOnTheClient
	_requestExecutor             *RequestExecutor

	_bulkInsertExecuteTask *CompletableFuture

	_reader        *io.PipeReader
	_currentWriter *io.PipeWriter

	_first       bool
	_operationID int

	useCompression bool

	_concurrentCheck atomicInteger

	_conventions *DocumentConventions
	err          error

	Command *BulkInsertCommand
}

// NewBulkInsertOperation returns new BulkInsertOperation
func NewBulkInsertOperation(database string, store *IDocumentStore) *BulkInsertOperation {
	re := store.GetRequestExecutorWithDatabase(database)
	f := func(entity interface{}) string {
		return re.GetConventions().GenerateDocumentID(database, entity)
	}

	reader, writer := io.Pipe()

	res := &BulkInsertOperation{
		_conventions:                 store.GetConventions(),
		_requestExecutor:             re,
		_generateEntityIDOnTheClient: newgenerateEntityIDOnTheClient(re.GetConventions(), f),
		_reader:                      reader,
		_currentWriter:               writer,
		_operationID:                 -1,
		_first:                       true,
	}
	return res
}

func (o *BulkInsertOperation) throwBulkInsertAborted(e error, flushEx error) error {
	err := error(o.getErrorFromOperation())
	if err == nil {
		err = e
	}
	if err == nil {
		err = flushEx
	}
	return newBulkInsertAbortedError("Failed to execute bulk insert, error: %s", err)
}

func (o *BulkInsertOperation) getErrorFromOperation() *BulkInsertAbortedError {
	stateRequest := NewGetOperationStateCommand(o._requestExecutor.GetConventions(), o._operationID)
	err := o._requestExecutor.ExecuteCommand(stateRequest)
	if err != nil {
		return nil // TODO: return an error?
	}

	if result, ok := stateRequest.Result["Result"]; ok {
		if result, ok := result.(ObjectNode); ok {
			typ, _ := jsonGetAsString(result, "$type")
			if strings.HasPrefix(typ, "Raven.Client.Documents.Operations.OperationExceptionResult") {
				errStr, _ := jsonGetAsString(result, "Error")
				return newBulkInsertAbortedError(errStr)
			}
		}
	}
	return nil
}

// WaitForID waits for operation id to finish
func (o *BulkInsertOperation) WaitForID() error {
	if o._operationID != -1 {
		return nil
	}

	bulkInsertGetIDRequest := NewGetNextOperationIDCommand()
	o.err = o._requestExecutor.ExecuteCommand(bulkInsertGetIDRequest)
	if o.err != nil {
		return o.err
	}
	o._operationID = bulkInsertGetIDRequest.Result
	return nil
}

// StoreWithID stores an entity with a given id
func (o *BulkInsertOperation) StoreWithID(entity interface{}, id string, metadata *MetadataAsDictionary) error {
	if !o._concurrentCheck.compareAndSet(0, 1) {
		return newIllegalStateError("Bulk Insert Store methods cannot be executed concurrently.")
	}
	defer o._concurrentCheck.set(0)

	// early exit if we failed previously
	if o.err != nil {
		return o.err
	}

	err := bulkInsertOperationVerifyValidID(id)
	if err != nil {
		return err
	}
	o.err = o.WaitForID()
	if o.err != nil {
		return o.err
	}
	o.err = o.ensureCommand()
	if o.err != nil {
		return o.err
	}

	if o._bulkInsertExecuteTask.IsCompletedExceptionally() {
		_, err = o._bulkInsertExecuteTask.Get()
		panicIf(err == nil, "err should not be nil")
		return o.throwBulkInsertAborted(err, nil)
	}

	if metadata == nil {
		metadata = &MetadataAsDictionary{}
	}

	if !metadata.ContainsKey(MetadataCollection) {
		collection := o._requestExecutor.GetConventions().GetCollectionName(entity)
		if collection != "" {
			metadata.Put(MetadataCollection, collection)
		}
	}
	if !metadata.ContainsKey(MetadataRavenGoType) {
		goType := o._requestExecutor.GetConventions().GetGoTypeName(entity)
		if goType != "" {
			metadata.Put(MetadataRavenGoType, goType)
		}
	}

	documentInfo := &documentInfo{}
	documentInfo.metadataInstance = metadata
	jsNode := convertEntityToJSON(entity, documentInfo)

	var b bytes.Buffer
	if o._first {
		b.WriteByte('[')
		o._first = false
	} else {
		b.WriteByte(',')
	}
	m := map[string]interface{}{}
	m["Id"] = id
	m["Type"] = "PUT"
	m["Document"] = jsNode

	d, err := jsonMarshal(m)
	if err != nil {
		return err
	}
	b.Write(d)

	_, o.err = o._currentWriter.Write(b.Bytes())
	if o.err != nil {
		err = o.getErrorFromOperation()
		if err != nil {
			o.err = err
			return o.err
		}
		// TODO:
		//o.err = o.throwOnUnavailableStream()
		return o.err
	}
	return o.err
}

func (o *BulkInsertOperation) ensureCommand() error {
	if o.Command != nil {
		return nil
	}
	bulkCommand := NewBulkInsertCommand(o._operationID, o._reader, o.useCompression)
	panicIf(o._bulkInsertExecuteTask != nil, "already started _bulkInsertExecuteTask")
	o._bulkInsertExecuteTask = NewCompletableFuture()
	go func() {
		err := o._requestExecutor.ExecuteCommand(bulkCommand)
		if err != nil {
			o._bulkInsertExecuteTask.CompleteExceptionally(err)
		} else {
			o._bulkInsertExecuteTask.Complete(nil)
		}
	}()

	o.Command = bulkCommand
	return nil
}

// Abort aborts insert operation
func (o *BulkInsertOperation) Abort() error {
	if o._operationID == -1 {
		return nil // nothing was done, nothing to kill
	}

	err := o.WaitForID()
	if err != nil {
		return err
	}

	command := NewKillOperationCommand(strconv.Itoa(o._operationID))
	err = o._requestExecutor.ExecuteCommand(command)
	//o._currentWriter.Close()
	if err != nil {
		return newBulkInsertAbortedError("%s", "Unable to kill ths bulk insert operation, because it was not found on the server.")
	}
	o._currentWriter.CloseWithError(newBulkInsertAbortedError("killed operation"))
	return nil
}

// Close closes operation
func (o *BulkInsertOperation) Close() error {
	if o._operationID == -1 {
		// closing without calling a single Store.
		return nil
	}

	d := []byte{']'}
	_, err := o._currentWriter.Write(d)
	errClose := o._currentWriter.Close()
	if o._bulkInsertExecuteTask != nil {
		_, err2 := o._bulkInsertExecuteTask.Get()
		if err2 != nil && err == nil {
			err = o.throwBulkInsertAborted(err, errClose)
		}
	}

	if err != nil {
		o.err = err
		return err
	}
	return nil
}

// Store stores entity. metadata can be nil
func (o *BulkInsertOperation) Store(entity interface{}, metadata *MetadataAsDictionary) (string, error) {
	var id string
	if metadata == nil || !metadata.ContainsKey(MetadataID) {
		id = o.GetID(entity)
	} else {
		idVal, ok := metadata.Get(MetadataID)
		panicIf(!ok, "didn't find %s key in meatadata", MetadataID)
		id = idVal.(string)
	}

	return id, o.StoreWithID(entity, id, metadata)
}

// GetID returns id for an entity
func (o *BulkInsertOperation) GetID(entity interface{}) string {
	idRef, ok := o._generateEntityIDOnTheClient.tryGetIDFromInstance(entity)
	if ok {
		return idRef
	}

	idRef = o._generateEntityIDOnTheClient.generateDocumentKeyForStorage(entity)

	// set id property if it was null
	o._generateEntityIDOnTheClient.trySetIdentity(entity, idRef)
	return idRef
}

func (o *BulkInsertOperation) throwOnUnavailableStream(id string, innerEx error) error {
	// TODO: not sure how this translates
	//_streamExposerContent.errorOnProcessingRequest(new BulkInsertAbortedError("Write to stream failed at document with id " + id, innerEx))

	_, err := o._bulkInsertExecuteTask.Get()
	if err != nil {
		return unwrapError(err)
	}
	return nil
}

func bulkInsertOperationVerifyValidID(id string) error {
	if stringIsEmpty(id) {
		return newIllegalStateError("Document id must have a non empty value")
	}

	if strings.HasSuffix(id, "|") {
		return newUnsupportedOperationError("Document ids cannot end with '|', but was called with %s", id)
	}
	return nil
}
