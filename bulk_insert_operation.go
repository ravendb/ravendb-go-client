package ravendb

import (
	"bytes"
	"io"
	"net/http"
	"strconv"
)

type StreamExposerContent struct {
}

/*
     static class StreamExposerContent extends AbstractHttpEntity {

         CompletableFuture<OutputStream> outputStream
          CompletableFuture<Void> _done

         StreamExposerContent() {
            setContentType(ContentType.APPLICATION_JSON.tostring())
            outputStream = new CompletableFuture<>()
            _done = new CompletableFuture()
        }

        @Override
         InputStream getContent() throws IOException, UnsupportedOperationException {
            throw new UnsupportedEncodingException()
        }

        @Override
         bool isStreaming() {
            return false
        }


        @Override
         bool isChunked() {
            return true
        }

        @Override
         bool isRepeatable() {
            return false
        }

        @Override
         long getContentLength() {
            return -1
        }

        @Override
          writeTo(OutputStream outputStream) {
            o.outputStream.complete(outputStream)
            try {
                _done.get()
            } catch (Exception e) {
                throw ExceptionsUtils.unwrapException(e)
            }
        }

          done() {
            _done.complete(null)
        }

          errorOnProcessingRequest(Exception exception) {
            _done.completeExceptionally(exception)
        }

          errorOnRequestStart(Exception exception) {
            outputStream.completeExceptionally(exception)
        }
	}
}
*/

var _ RavenCommand = &BulkInsertCommand{}

type BulkInsertCommand struct {
	*RavenCommandBase

	_stream *StreamExposerContent

	_id int

	useCompression bool

	Result *http.Response
}

func NewUpdateBulkInsertCommand(id int, stream *StreamExposerContent, useCompression bool) *BulkInsertCommand {
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
	// TODO: implement compression
	//message.setEntity(useCompression ? new GzipCompressingEntity(_stream) : _stream)
	return NewHttpPost(url, nil)
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
	_bulkInsertExecuteTask       *CompletableFuture // <>
	// objectMapper ObjectMapper

	_stream               io.Writer // OutputStream
	_streamExposerContent *StreamExposerContent

	_first       bool
	_operationId int

	useCompression bool

	_concurrentCheck AtomicInteger

	_conventions *DocumentConventions

	_requestBodyStream       io.Writer          // OutputStream
	_currentWriterBacking    *bytes.Buffer      // ByteArrayOutputStream
	_currentWriter           io.Writer          // Writer
	_backgroundWriterBacking *bytes.Buffer      // ByteArrayOutputStream
	_backgroundWriter        io.Writer          // Writer
	_asyncWrite              *CompletableFuture // void
	_maxSizeInBuffer         int
}

func NewBulkInsertOperation(database string, store *IDocumentStore) *BulkInsertOperation {
	re := store.GetRequestExecutorWithDatabase(database)
	f := func(entity Object) string {
		return re.getConventions().generateDocumentId(database, entity)
	}

	res := &BulkInsertOperation{
		_conventions:     store.getConventions(),
		_requestExecutor: re,
		//objectMapper: store.getConventions().getEntityMapper(),

		_currentWriterBacking: bytes.NewBuffer(nil), // new ByteArrayOutputStream()
		_currentWriter:        bytes.NewBuffer(nil), // = new OutputStreamWriter(_currentWriterBacking)
		//_backgroundWriterBacking: bytes.NewBUffer(), // = new ByteArrayOutputStream()
		//_backgroundWriter = new OutputStreamWriter(_backgroundWriterBacking)
		//_streamExposerContent = new StreamExposerContent()
		_maxSizeInBuffer:             1024 * 1024,
		_asyncWrite:                  NewCompletableFutureAlreadyCompleted(nil),
		_generateEntityIdOnTheClient: NewGenerateEntityIdOnTheClient(re.getConventions(), f),
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
	panicIf(true, "NYI")
	/*
		GetOperationStateOperation.GetOperationStateCommand stateRequest = new GetOperationStateOperation.GetOperationStateCommand(_requestExecutor.getConventions(), _operationId)
		_requestExecutor.execute(stateRequest)

		JsonNode result = stateRequest.getResult().get("Result")

		if (result.get("$type").asText().startsWith("Raven.Client.Documents.Operations.OperationExceptionResult")) {
			return new BulkInsertAbortedException(result.get("Error").asText())
		}

	*/
	return nil
}

func (o *BulkInsertOperation) waitForId() {
	if o._operationId != -1 {
		return
	}

	bulkInsertGetIdRequest := NewGetNextOperationIdCommand()
	o._requestExecutor.executeCommand(bulkInsertGetIdRequest)
	o._operationId = bulkInsertGetIdRequest.Result
}

/*
 class BulkInsertOperation implements CleanCloseable {

      store(Object entity, string id)  {
        store(entity, id, null)
    }

      store(Object entity, string id, IMetadataDictionary metadata) {
        if (!_concurrentCheck.compareAndSet(0, 1)) {
            throw new IllegalStateException("Bulk Insert store methods cannot be executed concurrently.")
        }

        try {
            verifyValidId(id)

            if (_stream == null) {
                waitForId()
                ensureStream()
            }

            if (_bulkInsertExecuteTask.isCompletedExceptionally()) {
                try {
                    _bulkInsertExecuteTask.get()
                } catch (Exception e) {
                    throwBulkInsertAborted(e, null)
                }
            }

            if (metadata == null) {
                metadata = new MetadataAsDictionary()
            }

            if (!metadata.containsKey(Constants.Documents.Metadata.COLLECTION)) {
                string collection = _requestExecutor.getConventions().getCollectionName(entity)
                if (collection != null) {
                    metadata.put(Constants.Documents.Metadata.COLLECTION, collection)
                }
            }

            if (!metadata.containsKey(Constants.Documents.Metadata.RAVEN_JAVA_TYPE)) {
                string javaType = _requestExecutor.getConventions().getJavaClassName(entity.getClass())
                if (javaType != null) {
                    metadata.put(Constants.Documents.Metadata.RAVEN_JAVA_TYPE, javaType)
                }
            }

            try {
                if (!_first) {
                    _currentWriter.write(",")
                }

                _first = false

                _currentWriter.write("{'Id':'")
                _currentWriter.write(id)
                _currentWriter.write("','Type':'PUT','Document':")

                DocumentInfo documentInfo = new DocumentInfo()
                documentInfo.setMetadataInstance(metadata)
                ObjectNode json = EntityToJson.convertEntityToJson(entity, _conventions, documentInfo)

                _currentWriter.flush()

                try (JsonGenerator generator =
                        objectMapper.getFactory().createGenerator(_currentWriter)) {
                    generator.configure(JsonGenerator.Feature.AUTO_CLOSE_TARGET, false)

                    generator.writeTree(json)
                }

                _currentWriter.write("}")
                _currentWriter.flush()

                if (_currentWriterBacking.size() > _maxSizeInBuffer || _asyncWrite.isDone()) {

                    _asyncWrite.get()

                    Writer tmp = _currentWriter
                    _currentWriter = _backgroundWriter
                    _backgroundWriter = tmp

                    ByteArrayOutputStream tmpBaos = _currentWriterBacking
                    _currentWriterBacking = _backgroundWriterBacking
                    _backgroundWriterBacking = tmpBaos

                    _currentWriterBacking.reset()

                     byte[] buffer = _backgroundWriterBacking.toByteArray()
                    _asyncWrite = CompletableFuture.supplyAsync(() -> {
                        try {
                            _requestBodyStream.write(buffer)

                            // send this chunk
                            _requestBodyStream.flush()
                        } catch (IOException e) {
                            throw new RuntimeException(e)
                        }
                        return null
                    })
                }
            } catch (Exception e) {
                RuntimeException error = getExceptionFromOperation()
                if (error != null) {
                    throw error
                }

                throwOnUnavailableStream(id, e)
            }
        } ly {
            _concurrentCheck.set(0)
        }

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

     static  verifyValidId(string id) {
        if (stringUtils.isEmpty(id)) {
            throw new IllegalStateException("Document id must have a non empty value")
        }

        if (id.endsWith("|")) {
            throw new UnsupportedOperationException("Document ids cannot end with '|', but was called with " + id)
        }
    }


      ensureStream() {
        try {
            BulkInsertCommand bulkCommand = new BulkInsertCommand(_operationId, _streamExposerContent, useCompression)

            _bulkInsertExecuteTask = CompletableFuture.supplyAsync(() -> {
                _requestExecutor.execute(bulkCommand)
                return null
            })

            _stream = _streamExposerContent.outputStream.get()

            _requestBodyStream = _stream

            _currentWriter.write('[')
        } catch (Exception e) {
            throw new RavenException("Unable to open bulk insert stream ", e)
        }
    }

      throwOnUnavailableStream(string id, Exception innerEx) {
        _streamExposerContent.errorOnProcessingRequest(new BulkInsertAbortedException("Write to stream failed at document with id " + id, innerEx))

        try {
            _bulkInsertExecuteTask.get()
        } catch (Exception e) {
            throw ExceptionsUtils.unwrapException(e)
        }
    }

      abort() {
        if (_operationId == -1) {
            return // nothing was done, nothing to kill
        }

        waitForId()

        try {
            _requestExecutor.execute(new KillOperationCommand(_operationId))
        } catch (RavenException e) {
            throw new BulkInsertAbortedException("Unable to kill ths bulk insert operation, because it was not found on the server.")
        }
    }

    @Override
      Close() {
        Exception flushEx = null

        if (_stream != null) {
            try {
                _currentWriter.write("]")
                _currentWriter.flush()

                _asyncWrite.get()

                byte[] buffer = _currentWriterBacking.toByteArray()
                _requestBodyStream.write(buffer)
                _stream.flush()
            } catch (Exception e) {
                flushEx = e
            }
        }

        _streamExposerContent.done()

        if (_operationId == -1) {
            // closing without calling a single store.
            return
        }

        if (_bulkInsertExecuteTask != null) {
            try {
                _bulkInsertExecuteTask.get()
            } catch (Exception e) {
                throwBulkInsertAborted(e, flushEx)
            }
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
