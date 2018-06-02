package ravendb

// SessionInfo describes a session
type SessionInfo struct {
	SessionID int
}

func (si *SessionInfo) getSessionId() int {
	return si.SessionID
}
