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
	nodeSelector NodeSelector
	updateTopologyLock sync.Mutex
	GlobalHttpClientTimeout time.Duration
	GlobalHttpClient http.Client
	ServerNode ServerNode

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
	topology Topology
	currentNodeIdx int
}

func NewRequestExecutor(dBName string, apiKey string) (*RequestExecutor, error){
	return &RequestExecutor{database:dBName, apiKey:apiKey, TopologyEtag:0, lastReturnedResponseTime:time.Now(), updateTopologyTickerStarted:false}, nil
}

func NewNodeSelector(topology Topology) (*NodeSelector, error){
	return &NodeSelector{topology, 0}, nil
}

func (executor RequestExecutor) Create(){

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

	//start of json operation conetext
	command, _ := commands.NewGetTopologyCommand()
	executor.Execute(node, context, command, shouldRetry: false)
	serverHash := GetServerHash(node.Url, executor.database)

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

func (executor RequestExecutor) firstTopologyUpdate(initialUrls []string){
	var errorList map[string]error
	var promises []chan error
	for url := range initialUrls{
		serverNode := NewServerNode(url, executor.database)
		promise := executor.UpdateTopologyAsync(serverNode, 0)
		initPeriodicTopologyUpdates()
	}
	res <-
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

func (executor RequestExecutor) Execute(chosenNode ServerNode, command RavenRequestable, shouldRetry bool) (interface{}, error){
	request := executor.createRequest(chosenNode, command, &executor.url)
	nodeIdx := executor.nodeSelector.GetCurrentNodeIndex()

	if executor.withoutTopology{
		request.Header["Topology-Etag"] = append(request.Header["Topology-Etag"], fmt.Sprintf("\"%s\"", executor.TopologyEtag))
	}

	requestStartTime := time.Now()
	client, err := executor.getHttpClientForCommand(command)
	if err != nil{
		return nil, err
	}
	timeout := command.GetTimeout()
	response, err := command.Send(client, request)
	if err != nil{
		if !shouldRetry {
			return nil, err
		}
		if
	}
	requestEndTime := time.Now()
}

func (executor RequestExecutor) createRequest(node ServerNode, command RavenRequestable, urlPtr *string) http.Request{
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
		return nil, errors.New(fmt.Sprintf("Maximum request timeout is set to '%s' but was '%s'.", executor.GlobalHttpClientTimeout, timeout))
	}
	return executor.GlobalHttpClient, nil
}

func (executor RequestExecutor) HandleServerDown(chosenNode ServerNode, nodeIdx int, command RavenRequestable, request http.Request, response http.Response){
	serverError, err := ravenErrors.NewServerError(response)
	if err != nil{
		return
	}
	executor.AddFailedResponseToCommand(chosenNode, command, serverError)
	nodeSelector = executor.nodeSelector
	SpawnHealthchecks()
}

func (executor RequestExecutor) AddFailedResponseToCommand(chosenNode ServerNode, command RavenRequestable, err error) error{
	command.SetFailedNode(chosenNode, err)
	return nil
}

func (selector NodeSelector) GetCurrentNodeIndex() int{
	return selector.currentNodeIdx
}

func (selector NodeSelector) GetCurrentNode() (ServerNode, error){
	if len(selector.topology.Nodes) == 0{
		return nil, errors.New("request_executor:Topology has no nodes")
	}
	return selector.topology.Nodes[selector.currentNodeIdx], nil
}
