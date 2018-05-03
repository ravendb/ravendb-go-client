package ravendb

// RequestsExecutor describes executor of HTTP requests
type RequestsExecutor struct {
	databaseName string
}

// NewRequestsExecutor creates a new executor
// https://sourcegraph.com/github.com/ravendb/RavenDB-Python-Client@v4.0/-/blob/pyravendb/connection/requests_executor.py#L21
// TODO: certificate, conventions
func NewRequestsExecutor(databaseName string) *RequestsExecutor {
	res := &RequestsExecutor{}
	return res
}

// CreateRequestsExecutor creates a RequestsExecutor
// https://sourcegraph.com/github.com/ravendb/RavenDB-Python-Client@v4.0/-/blob/pyravendb/connection/requests_executor.py#L52
// TODO: certificate, conventions
func CreateRequestsExecutor(urls []string, databaseName string) *RequestsExecutor {
	re := NewRequestsExecutor(databaseName)
	re.startFirstTopologyThread(urls)
	return re
}

// https://sourcegraph.com/github.com/ravendb/RavenDB-Python-Client@v4.0/-/blob/pyravendb/connection/requests_executor.py#L63
func (re *RequestsExecutor) startFirstTopologyThread(urls []string) {
	// re.firstTopologyUpdate = NewPropagatingThread(target=self.first_topology_update, args=(urls,), daemon=True)
	// re.firstTopologyUpdate.start()
}
