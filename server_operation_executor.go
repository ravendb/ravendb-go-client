package ravendb

type ServerOperationExecutor struct {
	requestExecutor *RequestExecutor // TODO: ClusterRequestExecutor
}

func NewServerOperationExecutor(store *DocumentStore) *ServerOperationExecutor {
	res := &ServerOperationExecutor{}
	urls := store.getUrls()
	dbName := store.getDatabase()
	conv := store.getConventions()
	if conv.isDisableTopologyUpdates() {
		// TODO: ClusterRequestExecutor_createForSingleNode()
		res.requestExecutor = RequestExecutor_createForSingleNodeWithoutConfigurationUpdates(urls[0], dbName, nil, conv)
	} else {
		// TODO: ClusterRequestExecutor_create()
		res.requestExecutor = RequestExecutor_create(urls, dbName, nil, conv)
	}
	fn := func(store *DocumentStore) {
		res.requestExecutor.close()
	}
	store.addAfterCloseListener(fn)
	return res
}

func (e *ServerOperationExecutor) send(operation IServerOperation) error {
	command := operation.getCommand(e.requestExecutor.getConventions())
	err := e.requestExecutor.executeCommand(command)
	return err
}

func (e *ServerOperationExecutor) sendAsync(operation IServerOperation) (*Operation, error) {
	requestExecutor := e.requestExecutor
	command := operation.getCommand(requestExecutor.getConventions())
	err := requestExecutor.executeCommand(command)
	if err != nil {
		return nil, err
	}
	result := getCommandOperationIdResult(command)
	return NewServerWideOperation(requestExecutor, requestExecutor.getConventions(), result.getOperationId()), nil
}

func (e *ServerOperationExecutor) close() {
	e.requestExecutor.close()
	e.requestExecutor = nil
}
