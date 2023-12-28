package ravendb

// SessionOptions describes session options
type SessionOptions struct {
	Database                                            string
	RequestExecutor                                     *RequestExecutor
	TransactionMode                                     int
	DisableAtomicDocumentWritesInClusterWideTransaction *bool
}

func assertTransactionMode(transactionMode int) error {
	if transactionMode == TransactionMode_SingleNode || transactionMode == TransactionMode_ClusterWide {
		return nil
	}

	return newIllegalStateError("transactionMode has to be set as `TransactionMode_SingleNode` or 'TransactionMode_ClusterWide`.")
}
