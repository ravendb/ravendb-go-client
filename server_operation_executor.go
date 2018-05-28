package ravendb

type ServerOperationExecutor struct {
	requestExecutor *RequestExecutor // TODO: ClusterRequestExecutor
}

func NewServerOperationExecutor(store *DocumentStore) *ServerOperationExecutor {
	res := &ServerOperationExecutor{}
	urls := store.getURLS()
	dbName := store.getDatabase()
	conv := store.getConventions()
	if conv.DisableTopologyUpdate {
		// TODO: ClusterRequestExecutor_createForSingleNode()
		res.requestExecutor = CreateRequestsExecutorForSingleNode(urls[0], dbName)
	} else {
		// TODO: ClusterRequestExecutor_create()
		res.requestExecutor = CreateRequestsExecutor(urls, dbName, conv)
	}
	return res
}
