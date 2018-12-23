package ravendb

// SessionOptions describes Session options
type SessionOptions struct {
	Database        string
	RequestExecutor *RequestExecutor
}
