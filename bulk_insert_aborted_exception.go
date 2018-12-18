package ravendb

type BulkInsertAbortedException struct {
	RavenError
}

func NewBulkInsertAbortedException(format string, args ...interface{}) *BulkInsertAbortedException {
	res := &BulkInsertAbortedException{
		RavenError: *newRavenError(format, args...),
	}
	return res
}
