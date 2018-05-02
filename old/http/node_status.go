package http

import (
	"time"

	"github.com/ravendb/ravendb-go-client/http/server_nodes"
)

type NodeStatus struct {
	NodeIndex int
	Node      server_nodes.IServerNode

	tickerIsRunning bool
	ticker          time.Ticker
}

func NewNodeStatus(nodeIndex int, node server_nodes.IServerNode) (*NodeStatus, error) {
	return &NodeStatus{NodeIndex: nodeIndex, Node: node}, nil
}

func (status NodeStatus) StartTicker() {
	period := time.Duration(time.Millisecond * 100)
	status.ticker = *time.NewTicker(period)
}

func (status NodeStatus) StopTicker() {
	status.ticker.Stop()
}

func (status NodeStatus) GetIsTickerRunning() bool {
	return status.tickerIsRunning
}
