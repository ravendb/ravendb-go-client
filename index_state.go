package ravendb

type IndexState = string

const (
	IndexState_NORMAL   = "NORMAL"
	IndexState_DISABLED = "DISABLED"
	IndexState_IDLE     = "IDLE"
	IndexState_ERROR    = "ERROR"
)
