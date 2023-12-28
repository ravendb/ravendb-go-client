package ravendb

// SessionInfo describes a session
type SessionInfo struct {
	SessionID                   int
	lastClusterTransactionIndex *int64
}
