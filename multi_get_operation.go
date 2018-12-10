package ravendb

type MultiGetOperation struct {
	_session *InMemoryDocumentSessionOperations
}

func NewMultiGetOperation(session *InMemoryDocumentSessionOperations) *MultiGetOperation {
	return &MultiGetOperation{
		_session: session,
	}
}

func (o *MultiGetOperation) createRequest(requests []*GetRequest) *MultiGetCommand {
	return NewMultiGetCommand(o._session.GetRequestExecutor().Cache, requests)
}

func (o *MultiGetOperation) setResult(result ObjectNode) {
}
