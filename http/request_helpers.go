package http

import (
	"time"
	"../data"
)

const MAX_RESPONSES uint8 = 5
const RATE_SUPRESSION_COEF = 0.75

type ServerNode struct{
	Url, Database, ApiKey, CurrentToken, ClusterToken string
	IsFailed, isRateSurpassed bool
	responseTime []time.Duration
}

type Topology struct{
	Etag int64
	Sla int
	LeaderNode ServerNode
	ReadBehavior data.ReadBehaviour
	WriteBehaviour data.WriteBehaviour
	Nodes []ServerNode
}

type NodeStatus struct{
	tickerPeriod time.Duration
	ticker time.Ticker
	shouldPerformNextTick chan(bool)
	requestExecutor RequestExecutor
	NodeIndex int
	Node ServerNode
}

func NewServerNode(url, database, apiKey, currentToken string, isFailed bool) *ServerNode{
	return &ServerNode{url, database, apiKey, currentToken, isFailed,
		false, nil}
}

func NewTopology(etag int64, leaderNode ServerNode, readBehaviour data.ReadBehaviour,
	writeBehaviour data.WriteBehaviour, nodes []ServerNode, sla int) *Topology{

	if readBehaviour.IsEmpty(){
		rb, _ := data.NewReadBehaviour(data.LEADER_ONLY)
		readBehaviour = *rb
	}
	if writeBehaviour.IsEmpty(){
		wb, _ := data.NewWriteBehaviour(data.LEADER_ONLY)
		writeBehaviour = *wb
	}
	if sla == 0{
		sla = 100 /1000
	}

	return &Topology{etag, sla, leaderNode,
		readBehaviour, writeBehaviour,
		nodes,
	}
}

func NewNodeStatus(executor RequestExecutor, nodeIndex int, node ServerNode) (*NodeStatus, error){
	period := time.Duration(time.Millisecond*100)
	ticker := time.NewTicker(period)
	return &NodeStatus{requestExecutor: executor, NodeIndex: nodeIndex, Node: node, timerPeriod:period, ticker:*ticker}, nil
}

func (sn ServerNode) ResponseTime() []time.Duration{
	return sn.responseTime
}

func (sn ServerNode) SetResponseTime(val time.Duration){
	sn.responseTime[uint8(len(sn.responseTime)) % MAX_RESPONSES] = val
}

func (sn ServerNode) Ewma() uint{
	responseTimeCount := uint(len(sn.responseTime))
	if responseTimeCount > 0{
		var totalTime uint
		for _, measurement := range sn.responseTime{
			totalTime += uint(measurement)
		}
		return totalTime / responseTimeCount
	}
	return 0
}

func (sn ServerNode) IsRateSupressed(requestTimeSlaTresholdInMilliseconds uint) bool{
	supressionThreshold := float64(requestTimeSlaTresholdInMilliseconds)
	if sn.isRateSurpassed {
		supressionThreshold *= RATE_SUPRESSION_COEF
	}
	sn.isRateSurpassed = float64(sn.Ewma()) >= supressionThreshold
	return sn.isRateSurpassed
}

func (ns NodeStatus) StartTicker(){
	go func(){
		for {
			select {
			case <-ns.ticker.C:
				ns.requestExecutor.CheckNodeStatusCallback(&ns)
			case <-ns.shouldPerformNextTick:
				ns.ticker.Stop()
				return
			default:
				time.Sleep(ns.getTickerPeriod())
			}
		}
	}()
}

func (ns NodeStatus) StopTicker(){
	close(ns.shouldPerformNextTick)
}

func (ns NodeStatus) getTickerPeriod() time.Duration{
	if ns.tickerPeriod >= time.Second * 5{
		return time.Second * 5
	}

	ns.tickerPeriod += time.Millisecond * 100
	return ns.tickerPeriod
}

func (ns NodeStatus) UpdateTicker(){
	ns.ticker.Stop()
	ns.ticker = *time.NewTicker(ns.getTickerPeriod())
}