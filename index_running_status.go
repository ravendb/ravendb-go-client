package ravendb

type IndexRunningStatus = string

const (
	IndexRunningStatusRunning  = "Running"
	IndexRunningStatusPaused   = "Paused"
	IndexRunningStatusDisabled = "Disabled"
)
