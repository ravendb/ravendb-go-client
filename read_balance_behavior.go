package ravendb

type ReadBalanceBehavior = string

const (
	ReadBalanceBehavior_NONE         = "None" // TODO: should this be "" ?
	ReadBalanceBehavior_ROUND_ROBIN  = "RoundRobin"
	ReadBalanceBehavior_FASTEST_NODE = "FastestNode"
)
