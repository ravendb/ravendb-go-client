package ravendb

type ServerOperationExecutor struct {
	requestExecutor *ClusterRequestExecutor
}

func NewServerOperationExecutor(store *DocumentStore) *ServerOperationExecutor {
	res := &ServerOperationExecutor{}
	urls := store.getUrls()
	cert := store.getCertificate()
	conv := store.getConventions()
	if conv.isDisableTopologyUpdates() {
		res.requestExecutor = ClusterRequestExecutor_createForSingleNode(urls[0], cert, conv)
	} else {
		res.requestExecutor = ClusterRequestExecutor_create(urls, cert, conv)
	}
	fn := func(store *DocumentStore) {
		res.requestExecutor.Close()
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

func (e *ServerOperationExecutor) Close() {
	e.requestExecutor.Close()
	e.requestExecutor = nil
}
