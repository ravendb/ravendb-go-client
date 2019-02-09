package ravendb

import (
	"bytes"
	"io"
	"net/http"
	"strings"
)

// Note: the implementation details are different from Java
// We take advantage of a pipe: a read end is passed as io.Reader
// to the request. A write end is what we use to write to the request.

var _ RavenCommand = &BulkInsertCommand{}

// BulkInsertCommand describes build insert command
type BulkInsertCommand struct {
	RavenCommandBase

	stream io.Reader

	id int64

	useCompression bool

	Result *http.Response
}

// NewBulkInsertCommand returns new BulkInsertCommand
func NewBulkInsertCommand(id int64, stream io.Reader, useCompression bool) *BulkInsertCommand {
	cmd := &BulkInsertCommand{
		RavenCommandBase: NewRavenCommandBase(),

		stream:         stream,
		id:             id,
		useCompression: useCompression,
	}
	return cmd
}

func (c *BulkInsertCommand) createRequest(node *ServerNode) (*http.Request, error) {
	url := node.URL + "/databases/" + node.Database + "/bulk_insert?id=" + i64toa(c.id)
	// TODO: implement compression. It must be attached to the writer
	//message.setEntity(useCompression ? new GzipCompressingEntity(_stream) : _stream)
	return newHttpPostReader(url, c.stream)
}

func (c *BulkInsertCommand) setResponse(response []byte, fromCache bool) error {
	return newNotImplementedError("Not implemented")
}

func (c *BulkInsertCommand) send(client *http.Client, req *http.Request) (*http.Response, error) {
	base := c.getBase()
	rsp, err := base.send(client, req)
	if err != nil {
		// TODO: don't know how/if this translates to Go
		// c.stream.errorOnRequestStart(err)
		return nil, err
	}
	return rsp, nil
}

// BulkInsertOperation represents bulk insert operation
type BulkInsertOperation struct {
	generateEntityIDOnTheClient *generateEntityIDOnTheClient
	requestExecutor             *RequestExecutor

	bulkInsertExecuteTask *completableFuture

	reader        *io.PipeReader
	currentWriter *io.PipeWriter

	first       bool
	operationID int64

	useCompression bool

	concurrentCheck atomicInteger

	conventions *DocumentConventions
	err         error

	Command *BulkInsertCommand
}

// NewBulkInsertOperation returns new BulkInsertOperation
func NewBulkInsertOperation(database string, store *DocumentStore) *BulkInsertOperation {
	re := store.GetRequestExecutor(database)
	f := func(entity interface{}) (string, error) {
		return re.GetConventions().GenerateDocumentID(database, entity)
	}

	reader, writer := io.Pipe()

	res := &BulkInsertOperation{
		conventions:                 store.GetConventions(),
		requestExecutor:             re,
		generateEntityIDOnTheClient: newGenerateEntityIDOnTheClient(re.GetConventions(), f),
		reader:                      reader,
		currentWriter:               writer,
		operationID:                 -1,
		first:                       true,
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

func (o *BulkInsertOperation) getErrorFromOperation() error {
	stateRequest := NewGetOperationStateCommand(o.requestExecutor.GetConventions(), o.operationID)
	err := o.requestExecutor.ExecuteCommand(stateRequest, nil)
	if err != nil {
		return err
	}

	status, _ := jsonGetAsText(stateRequest.Result, "Status")
	if status != "Faulted" {
		return nil
	}

	if result, ok := stateRequest.Result["Result"]; ok {
		if result, ok := result.(map[string]interface{}); ok {
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
	if o.operationID != -1 {
		return nil
	}

	bulkInsertGetIDRequest := NewGetNextOperationIDCommand()
	o.err = o.requestExecutor.ExecuteCommand(bulkInsertGetIDRequest, nil)
	if o.err != nil {
		return o.err
	}
	o.operationID = bulkInsertGetIDRequest.Result
	return nil
}

// StoreWithID stores an entity with a given id
func (o *BulkInsertOperation) StoreWithID(entity interface{}, id string, metadata *MetadataAsDictionary) error {
	if !o.concurrentCheck.compareAndSet(0, 1) {
		return newIllegalStateError("Bulk Insert Store methods cannot be executed concurrently.")
	}
	defer o.concurrentCheck.set(0)

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

	if o.bulkInsertExecuteTask.IsCompletedExceptionally() {
		_, err = o.bulkInsertExecuteTask.Get()
		panicIf(err == nil, "err should not be nil")
		return o.throwBulkInsertAborted(err, nil)
	}

	if metadata == nil {
		metadata = &MetadataAsDictionary{}
	}

	if !metadata.ContainsKey(MetadataCollection) {
		collection := o.requestExecutor.GetConventions().GetCollectionName(entity)
		if collection != "" {
			metadata.Put(MetadataCollection, collection)
		}
	}
	if !metadata.ContainsKey(MetadataRavenGoType) {
		goType := o.requestExecutor.GetConventions().getGoTypeName(entity)
		if goType != "" {
			metadata.Put(MetadataRavenGoType, goType)
		}
	}

	documentInfo := &documentInfo{}
	documentInfo.metadataInstance = metadata
	jsNode := convertEntityToJSON(entity, documentInfo)

	var b bytes.Buffer
	if o.first {
		b.WriteByte('[')
		o.first = false
	} else {
		b.WriteByte(',')
	}
	m := map[string]interface{}{}
	m["Id"] = o.escapeID(id)
	m["Type"] = "PUT"
	m["Document"] = jsNode

	d, err := jsonMarshal(m)
	if err != nil {
		return err
	}
	b.Write(d)

	_, o.err = o.currentWriter.Write(b.Bytes())
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

func (o *BulkInsertOperation) escapeID(input string) string {
	if !strings.Contains(input, `"`) {
		return input
	}
	var res bytes.Buffer
	for i := 0; i < len(input); i++ {
		c := input[i]
		if c == '"' {
			if i == 0 || input[i-1] != '\\' {
				res.WriteByte('\\')
			}
		}
		res.WriteByte(c)
	}
	return res.String()
}

func (o *BulkInsertOperation) ensureCommand() error {
	if o.Command != nil {
		return nil
	}
	bulkCommand := NewBulkInsertCommand(o.operationID, o.reader, o.useCompression)
	panicIf(o.bulkInsertExecuteTask != nil, "already started _bulkInsertExecuteTask")
	o.bulkInsertExecuteTask = newCompletableFuture()
	go func() {
		err := o.requestExecutor.ExecuteCommand(bulkCommand, nil)
		if err != nil {
			o.bulkInsertExecuteTask.completeWithError(err)
		} else {
			o.bulkInsertExecuteTask.complete(nil)
		}
	}()

	o.Command = bulkCommand
	return nil
}

// Abort aborts insert operation
func (o *BulkInsertOperation) Abort() error {
	if o.operationID == -1 {
		return nil // nothing was done, nothing to kill
	}

	if err := o.WaitForID(); err != nil {
		return err
	}

	command, err := NewKillOperationCommand(i64toa(o.operationID))
	if err != nil {
		return err
	}
	err = o.requestExecutor.ExecuteCommand(command, nil)
	if err != nil {
		if _, ok := err.(*RavenError); ok {
			return newBulkInsertAbortedError("Unable to kill ths bulk insert operation, because it was not found on the server.")
		}
		return err
	}
	return nil
}

// Close closes operation
func (o *BulkInsertOperation) Close() error {
	if o.operationID == -1 {
		// closing without calling a single Store.
		return nil
	}

	d := []byte{']'}
	_, err := o.currentWriter.Write(d)
	errClose := o.currentWriter.Close()
	if o.bulkInsertExecuteTask != nil {
		_, err2 := o.bulkInsertExecuteTask.Get()
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
	var err error
	var id string
	if metadata == nil || !metadata.ContainsKey(MetadataID) {
		if id, err = o.GetID(entity); err != nil {
			return "", err
		}
	} else {
		idVal, ok := metadata.Get(MetadataID)
		panicIf(!ok, "didn't find %s key in meatadata", MetadataID)
		id = idVal.(string)
	}

	return id, o.StoreWithID(entity, id, metadata)
}

// GetID returns id for an entity
func (o *BulkInsertOperation) GetID(entity interface{}) (string, error) {
	var err error
	idRef, ok := o.generateEntityIDOnTheClient.tryGetIDFromInstance(entity)
	if ok {
		return idRef, nil
	}

	idRef, err = o.generateEntityIDOnTheClient.generateDocumentKeyForStorage(entity)
	if err != nil {
		return "", err
	}

	// set id property if it was null
	o.generateEntityIDOnTheClient.trySetIdentity(entity, idRef)
	return idRef, nil
}

func (o *BulkInsertOperation) throwOnUnavailableStream(id string, innerEx error) error {
	// TODO: don't know how/if this translates to Go
	//_streamExposerContent.errorOnProcessingRequest(new BulkInsertAbortedError("Write to stream failed at document with id " + id, innerEx))

	_, err := o.bulkInsertExecuteTask.Get()
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
