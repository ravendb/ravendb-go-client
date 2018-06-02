package ravendb

type ReadBalanceBehavior = string

const (
	ReadBalanceBehavior_NONE         = "NONE" // TODO: should this be "" ?
	ReadBalanceBehavior_ROUND_ROBIN  = "ROUND_ROBIN"
	ReadBalanceBehavior_FASTEST_NODE = "FASTEST_NODE"
)
