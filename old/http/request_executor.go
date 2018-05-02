package http

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/golang/glog"
	"github.com/ravendb/ravendb-go-client/http/commands"
	"github.com/ravendb/ravendb-go-client/http/server_nodes"
)

type RequestExecutor struct {
	NodeSelector NodeSelector
	TopologyEtag int64
	databaseName string

	lastKnownUrls            []string
	lastReturnedResponseTime time.Time

	Headers          map[string]string
	GlobalHttpClient http.Client

	updateTopologyTickerRunning    bool
	updateTopologyTicker           time.Ticker
	updateTopologyLock             sync.Mutex
	updateTickerLock               sync.Mutex
	failedNodesTickers             map[server_nodes.IServerNode]NodeStatus
	DisableTopologyUpdates, closed bool
}

func Create(urls []string, databaseName string) (*RequestExecutor, error) {
	executor := RequestExecutor{databaseName: databaseName}
	go executor.FirstTopologyUpdate(urls)
	return &executor, nil
}

func CreateForSingleNode(url string, databaseName string) (*RequestExecutor, error) {
	nodePtr, _ := server_nodes.NewServerNode(url, databaseName)
	topology, _ := NewTopology(-1, []server_nodes.IServerNode{*nodePtr})
	nodeSelectorPtr, _ := NewNodeSelector(topology)
	return &RequestExecutor{
		databaseName:           databaseName,
		NodeSelector:           *nodeSelectorPtr,
		TopologyEtag:           -2,
		DisableTopologyUpdates: true,
	}, nil
}

func (executor RequestExecutor) FirstTopologyUpdate(initialUrls []string) (bool, error) {
	errorList := make(map[string]error)
	for _, url := range initialUrls {
		serverNodePtr, _ := server_nodes.NewServerNode(url, executor.databaseName)
		_, err := executor.UpdateTopology(*serverNodePtr)
		if err != nil {
			glog.Info(fmt.Sprintf("Cannot get topology from server: %s %s", url, err))
			errorList[url] = err
		} else {
			executor.initPeriodicTopologyUpdates()
			return true, nil
		}
	}

	executor.lastKnownUrls = initialUrls
	return false, TopologyUpdateError{"Failed to retrieve cluster topology from all known nodes", errorList}
}

// ExecuteOnCurrentNode - shouldRetry must be true default
func (executor RequestExecutor) ExecuteOnCurrentNode(command commands.IRavenRequestable, shouldRetry bool) ([]byte, error) {
	//topologyUpdate := executor.updateTopologyTickerRunning
	if !executor.DisableTopologyUpdates {
		if !executor.updateTopologyTickerRunning {
			if len(executor.lastKnownUrls) == 0 {
				return []byte{}, errors.New("No known topology and no previously known one, cannot proceed, likely a bug")
			}
			executor.FirstTopologyUpdate(executor.lastKnownUrls)
		}
	}

	if &executor.NodeSelector == nil {
		return []byte{}, errors.New("A connection with the server could not be established\nnode_selector cannot be Nil, please check your connection\nor supply a valid node_selector")
	}
	node := executor.NodeSelector.GetCurrentNode()
	return executor.Execute(node, command, shouldRetry)
}

func (executor RequestExecutor) Execute(node server_nodes.IServerNode, command commands.IRavenRequestable, shouldRetry bool) ([]byte, error) {
	for {
		command.CreateRequest(node)
		var nodeIndex int
		if &executor.NodeSelector != nil {
			nodeIndex = executor.NodeSelector.CurrentNodeIndex
		}

		//open session?
		command.SetHeaders(executor.Headers)
		if !executor.DisableTopologyUpdates {
			headers := command.GetHeaders()
			headers["Topology-Etag"] = fmt.Sprintf("\"%d\"", executor.TopologyEtag)
			command.SetHeaders(headers)
		}

		rawData := command.GetData()
		if &rawData != nil {
			data, err := json.Marshal(command.GetData())
			if err != nil {
				return nil, err
			}
			requestPtr, err := http.NewRequest(command.GetMethod(), command.GetUrl(), bytes.NewBuffer(data))
			if err != nil {
				return nil, nil
			}
			client, err := executor.getHttpClientForCommand(command)
			if err != nil {
				return nil, err
			}
			startTime := time.Now()
			var endTime time.Time
			respPtr, err := client.Do(requestPtr)
			if err != nil {
				endTime = time.Now()
				if !shouldRetry {
					return nil, err
				}
				handled, err := executor.HandleServerDown(node, nodeIndex, command, err)
				if !handled || err != nil {
					topologyErrPtr, err2 := NewAllTopologyNodesDownError("Tried to send request to all configured nodes in the topology,\nall of them seem to be down or not responding.", executor.NodeSelector.Topology)
					if err2 != nil {
						return nil, err
					}

					return nil, topologyErrPtr
				}
				node = executor.NodeSelector.GetCurrentNode()
				continue
			}
			for headerName, headerVal := range command.GetHeaders() {
				requestPtr.Header.Add(headerName, headerVal)
			}
			if &endTime == nil {
				endTime = time.Now()
			}
			elapsedTime := endTime.Sub(startTime)
			node.SetResponseTime(elapsedTime)

			if respPtr.StatusCode == 404 {
				// nil зачем передавать ? Это ж точно лишний вызов?
				return command.GetResponseRaw(nil)
			} else if respPtr.StatusCode == 403 {
				//todo handle cert
			} else if respPtr.StatusCode == 408 || respPtr.StatusCode == 502 || respPtr.StatusCode == 503 || respPtr.StatusCode == 504 {
				failedNodes := command.GetFailedNodes()
				if len(failedNodes) == 1 {
					node = failedNodes[0]
					reqError, err2 := NewUnsuccessfulRequestError(command.GetUrl(), node)
					if err2 != nil {
						return nil, err2
					}
					return nil, reqError
				}
			} else if respPtr.StatusCode == 409 {
				//todo
			}

			if respPtr.Header.Get("Refresh-Topology") != "" {
				newNode, _ := server_nodes.NewServerNode(node.GetUrl(), executor.databaseName)
				executor.UpdateTopology(*newNode)
			}
			executor.lastReturnedResponseTime = time.Now()
			return command.GetResponseRaw(respPtr)
		}
	}
}

func (executor RequestExecutor) getHttpClientForCommand(command commands.IRavenRequestable) (http.Client, error) {
	return executor.GlobalHttpClient, nil
}

func (executor RequestExecutor) UpdateTopology(node server_nodes.IServerNode) (bool, error) {
	if executor.closed {
		return false, errors.New("Request executor is closed")
	}

	executor.updateTopologyLock.Lock()
	defer executor.updateTopologyLock.Unlock()

	if executor.closed {
		return false, errors.New("Request executor is closed")
	}

	command, _ := commands.NewGetTopologyCommand()

	response, err := executor.Execute(node, command, false)
	if err != nil {
		return false, err
	}

	topologyPtr, err := CreateFromJSON(response) //Todo: Save topology to local cache
	if err != nil {
		return false, err
	}
	if &executor.NodeSelector == nil {
		nodesSelectorPtr, err := NewNodeSelector(topologyPtr)
		if err != nil {
			return false, err
		}
		executor.NodeSelector = *nodesSelectorPtr
	} else if executor.NodeSelector.OnUpdateTopology(topologyPtr) {
		executor.stopAllFailedNodesTickers()
	}

	executor.TopologyEtag = executor.NodeSelector.Topology.Etag

	return true, nil
}

func (executor RequestExecutor) HandleServerDown(node server_nodes.IServerNode, nodeIndex int, command commands.IRavenRequestable, err error) (bool, error) {
	command.AddFailedNode(node, err)

	if _, nodeIsFailed := executor.failedNodesTickers[node]; &executor.NodeSelector != nil && !nodeIsFailed {
		nodeStatusPtr, _ := NewNodeStatus(nodeIndex, node)

		executor.updateTickerLock.Lock()
		defer executor.updateTickerLock.Unlock()

		if _, nodeIsFailed := executor.failedNodesTickers[node]; !nodeIsFailed {
			executor.failedNodesTickers[node] = *nodeStatusPtr
			nodeStatusPtr.StartTicker()
		}

		executor.NodeSelector.OnFailedRequest(node)
		currentNode := executor.NodeSelector.GetCurrentNode()
		if command.HasFailedWithNode(currentNode) {
			return false, nil
		}
	}
	return true, nil
}

func (executor RequestExecutor) CheckNodeStatus(status NodeStatus) error {
	if &executor.NodeSelector != nil {
		nodes := executor.NodeSelector.Topology.Nodes
		if status.NodeIndex >= len(nodes) {
			return nil
		}
		node := nodes[status.NodeIndex]
		if node != status.Node {
			return executor.PerformHealthCheck(node, status)
		}
	}
	return nil
}

func (executor RequestExecutor) PerformHealthCheck(node server_nodes.IServerNode, status NodeStatus) error {
	commandPtr, err := commands.NewGetStatisticsCommand()
	if err != nil {
		return err
	}
	_, err = executor.Execute(node, commandPtr, false)
	if err != nil {
		glog.Info(fmt.Sprintf("%s is still down", node.GetClusterTag()))
		if nodeStatus, ok := executor.failedNodesTickers[node]; ok {
			nodeStatus.StartTicker()
		}
	}
	if _, ok := executor.failedNodesTickers[node]; ok {
		delete(executor.failedNodesTickers, node)
	}
	executor.NodeSelector.RestoreNodeIndex(status.NodeIndex)
	return nil
}

func (executor RequestExecutor) initPeriodicTopologyUpdates() error {
	if executor.updateTopologyTickerRunning {
		return nil
	}

	executor.updateTopologyTicker = *time.NewTicker(time.Minute * 5)
	go func() {
		for t := range executor.updateTopologyTicker.C {
			if t.Sub(executor.lastReturnedResponseTime) < time.Duration(5*time.Minute) {
				return
			}
			node := executor.NodeSelector.GetCurrentNode()
			_, err := executor.UpdateTopology(node)
			if err != nil {
				glog.Info("Couldn't Update Topology during periodic updates")
			}
		}
	}()
	return nil
}

func (executor RequestExecutor) stopAllFailedNodesTickers() {
	for _, nodeStatus := range executor.failedNodesTickers {
		nodeStatus.StopTicker()
	}
}

func (executor RequestExecutor) Close() {
	if executor.closed {
		return
	}

	executor.closed = true
	executor.stopAllFailedNodesTickers()
	if executor.updateTopologyTickerRunning {
		executor.updateTopologyTickerRunning = false
		executor.updateTopologyTicker.Stop()
	}
}
