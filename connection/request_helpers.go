package connection

import (
	"time"
	"../data"
)

const MAX_RESPONSES uint8 = 5

type ServerNode struct{
	Url, Database, ApiKey, CurrentToken, IsFailed,
	isRateSurpassed string
	responseTime []time.Duration
}

type Topology struct{
	Etag, Sla int
	LeaderNode ServerNode
	ReadBehavior data.ReadBehaviour
	WriteBehaviour data.WriteBehaviour
	Nodes []ServerNode
}

func NewServerNode(url, database, apiKey, currentToken, isFailed string) *ServerNode{

}

func NewTopology(etag int, leaderNode ServerNode, readBehaviour data.ReadBehaviour,
	writeBehaviour data.WriteBehaviour, nodes []ServerNode, sla int) *Topology{
	if readBehaviour == data.ReadBehaviour{}{
		readBehaviour := data.NewReadBehaviour(data.LEADER_ONLY)
	}
	if writeBehaviour == data.WriteBehaviour{}{
		writeBehaviour := data.NewWriteBehaviour(data.LEADER_ONLY)
	}
	if sla == 0{
		sla = 100 /1000
	}

	return &Topology{etag, sla, leaderNode,
		readBehaviour, writeBehaviour,
		nodes,
	}
}

func (sn ServerNode) ResponseTime() []time.Duration{
	return sn.responseTime
}

func (sn ServerNode) SetResponseTime(val time.Duration){
	sn.responseTime[uint8(len(sn.responseTime)) % MAX_RESPONSES] = val
}
