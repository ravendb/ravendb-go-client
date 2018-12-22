package ravendb

// MultiGetOperation represents multi-get operation
type MultiGetOperation struct {
	_session *InMemoryDocumentSessionOperations
}

func (o *MultiGetOperation) createRequest(requests []*GetRequest) *MultiGetCommand {
	return NewMultiGetCommand(o._session.GetRequestExecutor().Cache, requests)
}

/* TODO: not used anywhere
func (o *MultiGetOperation) setResult(result ObjectNode) {
}
*/
