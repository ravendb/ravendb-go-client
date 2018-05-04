package ravendb

import (
	"net/http"
	"sync"
	"time"
)

// NodeSelector describes node selector
type NodeSelector struct {
}

// RequestsExecutor describes executor of HTTP requests
type RequestsExecutor struct {
	databaseName string
	TopologyEtag int

	urls                   []string // TODO: temporary
	lastReturnResponse     time.Time
	Conventions            *DocumentConventions
	nodeSelector           *NodeSelector
	lastKnownUrls          []string
	headers                map[string]string
	updateTopologyLock     sync.Mutex
	updateTimerLock        sync.Mutex
	lock                   sync.Mutex
	disableTopologyUpdates bool

	// TODO:
	// failedNodesTimers
	updateTopologyTimer *time.Timer
	firstTopologyUpdate chan bool  // TODO: something other than bool
	topologyNodes       []struct{} // TODO: something other than struct{}
	closed              bool
}

// NewRequestsExecutor creates a new executor
// https://sourcegraph.com/github.com/ravendb/RavenDB-Python-Client@v4.0/-/blob/pyravendb/connection/requests_executor.py#L21
// TODO: certificate
func NewRequestsExecutor(databaseName string, conventions *DocumentConventions) *RequestsExecutor {
	if conventions == nil {
		conventions = NewDocumentConventions()
	}
	res := &RequestsExecutor{
		Conventions: conventions,
		headers:     map[string]string{},
	}
	return res
}

// CreateRequestsExecutor creates a RequestsExecutor
// https://sourcegraph.com/github.com/ravendb/RavenDB-Python-Client@v4.0/-/blob/pyravendb/connection/requests_executor.py#L52
// TODO: certificate, conventions
func CreateRequestsExecutor(urls []string, databaseName string, conventions *DocumentConventions) *RequestsExecutor {
	re := NewRequestsExecutor(databaseName, conventions)
	re.urls = urls
	re.startFirstTopologyThread(urls)
	return re
}

// https://sourcegraph.com/github.com/ravendb/RavenDB-Python-Client@v4.0/-/blob/pyravendb/connection/requests_executor.py#L63
func (re *RequestsExecutor) startFirstTopologyThread(urls []string) {
	// re.firstTopologyUpdate = NewPropagatingThread(target=self.first_topology_update, args=(urls,), daemon=True)
	// re.firstTopologyUpdate.start()
}

// GetExecutor returns command executor function
func (re *RequestsExecutor) GetExecutor() CommandExecutorFunc {
	fn := func(cmd *RavenCommand, shouldRetry bool) (*http.Response, error) {
		return nil, nil
	}
	return fn
}

// Execute executes a command
func (re *RequestsExecutor) Execute(cmd *RavenCommand, shouldRetry bool) {
	/*node := &ServerNode{
		URL:        re.urls[0],
		Database:   re.databaseName,
		ClusterTag: "0", // TODO: is it re.TopologyEtag?
	}
	exec := MakeSimpleExecutor(node)
	ExecuteCommand(exec, cmd)
	*/
}
