package ravendb

type BulkInsertAbortedException struct {
	RavenException
}

func NewBulkInsertAbortedException(format string, args ...interface{}) *BulkInsertAbortedException {
	res := &BulkInsertAbortedException{
		RavenException: *NewRavenException(format, args),
	}
	return res
}
