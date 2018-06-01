package ravendb

import (
	"errors"
	"fmt"
	"math"
	"net/http"
	"strconv"
	"sync"
	"time"
)

//     private static final GetStatisticsOperation failureCheckOperation = new GetStatisticsOperation("failure=check");
// public static Consumer<HttpRequestBase> requestPostProcessor = null;

const (
	goClientVersion = "0.1"
)

// RequestExecutor describes executor of HTTP requests
type RequestExecutor struct {
	certificate           *KeyStore
	_databaseName         string
	_lastReturnedResponse time.Time

	_updateTopologyTimer *time.Timer
	_nodeSelector        *NodeSelector

	numberOfServerRequests  AtomicInteger
	topologyEtag            int
	clientConfigurationEtag int
	Conventions             *DocumentConventions

	_disableTopologyUpdates            bool
	_disableClientConfigurationUpdates bool

	_firstTopologyUpdate *CompletableFuture

	_readBalanceBehavior   ReadBalanceBehavior
	_topologyTakenFromNode *ServerNode

	_lastKnownUrls []string

	mu sync.Mutex

	_disposed bool
	// old stuff

	urls               []string // TODO: temporary
	lastKnownUrls      []string
	headers            map[string]string
	updateTopologyLock sync.Mutex
	updateTimerLock    sync.Mutex
	lock               sync.Mutex
}

func (re *RequestExecutor) getTopology() *Topology {
	if re._nodeSelector != nil {
		return re._nodeSelector.getTopology()
	}
	return nil
}

func (re *RequestExecutor) getTopologyNodes() []*ServerNode {
	var res []*ServerNode
	nodes := re.getTopology().getNodes()
	for _, n := range nodes {
		// TODO: is this really filtered. I don't quite get Java code
		if n != nil {
			res = append(res, n)
		}
	}
	return res
}

func (re *RequestExecutor) getUrl() String {
	if re._nodeSelector == nil {
		return ""
	}

	preferredNode := re._nodeSelector.getPreferredNode()
	if preferredNode != nil {
		return preferredNode.currentNode.getUrl()
	}
	return ""
}

func (re *RequestExecutor) getTopologyEtag() int {
	return re.topologyEtag
}

func (re *RequestExecutor) getClientConfigurationEtag() int {
	return re.clientConfigurationEtag
}

func (re *RequestExecutor) getConventions() *DocumentConventions {
	return re.Conventions
}

func (re *RequestExecutor) getCertificate() *KeyStore {
	return re.certificate
}

// NewRequestExecutor creates a new executor
// https://sourcegraph.com/github.com/ravendb/RavenDB-Python-Client@v4.0/-/blob/pyravendb/connection/requests_executor.py#L21
// TODO: certificate
func NewRequestExecutor(databaseName string, certificate *KeyStore, conventions *DocumentConventions, initialUrls []string) *RequestExecutor {
	if conventions == nil {
		conventions = NewDocumentConventions()
	}
	res := &RequestExecutor{
		_readBalanceBehavior:  conventions.getReadBalanceBehavior(),
		_databaseName:         databaseName,
		certificate:           certificate,
		_lastReturnedResponse: time.Now(),
		Conventions:           conventions.clone(),

		headers: map[string]string{},
	}
	return res
}

// TODO: only used for http cache?
//private String extractThumbprintFromCertificate(KeyStore certificate) {

func RequestExecutor_create(initialUrls []string, databaseName string, certificate *KeyStore, conventions *DocumentConventions) *RequestExecutor {
	re := NewRequestExecutor(databaseName, certificate, conventions, initialUrls)
	re._firstTopologyUpdate = re.firstTopologyUpdate(initialUrls)
	return re
}

func RequestExecutor_createForSingleNodeWithConfigurationUpdates(url string, databaseName string, certificate *KeyStore, conventions *DocumentConventions) *RequestExecutor {
	executor := RequestExecutor_createForSingleNodeWithoutConfigurationUpdates(url, databaseName, certificate, conventions)
	executor._disableClientConfigurationUpdates = false
	return executor
}

func RequestExecutor_createForSingleNodeWithoutConfigurationUpdates(url string, databaseName string, certificate *KeyStore, conventions *DocumentConventions) *RequestExecutor {
	initialUrls := RequestExecutor_validateUrls([]string{url}, certificate)
	executor := NewRequestExecutor(databaseName, certificate, conventions, initialUrls)

	topology := NewTopology()
	topology.setEtag(-1)

	serverNode := NewServerNode()
	serverNode.setDatabase(databaseName)
	serverNode.setUrl(initialUrls[0])
	// TODO: is Collections.singletonList in Java code subtly significant?
	topology.setNodes([]*ServerNode{serverNode})

	executor._nodeSelector = NewNodeSelector(topology)
	executor.topologyEtag = -2
	executor._disableTopologyUpdates = true
	executor._disableClientConfigurationUpdates = true

	return executor
}

func (re *RequestExecutor) updateClientConfigurationAsync() *CompletableFuture {
	// TODO: implement me
	panicIf(true, "NYI")
	return nil
}

func (re *RequestExecutor) updateTopologyAsync(node *ServerNode, timeout int) *CompletableFuture {
	return re.updateTopologyAsyncWithForceUpdate(node, timeout, false)
}

func (re *RequestExecutor) updateTopologyAsyncWithForceUpdate(node *ServerNode, timeout int, forceUpdate bool) *CompletableFuture {
	// TODO: handle _disposed
	// TODO: locking with _updateDatabaseTopologySemaphore
	future := NewCompletableFuture()
	f := func() {
		command := NewGetDatabaseTopologyCommand()
		err := re.execute(node, 0, command, false, nil)
		if err != nil {
			future.markAsDoneWithError(err)
			return
		}
		result := command.result.(*Topology)
		if re._nodeSelector == nil {
			re._nodeSelector = NewNodeSelector(result)
			if re._readBalanceBehavior == ReadBalanceBehavior_FASTEST_NODE {
				re._nodeSelector.scheduleSpeedTest()
			}
		} else if re._nodeSelector.onUpdateTopology(result, forceUpdate) {
			re.disposeAllFailedNodesTimers()
			if re._readBalanceBehavior == ReadBalanceBehavior_FASTEST_NODE {
				re._nodeSelector.scheduleSpeedTest()
			}
		}
	}
	go f()
	return future
}

func (re *RequestExecutor) disposeAllFailedNodesTimers() {
	// TODO: implement me
	panicIf(true, "NYI")
}

// execute(command) in java
func (re *RequestExecutor) executeCommand(command *RavenCommand) error {
	return re.executeCommandWithSessionInfo(command, nil)
}

// execute(command, session) in java
func (re *RequestExecutor) executeCommandWithSessionInfo(command *RavenCommand, sessionInfo *SessionInfo) error {
	topologyUpdate := re._firstTopologyUpdate
	if (topologyUpdate != nil && topologyUpdate.isDone()) || re._disableTopologyUpdates {
		currentIndexAndNode := re.chooseNodeForRequest(command, sessionInfo)
		return re.execute(currentIndexAndNode.currentNode, currentIndexAndNode.currentIndex, command, true, sessionInfo)
	} else {
		return re.unlikelyExecute(command, topologyUpdate, sessionInfo)
	}
}

func (re *RequestExecutor) chooseNodeForRequest(cmd *RavenCommand, sessionInfo *SessionInfo) *CurrentIndexAndNode {
	if !cmd.isReadRequest() {
		return re._nodeSelector.getPreferredNode()
	}

	switch re._readBalanceBehavior {
	case ReadBalanceBehavior_NONE:
		return re._nodeSelector.getPreferredNode()
	case ReadBalanceBehavior_ROUND_ROBIN:
		sessionID := 0
		if sessionInfo != nil {
			sessionID = sessionInfo.SessionID
		}
		return re._nodeSelector.getNodeBySessionId(sessionID)
	case ReadBalanceBehavior_FASTEST_NODE:
		return re._nodeSelector.getFastestNode()
	default:
		panicIf(true, "Unknown re._readBalanceBehavior: '%s'", re._readBalanceBehavior)
	}
	return nil
}

func (re *RequestExecutor) unlikelyExecuteInner(command *RavenCommand, topologyUpdate *CompletableFuture, sessionInfo *SessionInfo) error {

	if topologyUpdate == nil {
		re.mu.Lock()
		defer re.mu.Unlock()
		if re._firstTopologyUpdate == nil {
			if len(re._lastKnownUrls) == 0 {
				return NewIllegalStateException("No known topology and no previously known one, cannot proceed, likely a bug")
			}

			re._firstTopologyUpdate = re.firstTopologyUpdate(re._lastKnownUrls)
		}

		topologyUpdate = re._firstTopologyUpdate
	}

	topologyUpdate.get()
	return nil
}

func (re *RequestExecutor) unlikelyExecute(command *RavenCommand, topologyUpdate *CompletableFuture, sessionInfo *SessionInfo) error {
	err := re.unlikelyExecuteInner(command, topologyUpdate, sessionInfo)
	if err != nil {
		re.mu.Lock()
		if re._firstTopologyUpdate == topologyUpdate {
			re._firstTopologyUpdate = nil // next request will raise it
		}
		re.mu.Unlock()
		return err
	}

	currentIndexAndNode := re.chooseNodeForRequest(command, sessionInfo)
	re.execute(currentIndexAndNode.currentNode, currentIndexAndNode.currentIndex, command, true, sessionInfo)
	return nil
}

func (re *RequestExecutor) updateTopologyCallback() {
	dur := time.Since(re._lastReturnedResponse)
	if dur < time.Minute {
		return
	}

	var serverNode *ServerNode

	// TODO: early exist if getPreferredNode() returns an error
	preferredNode := re._nodeSelector.getPreferredNode()
	serverNode = preferredNode.currentNode

	re.updateTopologyAsync(serverNode, 0)
}

type Tuple_String_Error struct {
	S   string
	Err error
}

func (re *RequestExecutor) firstTopologyUpdate(inputUrls []string) *CompletableFuture {
	initialUrls := RequestExecutor_validateUrls(inputUrls, re.certificate)

	res := NewCompletableFuture()
	//var list []*Tuple_String_Error
	f := func() {
		for _, url := range initialUrls {
			var err error
			serverNode := NewServerNode()
			serverNode.setUrl(url)
			serverNode.setDatabase(re._databaseName)

			re.updateTopologyAsync(serverNode, math.MaxInt32).get()

			re.initializeUpdateTopologyTimer()

			re._topologyTakenFromNode = serverNode
			if err == nil {
				res.markAsDone(nil)
				return
			}
		}
		/* TODO:
		catch (Exception e) {
			if (e instanceof ExecutionException && e.getCause() instanceof DatabaseDoesNotExistException) {
				// Will happen on all node in the cluster,
				// so errors immediately
				_lastKnownUrls = initialUrls;
				throw (DatabaseDoesNotExistException) e.getCause();
			}
			if (initialUrls.length == 0) {
				_lastKnownUrls = initialUrls;
				throw new IllegalStateException("Cannot get topology from server: " + url, e);
			}

			list.add(Tuple.create(url, e));
		}
		*/

		/* TODO:
		       Topology topology = new Topology();
		       topology.setEtag(topologyEtag);

		       List<ServerNode> topologyNodes = getTopologyNodes();
		       if (topologyNodes == null) {
		           topologyNodes = Arrays.stream(initialUrls)
		                   .map(url -> {
		                       ServerNode serverNode = new ServerNode();
		                       serverNode.setUrl(url);
		                       serverNode.setDatabase(_databaseName);
		                       serverNode.setClusterTag("!");
		                       return serverNode;
		                   }).collect(Collectors.toList());
		       }

		       topology.setNodes(topologyNodes);

		       _nodeSelector = new NodeSelector(topology);

		       if (initialUrls != null && initialUrls.length > 0) {
		           initializeUpdateTopologyTimer();
		           return;
		       }

		       _lastKnownUrls = initialUrls;
		       String details = list.stream().map(x -> x.first + " -> " + Optional.ofNullable(x.second).map(m -> m.getMessage()).orElse("")).collect(Collectors.joining(", "));
		       throwExceptions(details);
		   });
		*/
	}
	go f()
	return res
}

// TODO: return an error
func (re *RequestExecutor) throwExceptions(details String) {
	err := NewIllegalStateException("Failed to retrieve database topology from all known nodes \n" + details)
	panicIf(true, "%s", err.Error())
}

func RequestExecutor_validateUrls(initialUrls []string, certificate *KeyStore) []string {
	// TODO: implement me
	return initialUrls
}

func (re *RequestExecutor) initializeUpdateTopologyTimer() {
	re.mu.Lock()
	defer re.mu.Unlock()
	if re._updateTopologyTimer != nil {
		return
	}
	// TODO: make it into an infinite goroutine instead
	f := func() {
		re.updateTopologyCallback()
		// Go doesn't have repeatable timer, so re-trigger ourselves
		re._updateTopologyTimer = nil
		re.initializeUpdateTopologyTimer()
	}
	re._updateTopologyTimer = time.AfterFunc(time.Minute, f)
}

func (re *RequestExecutor) execute(chosenNode *ServerNode, nodeIndex int, command *RavenCommand, shouldRetry bool, sessionInfo *SessionInfo) error {
	fmt.Printf("RequestExecutor.execute cmd: %#v\n", command)
	request, err := re.createRequest(chosenNode, command)
	if err != nil {
		return err
	}
	// TODO: caching

	if !re._disableClientConfigurationUpdates {
		etag := `"` + strconv.Itoa(re.clientConfigurationEtag) + `"`
		request.Header.Set(Constants_Headers_CLIENT_CONFIGURATION_ETAG, etag)
	}

	if !re._disableTopologyUpdates {
		etag := `"` + strconv.Itoa(re.topologyEtag) + `"`
		request.Header.Set(Constants_Headers_TOPOLOGY_ETAG, etag)
	}

	httpClient := &http.Client{
		Timeout: time.Second * 5,
	}

	//sp := time.Now()
	//responseDispose := ResponseDisposeHandling_AUTOMATIC
	var response *http.Response
	re.numberOfServerRequests.incrementAndGet()
	if re.shouldExecuteOnAll(chosenNode, command) {
		//response, err =
	} else {
		response, err = command.send(httpClient, request)
	}
	if err != nil {
		if !shouldRetry {
			return err
		}
		urlRef := request.RequestURI
		if !re.handleServerDown(urlRef, chosenNode, nodeIndex, command, request, response, err, sessionInfo) {
			return re.throwFailedToContactAllNodes(command, request, err, nil)
		}
		return nil
	}

	command.statusCode = response.StatusCode

	//refreshTopology := HttpExtensions_getBooleanHeader(response, Constants_Headers_REFRESH_TOPOLOGY)
	//refreshClientConfiguration := HttpExtensions_getBooleanHeader(response, Constants_Headers_REFRESH_CLIENT_CONFIGURATION)

	panicIf(true, "NYI")
	return nil
	// TODO: implement meinSpeedTestPhase
}

func (re *RequestExecutor) throwFailedToContactAllNodes(command *RavenCommand, request *http.Request, e error, timeoutException error) error {
	// TODO: implement me
	panicIf(true, "NYI")
	return errors.New("throwFailedToContactAllNodes")
}

func (re *RequestExecutor) inSpeedTestPhase() bool {
	return (re._nodeSelector != nil) && re._nodeSelector.inSpeedTestPhase()
}

func (re *RequestExecutor) shouldExecuteOnAll(chosenNode *ServerNode, command *RavenCommand) bool {
	multipleNodes := (re._nodeSelector != nil) && (len(re._nodeSelector.getTopology().getNodes()) > 1)

	return re._readBalanceBehavior == ReadBalanceBehavior_FASTEST_NODE &&
		re._nodeSelector != nil &&
		re._nodeSelector.inSpeedTestPhase() &&
		multipleNodes &&
		command.isReadRequest() &&
		command.getResponseType() == RavenCommandResponseType_OBJECT &&
		chosenNode != nil
}

/*
// ExecuteWithNode sends a command to the server via http and parses a result
func (re *RequestExecutor) ExecuteWithNode(chosenNode *ServerNode, ravenCommand *RavenCommand, shouldRetry bool) (*http.Response, error) {
	for {
		nodeIndex := 0
		if re._nodeSelector != nil {
			nodeIndex = re._nodeSelector.CurrentNodeIndex
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
			fmt.Printf("ExecuteWithNode: client.Do() failed with %s\n", err)
			// if asked, retry network-level errors
			if shouldRetry == false {
				return nil, err
			}
			if !re.handleServerDown(chosenNode, nodeIndex, ravenCommand, err) {
				// TODO: wrap in AllTopologyNodesDownError
				return nil, err
			}
			chosenNode = re._nodeSelector.GetCurrentNode()
			continue
		}

		body, _ := getCopyOfResponseBody(rsp)
		dumpHTTPResponse(rsp, body)

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
				chosenNode = re._nodeSelector.GetCurrentNode()
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
			node := NewServerNode(chosenNode.URL, re._databaseName)
			re.updateTopology(node, false)
		}
		re._lastReturnedResponse = time.Now()
		return rsp, nil
	}
}
*/

func (re *RequestExecutor) createRequest(node *ServerNode, command *RavenCommand) (*http.Request, error) {
	request, err := command.createRequest(node)
	if err != nil {
		return nil, err
	}
	request.Header.Set("Raven-Client-Version", goClientVersion)
	return request, err
}

// private <TResult> boolean handleUnsuccessfulResponse(ServerNode chosenNode, Integer nodeIndex, RavenCommand<TResult> command, HttpRequestBase request, CloseableHttpResponse response, String url, SessionInfo sessionInfo, boolean shouldRetry) {

func RequestExecutor_handleConflict(response *http.Response) error {
	// ExceptionDispatcher.throwException(response);
	return errors.New("ExceptionDispatcher")
}

//     public static InputStream readAsStream(CloseableHttpResponse response) throws IOException {

func (re *RequestExecutor) handleServerDown(url String, chosenNode *ServerNode, nodeIndex int, command *RavenCommand, request *http.Request, response *http.Response, e error, sessionInfo *SessionInfo) bool {
	if command.getFailedNodes() == nil {
		command.setFailedNodes(make(map[*ServerNode]error))
	}

	re.addFailedResponseToCommand(chosenNode, command, request, response, e)

	// TODO: -1 ?
	if nodeIndex == 0 {
		//We executed request over a node not in the topology. This means no failover...
		return false
	}

	re.spawnHealthChecks(chosenNode, nodeIndex)

	if re._nodeSelector == nil {
		return false
	}

	re._nodeSelector.onFailedRequest(nodeIndex)

	currentIndexAndNode := re._nodeSelector.getPreferredNode()
	if _, ok := command.getFailedNodes()[currentIndexAndNode.currentNode]; ok {
		return false //we tried all the nodes...nothing left to do
	}

	re.execute(currentIndexAndNode.currentNode, currentIndexAndNode.currentIndex, command, false, sessionInfo)

	return true
}

func (re *RequestExecutor) spawnHealthChecks(chosenNode *ServerNode, nodeIndex int) {
	panicIf(true, "NYI")
}

func (re *RequestExecutor) checkNodeStatusCallback(nodeStatus *NodeStatus) {
	panicIf(true, "NYI")
}

func (re *RequestExecutor) performHealthCheck(serverNode *ServerNode, nodeIndex int) error {
	panicIf(true, "NYI")
	//execute(serverNode, nodeIndex, failureCheckOperation.getCommand(conventions), false, null);
	return nil
}

func (re *RequestExecutor) addFailedResponseToCommand(chosenNode *ServerNode, command *RavenCommand, request *http.Request, response *http.Response, e error) {
	panicIf(true, "NYI")
}

// TODO: write me. this should be configurable by the user
func (re *RequestExecutor) tryLoadFromCache(url string) {
}

// TODO: write me. this should be configurable by the user
func writeToCache(topology *Topology, node *ServerNode) {
}

// Close should be called when deleting executor
func (re *RequestExecutor) Close() {
	// TODO: implement me
	panicIf(true, "NYI")
}

type NodeStatus struct {
	_timerPeriod     time.Duration
	_requestExecutor *RequestExecutor
	nodeIndex        int
	node             *ServerNode
	_timer           *time.Timer
}

func NewNodeStatus(requestExecutor *RequestExecutor, nodeIndex int, node *ServerNode) *NodeStatus {
	return &NodeStatus{
		_requestExecutor: requestExecutor,
		nodeIndex:        nodeIndex,
		node:             node,
		_timerPeriod:     time.Millisecond * 100,
	}
}

func (s *NodeStatus) nextTimerPeriod() time.Duration {
	if s._timerPeriod > time.Second*5 {
		return time.Second * 5
	}
	s._timerPeriod = s._timerPeriod + (time.Millisecond * 100)
	return s._timerPeriod
}

func (s *NodeStatus) startTimer() {
	f := func() {
		s.timerCallback()
	}
	s._timer = time.AfterFunc(s._timerPeriod, f)
}

func (s *NodeStatus) updateTimer() {
	// TODO: not sure if Reset
	s._timer.Reset(s.nextTimerPeriod())
}

func (s *NodeStatus) timerCallback() {
	if !s._requestExecutor._disposed {
		s._requestExecutor.checkNodeStatusCallback(s)
	}
}

func (s *NodeStatus) close() {
	if s._timer != nil {
		s._timer.Stop()
		s._timer = nil
	}
}
