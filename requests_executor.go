package ravendb

import (
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"
)

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

	waitForFirstTopologyUpdate sync.WaitGroup

	failedNodesTimers   map[*ServerNode]*NodeStatus
	updateTopologyTimer *time.Timer
	topologyNodes       []*ServerNode
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
		Conventions:        conventions,
		headers:            map[string]string{},
		failedNodesTimers:  map[*ServerNode]*NodeStatus{},
		lastReturnResponse: time.Now(),
		databaseName:       databaseName,
	}
	return res
}

// CreateRequestsExecutor creates a RequestsExecutor
// https://sourcegraph.com/github.com/ravendb/RavenDB-Python-Client@v4.0/-/blob/pyravendb/connection/requests_executor.py#L52
// TODO: certificate, conventions
func CreateRequestsExecutor(urls []string, dbName string, conventions *DocumentConventions) *RequestsExecutor {
	re := NewRequestsExecutor(dbName, conventions)
	re.urls = urls
	re.startFirstTopologyThread(urls)
	return re
}

// CreateRequestsExecutorForSingleNode creates RequestsExecutor for a single server
// TODO: certificate
func CreateRequestsExecutorForSingleNode(url string, dbName string) *RequestsExecutor {
	topology := NewTopology()
	topology.Etag = -1
	node := NewServerNode(url, dbName)
	topology.Nodes = []*ServerNode{node}

	re := NewRequestsExecutor(dbName, nil)
	re.nodeSelector = NewNodeSelector(topology)
	re.disableTopologyUpdates = true
	return re
}

// https://sourcegraph.com/github.com/ravendb/RavenDB-Python-Client@v4.0/-/blob/pyravendb/connection/requests_executor.py#L63
func (re *RequestsExecutor) startFirstTopologyThread(urls []string) {
	fmt.Printf("startFirstTopologyThread\n")
	initialUrls := re.urls
	re.waitForFirstTopologyUpdate.Add(1)
	go func() {
		re.firstTopologyUpdate(initialUrls)
		re.waitForFirstTopologyUpdate.Done()
	}()
}

func (re *RequestsExecutor) ensureNodeSelector() {
	re.waitForFirstTopologyUpdate.Wait()
	if re.nodeSelector != nil {
		return
	}
	t := NewTopology()
	t.Etag = re.TopologyEtag
	t.Nodes = re.topologyNodes
	re.nodeSelector = NewNodeSelector(t)
}

func (re *RequestsExecutor) getPreferredNode() *ServerNode {
	re.ensureNodeSelector()
	return re.nodeSelector.GetCurrentNode()
}

// Execute sends a command to the server via http and parses a result
func (re *RequestsExecutor) Execute(ravenCommand *RavenCommand, shouldRetry bool) (*http.Response, error) {
	//fmt.Printf("waiting for firstTopologyUpdate() to finish\n")
	re.waitForFirstTopologyUpdate.Wait()
	//fmt.Printf("firstTopologyUpdate() finished\n")
	chosenNode := re.nodeSelector.GetCurrentNode()
	return re.ExecuteWithNode(chosenNode, ravenCommand, shouldRetry)
}

// ExecuteWithNode sends a command to the server via http and parses a result
func (re *RequestsExecutor) ExecuteWithNode(chosenNode *ServerNode, ravenCommand *RavenCommand, shouldRetry bool) (*http.Response, error) {
	for {
		nodeIndex := 0
		if re.nodeSelector != nil {
			nodeIndex = re.nodeSelector.CurrentNodeIndex
		}
		req, err := makeHTTPRequest(chosenNode, ravenCommand)
		if !re.disableTopologyUpdates {
			etagStr := fmt.Sprintf(`"%d"`, re.TopologyEtag)
			req.Header.Add("Topology-Etag", etagStr)
		}

		// TODO: handle an error?
		must(err)
		client := &http.Client{
			Timeout: time.Second * 5,
		}
		rsp, err := client.Do(req)
		// this is for network-level errors when we don't get response
		if err != nil {
			// if asked, retry network-level errors
			if shouldRetry == false {
				return nil, err
			}
			if !re.handleServerDown(chosenNode, nodeIndex, ravenCommand, err) {
				// TODO: wrap in AllTopologyNodesDownError
				return nil, err
			}
			chosenNode = re.nodeSelector.GetCurrentNode()
			continue
		}

		code := rsp.StatusCode

		// convert 404 Not Found to NotFoundError
		if rsp.StatusCode == http.StatusNotFound {
			// TODO: does it ever return non-empty response?
			res := NotFoundError{
				URL: req.URL.String(),
			}
			return nil, &res
		}

		// 403
		if code == http.StatusForbidden {
			// TOOD: if certificate is nil, load certificate and retry
			panicIf(true, "NYI")
			return nil, err
		}

		// 410
		if code == http.StatusGone {
			if shouldRetry {
				re.updateTopology(chosenNode, true)
				continue
			} else {
				// TODO: python code always retries
				return nil, err
			}
		}

		// 408, 502, 503, 504
		if code == http.StatusRequestTimeout || code == http.StatusBadGateway || code == http.StatusServiceUnavailable || code == http.StatusGatewayTimeout {
			if len(ravenCommand.failedNodes) == 1 {
				panicIf(true, "NYI")
				databaseMissing := rsp.Header.Get("Database-Missing")
				if databaseMissing != "" {
					// TODO: return DatabaseDoesNotExistException
					return nil, err
				}
				// TODO: return UnsuccessfulRequestException
				// node := ravenCommand.failedNodes[0]
				return nil, err
			}

			// TODO: e = response.json()["Message"]
			if re.handleServerDown(chosenNode, nodeIndex, ravenCommand, nil) {
				chosenNode = re.nodeSelector.GetCurrentNode()
			}
			continue
		}

		// 409
		if code == http.StatusConflict {
			// TODO: conflict resolution
			return nil, err
		}

		// convert 400 Bad Request response to BadReqeustError
		// TODO: in python code this only happends for some commands
		if rsp.StatusCode == http.StatusBadRequest {
			var res BadRequestError
			err = decodeJSONFromReader(rsp.Body, &res)
			if err != nil {
				return nil, err
			}
			return nil, &res
		}

		if rsp.Header.Get("Refresh-Topology") != "" {
			node := NewServerNode(chosenNode.URL, re.databaseName)
			re.updateTopology(node, false)
		}
		re.lastReturnResponse = time.Now()
		return rsp, nil
	}
}

// GetCommandExecutorWithNode returns command executor for a given node
func (re *RequestsExecutor) GetCommandExecutorWithNode(node *ServerNode, shouldRetry bool) CommandExecutorFunc {
	f := func(cmd *RavenCommand) (*http.Response, error) {
		return re.ExecuteWithNode(node, cmd, shouldRetry)
	}
	return f
}

// GetCommandExecutor returns command executor
func (re *RequestsExecutor) GetCommandExecutor(shouldRetry bool) CommandExecutorFunc {
	f := func(cmd *RavenCommand) (*http.Response, error) {
		return re.Execute(cmd, shouldRetry)
	}
	return f
}

func (re *RequestsExecutor) firstTopologyUpdate(initialUrls []string) error {
	var errorList []error
	for _, url := range initialUrls {
		panicIf(re.databaseName == "", "re.databaseName is empty")
		node := NewServerNode(url, re.databaseName)
		err := re.updateTopology(node, false)
		// TODO: if DatabaseDoesNotExistException
		if err != nil {
			errorList = append(errorList, err)
			continue
		}
		cb := func() {
			re.updateTopologyCallback()
		}
		dur := time.Second * 60 * 5 // TODO: verify this the time
		re.updateTopologyTimer = time.AfterFunc(dur, cb)
		re.topologyNodes = re.nodeSelector.Topology.Nodes
	}
	// TODO: try to load from cache
	/*
		for url in initial_urls:
			if self.try_load_from_cache(url):
				self.topology_nodes = self._node_selector.topology.nodes
				return
	*/
	re.lastKnownUrls = initialUrls
	if len(errorList) == 0 {
		return nil
	}
	return errorList[0]
}

// TODO: write me. this should be configurable by the user
func (re *RequestsExecutor) tryLoadFromCache(url string) {
	/*
	   server_hash = hashlib.md5(
	       "{0}{1}".format(url, self._database_name).encode(
	           'utf-8')).hexdigest()
	   topology_file_path = "{0}\{1}.raven-topology".format(os.getcwd(), server_hash)
	   try:
	       with open(topology_file_path, 'r') as topology_file:
	           json_file = json.load(topology_file)
	           self._node_selector = NodeSelector(
	               Topology.convert_json_topology_to_entity(json_file))
	           self.topology_etag = -2
	           self.update_topology_timer = Utils.start_a_timer(60 * 5, self.update_topology_callback, daemon=True)
	           return True
	   except (FileNotFoundError, json.JSONDecodeError) as e:
	       log.info(e)
	   return False
	*/
}

// TODO: write me. this should be configurable by the user
func writeToCache(topology *Topology, node *ServerNode) {
	/*
		hash_name = hashlib.md5(
			"{0}{1}".format(node.url, node.database).encode(
				'utf-8')).hexdigest()

		topology_file = "{0}\{1}.raven-topology".format(os.getcwd(), hash_name)
		try:
			with open(topology_file, 'w') as outfile:
				json.dump(response, outfile, ensure_ascii=False)
		except (IOError, json.JSONDecodeError):
			pass
	*/
}

func (re *RequestsExecutor) updateTopology(node *ServerNode, forceUpdate bool) error {
	panicIf(node.Database == "", "node.Database is empty in %#v", node)
	fmt.Printf("updateTopology\n")
	if re.closed {
		return errors.New("RequestsExecutor is closed")
	}
	re.updateTopologyLock.Lock()
	defer re.updateTopologyLock.Unlock()

	cmd := NewGetTopologyCommand()
	exec := re.GetCommandExecutorWithNode(node, false)
	topology, err := ExecuteGetTopologyCommand(exec, cmd)
	if err != nil {
		return err
	}
	writeToCache(topology, node)
	if re.nodeSelector == nil {
		re.nodeSelector = NewNodeSelector(topology)
	} else {
		re.nodeSelector.OnUpdateTopology(topology, forceUpdate)
	}
	re.TopologyEtag = re.nodeSelector.Topology.Etag
	return nil
}

func (re *RequestsExecutor) handleServerDown(chosenNode *ServerNode, nodeIndex int, command *RavenCommand, err error) bool {
	command.addFailedNode(chosenNode)
	nodeSelector := re.nodeSelector

	if nodeSelector == nil {
		return false
	}

	re.updateTimerLock.Lock()
	re.updateTimerLock.Unlock()

	nodeStatus, ok := re.failedNodesTimers[chosenNode]
	if ok {
		re.updateTimerLock.Unlock()
		return true
	}
	nodeStatus = NewNodeStatus(re, nodeIndex, chosenNode)
	re.failedNodesTimers[chosenNode] = nodeStatus
	nodeStatus.startTimer()
	re.updateTimerLock.Unlock()

	nodeSelector.OnFailedRequest(nodeIndex)
	currentNode := nodeSelector.GetCurrentNode()
	return command.isFailedWithNode(currentNode)
}

func (re *RequestsExecutor) cancelAllFailedNodesTimers() {
	for k, t := range re.failedNodesTimers {
		t.Cancel()
		delete(re.failedNodesTimers, k)
	}
}

func (re *RequestsExecutor) checkNodeStatus(nodeStatus *NodeStatus) {
	if re.nodeSelector == nil {
		return
	}
	nodes := re.nodeSelector.Topology.Nodes
	if nodeStatus.nodeIndex > len(nodes) {
		return
		serverNode := nodes[nodeStatus.nodeIndex]
		if serverNode != nodeStatus.node {
			re.performHealthCheck(serverNode, nodeStatus)
		}
	}
}

func (re *RequestsExecutor) performHealthCheck(node *ServerNode, nodeStatus *NodeStatus) {
	command := NewGetStatisticsCommand("failure=check")
	exec := re.GetCommandExecutorWithNode(node, false)
	_, err := ExecuteGetStatisticsCommand(exec, command)
	if err != nil {
		failedNodeTimer, ok := re.failedNodesTimers[nodeStatus.node]
		if ok {
			failedNodeTimer.startTimer()
			return
		}
	}

	failedNodeTimer, ok := re.failedNodesTimers[nodeStatus.node]
	if ok {
		failedNodeTimer.Cancel()
		delete(re.failedNodesTimers, nodeStatus.node)
	}
	re.nodeSelector.RestoreNodeIndex(nodeStatus.nodeIndex)
}

func (re *RequestsExecutor) updateTopologyCallback() {
	now := time.Now()
	if now.Sub(re.lastReturnResponse) < time.Minute*5 {
		return
	}
	re.updateTopology(re.nodeSelector.GetCurrentNode(), false)
}

// Close should be called when deleting executor
func (re *RequestsExecutor) Close() {
	if re.closed {
		return
	}
	re.closed = true
	re.cancelAllFailedNodesTimers()
	if re.updateTopologyTimer != nil {
		re.updateTopologyTimer.Stop()
	}
}
