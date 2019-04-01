package ravendb

// SessionInfo describes a session
type SessionInfo struct {
	LastClusterTransactionIndex int64
	SessionID                   int
	NoCaching                   bool
}
