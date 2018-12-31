package ravendb

// ReadBalanceBehavior defines the type of read balancing
type ReadBalanceBehavior = string

const (
	ReadBalanceBehaviorNone        = "None"
	ReadBalanceBehaviorRoundRobin  = "RoundRobin"
	ReadBalanceBehaviorFastestNode = "FastestNode"
)
