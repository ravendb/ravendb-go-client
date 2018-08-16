package ravendb

import "time"

type QueryOperationOptions struct {
	_maxOpsPerSecond int

	allowStale bool

	staleTimeout time.Duration

	retrieveDetails bool
}

func NewQueryOperationOptions() *QueryOperationOptions {
	return &QueryOperationOptions{}
}

func (o *QueryOperationOptions) isAllowStale() bool {
	return o.allowStale
}

func (o *QueryOperationOptions) setAllowStale(allowStale bool) {
	o.allowStale = allowStale
}

func (o *QueryOperationOptions) getStaleTimeout() time.Duration {
	return o.staleTimeout
}

func (o *QueryOperationOptions) setStaleTimeout(staleTimeout time.Duration) {
	o.staleTimeout = staleTimeout
}

func (o *QueryOperationOptions) getMaxOpsPerSecond() int {
	return o._maxOpsPerSecond
}

// set to 0 to disable
func (o *QueryOperationOptions) setMaxOpsPerSecond(maxOpsPerSecond int) {
	o._maxOpsPerSecond = maxOpsPerSecond
}

func (o *QueryOperationOptions) isRetrieveDetails() bool {
	return o.retrieveDetails
}

func (o *QueryOperationOptions) setRetrieveDetails(retrieveDetails bool) {
	o.retrieveDetails = retrieveDetails
}
