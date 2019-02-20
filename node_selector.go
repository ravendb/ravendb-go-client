package ravendb

import (
	"time"
)

// NodeSelector describes node selector
type NodeSelector struct {
	_updateFastestNodeTimer *time.Timer
	_state                  *NodeSelectorState
}

// NewNodeSelector creates a new NodeSelector
func NewNodeSelector(t *Topology) *NodeSelector {
	state := NewNodeSelectorState(t)
	return &NodeSelector{
		_state: state,
	}
}

func (s *NodeSelector) getTopology() *Topology {
	return s._state.topology
}

func (s *NodeSelector) onFailedRequest(nodeIndex int) {
	state := s._state
	if nodeIndex < 0 || nodeIndex >= len(state.failures) {
		return // probably already changed
	}

	state.failures[nodeIndex].incrementAndGet()
}

func (s *NodeSelector) onUpdateTopology(topology *Topology, forceUpdate bool) bool {
	if topology == nil {
		return false
	}

	stateEtag := s._state.topology.Etag
	topologyEtag := topology.Etag

	if stateEtag >= topologyEtag && !forceUpdate {
		return false
	}

	s._state = NewNodeSelectorState(topology)

	return true
}

func (s *NodeSelector) getPreferredNode() (*CurrentIndexAndNode, error) {
	state := s._state
	stateFailures := state.failures
	serverNodes := state.nodes
	n := min(len(serverNodes), len(stateFailures))
	for i := 0; i < n; i++ {
		if stateFailures[i].get() == 0 && serverNodes[i].URL != "" {
			return NewCurrentIndexAndNode(i, serverNodes[i]), nil
		}
	}
	return s.unlikelyEveryoneFaultedChoice(state)
}

func (s *NodeSelector) unlikelyEveryoneFaultedChoice(state *NodeSelectorState) (*CurrentIndexAndNode, error) {
	// if there are all marked as failed, we'll chose the first
	// one so the user will get an error (or recover :-) );
	if len(state.nodes) == 0 {
		return nil, newAllTopologyNodesDownError("There are no nodes in the topology at all")
	}

	return NewCurrentIndexAndNode(0, state.nodes[0]), nil
}

func (s *NodeSelector) getNodeBySessionID(sessionId int) (*CurrentIndexAndNode, error) {
	state := s._state
	index := sessionId % len(state.topology.Nodes)

	for i := index; i < len(state.failures); i++ {
		if state.failures[i].get() == 0 && state.nodes[i].ServerRole == ServerNodeRoleMember {
			return NewCurrentIndexAndNode(i, state.nodes[i]), nil
		}
	}

	for i := 0; i < index; i++ {
		if state.failures[i].get() == 0 && state.nodes[i].ServerRole == ServerNodeRoleMember {
			return NewCurrentIndexAndNode(i, state.nodes[i]), nil
		}
	}

	return s.getPreferredNode()
}

func (s *NodeSelector) getFastestNode() (*CurrentIndexAndNode, error) {
	state := s._state
	if state.failures[state.fastest].get() == 0 && state.nodes[state.fastest].ServerRole == ServerNodeRoleMember {
		return NewCurrentIndexAndNode(state.fastest, state.nodes[state.fastest]), nil
	}

	// if the fastest node has failures, we'll immediately schedule
	// another run of finding who the fastest node is, in the meantime
	// we'll just use the server preferred node or failover as usual

	s.switchToSpeedTestPhase()
	return s.getPreferredNode()
}

func (s *NodeSelector) restoreNodeIndex(nodeIndex int) {
	state := s._state
	if len(state.failures) < nodeIndex {
		return // the state was changed and we no longer have it?
	}

	state.failures[nodeIndex].set(0)
}

/*
TODO: not used, remove?
func nodeSelectorThrowEmptyTopology() error {
	return newIllegalStateError("Empty database topology, this shouldn't happen.")
}
*/

func (s *NodeSelector) switchToSpeedTestPhase() {
	state := s._state

	if !state.speedTestMode.compareAndSet(0, 1) {
		return
	}

	for i := 0; i < len(state.fastestRecords); i++ {
		state.fastestRecords[i] = 0
	}

	state.speedTestMode.incrementAndGet()
}

func (s *NodeSelector) inSpeedTestPhase() bool {
	return s._state.speedTestMode.get() > 1
}

func (s *NodeSelector) recordFastest(index int, node *ServerNode) {
	state := s._state
	stateFastest := state.fastestRecords

	// the following two checks are to verify that things didn't move
	// while we were computing the fastest node, we verify that the index
	// of the fastest node and the identity of the node didn't change during
	// our check
	if index < 0 || index >= len(stateFastest) {
		return
	}

	if node != state.nodes[index] {
		return
	}

	stateFastest[index]++
	if stateFastest[index] >= 10 {
		s.selectFastest(state, index)
	}

	if state.speedTestMode.incrementAndGet() <= len(state.nodes)*10 {
		return
	}

	//too many concurrent speed tests are happening
	maxIndex := s.findMaxIndex(state)
	s.selectFastest(state, maxIndex)
}

func (s *NodeSelector) findMaxIndex(state *NodeSelectorState) int {
	stateFastest := state.fastestRecords
	maxIndex := 0
	maxValue := 0

	for i := 0; i < len(stateFastest); i++ {
		if maxValue >= stateFastest[i] {
			continue
		}

		maxIndex = i
		maxValue = stateFastest[i]
	}

	return maxIndex
}

func (s *NodeSelector) selectFastest(state *NodeSelectorState, index int) {
	state.fastest = index
	state.speedTestMode.set(0)

	if s._updateFastestNodeTimer != nil {
		s._updateFastestNodeTimer.Reset(time.Minute)
	} else {
		f := func() {
			s._updateFastestNodeTimer = nil
			s.switchToSpeedTestPhase()
		}
		s._updateFastestNodeTimer = time.AfterFunc(time.Minute, f)
	}
}

func (s *NodeSelector) scheduleSpeedTest() {
	s.switchToSpeedTestPhase()
}

func (s *NodeSelector) Close() {
	if s._updateFastestNodeTimer != nil {
		s._updateFastestNodeTimer.Stop()
		s._updateFastestNodeTimer = nil
	}
}

type NodeSelectorState struct {
	topology       *Topology
	nodes          []*ServerNode
	failures       []atomicInteger
	fastestRecords []int
	fastest        int
	speedTestMode  atomicInteger
}

func NewNodeSelectorState(topology *Topology) *NodeSelectorState {
	nodes := topology.Nodes
	res := &NodeSelectorState{
		topology: topology,
		nodes:    nodes,
	}
	failures := make([]atomicInteger, len(nodes))
	res.failures = failures
	res.fastestRecords = make([]int, len(nodes))
	return res
}
