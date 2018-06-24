package ravendb

type IndexRunningStatus = string

const (
	IndexRunningStatus_RUNNING  = "Running"
	IndexRunningStatus_PAUSED   = "Paused"
	IndexRunningStatus_DISABLED = "Disabled"
)
