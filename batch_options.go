package ravendb

// BatchOptions describes options for batch operations
type BatchOptions struct {
	replicationOptions *ReplicationBatchOptions
	indexOptions       *IndexBatchOptions
}

// NewBatchOptions returns new BatchOptions
func NewBatchOptions() *BatchOptions {
	return &BatchOptions{}
}
