package ravendb

// SessionOptions describes session options
type SessionOptions struct {
	Database        string
	RequestExecutor *RequestExecutor
}
