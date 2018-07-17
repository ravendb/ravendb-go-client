package ravendb

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"strings"
)

// Note: the implementation details are different from Java
// We take advantage of a pipe: a read end is passed as io.Reader
// to the request. A write end is what we use to write to the request.

var _ RavenCommand = &BulkInsertCommand{}

type BulkInsertCommand struct {
	*RavenCommandBase

	_stream io.Reader

	_id int

	useCompression bool

	Result *http.Response
}

func NewBulkInsertCommand(id int, stream io.Reader, useCompression bool) *BulkInsertCommand {
	cmd := &BulkInsertCommand{
		RavenCommandBase: NewRavenCommandBase(),

		_stream:        stream,
		_id:            id,
		useCompression: useCompression,
	}
	return cmd
}

func (c *BulkInsertCommand) createRequest(node *ServerNode) (*http.Request, error) {
	url := node.getUrl() + "/databases/" + node.getDatabase() + "/bulk_insert?id=" + strconv.Itoa(c._id)
	// TODO: implement compression. It must be attached to the writer
	//message.setEntity(useCompression ? new GzipCompressingEntity(_stream) : _stream)
	return NewHttpPostReader(url, c._stream)
}

func (c *BulkInsertCommand) setResponse(response []byte, fromCache bool) error {
	return NewNotImplementedException("Not implemented")
}

// TODO: port this. Currenlty send is not over-rideable
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

type BulkInsertOperation struct {
	_generateEntityIdOnTheClient *GenerateEntityIdOnTheClient
	_requestExecutor             *RequestExecutor

	_bulkInsertExecuteTask *CompletableFuture

	_reader        *io.PipeReader
	_currentWriter *io.PipeWriter

	_first       bool
	_operationId int

	useCompression bool

	_concurrentCheck AtomicInteger

	_conventions *DocumentConventions
	err          error

	Command *BulkInsertCommand
}

func NewBulkInsertOperation(database string, store *IDocumentStore) *BulkInsertOperation {
	re := store.GetRequestExecutorWithDatabase(database)
	f := func(entity Object) string {
		return re.getConventions().generateDocumentId(database, entity)
	}

	reader, writer := io.Pipe()

	res := &BulkInsertOperation{
		_conventions:                 store.getConventions(),
		_requestExecutor:             re,
		_generateEntityIdOnTheClient: NewGenerateEntityIdOnTheClient(re.getConventions(), f),
		_reader:        reader,
		_currentWriter: writer,
		_operationId:   -1,
	}
	return res
}

func (o *BulkInsertOperation) isUseCompression() bool {
	return o.useCompression
}

func (o *BulkInsertOperation) setUseCompression(useCompression bool) {
	o.useCompression = useCompression
}

func (o *BulkInsertOperation) throwBulkInsertAborted(e error, flushEx error) error {
	err := error(o.getExceptionFromOperation())
	if err == nil {
		err = e
	}
	if err == nil {
		err = flushEx
	}
	return NewBulkInsertAbortedException("Failed to execute bulk insert, error: %s", err)
}

func (o *BulkInsertOperation) getExceptionFromOperation() *BulkInsertAbortedException {
	stateRequest := NewGetOperationStateCommand(o._requestExecutor.getConventions(), o._operationId)
	err := o._requestExecutor.executeCommand(stateRequest)
	if err != nil {
		return nil // TODO: return an error?
	}

	if result, ok := stateRequest.Result["Result"]; ok {
		if result, ok := result.(ObjectNode); ok {
			typ, _ := jsonGetAsString(result, "$type")
			if strings.HasPrefix(typ, "Raven.Client.Documents.Operations.OperationExceptionResult") {
				errStr, _ := jsonGetAsString(result, "Error")
				return NewBulkInsertAbortedException(errStr)
			}
		}
	}
	return nil
}

func (o *BulkInsertOperation) waitForId() error {
	if o._operationId != -1 {
		return nil
	}

	bulkInsertGetIdRequest := NewGetNextOperationIdCommand()
	o.err = o._requestExecutor.executeCommand(bulkInsertGetIdRequest)
	if o.err != nil {
		return o.err
	}
	o._operationId = bulkInsertGetIdRequest.Result
	return nil
}

func (o *BulkInsertOperation) storeWithID(entity Object, id string, metadata *IMetadataDictionary) error {
	if !o._concurrentCheck.compareAndSet(0, 1) {
		return NewIllegalStateException("Bulk Insert store methods cannot be executed concurrently.")
	}
	defer o._concurrentCheck.set(0)

	// early exit if we failed previously
	if o.err != nil {
		return o.err
	}

	err := BulkInsertOperation_verifyValidId(id)
	if err != nil {
		return err
	}
	o.err = o.waitForId()
	if o.err != nil {
		return o.err
	}
	o.err = o.ensureCommand()
	if o.err != nil {
		return o.err
	}

	if o._bulkInsertExecuteTask.isCompletedExceptionally() {
		_, err := o._bulkInsertExecuteTask.get()
		panicIf(err == nil, "err should not be nil")
		return o.throwBulkInsertAborted(err, nil)
	}

	if metadata == nil {
		metadata = &MetadataAsDictionary{}
	}

	if !metadata.containsKey(Constants_Documents_Metadata_COLLECTION) {
		collection := o._requestExecutor.getConventions().getCollectionName(entity)
		if collection != "" {
			metadata.put(Constants_Documents_Metadata_COLLECTION, collection)
		}
	}
	if !metadata.containsKey(Constants_Documents_Metadata_RAVEN_GO_TYPE) {
		goType := o._requestExecutor.getConventions().getGoTypeName(entity)
		if goType != "" {
			metadata.put(Constants_Documents_Metadata_RAVEN_GO_TYPE, goType)
		}
	}

	documentInfo := NewDocumentInfo()
	documentInfo.setMetadataInstance(metadata)
	jsNode := EntityToJson_convertEntityToJson(entity, documentInfo)

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

	d, err := json.Marshal(m)
	if err != nil {
		return err
	}
	b.Write(d)
	_, o.err = b.WriteTo(o._currentWriter)
	if o.err != nil {
		err := o.getExceptionFromOperation()
		if err != nil {
			o.err = err
			return o.err
		}
		// TODO:
		//o.err = throwOnUnavailableStream()
		return o.err
	}
	return o.err
}

func (o *BulkInsertOperation) ensureCommand() error {
	if o.Command != nil {
		return nil
	}
	bulkCommand := NewBulkInsertCommand(o._operationId, o._reader, o.useCompression)
	panicIf(o._bulkInsertExecuteTask != nil, "already started _bulkInsertExecuteTask")
	o._bulkInsertExecuteTask = NewCompletableFuture()
	go func() {
		err := o._requestExecutor.executeCommand(bulkCommand)
		if err != nil {
			o._bulkInsertExecuteTask.markAsDoneWithError(err)
		} else {
			o._bulkInsertExecuteTask.markAsDone(nil)
		}
	}()

	_, err := o._currentWriter.Write([]byte{'['})
	if err != nil {
		return NewRavenException("Unable to open bulk insert stream %s", err)
	}
	return nil
}

func (o *BulkInsertOperation) abort() error {
	if o._operationId == -1 {
		return nil // nothing was done, nothing to kill
	}

	err := o.waitForId()
	if err != nil {
		return err
	}

	command := NewKillOperationCommand(strconv.Itoa(o._operationId))
	err = o._requestExecutor.executeCommand(command)
	if err != nil {
		return NewBulkInsertAbortedException("Unable to kill ths bulk insert operation, because it was not found on the server.")
	}
	return nil
}

func (o *BulkInsertOperation) Close() error {
	if o._operationId == -1 {
		// closing without calling a single store.
		return nil
	}

	d := []byte{']'}
	_, err := o._currentWriter.Write(d)
	errClose := o._currentWriter.Close()
	o._operationId = -1
	if o._bulkInsertExecuteTask != nil {
		_, err2 := o._bulkInsertExecuteTask.get()
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

/*
 class BulkInsertOperation implements CleanCloseable {

      store(Object entity, string id)  {
        store(entity, id, null)
    }

     string store(Object entity) {
        return store(entity, (IMetadataDictionary) null)
    }

     string store(Object entity, IMetadataDictionary metadata) {
        string id
        if (metadata == null || !metadata.containsKey(Constants.Documents.Metadata.ID)) {
            id = getId(entity)
        } else {
            id = (string) metadata.get(Constants.Documents.Metadata.ID)
        }

        store(entity, id, metadata)

        return id
    }

      throwOnUnavailableStream(string id, Exception innerEx) {
        _streamExposerContent.errorOnProcessingRequest(new BulkInsertAbortedException("Write to stream failed at document with id " + id, innerEx))

        try {
            _bulkInsertExecuteTask.get()
        } catch (Exception e) {
            throw ExceptionsUtils.unwrapException(e)
        }
    }

     string getId(Object entity) {
        Reference<string> idRef = new Reference<>()
        if (_generateEntityIdOnTheClient.tryGetIdFromInstance(entity, idRef)) {
            return idRef.value
        }

        idRef.value = _generateEntityIdOnTheClient.generateDocumentKeyForStorage(entity)

        _generateEntityIdOnTheClient.trySetIdentity(entity, idRef.value) // set id property if it was null
        return idRef.value
    }
}
*/

func BulkInsertOperation_verifyValidId(id string) error {
	if StringUtils_isEmpty(id) {
		return NewIllegalStateException("Document id must have a non empty value")
	}

	if strings.HasSuffix(id, "|") {
		return NewUnsupportedOperationException("Document ids cannot end with '|', but was called with %s", id)
	}
	return nil
}
