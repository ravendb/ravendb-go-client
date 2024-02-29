package ravendb

import "time"

// QueryOperationOptions represents options for query operation
type QueryOperationOptions struct {
	MaxOpsPerSecond int
	AllowStale      bool
	StaleTimeout    time.Duration
	RetrieveDetails bool
}
