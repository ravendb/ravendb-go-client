package ravendb

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

var (
	RequestExecutor_failureCheckOperation *GetStatisticsOperation = NewGetStatisticsOperationWithDebugTag("failure=check")
)

// public static Consumer<HttpRequestBase> requestPostProcessor = null;

const (
	goClientVersion = "4.0.0"
)

// RequestExecutor describes executor of HTTP requests
type RequestExecutor struct {
	_updateDatabaseTopologySemaphore    *Semaphore
	_updateClientConfigurationSemaphore *Semaphore

	_failedNodesTimers sync.Map // *ServerNode => *NodeStatus

	certificate           *KeyStore
	_databaseName         string
	_lastReturnedResponse atomic.Value // atomic to avoid data races

	_updateTopologyTimer *time.Timer
	_nodeSelector        *NodeSelector

	numberOfServerRequests  AtomicInteger
	topologyEtag            int
	clientConfigurationEtag int
	conventions             *DocumentConventions

	_disableTopologyUpdates            bool
	_disableClientConfigurationUpdates bool

	_firstTopologyUpdate *CompletableFuture

	_readBalanceBehavior   ReadBalanceBehavior
	cache                  *HttpCache
	httpClient             *http.Client
	_topologyTakenFromNode *ServerNode

	_lastKnownUrls []string

	mu sync.Mutex

	_disposed bool

	// those are needed to implement ClusterRequestExecutor logic
	isCluster                bool
	clusterTopologySemaphore *Semaphore

	/// TODO: in Java this thread local
	aggressiveCaching *AggressiveCacheOptions
}

func (re *RequestExecutor) getCache() *HttpCache {
	return re.cache
}

func (re *RequestExecutor) getTopology() *Topology {
	if re._nodeSelector != nil {
		return re._nodeSelector.getTopology()
	}
	return nil
}

func (re *RequestExecutor) getTopologyNodes() []*ServerNode {
	if re.getTopology() == nil {
		return nil
	}
	var res []*ServerNode
	nodes := re.getTopology().getNodes()
	for _, n := range nodes {
		res = append(res, n)
	}
	return res
}

func (re *RequestExecutor) getUrl() string {
	if re._nodeSelector == nil {
		return ""
	}

	// TODO: propagate error
	preferredNode, _ := re.getPreferredNode()
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
	return re.conventions
}

func (re *RequestExecutor) getCertificate() *KeyStore {
	return re.certificate
}

var (
	globalHTTPClient *http.Client
)

func getGlobalHTTPClientNoKeepAlive() *http.Client {
	if globalHTTPClient == nil {
		// TODO: certificate

		// based on http.DefaultTransport with DisableKeepAlives set to true
		tr := &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			DialContext: (&net.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 30 * time.Second,
				DualStack: true,
			}).DialContext,
			MaxIdleConns:          100,
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
			DisableKeepAlives: true,
		}

		client := &http.Client{
			Timeout: time.Second * 15,
			Transport: tr,
		}
		globalHTTPClient = client
	}
	return globalHTTPClient
}

func getGlobalHTTPClient() *http.Client {
	if globalHTTPClient == nil {
		// TODO: certificate
		client := &http.Client{
			Timeout: time.Second * 15,
		}
		globalHTTPClient = client
	}
	return globalHTTPClient
}

// NewRequestExecutor creates a new executor
func NewRequestExecutor(databaseName string, certificate *KeyStore, conventions *DocumentConventions, initialUrls []string) *RequestExecutor {
	if conventions == nil {
		conventions = NewDocumentConventions()
	}
	res := &RequestExecutor{
		_updateDatabaseTopologySemaphore:    NewSemaphore(1),
		_updateClientConfigurationSemaphore: NewSemaphore(1),

		cache:                NewHttpCache(),
		_readBalanceBehavior: conventions.getReadBalanceBehavior(),
		_databaseName:        databaseName,
		certificate:          certificate,

		conventions: conventions.clone(),
	}
	res._lastReturnedResponse.Store(time.Now())
	// TODO: create a different client if settings like compression
	// or certificate differ
	//res.httpClient = res.createClient()
	res.httpClient = getGlobalHTTPClient()
	return res
}

func NewClusterRequestExecutor(certificate *KeyStore, conventions *DocumentConventions, initialUrls []string) *RequestExecutor {
	res := NewRequestExecutor("", certificate, conventions, initialUrls)
	res.makeCluster()

	return res
}

// TODO: only used for http cache?
//private string extractThumbprintFromCertificate(KeyStore certificate) {

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

func ClusterRequestExecutor_createForSingleNode(url string, certificate *KeyStore, conventions *DocumentConventions) *RequestExecutor {

	initialUrls := []string{url}
	url = RequestExecutor_validateUrls(initialUrls, certificate)[0]

	if conventions == nil {
		conventions = DocumentConventions_defaultConventions()
	}
	executor := NewClusterRequestExecutor(certificate, conventions, initialUrls)
	executor.makeCluster()

	serverNode := NewServerNode()
	serverNode.setUrl(url)

	topology := NewTopology()
	topology.setEtag(-1)
	topology.setNodes([]*ServerNode{serverNode})

	nodeSelector := NewNodeSelector(topology)

	executor._nodeSelector = nodeSelector
	executor.topologyEtag = -2
	executor._disableClientConfigurationUpdates = true
	executor._disableTopologyUpdates = true

	return executor
}

func (re *RequestExecutor) makeCluster() {
	re.isCluster = true
	re.clusterTopologySemaphore = NewSemaphore(1)
}

func ClusterRequestExecutor_create(initialUrls []string, certificate *KeyStore, conventions *DocumentConventions) *RequestExecutor {
	if conventions == nil {
		conventions = DocumentConventions_defaultConventions()
	}
	executor := NewClusterRequestExecutor(certificate, conventions, initialUrls)
	executor.makeCluster()

	executor._disableClientConfigurationUpdates = true
	executor._firstTopologyUpdate = executor.firstTopologyUpdate(initialUrls)

	return executor
}

func (re *RequestExecutor) clusterUpdateClientConfigurationAsync() *CompletableFuture {
	panicIf(!re.isCluster, "clusterUpdateClientConfigurationAsync() called on non-cluster RequestExecutor")
	return NewCompletableFutureAlreadyCompleted(nil)
}

func (re *RequestExecutor) updateClientConfigurationAsync() *CompletableFuture {
	// Note: in Java this is done via virtual functions
	if re.isCluster {
		return re.clusterUpdateClientConfigurationAsync()
	}

	if re._disposed {
		return NewCompletableFutureAlreadyCompleted(nil)
	}

	future := NewCompletableFuture()
	f := func() {
		var err error

		defer func() {
			if err != nil {
				future.markAsDoneWithError(err)
			} else {
				future.markAsDone(nil)
			}
		}()

		re._updateClientConfigurationSemaphore.acquire()
		defer re._updateClientConfigurationSemaphore.release()

		oldDisableClientConfigurationUpdates := re._disableClientConfigurationUpdates
		re._disableClientConfigurationUpdates = true

		defer func() {
			re._disableClientConfigurationUpdates = oldDisableClientConfigurationUpdates
		}()

		command := NewGetClientConfigurationCommand()
		currentIndexAndNode, err := re.chooseNodeForRequest(command, nil)
		if err != nil {
			return
		}
		err = re.execute(currentIndexAndNode.currentNode, currentIndexAndNode.currentIndex, command, false, nil)
		if err != nil {
			return
		}

		result := command.Result
		if result == nil {
			return
		}

		re.conventions.updateFrom(result.getConfiguration())
		re.clientConfigurationEtag = result.getEtag()

		if re._disposed {
			return
		}
	}

	go f()
	return future
}

func (re *RequestExecutor) updateTopologyAsync(node *ServerNode, timeout int) *CompletableFuture {
	return re.updateTopologyAsyncWithForceUpdate(node, timeout, false)
}

func (re *RequestExecutor) clusterUpdateTopologyAsyncWithForceUpdate(node *ServerNode, timeout int, forceUpdate bool) *CompletableFuture {
	panicIf(!re.isCluster, "clusterUpdateTopologyAsyncWithForceUpdate() called on non-cluster RequestExecutor")

	if re._disposed {
		return NewCompletableFutureAlreadyCompleted(false)
	}

	future := NewCompletableFuture()
	f := func() {
		var err error
		var res bool
		defer func() {
			if err != nil {
				future.markAsDoneWithError(err)
			} else {
				future.markAsDone(res)
			}
			re.clusterTopologySemaphore.release()
		}()

		re.clusterTopologySemaphore.acquire()
		if re._disposed {
			res = false
			return
		}

		command := NewGetClusterTopologyCommand()
		err = re.execute(node, -1, command, false, nil)
		if err != nil {
			return
		}
		results := command.Result
		members := results.getTopology().Members
		var nodes []*ServerNode
		for key, value := range members {
			serverNode := NewServerNode()
			serverNode.setUrl(value)
			serverNode.setClusterTag(key)
			nodes = append(nodes, serverNode)
		}
		newTopology := NewTopology()
		newTopology.setNodes(nodes)

		if re._nodeSelector == nil {
			re._nodeSelector = NewNodeSelector(newTopology)

			if re._readBalanceBehavior == ReadBalanceBehavior_FASTEST_NODE {
				re._nodeSelector.scheduleSpeedTest()
			}
		} else if re._nodeSelector.onUpdateTopology(newTopology, forceUpdate) {
			re.disposeAllFailedNodesTimers()

			if re._readBalanceBehavior == ReadBalanceBehavior_FASTEST_NODE {
				re._nodeSelector.scheduleSpeedTest()
			}
		}
	}

	go f()
	return future
}

func (re *RequestExecutor) updateTopologyAsyncWithForceUpdate(node *ServerNode, timeout int, forceUpdate bool) *CompletableFuture {
	// Note: in Java this is done via virtual functions
	if re.isCluster {
		return re.clusterUpdateTopologyAsyncWithForceUpdate(node, timeout, forceUpdate)
	}
	//fmt.Printf("updateTopologyAsyncWithForceUpdate\n")
	future := NewCompletableFuture()
	f := func() {
		var err error
		var res bool
		defer func() {
			if err != nil {
				future.markAsDoneWithError(err)
			} else {
				future.markAsDone(res)
			}
		}()
		if re._disposed {
			res = false
			return
		}
		re._updateDatabaseTopologySemaphore.acquire()
		defer re._updateDatabaseTopologySemaphore.release()
		command := NewGetDatabaseTopologyCommand()
		err = re.execute(node, 0, command, false, nil)
		if err != nil {
			return
		}
		result := command.Result
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
		re.topologyEtag = re._nodeSelector.getTopology().getEtag()
		res = true
	}

	go f()
	return future
}

func (re *RequestExecutor) disposeAllFailedNodesTimers() {
	f := func(key, val interface{}) bool {
		status := val.(*NodeStatus)
		status.Close()
		return true
	}
	re._failedNodesTimers.Range(f)
	re._failedNodesTimers = sync.Map{}
}

// execute(command) in java
func (re *RequestExecutor) executeCommand(command RavenCommand) error {
	return re.executeCommandWithSessionInfo(command, nil)
}

// execute(command, session) in java
func (re *RequestExecutor) executeCommandWithSessionInfo(command RavenCommand, sessionInfo *SessionInfo) error {
	topologyUpdate := re._firstTopologyUpdate
	if (topologyUpdate != nil && topologyUpdate.isDone()) || re._disableTopologyUpdates {
		currentIndexAndNode, err := re.chooseNodeForRequest(command, sessionInfo)
		if err != nil {
			return err
		}
		return re.execute(currentIndexAndNode.currentNode, currentIndexAndNode.currentIndex, command, true, sessionInfo)
	} else {
		return re.unlikelyExecute(command, topologyUpdate, sessionInfo)
	}
}

func (re *RequestExecutor) chooseNodeForRequest(cmd RavenCommand, sessionInfo *SessionInfo) (*CurrentIndexAndNode, error) {
	if !cmd.GetBase().isReadRequest() {
		return re.getPreferredNode()
	}

	switch re._readBalanceBehavior {
	case ReadBalanceBehavior_NONE:
		return re.getPreferredNode()
	case ReadBalanceBehavior_ROUND_ROBIN:
		sessionID := 0
		if sessionInfo != nil {
			sessionID = sessionInfo.SessionID
		}
		return re.getNodeBySessionId(sessionID)
	case ReadBalanceBehavior_FASTEST_NODE:
		return re.getFastestNode()
	default:
		panicIf(true, "Unknown re._readBalanceBehavior: '%s'", re._readBalanceBehavior)
	}
	return nil, nil
}

func (re *RequestExecutor) unlikelyExecuteInner(command RavenCommand, topologyUpdate *CompletableFuture, sessionInfo *SessionInfo) (*CompletableFuture, error) {

	if topologyUpdate == nil {
		re.mu.Lock()
		if re._firstTopologyUpdate == nil {
			if len(re._lastKnownUrls) == 0 {
				re.mu.Unlock()
				return topologyUpdate, NewIllegalStateException("No known topology and no previously known one, cannot proceed, likely a bug")
			}

			re._firstTopologyUpdate = re.firstTopologyUpdate(re._lastKnownUrls)
		}
		topologyUpdate = re._firstTopologyUpdate
		re.mu.Unlock()
	}

	_, err := topologyUpdate.get()
	return topologyUpdate, err
}

func (re *RequestExecutor) unlikelyExecute(command RavenCommand, topologyUpdate *CompletableFuture, sessionInfo *SessionInfo) error {
	var err error
	topologyUpdate, err = re.unlikelyExecuteInner(command, topologyUpdate, sessionInfo)
	if err != nil {
		re.mu.Lock()
		if re._firstTopologyUpdate == topologyUpdate {
			re._firstTopologyUpdate = nil // next request will raise it
		}
		re.mu.Unlock()
		return err
	}

	currentIndexAndNode, err := re.chooseNodeForRequest(command, sessionInfo)
	if err != nil {
		return err
	}
	err = re.execute(currentIndexAndNode.currentNode, currentIndexAndNode.currentIndex, command, true, sessionInfo)
	return err
}

func (re *RequestExecutor) updateTopologyCallback() {
	last := re._lastReturnedResponse.Load().(time.Time)
	dur := time.Since(last)
	if dur < time.Minute {
		return
	}

	var serverNode *ServerNode

	// TODO: early exist if getPreferredNode() returns an error
	preferredNode, err := re.getPreferredNode()
	if err != nil {
		return
	}
	serverNode = preferredNode.currentNode

	re.updateTopologyAsync(serverNode, 0)
}

type Tuple_String_Error struct {
	S   string
	Err error
}

func (re *RequestExecutor) firstTopologyUpdate(inputUrls []string) *CompletableFuture {
	initialUrls := RequestExecutor_validateUrls(inputUrls, re.certificate)

	future := NewCompletableFuture()
	var list []*Tuple_String_Error
	f := func() {
		var err error
		defer func() {
			if err != nil {
				future.markAsDoneWithError(err)
			} else {
				future.markAsDone(nil)
			}
		}()

		for _, url := range initialUrls {
			{
				serverNode := NewServerNode()
				serverNode.setUrl(url)
				serverNode.setDatabase(re._databaseName)

				res := re.updateTopologyAsync(serverNode, math.MaxInt32)
				_, err = res.get()
				if err == nil {
					re.initializeUpdateTopologyTimer()
					re._topologyTakenFromNode = serverNode
					return
				}
			}

			if _, ok := (err).(*DatabaseDoesNotExistException); ok {
				// Will happen on all node in the cluster,
				// so errors immediately
				re._lastKnownUrls = initialUrls
				return
			}

			if len(initialUrls) == 0 {
				re._lastKnownUrls = initialUrls
				err = NewIllegalStateException("Cannot get topology from server: %s", url)
				return
			}
			list = append(list, &Tuple_String_Error{url, err})
		}
		topology := NewTopology()
		topology.setEtag(re.topologyEtag)
		topologyNodes := re.getTopologyNodes()
		if len(topologyNodes) == 0 {
			for _, uri := range initialUrls {
				serverNode := NewServerNode()
				serverNode.setUrl(uri)
				serverNode.setDatabase(re._databaseName)
				serverNode.setClusterTag("!")
				topologyNodes = append(topologyNodes, serverNode)
			}
		}
		topology.setNodes(topologyNodes)
		re._nodeSelector = NewNodeSelector(topology)
		if len(initialUrls) > 0 {
			re.initializeUpdateTopologyTimer()
			return
		}
		re._lastKnownUrls = initialUrls

		var a []string
		for _, el := range list {
			first := el.S
			second := el.Err
			s := first + " -> " + second.Error()
			a = append(a, s)
		}
		details := strings.Join(a, ", ")
		err = re.throwExceptions(details)
		return
	}
	go f()
	return future
}

func (re *RequestExecutor) throwExceptions(details string) error {
	err := NewIllegalStateException("Failed to retrieve database topology from all known nodes \n" + details)
	return err
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
		re.mu.Lock()
		re._updateTopologyTimer = nil
		re.mu.Unlock()
		re.initializeUpdateTopologyTimer()
	}
	re._updateTopologyTimer = time.AfterFunc(time.Minute, f)
}

func isNetworkTimeoutError(err error) bool {
	// TODO: implement me
	// can test it by setting very low timeout in http.Client
	return false
}

func (re *RequestExecutor) execute(chosenNode *ServerNode, nodeIndex int, command RavenCommand, shouldRetry bool, sessionInfo *SessionInfo) error {
	//fmt.Printf("RequestExecutor.execute cmd: %#v\n", command)
	request, err := re.CreateRequest(chosenNode, command)
	if err != nil {
		return err
	}
	// TODO: caching

	urlRef := request.URL.String()

	if !re._disableClientConfigurationUpdates {
		etag := `"` + strconv.Itoa(re.clientConfigurationEtag) + `"`
		request.Header.Set(Constants_Headers_CLIENT_CONFIGURATION_ETAG, etag)
	}

	if !re._disableTopologyUpdates {
		etag := `"` + strconv.Itoa(re.topologyEtag) + `"`
		request.Header.Set(Constants_Headers_TOPOLOGY_ETAG, etag)
	}

	//sp := time.Now()
	responseDispose := ResponseDisposeHandling_AUTOMATIC
	var response *http.Response
	re.numberOfServerRequests.incrementAndGet()
	if re.shouldExecuteOnAll(chosenNode, command) {
		response, err = re.executeOnAllToFigureOutTheFastest(chosenNode, command)
	} else {
		response, err = command.GetBase().Send(re.httpClient, request)
	}

	if err != nil {
		if !shouldRetry && isNetworkTimeoutError(err) {
			return err
		}
		// Note: Java here re-throws if err is IOException and !shouldRetry
		// but for us that propagates the wrong error to RequestExecutorTest_failsWhenServerIsOffline
		urlRef := request.URL.String()
		if !re.handleServerDown(urlRef, chosenNode, nodeIndex, command, request, response, err, sessionInfo) {
			return re.throwFailedToContactAllNodes(command, request, err, nil)
		}
		return nil
	}

	command.GetBase().statusCode = response.StatusCode

	refreshTopology := HttpExtensions_getBooleanHeader(response, Constants_Headers_REFRESH_TOPOLOGY)
	refreshClientConfiguration := HttpExtensions_getBooleanHeader(response, Constants_Headers_REFRESH_CLIENT_CONFIGURATION)

	// TODO: handle not modified

	if response.StatusCode >= 400 {
		ok, err := re.handleUnsuccessfulResponse(chosenNode, nodeIndex, command, request, response, urlRef, sessionInfo, shouldRetry)
		if err != nil {
			return err
		}

		if !ok {
			dbMissingHeader := response.Header.Get("Database-Missing")
			if dbMissingHeader != "" {
				return NewDatabaseDoesNotExistException(dbMissingHeader)
			}

			if len(command.GetBase().GetFailedNodes()) == 0 {
				return NewIllegalStateException("Received unsuccessful response and couldn't recover from it. Also, no record of exceptions per failed nodes. This is weird and should not happen.")
			}

			if len(command.GetBase().GetFailedNodes()) == 1 {
				// return first error
				failedNodes := command.GetBase().GetFailedNodes()
				for _, err := range failedNodes {
					panicIf(err == nil, "err is nil")
					return err
				}
			}

			return NewAllTopologyNodesDownException("Received unsuccessful response from all servers and couldn't recover from it.")
		}
		return nil // we either handled this already in the unsuccessful response or we are throwing
	}

	responseDispose, err = processCommandResponse(command, re.cache, response, urlRef)
	re._lastReturnedResponse.Store(time.Now())
	if err != nil {
		return err
	}

	if responseDispose == ResponseDisposeHandling_AUTOMATIC {
		// TODO: not sure if it translates
		response.Body.Close()
		//IOUtils.closeQuietly(response)
	}

	if refreshTopology || refreshClientConfiguration {

		serverNode := NewServerNode()
		serverNode.setUrl(chosenNode.getUrl())
		serverNode.setDatabase(re._databaseName)

		var topologyTask *CompletableFuture
		if refreshTopology {
			topologyTask = re.updateTopologyAsync(serverNode, 0)
		} else {
			topologyTask = NewCompletableFutureAlreadyCompleted(false)
		}
		var clientConfiguration *CompletableFuture
		if refreshClientConfiguration {
			clientConfiguration = re.updateClientConfigurationAsync()
		} else {
			clientConfiguration = NewCompletableFutureAlreadyCompleted(nil)
		}
		_, err1 := topologyTask.get()
		_, err2 := clientConfiguration.get()
		if err1 != nil {
			return err1
		}
		if err2 != nil {
			return err2
		}
	}
	return nil
}

func (re *RequestExecutor) throwFailedToContactAllNodes(command RavenCommand, request *http.Request, e error, timeoutException error) error {
	// TODO: after transition to RavenCommand as interface, this will
	// be command name via type
	commandName := "command"
	message := "Tried to send " + commandName + " request via " + request.Method + " " + request.URL.String() + " to all configured nodes in the topology, " +
		"all of them seem to be down or not responding. I've tried to access the following nodes: "

	var urls []string
	if re._nodeSelector != nil {
		for _, node := range re._nodeSelector.getTopology().getNodes() {
			url := node.getUrl()
			urls = append(urls, url)
		}
	}
	message += strings.Join(urls, ", ")

	if re._topologyTakenFromNode != nil {
		nodes := re._nodeSelector.getTopology().getNodes()
		var a []string
		for _, n := range nodes {
			s := "( url: " + n.getUrl() + ", clusterTag: " + n.getClusterTag() + ", serverRole: " + n.getServerRole() + ")"
			a = append(a, s)
		}
		nodesStr := strings.Join(a, ", ")

		message += "\nI was able to fetch " + re._topologyTakenFromNode.getDatabase() + " topology from " + re._topologyTakenFromNode.getUrl() + ".\n" + "Fetched topology: " + nodesStr
	}

	return NewAllTopologyNodesDownException("%s", message)
}

func (re *RequestExecutor) inSpeedTestPhase() bool {
	return (re._nodeSelector != nil) && re._nodeSelector.inSpeedTestPhase()
}

func (re *RequestExecutor) shouldExecuteOnAll(chosenNode *ServerNode, command RavenCommand) bool {
	multipleNodes := (re._nodeSelector != nil) && (len(re._nodeSelector.getTopology().getNodes()) > 1)

	return re._readBalanceBehavior == ReadBalanceBehavior_FASTEST_NODE &&
		re._nodeSelector != nil &&
		re._nodeSelector.inSpeedTestPhase() &&
		multipleNodes &&
		command.GetBase().isReadRequest() &&
		command.GetBase().GetResponseType() == RavenCommandResponseType_OBJECT &&
		chosenNode != nil
}

func (re *RequestExecutor) executeOnAllToFigureOutTheFastest(chosenNode *ServerNode, command RavenCommand) (*http.Response, error) {
	panic("NYI")
	return nil, nil
}

func (re *RequestExecutor) getFromCache(command RavenCommand, url string, cachedChangeVector *string, cachedValue *string) *ReleaseCacheItem {
	panic("NYI")
	return nil
}

func (re *RequestExecutor) CreateRequest(node *ServerNode, command RavenCommand) (*http.Request, error) {
	request, err := command.CreateRequest(node)
	if err != nil {
		return nil, err
	}
	request.Header.Set("Raven-Client-Version", goClientVersion)
	return request, err
}

func (re *RequestExecutor) handleUnsuccessfulResponse(chosenNode *ServerNode, nodeIndex int, command RavenCommand, request *http.Request, response *http.Response, url string, sessionInfo *SessionInfo, shouldRetry bool) (bool, error) {
	var err error
	switch response.StatusCode {
	case http.StatusNotFound:
		re.cache.setNotFound(url)
		switch command.GetBase().GetResponseType() {
		case RavenCommandResponseType_EMPTY:
			return true, nil
		case RavenCommandResponseType_OBJECT:
			command.SetResponse(nil, false)
			break
		default:
			command.SetResponseRaw(response, nil)
			break
		}
		return true, nil
	case http.StatusForbidden:
		err = NewAuthorizationException("Forbidden access to " + chosenNode.getDatabase() + "@" + chosenNode.getUrl() + ", " + request.Method + " " + request.URL.String())
	case http.StatusGone: // request not relevant for the chosen node - the database has been moved to a different one
		if !shouldRetry {
			return false, nil
		}

		updateFuture := re.updateTopologyAsyncWithForceUpdate(chosenNode, int(math.MaxInt32), true)
		_, err := updateFuture.get()
		if err != nil {
			return false, err
		}

		currentIndexAndNode, err := re.chooseNodeForRequest(command, sessionInfo)
		if err != nil {
			return false, err
		}
		err = re.execute(currentIndexAndNode.currentNode, currentIndexAndNode.currentIndex, command, false, sessionInfo)
		return false, err
	case http.StatusGatewayTimeout, http.StatusRequestTimeout,
		http.StatusBadGateway, http.StatusServiceUnavailable:
		ok := re.handleServerDown(url, chosenNode, nodeIndex, command, request, response, nil, sessionInfo)
		return ok, nil
	case http.StatusConflict:
		err = RequestExecutor_handleConflict(response)
		break
	default:
		if gLogHTTP {
			dumpRequestAndResponse(request, response, err)
		}
		command.GetBase().OnResponseFailure(response)
		err = ExceptionDispatcher_throwException(response)
		break
	}
	return false, err
}

func RequestExecutor_handleConflict(response *http.Response) error {
	//fmt.Printf("RequestExecutor_handleConflict\n")
	return ExceptionDispatcher_throwException(response)
}

//     public static InputStream readAsStream(CloseableHttpResponse response) throws IOException {

func (re *RequestExecutor) handleServerDown(url string, chosenNode *ServerNode, nodeIndex int, command RavenCommand, request *http.Request, response *http.Response, e error, sessionInfo *SessionInfo) bool {
	if command.GetBase().GetFailedNodes() == nil {
		command.GetBase().SetFailedNodes(make(map[*ServerNode]error))
	}

	re.addFailedResponseToCommand(chosenNode, command, request, response, e)

	// TODO: Java checks for nodeIndex != null, don't know how that could happen
	// TODO: change to false
	if true && nodeIndex == 0 {
		//We executed request over a node not in the topology. This means no failover...
		return false
	}

	re.spawnHealthChecks(chosenNode, nodeIndex)

	if re._nodeSelector == nil {
		return false
	}

	re._nodeSelector.onFailedRequest(nodeIndex)

	currentIndexAndNode, err := re.getPreferredNode()
	if err != nil {
		return false
	}

	if _, ok := command.GetBase().GetFailedNodes()[currentIndexAndNode.currentNode]; ok {
		return false //we tried all the nodes...nothing left to do
	}

	// TODO: propagate error?
	re.execute(currentIndexAndNode.currentNode, currentIndexAndNode.currentIndex, command, false, sessionInfo)

	return true
}

func (re *RequestExecutor) spawnHealthChecks(chosenNode *ServerNode, nodeIndex int) {
	nodeStatus := NewNodeStatus(re, nodeIndex, chosenNode)
	_, loaded := re._failedNodesTimers.LoadOrStore(chosenNode, nodeStatus)
	if !loaded {
		nodeStatus.startTimer()
	}
}

func (re *RequestExecutor) checkNodeStatusCallback(nodeStatus *NodeStatus) {
	nodesCopy := re.getTopologyNodes()
	idx := nodeStatus.nodeIndex
	// TODO: idx < 0 probably shouldn't happen but it's the only cause of
	// https://travis-ci.org/kjk/ravendb-go-client/builds/404760557
	// that I can think of
	if idx < 0 || idx >= len(nodesCopy) {
		return // topology index changed / removed
	}

	serverNode := nodesCopy[idx]
	if serverNode != nodeStatus.node {
		return // topology changed, nothing to check
	}

	err := re.performHealthCheck(serverNode, idx)
	if err != nil {
		// TODO: logging
		_, ok := re._failedNodesTimers.Load(nodeStatus.node)
		if !ok {
			nodeStatus.updateTimer()
		}

		return // will wait for the next timer call
	}
	statusI, ok := re._failedNodesTimers.Load(nodeStatus.node)
	if ok {
		status := statusI.(*NodeStatus)
		re._failedNodesTimers.Delete(nodeStatus.node)
		status.Close()
	}

	if re._nodeSelector != nil {
		re._nodeSelector.restoreNodeIndex(idx)
	}
}

func (re *RequestExecutor) clusterPerformHealthCheck(serverNode *ServerNode, nodeIndex int) error {
	panicIf(!re.isCluster, "clusterPerformHealthCheck() called on non-cluster RequestExector")
	command := NewGetTcpInfoCommand("health-check")
	return re.execute(serverNode, nodeIndex, command, false, nil)
}

func (re *RequestExecutor) performHealthCheck(serverNode *ServerNode, nodeIndex int) error {
	// Note: in Java this is done via virtual functions
	if re.isCluster {
		return re.clusterPerformHealthCheck(serverNode, nodeIndex)
	}
	command := RequestExecutor_failureCheckOperation.getCommand(re.conventions)
	return re.execute(serverNode, nodeIndex, command, false, nil)
}

// TODO: this is static
func (re *RequestExecutor) addFailedResponseToCommand(chosenNode *ServerNode, command RavenCommand, request *http.Request, response *http.Response, e error) {
	failedNodes := command.GetBase().GetFailedNodes()

	if response != nil && response.Body != nil {
		responseJson, err := ioutil.ReadAll(response.Body)
		if err == nil {
			var schema ExceptionSchema
			json.Unmarshal(responseJson, &schema)
			readException := ExceptionDispatcher_get(&schema, response.StatusCode)
			failedNodes[chosenNode] = readException
		} else {
			exceptionSchema := NewExceptionSchema()
			exceptionSchema.setUrl(request.URL.String())
			exceptionSchema.setMessage("Get unrecognized response from the server")
			exceptionSchema.setError(string(responseJson))
			exceptionSchema.setType("Unparsable Server Response")
			exceptionToUse := ExceptionDispatcher_get(exceptionSchema, response.StatusCode)

			failedNodes[chosenNode] = exceptionToUse
		}
	}

	// this would be connections that didn't have response, such as "couldn't connect to remote server"
	if e == nil {
		// TODO: not sure if this is needed or a sign of a buf
		e = NewRavenException("")
	}
	exceptionSchema := NewExceptionSchema()
	exceptionSchema.setUrl(request.URL.String())
	exceptionSchema.setMessage(e.Error())
	exceptionSchema.setError(e.Error())
	errorType := fmt.Sprintf("%T", e)
	exceptionSchema.setType(errorType)

	exceptionToUse := ExceptionDispatcher_get(exceptionSchema, http.StatusInternalServerError)
	failedNodes[chosenNode] = exceptionToUse
}

// TODO: write me. this should be configurable by the user
func (re *RequestExecutor) tryLoadFromCache(url string) {
}

// TODO: write me. this should be configurable by the user
func writeToCache(topology *Topology, node *ServerNode) {
}

// Close should be called when deleting executor
func (re *RequestExecutor) Close() {
	if re._disposed {
		return
	}

	if re.isCluster {
		// make sure that a potentially pending updateTopologyAsync() has
		// finished
		re.clusterTopologySemaphore.acquire()
	}

	re._disposed = true
	//re.cache.close()

	re.mu.Lock()
	defer re.mu.Unlock()

	if re._updateTopologyTimer != nil {
		re._updateTopologyTimer.Stop()
		re._updateTopologyTimer = nil
	}
	re.disposeAllFailedNodesTimers()
}

var (
	envProxyURL string
)

func buildProxyURL(req *http.Request) (*url.URL, error) {
	proxy := envProxyURL
	proxyURL, err := url.Parse(proxy)
	if err != nil ||
		(proxyURL.Scheme != "http" &&
			proxyURL.Scheme != "https" &&
			proxyURL.Scheme != "socks5") {
		// proxy was bogus. Try prepending "http://" to it and
		// see if that parses correctly. If not, we fall
		// through and complain about the original one.
		if proxyURL, err := url.Parse("http://" + proxy); err == nil {
			return proxyURL, nil
		}

	}
	if err != nil {
		return nil, fmt.Errorf("invalid proxy address %q: %v", proxy, err)
	}
	return proxyURL, nil
}

var proxyTransport http.RoundTripper = &http.Transport{
	Proxy: buildProxyURL,
	DialContext: (&net.Dialer{
		Timeout:   30 * time.Second,
		KeepAlive: 30 * time.Second,
		DualStack: true,
	}).DialContext,
	MaxIdleConns:          100,
	IdleConnTimeout:       90 * time.Second,
	TLSHandshakeTimeout:   10 * time.Second,
	ExpectContinueTimeout: 1 * time.Second,
}

func (re *RequestExecutor) createClient() *http.Client {
	// TODO: certificate
	client := &http.Client{
		Timeout: time.Second * 5,
	}
	return client
}

func (re *RequestExecutor) getPreferredNode() (*CurrentIndexAndNode, error) {
	re.ensureNodeSelector()

	return re._nodeSelector.getPreferredNode()
}

func (re *RequestExecutor) getNodeBySessionId(sessionId int) (*CurrentIndexAndNode, error) {
	re.ensureNodeSelector()

	return re._nodeSelector.getNodeBySessionId(sessionId)
}

func (re *RequestExecutor) getFastestNode() (*CurrentIndexAndNode, error) {
	re.ensureNodeSelector()

	return re._nodeSelector.getFastestNode()
}

func (re *RequestExecutor) ensureNodeSelector() error {
	if re._firstTopologyUpdate != nil && !re._firstTopologyUpdate.isDone() {
		_, err := re._firstTopologyUpdate.get()
		if err != nil {
			return err
		}
	}

	if re._nodeSelector == nil {
		topology := NewTopology()

		topology.setNodes(re.getTopologyNodes())
		topology.setEtag(re.topologyEtag)

		re._nodeSelector = NewNodeSelector(topology)
	}
	return nil
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

func (s *NodeStatus) Close() {
	if s._timer != nil {
		s._timer.Stop()
		s._timer = nil
	}
}
