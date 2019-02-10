package ravendb

// MultiGetOperation represents multi-get operation
type MultiGetOperation struct {
	session *InMemoryDocumentSessionOperations
}

func (o *MultiGetOperation) createRequest(requests []*getRequest) *MultiGetCommand {
	return newMultiGetCommand(o.session.GetRequestExecutor().Cache, requests)
}

// Note: not used
func (o *MultiGetOperation) setResult(result map[string]interface{}) {
	// no-op
}
