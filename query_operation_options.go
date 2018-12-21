package ravendb

import "time"

// QueryOperationOptions represents options for query operation
type QueryOperationOptions struct {
	maxOpsPerSecond int
	allowStale      bool
	staleTimeout    time.Duration
	retrieveDetails bool
}
