package ravendb

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"math"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

var (
	// Note: unlike Java GetStatisticsOperation.GetCommand() is not thread safe
	// requestExecutorFailureCheckOperation *GetStatisticsOperation = NewGetStatisticsOperation("failure=check")

	// HTTPClientPostProcessor allows to tweak http client after it has been created
	// this allows replacing Transport with a custom transport that does logging,
	// proxying or tweaks each http request
	HTTPClientPostProcessor func(*http.Client)

	// if true, adds lots of logging to track bugs in request executor
	DebugLogRequestExecutor bool = false
)

const (
	goClientVersion = "4.0.0"
)

func redbg(format string, args ...interface{}) {
	if DebugLogRequestExecutor {
		fmt.Printf(format, args...)
	}
}

// Note: for simplicity ClusterRequestExecutor logic is implemented in RequestExecutor
// because Go doesn't support inheritance
type ClusterRequestExecutor = RequestExecutor

// RequestExecutor describes executor of HTTP requests
type RequestExecutor struct {
	updateDatabaseTopologySemaphore    *Semaphore
	updateClientConfigurationSemaphore *Semaphore

	failedNodesTimers sync.Map // *ServerNode => *NodeStatus

	Certificate          *tls.Certificate
	TrustStore           *x509.Certificate
	databaseName         string
	lastReturnedResponse atomic.Value // atomic to avoid data races

	updateTopologyTimer *time.Timer
	nodeSelector        atomic.Value // atomic to avoid data races

	NumberOfServerRequests  atomicInteger
	TopologyEtag            int64
	ClientConfigurationEtag int64
	conventions             *DocumentConventions

	disableTopologyUpdates            bool
	disableClientConfigurationUpdates bool

	firstTopologyUpdateFuture *completableFuture

	readBalanceBehavior ReadBalanceBehavior
	// TODO: mulit-threaded access, protect
	Cache                 *httpCache
	httpClient            *http.Client
	topologyTakenFromNode *ServerNode

	lastKnownUrls []string

	mu sync.Mutex

	disposed int32 // atomic

	// those are needed to implement ClusterRequestExecutor logic
	isCluster                bool
	clusterTopologySemaphore *Semaphore

	/// Note: in Java this is thread local but Go doesn't have equivalent
	// of thread local data
	aggressiveCaching *AggressiveCacheOptions
}

func (re *RequestExecutor) getFailedNodeTimer(n *ServerNode) *NodeStatus {
	v, ok := re.failedNodesTimers.Load(n)
	if !ok {
		return nil
	}
	return v.(*NodeStatus)
}

func (re *RequestExecutor) getNodeSelector() *NodeSelector {
	return re.nodeSelector.Load().(*NodeSelector)
}

func (re *RequestExecutor) setNodeSelector(s *NodeSelector) {
	re.nodeSelector.Store(s)
}

func (re *RequestExecutor) markDisposed() {
	atomic.StoreInt32(&re.disposed, 1)
}

func (re *RequestExecutor) isDisposed() bool {
	v := atomic.LoadInt32(&re.disposed)
	return v > 0
}

func (re *RequestExecutor) GetTopology() *Topology {
	nodeSelector := re.getNodeSelector()
	if nodeSelector != nil {
		return nodeSelector.getTopology()
	}
	return nil
}

// GetTopologyNodes returns a copy of topology nodes
func (re *RequestExecutor) GetTopologyNodes() []*ServerNode {
	t := re.GetTopology()
	if t == nil || len(t.Nodes) == 0 {
		return nil
	}
	return append([]*ServerNode{}, t.Nodes...)
}

// GetURL returns an URL
func (re *RequestExecutor) GetURL() (string, error) {
	if re.getNodeSelector() == nil {
		return "", nil
	}

	preferredNode, err := re.getPreferredNode()
	if err != nil {
		return "", err
	}
	if preferredNode != nil {
		return preferredNode.currentNode.URL, nil
	}
	return "", nil
}

func (re *RequestExecutor) GetConventions() *DocumentConventions {
	return re.conventions
}

// NewRequestExecutor creates a new executor
func NewRequestExecutor(databaseName string, certificate *tls.Certificate, trustStore *x509.Certificate, conventions *DocumentConventions, initialUrls []string) *RequestExecutor {
	if conventions == nil {
		conventions = NewDocumentConventions()
	}
	redbg("NewRequestExecutor: db: %s, urls: %v, read balance: %s\n", databaseName, initialUrls, conventions.ReadBalanceBehavior)
	res := &RequestExecutor{
		updateDatabaseTopologySemaphore:    NewSemaphore(1),
		updateClientConfigurationSemaphore: NewSemaphore(1),

		Cache:               newHttpCache(conventions.getMaxHttpCacheSize()),
		readBalanceBehavior: conventions.ReadBalanceBehavior,
		databaseName:        databaseName,
		Certificate:         certificate,
		TrustStore:          trustStore,

		conventions: conventions.Clone(),
	}
	res.lastReturnedResponse.Store(time.Now())
	res.setNodeSelector(nil)
	// TODO: handle an error
	// TODO: java globally caches http clients
	res.httpClient, _ = res.createClient()
	return res
}

// GetHTTPClient returns http client for sending the requests
func (re *RequestExecutor) GetHTTPClient() (*http.Client, error) {
	if re.httpClient != nil {
		return re.httpClient, nil
	}
	c, err := re.createClient()
	if err != nil {
		return nil, err
	}
	re.httpClient = c
	return re.httpClient, nil
}
func NewClusterRequestExecutor(certificate *tls.Certificate, trustStore *x509.Certificate, conventions *DocumentConventions, initialUrls []string) *RequestExecutor {
	res := NewRequestExecutor("", certificate, trustStore, conventions, initialUrls)
	res.MakeCluster()

	return res
}

// TODO: only used for http cache?
//private string extractThumbprintFromCertificate(KeyStore certificate) {

func RequestExecutorCreate(initialUrls []string, databaseName string, certificate *tls.Certificate, trustStore *x509.Certificate, conventions *DocumentConventions) *RequestExecutor {
	re := NewRequestExecutor(databaseName, certificate, trustStore, conventions, initialUrls)
	re.mu.Lock()
	re.firstTopologyUpdateFuture = re.firstTopologyUpdate(initialUrls)
	re.mu.Unlock()
	return re
}

func RequestExecutorCreateForSingleNodeWithConfigurationUpdates(url string, databaseName string, certificate *tls.Certificate, trustStore *x509.Certificate, conventions *DocumentConventions) *RequestExecutor {
	executor := RequestExecutorCreateForSingleNodeWithoutConfigurationUpdates(url, databaseName, certificate, trustStore, conventions)
	executor.disableClientConfigurationUpdates = false
	return executor
}

func RequestExecutorCreateForSingleNodeWithoutConfigurationUpdates(url string, databaseName string, certificate *tls.Certificate, trustStore *x509.Certificate, conventions *DocumentConventions) *RequestExecutor {
	initialUrls := requestExecutorValidateUrls([]string{url}, certificate)
	executor := NewRequestExecutor(databaseName, certificate, trustStore, conventions, initialUrls)

	topology := &Topology{
		Etag: -1,
	}

	serverNode := NewServerNode()
	serverNode.Database = databaseName
	serverNode.URL = initialUrls[0]
	// TODO: is Collections.singletonList in Java code subtly significant?
	topology.Nodes = []*ServerNode{serverNode}

	executor.setNodeSelector(NewNodeSelector(topology))
	executor.TopologyEtag = -2
	executor.disableTopologyUpdates = true
	executor.disableClientConfigurationUpdates = true

	return executor
}

func ClusterRequestExecutorCreateForSingleNode(url string, certificate *tls.Certificate, trustStore *x509.Certificate, conventions *DocumentConventions) *RequestExecutor {

	initialUrls := []string{url}
	url = requestExecutorValidateUrls(initialUrls, certificate)[0]

	if conventions == nil {
		conventions = getDefaultConventions()
	}
	executor := NewClusterRequestExecutor(certificate, trustStore, conventions, initialUrls)
	executor.MakeCluster()

	serverNode := NewServerNode()
	serverNode.URL = url

	topology := &Topology{
		Etag:  -1,
		Nodes: []*ServerNode{serverNode},
	}

	nodeSelector := NewNodeSelector(topology)

	executor.setNodeSelector(nodeSelector)
	executor.TopologyEtag = -2
	executor.disableClientConfigurationUpdates = true
	executor.disableTopologyUpdates = true

	return executor
}

func (re *RequestExecutor) MakeCluster() {
	re.isCluster = true
	re.clusterTopologySemaphore = NewSemaphore(1)
}

func ClusterRequestExecutorCreate(initialUrls []string, certificate *tls.Certificate, trustStore *x509.Certificate, conventions *DocumentConventions) *RequestExecutor {
	if conventions == nil {
		conventions = getDefaultConventions()
	}
	executor := NewClusterRequestExecutor(certificate, trustStore, conventions, initialUrls)
	executor.MakeCluster()

	executor.disableClientConfigurationUpdates = true
	executor.mu.Lock()
	executor.firstTopologyUpdateFuture = executor.firstTopologyUpdate(initialUrls)
	executor.mu.Unlock()

	return executor
}

func (re *RequestExecutor) clusterUpdateClientConfigurationAsync() *completableFuture {
	panicIf(!re.isCluster, "clusterUpdateClientConfigurationAsync() called on non-cluster RequestExecutor")
	return newCompletableFutureAlreadyCompleted(nil)
}

func (re *RequestExecutor) updateClientConfigurationAsync() *completableFuture {
	// Note: in Java this is done via virtual functions
	if re.isCluster {
		return re.clusterUpdateClientConfigurationAsync()
	}

	if re.isDisposed() {
		return newCompletableFutureAlreadyCompleted(nil)
	}

	future := newCompletableFuture()
	f := func() {
		var err error

		defer func() {
			if err != nil {
				future.completeWithError(err)
			} else {
				future.complete(nil)
			}
		}()

		re.updateClientConfigurationSemaphore.acquire()
		defer re.updateClientConfigurationSemaphore.release()

		oldDisableClientConfigurationUpdates := re.disableClientConfigurationUpdates
		re.disableClientConfigurationUpdates = true

		defer func() {
			re.disableClientConfigurationUpdates = oldDisableClientConfigurationUpdates
		}()

		command := NewGetClientConfigurationCommand()
		currentIndexAndNode, err := re.chooseNodeForRequest(command, nil)
		if err != nil {
			return
		}
		err = re.Execute(currentIndexAndNode.currentNode, currentIndexAndNode.currentIndex, command, false, nil)
		if err != nil {
			return
		}

		result := command.Result
		if result == nil {
			return
		}

		re.conventions.UpdateFrom(result.Configuration)
		re.ClientConfigurationEtag = result.Etag

		if re.isDisposed() {
			return
		}
	}

	go f()
	return future
}

func (re *RequestExecutor) UpdateTopologyAsync(node *ServerNode, timeout int) chan *ClusterUpdateAsyncResult {
	return re.updateTopologyAsyncWithForceUpdate(node, timeout, false)
}

type ClusterUpdateAsyncResult struct {
	Ok  bool
	Err error
}

func (re *RequestExecutor) clusterUpdateTopologyAsyncWithForceUpdate(node *ServerNode, timeout int, forceUpdate bool) chan *ClusterUpdateAsyncResult {
	panicIf(!re.isCluster, "clusterUpdateTopologyAsyncWithForceUpdate() called on non-cluster RequestExecutor")

	future := make(chan *ClusterUpdateAsyncResult, 1)
	if re.isDisposed() {
		future <- &ClusterUpdateAsyncResult{Ok: false}
		close(future)
		return future
	}

	f := func() {
		var err error
		var res bool
		defer func() {
			if err != nil && !re.isDisposed() {
				err = nil
			}

			if err != nil {
				future <- &ClusterUpdateAsyncResult{Err: err}
			} else {
				future <- &ClusterUpdateAsyncResult{Ok: res}
			}
			close(future)
			re.clusterTopologySemaphore.release()
		}()

		re.clusterTopologySemaphore.acquire()
		if re.isDisposed() {
			res = false
			return
		}

		command := NewGetClusterTopologyCommand()
		err = re.Execute(node, -1, command, false, nil)
		if err != nil {
			return
		}
		results := command.Result
		members := results.Topology.Members
		var nodes []*ServerNode
		for key, value := range members {
			serverNode := NewServerNode()
			serverNode.URL = value
			serverNode.ClusterTag = key
			nodes = append(nodes, serverNode)
		}
		newTopology := &Topology{
			Nodes: nodes,
		}

		nodeSelector := re.getNodeSelector()
		if nodeSelector == nil {
			nodeSelector = NewNodeSelector(newTopology)
			re.setNodeSelector(nodeSelector)

			if re.readBalanceBehavior == ReadBalanceBehaviorFastestNode {
				nodeSelector.scheduleSpeedTest()
			}
		} else if nodeSelector.onUpdateTopology(newTopology, forceUpdate) {
			re.disposeAllFailedNodesTimers()

			if re.readBalanceBehavior == ReadBalanceBehaviorFastestNode {
				nodeSelector.scheduleSpeedTest()
			}
		}
	}

	go f()
	return future
}

func (re *RequestExecutor) updateTopologyAsyncWithForceUpdate(node *ServerNode, timeout int, forceUpdate bool) chan *ClusterUpdateAsyncResult {
	// Note: in Java this is done via virtual functions
	if re.isCluster {
		return re.clusterUpdateTopologyAsyncWithForceUpdate(node, timeout, forceUpdate)
	}
	future := make(chan *ClusterUpdateAsyncResult, 1)
	f := func() {
		var err error
		var res bool
		defer func() {
			result := &ClusterUpdateAsyncResult{
				Ok:  res,
				Err: err,
			}
			future <- result
			close(future)
		}()

		if re.isDisposed() {
			res = false
			return
		}
		re.updateDatabaseTopologySemaphore.acquire()
		defer re.updateDatabaseTopologySemaphore.release()
		command := NewGetDatabaseTopologyCommand()
		err = re.Execute(node, 0, command, false, nil)
		if err != nil {
			return
		}
		result := command.Result
		nodeSelector := re.getNodeSelector()
		if nodeSelector == nil {
			nodeSelector = NewNodeSelector(result)
			re.setNodeSelector(nodeSelector)
			if re.readBalanceBehavior == ReadBalanceBehaviorFastestNode {
				nodeSelector.scheduleSpeedTest()
			}
		} else if nodeSelector.onUpdateTopology(result, forceUpdate) {
			re.disposeAllFailedNodesTimers()
			if re.readBalanceBehavior == ReadBalanceBehaviorFastestNode {
				nodeSelector.scheduleSpeedTest()
			}
		}
		re.TopologyEtag = nodeSelector.getTopology().Etag
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
	re.failedNodesTimers.Range(f)
	re.failedNodesTimers = sync.Map{}
}

// sessionInfo can be nil
func (re *RequestExecutor) ExecuteCommand(command RavenCommand, sessionInfo *SessionInfo) error {
	redbg("RequestExector.ExecuteCommand: %T\n", command)
	topologyUpdate := re.firstTopologyUpdateFuture
	isDone := topologyUpdate != nil && topologyUpdate.IsDone() && !topologyUpdate.IsCompletedExceptionally() && !topologyUpdate.isCancelled()
	if isDone || re.disableTopologyUpdates {
		currentIndexAndNode, err := re.chooseNodeForRequest(command, sessionInfo)
		if err != nil {
			return err
		}
		return re.Execute(currentIndexAndNode.currentNode, currentIndexAndNode.currentIndex, command, true, sessionInfo)
	} else {
		return re.unlikelyExecute(command, topologyUpdate, sessionInfo)
	}
}

func (re *RequestExecutor) chooseNodeForRequest(cmd RavenCommand, sessionInfo *SessionInfo) (*CurrentIndexAndNode, error) {
	if !cmd.getBase().IsReadRequest {
		return re.getPreferredNode()
	}

	switch re.readBalanceBehavior {
	case ReadBalanceBehaviorNone:
		return re.getPreferredNode()
	case ReadBalanceBehaviorRoundRobin:
		sessionID := 0
		if sessionInfo != nil {
			sessionID = sessionInfo.SessionID
		}
		return re.getNodeBySessionID(sessionID)
	case ReadBalanceBehaviorFastestNode:
		return re.getFastestNode()
	default:
		panicIf(true, "Unknown re.ReadBalanceBehavior: '%s'", re.readBalanceBehavior)
	}
	return nil, nil
}

func (re *RequestExecutor) unlikelyExecuteInner(command RavenCommand, topologyUpdate *completableFuture, sessionInfo *SessionInfo) (*completableFuture, error) {

	if topologyUpdate == nil {
		re.mu.Lock()
		if re.firstTopologyUpdateFuture == nil {
			if len(re.lastKnownUrls) == 0 {
				re.mu.Unlock()
				return nil, newIllegalStateError("No known topology and no previously known one, cannot proceed, likely a bug")
			}

			re.firstTopologyUpdateFuture = re.firstTopologyUpdate(re.lastKnownUrls)
		}
		topologyUpdate = re.firstTopologyUpdateFuture
		re.mu.Unlock()
	}

	_, err := topologyUpdate.Get()
	return topologyUpdate, err
}

func (re *RequestExecutor) unlikelyExecute(command RavenCommand, topologyUpdate *completableFuture, sessionInfo *SessionInfo) error {
	var err error
	topologyUpdate, err = re.unlikelyExecuteInner(command, topologyUpdate, sessionInfo)
	if err != nil {
		re.mu.Lock()
		if re.firstTopologyUpdateFuture == topologyUpdate {
			re.firstTopologyUpdateFuture = nil // next request will raise it
		}
		re.mu.Unlock()
		return err
	}

	currentIndexAndNode, err := re.chooseNodeForRequest(command, sessionInfo)
	if err != nil {
		return err
	}
	err = re.Execute(currentIndexAndNode.currentNode, currentIndexAndNode.currentIndex, command, true, sessionInfo)
	return err
}

func (re *RequestExecutor) updateTopologyCallback() {
	last := re.lastReturnedResponse.Load().(time.Time)
	dur := time.Since(last)
	if dur < time.Minute {
		return
	}

	var serverNode *ServerNode

	selector := re.getNodeSelector()
	if selector == nil {
		return
	}
	preferredNode, err := re.getPreferredNode()
	if err != nil {
		return
	}
	serverNode = preferredNode.currentNode

	re.UpdateTopologyAsync(serverNode, 0)
}

type tupleStringError struct {
	S   string
	Err error
}

func (re *RequestExecutor) firstTopologyUpdate(inputUrls []string) *completableFuture {
	initialUrls := requestExecutorValidateUrls(inputUrls, re.Certificate)

	future := newCompletableFuture()
	var list []*tupleStringError
	f := func() {
		var err error
		defer func() {
			if err != nil {
				future.completeWithError(err)
			} else {
				future.complete(nil)
			}
		}()

		for _, url := range initialUrls {
			{
				serverNode := NewServerNode()
				serverNode.URL = url
				serverNode.Database = re.databaseName

				chRes := re.UpdateTopologyAsync(serverNode, math.MaxInt32)
				res := <-chRes
				err = res.Err
				if err == nil {
					re.initializeUpdateTopologyTimer()
					re.topologyTakenFromNode = serverNode
					return
				}
			}

			if _, ok := (err).(*DatabaseDoesNotExistError); ok {
				// Will happen on all node in the cluster,
				// so errors immediately
				re.lastKnownUrls = initialUrls
				return
			}

			if len(initialUrls) == 0 {
				re.lastKnownUrls = initialUrls
				err = newIllegalStateError("Cannot get topology from server: %s", url)
				return
			}
			list = append(list, &tupleStringError{url, err})
		}
		topology := &Topology{
			Etag: re.TopologyEtag,
		}
		topologyNodes := re.GetTopologyNodes()
		if len(topologyNodes) == 0 {
			for _, uri := range initialUrls {
				serverNode := NewServerNode()
				serverNode.URL = uri
				serverNode.Database = re.databaseName
				serverNode.ClusterTag = "!"
				topologyNodes = append(topologyNodes, serverNode)
			}
		}
		topology.Nodes = topologyNodes
		re.setNodeSelector(NewNodeSelector(topology))
		if len(initialUrls) > 0 {
			re.initializeUpdateTopologyTimer()
			return
		}
		re.lastKnownUrls = initialUrls

		var a []string
		for _, el := range list {
			first := el.S
			second := el.Err
			s := first + " -> " + second.Error()
			a = append(a, s)
		}
		details := strings.Join(a, ", ")
		err = re.throwError(details)
	}
	go f()
	return future
}

func (re *RequestExecutor) throwError(details string) error {
	err := newIllegalStateError("Failed to retrieve database topology from all known nodes \n" + details)
	return err
}

func requestExecutorValidateUrls(initialUrls []string, certificate *tls.Certificate) []string {
	// TODO: implement me
	return initialUrls
}

func (re *RequestExecutor) initializeUpdateTopologyTimer() {
	re.mu.Lock()
	defer re.mu.Unlock()

	if re.updateTopologyTimer != nil {
		return
	}
	// TODO: make it into an infinite goroutine instead
	f := func() {
		re.updateTopologyCallback()
		// Go doesn't have repeatable timer, so re-trigger ourselves
		re.mu.Lock()
		re.updateTopologyTimer = nil
		re.mu.Unlock()
		re.initializeUpdateTopologyTimer()
	}
	re.updateTopologyTimer = time.AfterFunc(time.Minute, f)
}

func isNetworkTimeoutError(err error) bool {
	// TODO: implement me
	// can test it by setting very low timeout in http.Client
	return false
}

// Execute executes a command on a given node
// If nodeIndex is -1, we don't know the index
func (re *RequestExecutor) Execute(chosenNode *ServerNode, nodeIndex int, command RavenCommand, shouldRetry bool, sessionInfo *SessionInfo) error {
	// nodeIndex -1 is equivalent to Java's null
	request, err := re.createRequest(chosenNode, command)
	if err != nil {
		return err
	}
	urlRef := request.URL.String()

	cachedItem, cachedChangeVector, cachedValue := re.getFromCache(command, urlRef)
	defer cachedItem.close()

	if cachedChangeVector != nil {
		aggressiveCacheOptions := re.aggressiveCaching
		if aggressiveCacheOptions != nil {
			expired := cachedItem.getAge() > aggressiveCacheOptions.Duration
			if !expired &&
				!cachedItem.getMightHaveBeenModified() &&
				command.getBase().CanCacheAggressively {
				return command.setResponse(cachedValue, true)
			}
		}

		request.Header.Set(headersIfNoneMatch, "\""+*cachedChangeVector+"\"")
	}

	if !re.disableClientConfigurationUpdates {
		etag := `"` + i64toa(re.ClientConfigurationEtag) + `"`
		request.Header.Set(headersClientConfigurationEtag, etag)
	}

	if !re.disableTopologyUpdates {
		etag := `"` + i64toa(re.TopologyEtag) + `"`
		request.Header.Set(headersTopologyEtag, etag)
	}

	//sp := time.Now()
	var response *http.Response
	re.NumberOfServerRequests.incrementAndGet()
	if re.shouldExecuteOnAll(chosenNode, command) {
		response, err = re.executeOnAllToFigureOutTheFastest(chosenNode, command)
	} else {
		response, err = command.send(re.httpClient, request)
	}

	if err != nil {
		if !shouldRetry && isNetworkTimeoutError(err) {
			return err
		}
		// Note: Java here re-throws if err is IOException and !shouldRetry
		// but for us that propagates the wrong error to RequestExecutorTest_failsWhenServerIsOffline
		urlRef = request.URL.String()
		var ok bool
		ok, err = re.handleServerDown(urlRef, chosenNode, nodeIndex, command, request, response, err, sessionInfo)
		if err != nil {
			return err
		}
		if !ok {
			return re.throwFailedToContactAllNodes(command, request, err, nil)
		}
		return nil
	}

	command.getBase().StatusCode = response.StatusCode

	refreshTopology := httpExtensionsGetBooleanHeader(response, headersRefreshTopology)
	refreshClientConfiguration := httpExtensionsGetBooleanHeader(response, headersRefreshClientConfiguration)

	if response.StatusCode == http.StatusNotModified {
		cachedItem.notModified()

		if command.getBase().ResponseType == RavenCommandResponseTypeObject {
			err = command.setResponse(cachedValue, true)
		}
		return err
	}

	var ok bool
	if response.StatusCode >= 400 {
		ok, err = re.handleUnsuccessfulResponse(chosenNode, nodeIndex, command, request, response, urlRef, sessionInfo, shouldRetry)
		if err != nil {
			return err
		}

		if !ok {
			dbMissingHeader := response.Header.Get("Database-Missing")
			if dbMissingHeader != "" {
				return newDatabaseDoesNotExistError(dbMissingHeader)
			}

			if len(command.getBase().FailedNodes) == 0 {
				return newIllegalStateError("Received unsuccessful response and couldn't recover from it. Also, no record of exceptions per failed nodes. This is weird and should not happen.")
			}

			if len(command.getBase().FailedNodes) == 1 {
				// return first error
				failedNodes := command.getBase().FailedNodes
				for _, err := range failedNodes {
					panicIf(err == nil, "err is nil")
					return err
				}
			}

			return newAllTopologyNodesDownError("Received unsuccessful response from all servers and couldn't recover from it.")
		}
		return nil // we either handled this already in the unsuccessful response or we are throwing
	}

	var responseDispose responseDisposeHandling
	responseDispose, err = ravenCommand_processResponse(command, re.Cache, response, urlRef)
	re.lastReturnedResponse.Store(time.Now())
	if err != nil {
		return err
	}

	if responseDispose == responseDisposeHandlingAutomatic {
		_ = response.Body.Close()
	}

	if refreshTopology || refreshClientConfiguration {

		serverNode := NewServerNode()
		serverNode.URL = chosenNode.URL
		serverNode.Database = re.databaseName

		var topologyTask chan *ClusterUpdateAsyncResult
		if refreshTopology {
			topologyTask = re.UpdateTopologyAsync(serverNode, 0)
		} else {
			topologyTask = make(chan *ClusterUpdateAsyncResult, 1)
			topologyTask <- &ClusterUpdateAsyncResult{Ok: false}
			close(topologyTask)
		}
		var clientConfiguration *completableFuture
		if refreshClientConfiguration {
			clientConfiguration = re.updateClientConfigurationAsync()
		} else {
			clientConfiguration = newCompletableFutureAlreadyCompleted(nil)
		}
		result := <-topologyTask
		err1 := result.Err
		_, err2 := clientConfiguration.Get()
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
	commandName := fmt.Sprintf("%T", command)
	message := "Tried to send " + commandName + " request via " + request.Method + " " + request.URL.String() + " to all configured nodes in the topology, " +
		"all of them seem to be down or not responding. I've tried to access the following nodes: "

	var urls []string
	nodeSelector := re.getNodeSelector()
	if nodeSelector != nil {
		for _, node := range nodeSelector.getTopology().Nodes {
			url := node.URL
			urls = append(urls, url)
		}
	}
	message += strings.Join(urls, ", ")

	if nodeSelector != nil && re.topologyTakenFromNode != nil {
		nodes := nodeSelector.getTopology().Nodes
		var a []string
		for _, n := range nodes {
			s := "( url: " + n.URL + ", clusterTag: " + n.ClusterTag + ", serverRole: " + n.ServerRole + ")"
			a = append(a, s)
		}
		nodesStr := strings.Join(a, ", ")

		message += "\nI was able to fetch " + re.topologyTakenFromNode.Database + " topology from " + re.topologyTakenFromNode.URL + ".\n" + "Fetched topology: " + nodesStr
	}

	return newAllTopologyNodesDownError("%s", message)
}

func (re *RequestExecutor) inSpeedTestPhase() bool {
	nodeSelector := re.getNodeSelector()
	return (nodeSelector != nil) && nodeSelector.inSpeedTestPhase()
}

func (re *RequestExecutor) shouldExecuteOnAll(chosenNode *ServerNode, command RavenCommand) bool {
	nodeSelector := re.getNodeSelector()
	multipleNodes := (nodeSelector != nil) && (len(nodeSelector.getTopology().Nodes) > 1)

	cmd := command.getBase()
	return re.readBalanceBehavior == ReadBalanceBehaviorFastestNode &&
		nodeSelector != nil &&
		nodeSelector.inSpeedTestPhase() &&
		multipleNodes &&
		cmd.IsReadRequest &&
		cmd.ResponseType == RavenCommandResponseTypeObject &&
		chosenNode != nil
}

type responseAndError struct {
	response *http.Response
	err      error
}

func (re *RequestExecutor) executeOnAllToFigureOutTheFastest(chosenNode *ServerNode, command RavenCommand) (*http.Response, error) {
	// note: implementation is intentionally different than Java

	var fastestWasRecorded int32 // atomic
	chanPreferredResponse := make(chan *responseAndError, 1)

	nPreferred := 0
	nodes := re.getNodeSelector().getTopology().Nodes
	for idx, node := range nodes {
		re.NumberOfServerRequests.incrementAndGet()

		isPreferred := node.ClusterTag == chosenNode.ClusterTag
		if isPreferred {
			nPreferred++
			panicIf(nPreferred > 1, "nPreferred is %d, should not be > 1", nPreferred)
		}

		go func(nodeIndex int, node *ServerNode) {
			var response *http.Response
			request, err := re.createRequest(node, command)
			if err == nil {
				response, err = command.send(re.httpClient, request)
				n := atomic.AddInt32(&fastestWasRecorded, 1)
				if n == 1 {
					// this is the first one, so record as fastest
					re.getNodeSelector().recordFastest(nodeIndex, node)
				}
			}
			// we return http response of the preferred node and close
			// all others
			if isPreferred {
				chanPreferredResponse <- &responseAndError{
					response: response,
					err:      err,
				}
			} else {
				if response != nil && err == nil {
					_ = response.Body.Close()
				}
			}
		}(idx, node)
	}

	select {
	case ret := <-chanPreferredResponse:
		// note: can be nil if there was an error
		return ret.response, ret.err
	case <-time.After(time.Second * 15):
		return nil, fmt.Errorf("request timed out")
	}
}

func (re *RequestExecutor) getFromCache(command RavenCommand, url string) (*releaseCacheItem, *string, []byte) {
	cmd := command.getBase()
	if cmd.CanCache && cmd.IsReadRequest && cmd.ResponseType == RavenCommandResponseTypeObject {
		return re.Cache.get(url)
	}

	return newReleaseCacheItem(nil), nil, nil
}

func (re *RequestExecutor) createRequest(node *ServerNode, command RavenCommand) (*http.Request, error) {
	request, err := command.createRequest(node)
	if err != nil {
		return nil, err
	}
	request.Header.Set(headersClientVersion, goClientVersion)
	return request, err
}

func (re *RequestExecutor) handleUnsuccessfulResponse(chosenNode *ServerNode, nodeIndex int, command RavenCommand, request *http.Request, response *http.Response, url string, sessionInfo *SessionInfo, shouldRetry bool) (bool, error) {
	var err error
	switch response.StatusCode {
	case http.StatusNotFound:
		re.Cache.setNotFound(url)
		switch command.getBase().ResponseType {
		case RavenCommandResponseTypeEmpty:
			return true, nil
		case RavenCommandResponseTypeObject:
			// TODO: should I propagate the error?
			_ = command.setResponse(nil, false)
		default:
			// TODO: should I propagate the error?
			_ = command.setResponseRaw(response, nil)
		}
		return true, nil
	case http.StatusForbidden:
		err = newAuthorizationError("Forbidden access to " + chosenNode.Database + "@" + chosenNode.URL + ", " + request.Method + " " + request.URL.String())
	case http.StatusGone: // request not relevant for the chosen node - the database has been moved to a different one
		if !shouldRetry {
			return false, nil
		}

		updateFuture := re.updateTopologyAsyncWithForceUpdate(chosenNode, int(math.MaxInt32), true)
		result := <-updateFuture
		if result.Err != nil {
			return false, result.Err
		}

		var currentIndexAndNode *CurrentIndexAndNode
		currentIndexAndNode, err = re.chooseNodeForRequest(command, sessionInfo)
		if err != nil {
			return false, err
		}
		err = re.Execute(currentIndexAndNode.currentNode, currentIndexAndNode.currentIndex, command, false, sessionInfo)
		return false, err
	case http.StatusGatewayTimeout, http.StatusRequestTimeout,
		http.StatusBadGateway, http.StatusServiceUnavailable:
		ok, err := re.handleServerDown(url, chosenNode, nodeIndex, command, request, response, nil, sessionInfo)
		return ok, err
	case http.StatusConflict:
		err = requestExecutorHandleConflict(response)
	default:
		command.getBase().onResponseFailure(response)
		err = exceptionDispatcherThrowError(response)
	}
	return false, err
}

func requestExecutorHandleConflict(response *http.Response) error {
	return exceptionDispatcherThrowError(response)
}

func (re *RequestExecutor) handleServerDown(url string, chosenNode *ServerNode, nodeIndex int, command RavenCommand, request *http.Request, response *http.Response, e error, sessionInfo *SessionInfo) (bool, error) {
	if command.getBase().FailedNodes == nil {
		command.getBase().FailedNodes = map[*ServerNode]error{}
	}

	re.addFailedResponseToCommand(chosenNode, command, request, response, e)

	if nodeIndex < 0 {
		// We executed request over a node not in the topology. This means no failover...
		return false, nil
	}

	re.spawnHealthChecks(chosenNode, nodeIndex)

	nodeSelector := re.getNodeSelector()
	if nodeSelector == nil {
		return false, nil
	}

	nodeSelector.onFailedRequest(nodeIndex)

	currentIndexAndNode, err := re.getPreferredNode()
	if err != nil {
		return false, err
	}

	if _, ok := command.getBase().FailedNodes[currentIndexAndNode.currentNode]; ok {
		//we tried all the nodes...nothing left to do
		return false, nil
	}

	err = re.Execute(currentIndexAndNode.currentNode, currentIndexAndNode.currentIndex, command, false, sessionInfo)
	if err != nil {
		return false, err
	}
	return true, nil
}

func (re *RequestExecutor) spawnHealthChecks(chosenNode *ServerNode, nodeIndex int) {
	nodeStatus := NewNodeStatus(re, nodeIndex, chosenNode)

	_, loaded := re.failedNodesTimers.LoadOrStore(chosenNode, nodeStatus)
	if !loaded {
		nodeStatus.startTimer()
	}
}

func (re *RequestExecutor) checkNodeStatusCallback(nodeStatus *NodeStatus) {
	nodesCopy := re.GetTopologyNodes()

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
		status := re.getFailedNodeTimer(nodeStatus.node)
		if status != nil {
			status.updateTimer()
		}

		return // will wait for the next timer call
	}

	status := re.getFailedNodeTimer(nodeStatus.node)
	if status != nil {
		re.failedNodesTimers.Delete(nodeStatus.node)
		status.Close()
	}

	nodeSelector := re.getNodeSelector()
	if nodeSelector != nil {
		nodeSelector.restoreNodeIndex(idx)
	}
}

func (re *RequestExecutor) clusterPerformHealthCheck(serverNode *ServerNode, nodeIndex int) error {
	panicIf(!re.isCluster, "clusterPerformHealthCheck() called on non-cluster RequestExector")
	command := NewGetTcpInfoCommand("health-check", "")
	return re.Execute(serverNode, nodeIndex, command, false, nil)
}

func (re *RequestExecutor) performHealthCheck(serverNode *ServerNode, nodeIndex int) error {
	// Note: in Java this is done via virtual functions
	if re.isCluster {
		return re.clusterPerformHealthCheck(serverNode, nodeIndex)
	}
	// Note: not reusing global singleton because in Go GetCommand() is not thread-safe
	op := NewGetStatisticsOperation("failure=check")
	command, err := op.GetCommand(re.conventions)
	if err != nil {
		return err
	}
	return re.Execute(serverNode, nodeIndex, command, false, nil)
}

// note: static
func (re *RequestExecutor) addFailedResponseToCommand(chosenNode *ServerNode, command RavenCommand, request *http.Request, response *http.Response, e error) {
	failedNodes := command.getBase().FailedNodes

	if response != nil && response.Body != nil {
		var schema exceptionSchema
		responseJson, err := ioutil.ReadAll(response.Body)
		if err == nil {
			err = jsonUnmarshal(responseJson, &schema)
		}
		if err == nil {
			readException := exceptionDispatcherGetFromSchema(&schema, response.StatusCode, e)
			failedNodes[chosenNode] = readException
		} else {
			exceptionSchema := &exceptionSchema{
				URL:     request.URL.String(),
				Type:    "Unparsable Server Response",
				Message: "Get unrecognized response from the server",
				Error:   string(responseJson),
			}
			exceptionToUse := exceptionDispatcherGetFromSchema(exceptionSchema, response.StatusCode, e)

			failedNodes[chosenNode] = exceptionToUse
		}
	}

	// this would be connections that didn't have response, such as "couldn't connect to remote server"

	if e == nil {
		e = newRavenError("")
	}

	exceptionSchema := &exceptionSchema{
		URL:     request.URL.String(),
		Type:    fmt.Sprintf("%T", e),
		Message: e.Error(),
		Error:   e.Error(),
	}

	exceptionToUse := exceptionDispatcherGetFromSchema(exceptionSchema, http.StatusInternalServerError, e)
	failedNodes[chosenNode] = exceptionToUse
}

// Close should be called when deleting executor
func (re *RequestExecutor) Close() {
	if re.isDisposed() {
		return
	}

	if re.isCluster {
		// make sure that a potentially pending UpdateTopologyAsync() has
		// finished
		re.clusterTopologySemaphore.acquire()
	}

	re.markDisposed()
	re.Cache.close()

	re.mu.Lock()
	defer re.mu.Unlock()

	if re.updateTopologyTimer != nil {
		re.updateTopologyTimer.Stop()
		re.updateTopologyTimer = nil
	}
	re.disposeAllFailedNodesTimers()
}

// TODO: create a different client if settings like compression
// or certificate differ
func (re *RequestExecutor) createClient() (*http.Client, error) {
	client := &http.Client{
		Timeout:   time.Second * 30,
		Transport: http.DefaultTransport,
	}
	if re.Certificate != nil || re.TrustStore != nil {
		tlsConfig, err := newTLSConfig(re.Certificate, re.TrustStore)
		if err != nil {
			return nil, err
		}
		client.Transport = &http.Transport{
			TLSClientConfig: tlsConfig,
		}
	}
	if HTTPClientPostProcessor != nil {
		HTTPClientPostProcessor(client)
	}
	return client, nil
}

func (re *RequestExecutor) getPreferredNode() (*CurrentIndexAndNode, error) {
	ns, err := re.ensureNodeSelector()
	if err != nil {
		return nil, err
	}

	return ns.getPreferredNode()
}

func (re *RequestExecutor) getNodeBySessionID(sessionID int) (*CurrentIndexAndNode, error) {
	ns, err := re.ensureNodeSelector()
	if err != nil {
		return nil, err
	}

	return ns.getNodeBySessionID(sessionID)
}

func (re *RequestExecutor) getFastestNode() (*CurrentIndexAndNode, error) {
	ns, err := re.ensureNodeSelector()
	if err != nil {
		return nil, err
	}

	return ns.getFastestNode()
}

func (re *RequestExecutor) ensureNodeSelector() (*NodeSelector, error) {
	re.mu.Lock()
	firstTopologyUpdate := re.firstTopologyUpdateFuture
	re.mu.Unlock()

	if firstTopologyUpdate != nil && (!firstTopologyUpdate.IsDone() || firstTopologyUpdate.IsCompletedExceptionally()) {
		_, err := firstTopologyUpdate.Get()
		if err != nil {
			return nil, err
		}
	}

	nodeSelector := re.getNodeSelector()
	if nodeSelector == nil {
		topology := &Topology{
			Nodes: re.GetTopologyNodes(),
			Etag:  re.TopologyEtag,
		}

		nodeSelector = NewNodeSelector(topology)
		re.setNodeSelector(nodeSelector)
	}
	return nodeSelector, nil
}

// NodeStatus represents status of server node
type NodeStatus struct {
	timerPeriod     time.Duration
	requestExecutor *RequestExecutor
	nodeIndex       int
	node            *ServerNode
	timer           *time.Timer
}

func NewNodeStatus(requestExecutor *RequestExecutor, nodeIndex int, node *ServerNode) *NodeStatus {
	return &NodeStatus{
		requestExecutor: requestExecutor,
		nodeIndex:       nodeIndex,
		node:            node,
		timerPeriod:     time.Millisecond * 100,
	}
}

func (s *NodeStatus) nextTimerPeriod() time.Duration {
	if s.timerPeriod > time.Second*5 {
		return time.Second * 5
	}
	s.timerPeriod = s.timerPeriod + (time.Millisecond * 100)
	return s.timerPeriod
}

func (s *NodeStatus) startTimer() {
	f := func() {
		s.timerCallback()
	}
	s.timer = time.AfterFunc(s.timerPeriod, f)
}

func (s *NodeStatus) updateTimer() {
	// TODO: not sure if Reset
	s.timer.Reset(s.nextTimerPeriod())
}

func (s *NodeStatus) timerCallback() {
	if !s.requestExecutor.isDisposed() {
		s.requestExecutor.checkNodeStatusCallback(s)
	}
}

func (s *NodeStatus) Close() {
	if s.timer != nil {
		s.timer.Stop()
		s.timer = nil
	}
}
