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
	return res
}

// TODO: make argument an IServerOperation
func (e *ServerOperationExecutor) send(command *RavenCommand) (interface{}, error) {
	err := e.requestExecutor.executeCommand(command)
	return command.getResult(), err
}
