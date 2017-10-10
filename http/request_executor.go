package http

import (
	"../data"
	"../tools"
	"./commands"
	ravenErrors "../errors"
	"net/http"
	"time"
	"sync"
	"errors"
	"fmt"
)

type RequestExecutor struct{

	url, database, apiKey, ClientVersion string
	TopologyEtag int64
	lastReturnedResponseTime time.Time
	firstTopologyUpdate chan(bool)
	disposed bool
	updateTopologyTickerStarted, withoutTopology bool
	updateTopologyTickerLock sync.Mutex
	nodeSelector *NodeSelector
	updateTopologyLock sync.Mutex
	GlobalHttpClientTimeout time.Duration
	GlobalHttpClient http.Client
	ServerNode ServerNode
	failedNodesTickers map[IServerNode]NodeStatus

	convention data.DocumentConvention
	topology Topology
	IsFirstTryToLoadFromTopologyCache bool
	VersionInfo string
	Headers []http.Header
	TopologyChangeCounter uint
	RequestCount uint
	authenticator tools.Authenticator
}

type NodeSelector struct{
	topology *Topology
	topologyLock sync.Mutex
	currentNodeIdx int
	nodeIndexLock sync.RWMutex
}

func NewRequestExecutor(dBName string, apiKey string) (*RequestExecutor, error){
	return &RequestExecutor{database:dBName, apiKey:apiKey, TopologyEtag:0, lastReturnedResponseTime:time.Now(), updateTopologyTickerStarted:false}, nil
}

func NewNodeSelector(topology *Topology) (*NodeSelector, error){
	return &NodeSelector{topology, sync.Mutex{}, nil, sync.Mutex{}}, nil
}

func (executor RequestExecutor) Create(urls []string, databaseName string, apiKey string){
	executor.firstTopologyUpdate = executor.doFirstTopologyUpdate(urls)
}

func (executor RequestExecutor) CreateForSingleNode(url string, databaseName string, apiKey string){
	node := NewServerNode(url, databaseName, apiKey, "", false)
	nodes := []IServerNode{node}
	topology := NewTopology(-1, ServerNode{}, data.ReadBehaviour{}, data.WriteBehaviour{}, nodes, 0)
	executor.nodeSelector, _ = NewNodeSelector(topology)
}

func (executor RequestExecutor) UpdateTopology(node ServerNode, timeout int) (bool, error){
	if executor.disposed{
		return false, nil
	}

	executor.updateTopologyLock.Lock()
	defer executor.updateTopologyLock.Unlock()

	if executor.disposed{
		return false, nil
	}

	//start of json operation context
	command, _ := commands.NewGetTopologyCommand("")
	executor.Execute(node, *command,  false)
	//serverHash := GetServerHashWithSeed(node.Url, executor.database)
	//Todo: Save topology to local cache
	if &executor.nodeSelector == nil {
		nodesSelectorPtr, _ := NewNodeSelector(command.Result)
		executor.nodeSelector = *nodesSelectorPtr
	}else if executor.nodeSelector.OnUpdateTopology(command.Result, false){
		executor.DisposeAllFailedNodesTickers()
	}
	executor.TopologyEtag = executor.nodeSelector.topology.Etag
	//end of json operation context

	return false
}

func (executor RequestExecutor) UpdateTopologyAsync(node ServerNode, timeout int) chan(error){
	promise := make(chan error, 1)
	go func(){
		_, err := executor.UpdateTopology(node, timeout)
		promise <- err
	}()
	return promise
}

func (executor RequestExecutor) DisposeAllFailedNodesTickers(){
	oldFailedNodesTickers := executor.failedNodesTickers
	executor.failedNodesTickers = make(map[ServerNode]NodeStatus)
	for node, status := range oldFailedNodesTickers{
		status.StopTicker()
	}
}

func (executor RequestExecutor) doFirstTopologyUpdate(initialUrls []string) chan(bool){
	var errorList map[string]error
	var promises []chan error
	for url := range initialUrls{
		serverNode := *NewServerNode(url, executor.database)
		promise := executor.UpdateTopologyAsync(serverNode, 0)
		executor.initPeriodicTopologyUpdates()
	}
}


func (executor RequestExecutor) initPeriodicTopologyUpdates(){
	if executor.updateTopologyTickerStarted{
		return
	}

	executor.updateTopologyTickerLock.Lock()
	defer executor.updateTopologyTickerLock.Unlock()

	if executor.updateTopologyTickerStarted{
		return
	}

	ticker := time.NewTicker(time.Minute * 5)
	go func() {
		for t := range ticker.C {
			if t.Sub(executor.lastReturnedResponseTime) < time.Duration(5*time.Minute){
				return
			}
			node, err := executor.nodeSelector.GetCurrentNode()
			if err != nil{
				//log it i guess
			}
			executor.UpdateTopology(node, 0)
		}
	}()
}

func (executor RequestExecutor) ExecuteOnCurrentNode(command RavenRequestable) error{
	//topologyUpdate := executor.firstTopologyUpdate

	currentNode, _ := executor.nodeSelector.GetCurrentNode()
	_, err := executor.Execute(currentNode, command, false)
	return err
}

func (executor RequestExecutor) Execute(chosenNode IServerNode, command RavenRequestable, shouldRetry bool) (interface{}, error){
	request := executor.createRequest(chosenNode, command, &executor.url)
	nodeIdx := executor.nodeSelector.GetCurrentNodeIndex()

	if executor.withoutTopology{
		request.Header["Topology-Etag"] = append(request.Header["Topology-Etag"], fmt.Sprintf("\"%s\"", executor.TopologyEtag))
	}

	client, err := executor.getHttpClientForCommand(command)
	if err != nil{
		return nil, err
	}
	timeout := command.GetTimeout()
	client.Timeout = timeout
	response, err := command.Send(client, request)
	command.SetStatusCode(response.StatusCode)
	if err != nil{
		if !shouldRetry {
			return nil, err
		}
		if executor.HandleServerDown(chosenNode, nodeIdx, command, request, response){
			topologyErrPtr, _ := ravenErrors.NewAllTopologyNodesDownError("Tried to send request to all configured nodes in the topology, all of them seem to be down or not responding.", executor.nodeSelector.topology)
			return nil, *topologyErrPtr
		}
	}
	executor.lastReturnedResponseTime = time.Now()
	command.ProcessResponse(response, executor.url)
	if command.ShouldRefreshTopology(){
		serverNode := NewServerNode(executor.url, executor.database, "", "", false)
		executor.UpdateTopology(*serverNode, 0)
	}
	return response, nil //get result before returning
}

func (executor RequestExecutor) createRequest(node IServerNode, command RavenRequestable, urlPtr *string) http.Request{
	request := command.CreateRequest(node, urlPtr)
	request.RequestURI = *urlPtr
	if node.ClusterToken != ""{
		request.Header.Add("Raven-Authorization", node.ClusterToken)
	}
	if request.Header.Get("Raven-Client-Version") == ""{
		request.Header.Add("Raven-Client-Version", executor.ClientVersion)
	}
	return request
}

func (executor RequestExecutor) getHttpClientForCommand(command RavenRequestable) (http.Client, error){
	timeout := command.GetTimeout()
	if timeout > executor.GlobalHttpClientTimeout{
		return executor.GlobalHttpClient, errors.New(fmt.Sprintf("Maximum request timeout is set to '%s' but was '%s'.", executor.GlobalHttpClientTimeout, timeout))
	}
	return executor.GlobalHttpClient, nil
}

func (executor RequestExecutor) HandleServerDown(chosenNode IServerNode, nodeIdx int, command RavenRequestable, request http.Request, response http.Response) (bool){
	serverError, err := ravenErrors.NewServerError(response)
	if err != nil{
		return false
	}
	executor.AddFailedResponseToCommand(chosenNode, command, serverError)
	nodeSelector := executor.nodeSelector
	executor.SpawnHealthChecks(chosenNode, nodeIdx)
	if &nodeSelector != nil{
		nodeSelector.OnFailedRequest(nodeIdx)
	}
	currentNode, _ := executor.nodeSelector.GetCurrentNode()
	if _, ok := command.GetFailedNodes()[currentNode]; ok{
		return false
	}
	
}

func (executor RequestExecutor) AddFailedResponseToCommand(chosenNode IServerNode, command RavenRequestable, err error) error{
	command.SetFailedNode(chosenNode, err)
	return nil
}

func (executor RequestExecutor) CheckNodeStatusCallback(ns *NodeStatus){
	copy := executor.nodeSelector.topology.Nodes
	if ns.NodeIndex >= len(copy){
		return// topology index changed / removed
	}
	serverNode := copy[ns.NodeIndex]
	if &serverNode != ns.Node{
		return// topology changed, nothing to check
	}
	_, err := executor.PerformHealthCheck(serverNode)
	if err != nil{
		//log
		if val ,ok := executor.failedNodesTickers[ns.Node]; ok{
			val.UpdateTicker()
		}
		return
	}

	if val ,ok := executor.failedNodesTickers[ns.Node]; ok{
		val.StopTicker()
		delete(executor.failedNodesTickers, ns.Node)
	}
	executor.nodeSelector.RestoreNodeIndex(ns.NodeIndex)
}

func (executor RequestExecutor) PerformHealthCheck(serverNode IServerNode) (interface{}, error){
	getStatisticsCommand := commands.NewGetStatisticsCommand()
	return executor.Execute(serverNode, getStatisticsCommand, false)
}

func (executor RequestExecutor) SpawnHealthChecks(chosenNode IServerNode, nodeIndex int){
	nodeStatus, _ := NewNodeStatus(executor, nodeIndex, chosenNode)
	if _, ok := executor.failedNodesTickers[chosenNode]; !ok{
		executor.failedNodesTickers[chosenNode] = *nodeStatus
		nodeStatus.StartTicker()
	}
}

func (selector NodeSelector) GetCurrentNodeIndex() int{
	selector.nodeIndexLock.RLock()
	defer selector.nodeIndexLock.RUnlock()
	return selector.currentNodeIdx
}

func (selector NodeSelector) GetCurrentNode() (ServerNode, error){
	if len(selector.topology.Nodes) == 0{
		return nil, errors.New("request_executor:Topology has no nodes")
	}
	return selector.topology.Nodes[selector.currentNodeIdx], nil
}

func (selector NodeSelector) OnUpdateTopology(topology Topology, forceUpdate bool) bool{
	if &topology == nil{
		return false
	}

	oldTopology := selector.topology
	if oldTopology.Etag >= topology.Etag && !forceUpdate{
		return false
	}

	if &selector.topology == &oldTopology{
		selector.topology = topology
	}
	return &selector.topology == &topology
}

func (selector NodeSelector) RestoreNodeIndex(nodeIndex int){
	selector.nodeIndexLock.Lock()
	defer selector.nodeIndexLock.Unlock()
	selector.currentNodeIdx = nodeIndex
}

func (selector NodeSelector) OnFailedRequest(nodeIdx int){
	if len(selector.topology.Nodes) == 0{
		return
	}

	if nodeIdx < len(selector.topology.Nodes) - 1{
		nodeIdx = nodeIdx+1
	}else{
		nodeIdx = 0
	}
	selector.RestoreNodeIndex(nodeIdx)
}