package ravendb

type ServerOperationExecutor struct {
	requestExecutor *ClusterRequestExecutor
}

func NewServerOperationExecutor(store *DocumentStore) *ServerOperationExecutor {
	res := &ServerOperationExecutor{}
	urls := store.GetUrls()
	cert := store.GetCertificate()
	conv := store.GetConventions()
	if conv.IsDisableTopologyUpdates() {
		res.requestExecutor = ClusterRequestExecutor_createForSingleNode(urls[0], cert, conv)
	} else {
		res.requestExecutor = ClusterRequestExecutor_create(urls, cert, conv)
	}
	fn := func(store *DocumentStore) {
		res.requestExecutor.Close()
	}
	store.AddAfterCloseListener(fn)
	return res
}

func (e *ServerOperationExecutor) Send(operation IServerOperation) error {
	command := operation.GetCommand(e.requestExecutor.GetConventions())
	err := e.requestExecutor.ExecuteCommand(command)
	return err
}

func (e *ServerOperationExecutor) SendAsync(operation IServerOperation) (*Operation, error) {
	requestExecutor := e.requestExecutor
	command := operation.GetCommand(requestExecutor.GetConventions())
	err := requestExecutor.ExecuteCommand(command)
	if err != nil {
		return nil, err
	}
	result := getCommandOperationIDResult(command)
	return NewServerWideOperation(requestExecutor, requestExecutor.GetConventions(), result.getOperationId()), nil
}

func (e *ServerOperationExecutor) Close() {
	e.requestExecutor.Close()
	e.requestExecutor = nil
}
