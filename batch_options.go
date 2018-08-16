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

// TODO: remove getters/setters
func (o *BatchOptions) isWaitForReplicas() bool {
	return o.waitForReplicas
}

func (o *BatchOptions) setWaitForReplicas(waitForReplicas bool) {
	o.waitForReplicas = waitForReplicas
}

func (o *BatchOptions) getNumberOfReplicasToWaitFor() int {
	return o.numberOfReplicasToWaitFor
}

func (o *BatchOptions) setNumberOfReplicasToWaitFor(numberOfReplicasToWaitFor int) {
	o.numberOfReplicasToWaitFor = numberOfReplicasToWaitFor
}

func (o *BatchOptions) getWaitForReplicasTimeout() time.Duration {
	return o.waitForReplicasTimeout
}

func (o *BatchOptions) setWaitForReplicasTimeout(waitForReplicasTimeout time.Duration) {
	o.waitForReplicasTimeout = waitForReplicasTimeout
}

func (o *BatchOptions) isMajority() bool {
	return o.majority
}

func (o *BatchOptions) setMajority(majority bool) {
	o.majority = majority
}

func (o *BatchOptions) isThrowOnTimeoutInWaitForReplicas() bool {
	return o.throwOnTimeoutInWaitForReplicas
}

func (o *BatchOptions) setThrowOnTimeoutInWaitForReplicas(throwOnTimeoutInWaitForReplicas bool) {
	o.throwOnTimeoutInWaitForReplicas = throwOnTimeoutInWaitForReplicas
}

func (o *BatchOptions) isWaitForIndexes() bool {
	return o.waitForIndexes
}

func (o *BatchOptions) setWaitForIndexes(waitForIndexes bool) {
	o.waitForIndexes = waitForIndexes
}

func (o *BatchOptions) getWaitForIndexesTimeout() time.Duration {
	return o.waitForIndexesTimeout
}

func (o *BatchOptions) setWaitForIndexesTimeout(waitForIndexesTimeout time.Duration) {
	o.waitForIndexesTimeout = waitForIndexesTimeout
}

func (o *BatchOptions) isThrowOnTimeoutInWaitForIndexes() bool {
	return o.throwOnTimeoutInWaitForIndexes
}

func (o *BatchOptions) setThrowOnTimeoutInWaitForIndexes(throwOnTimeoutInWaitForIndexes bool) {
	o.throwOnTimeoutInWaitForIndexes = throwOnTimeoutInWaitForIndexes
}

func (o *BatchOptions) getWaitForSpecificIndexes() []string {
	return o.waitForSpecificIndexes
}

func (o *BatchOptions) setWaitForSpecificIndexes(waitForSpecificIndexes []string) {
	o.waitForSpecificIndexes = waitForSpecificIndexes
}
