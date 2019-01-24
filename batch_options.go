package ravendb

import "time"

// BatchOptions describes options for batch operations
type BatchOptions struct {
	waitForReplicas                 bool
	numberOfReplicasToWaitFor       int
	waitForReplicasTimeout          time.Duration
	majority                        bool
	throwOnTimeoutInWaitForReplicas bool

	waitForIndexes                 bool
	waitForIndexesTimeout          time.Duration
	throwOnTimeoutInWaitForIndexes bool
	waitForSpecificIndexes         []string
}

// NewBatchOptions returns new BatchOptions
func NewBatchOptions() *BatchOptions {
	return &BatchOptions{
		throwOnTimeoutInWaitForReplicas: true,
	}
}
