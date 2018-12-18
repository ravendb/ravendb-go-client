package ravendb

// BulkInsertAbortedError represents "bulk insert aborted" error
type BulkInsertAbortedError struct {
	RavenError
}

func newBulkInsertAbortedError(format string, args ...interface{}) *BulkInsertAbortedError {
	res := &BulkInsertAbortedError{
		RavenError: *newRavenError(format, args...),
	}
	return res
}
