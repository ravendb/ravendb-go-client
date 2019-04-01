package ravendb

type TransactionMode string

const (
	TransactionMode_SINGLE_NODE  = "SINGLE_NODE"
	TransactionMode_CLUSTER_WIDE = "CLUSTER_WIDE"
)
