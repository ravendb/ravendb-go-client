package ravendb

type IndexState = string

const (
	IndexState_NORMAL   = "Normal"
	IndexState_DISABLED = "Disabled"
	IndexState_IDLE     = "Idle"
	IndexState_ERROR    = "Error"
)
