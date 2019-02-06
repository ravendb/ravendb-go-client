package ravendb

// MultiGetOperation represents multi-get operation
type MultiGetOperation struct {
	_session *InMemoryDocumentSessionOperations
}

func (o *MultiGetOperation) createRequest(requests []*getRequest) *MultiGetCommand {
	return NewMultiGetCommand(o._session.GetRequestExecutor().Cache, requests)
}

// Note: not used
func (o *MultiGetOperation) setResult(result map[string]interface{}) {
	// no-op
}
