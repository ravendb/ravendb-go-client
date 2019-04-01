package ravendb

import "time"

type ReplicationBatchOptions struct {
	waitForReplicas                 bool
	numberOfReplicasToWaitFor       int
	waitForReplicasTimeout          time.Duration
	majority                        bool
	throwOnTimeoutInWaitForReplicas bool
}

func NewReplicationBatchOptions() *ReplicationBatchOptions {
	return &ReplicationBatchOptions{
		throwOnTimeoutInWaitForReplicas: true,
	}
}
