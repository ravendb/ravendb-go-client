package ravendb

import "time"

// QueryOperationOptions represents options for query operation
type QueryOperationOptions struct {
	maxOpsPerSecond int
	AllowStale      bool
	staleTimeout    time.Duration
	retrieveDetails bool
}
