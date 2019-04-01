package ravendb

// SessionOptions describes session options
type SessionOptions struct {
	Database        string
	NoTracking      bool
	NoCaching       bool
	RequestExecutor *RequestExecutor
	TransactionMode TransactionMode
}
