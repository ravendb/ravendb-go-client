package ravendb

import (
	"sync"
	"time"
)

// ServerNode describes a single server node
type ServerNode struct {
	URL        string `json:"Url"`
	ClusterTag string `json:"ClusterTag"`
	ServerRole string `json:"ServerRole"`
	Database   string `json:"Database"`
}

// TODO: on ServerNode implement:
// response_time, ewma(), is_rate_surpassed()

// Topology describes server nodes
// Result of
// {"Nodes":[{"Url":"http://localhost:9999","ClusterTag":"A","Database":"PyRavenDB","ServerRole":"Rehab"}],"Etag":10}
type Topology struct {
	Nodes []*ServerNode `json:"Nodes"`
	Etag  int           `json:"Etag"`
}

// NodeSelector describes node selector
type NodeSelector struct {
	Topology         *Topology
	CurrentNodeIndex int
	// TODO: the use of this lock seems to be either unnecessary or broken
	lock sync.Mutex
}

// NewNodeSelector creates a new NodeSelector
func NewNodeSelector(t *Topology) *NodeSelector {
	return &NodeSelector{
		Topology: t,
	}
}

// OnFailedRequest is called when node fails
func (s *NodeSelector) OnFailedRequest(nodeIndex int) {
	topologyNodesLen := len(s.Topology.Nodes)
	panicIf(topologyNodesLen == 0, "Empty database topology, this shouldn't happen.")
	nextNodeIndex := 0
	if nodeIndex < topologyNodesLen-1 {
		nextNodeIndex = nodeIndex + 1
	}
	s.CurrentNodeIndex = nextNodeIndex
}

// OnUpdateTopology is called when topology changes
func (s *NodeSelector) OnUpdateTopology(topology *Topology, forceUpdate bool) bool {
	if topology == nil {
		return false
	}
	oldTopology := s.Topology
	for {
		if oldTopology.Etag >= s.Topology.Etag && !forceUpdate {
			return false
		}
		s.lock.Lock()
		if !forceUpdate {
			s.CurrentNodeIndex = 0
		}

		if oldTopology == s.Topology {
			s.Topology = topology
			s.lock.Unlock()
			return true
		}
		s.lock.Unlock()
		oldTopology = s.Topology
	}
}

// GetCurrentNode returns current node
func (s *NodeSelector) GetCurrentNode() *ServerNode {
	panicIf(len(s.Topology.Nodes) == 0, "There are no nodes in the topology at all")
	return s.Topology.Nodes[s.CurrentNodeIndex]
}

// RestoreNodeIndex restores a node
func (s *NodeSelector) RestoreNodeIndex(nodeIndex int) {
	currentNodeIndex := s.CurrentNodeIndex
	if currentNodeIndex > nodeIndex {
		s.lock.Lock()
		if currentNodeIndex == s.CurrentNodeIndex {
			s.CurrentNodeIndex = nodeIndex
		}
		s.lock.Unlock()
	}
}

// TODO: NodeStatus could be replaced by a continously running goroutine
// that sleeps

// NodeStatus describes a status of the node
type NodeStatus struct {
	requestsExecutor *RequestsExecutor
	nodeIndex        int
	node             *ServerNode
	timerPeriod      time.Duration
	timer            *time.Timer
}

func (s *NodeStatus) nextTimerPeriod() time.Duration {
	// TODO: verify those are the values that correspond to Python values
	if !(s.timerPeriod >= time.Minute*5) {
		s.timerPeriod += time.Millisecond * 100 // 0.1 of a second
	}
	if s.timerPeriod >= time.Minute*5 {
		return time.Minute * 5
	}
	return s.timerPeriod
}

func (s *NodeStatus) startTimer() {
	// this function is called on a separate goroutine
	f := func() {
		// TDOO: call the right function
		// s.requestsExecutor.CheckNodeStatus()

	}
	dur := s.nextTimerPeriod()
	s.timer = time.AfterFunc(dur, f)
}

// Cancel should be called when we delete NodeStatus to stop timer runs
// the node status check goroutine
func (s *NodeStatus) Cancel() {
	if s.timer != nil {
		s.timer.Stop()
		s.timer = nil
	}
}
