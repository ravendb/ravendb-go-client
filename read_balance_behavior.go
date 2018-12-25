package ravendb

// ReadBalanceBehavior defines the type of read balancing
type ReadBalanceBehavior = string

const (
	ReadBalanceBehaviorNone        = "None" // TODO: should this be "" ?
	ReadBalanceBehaviorRoundRobin  = "RoundRobin"
	ReadBalanceBehaviorFastestNode = "FastestNode"
)
