package ravendb

type IndexState = string

const (
	IndexStateNormal   = "Normal"
	IndexStateDisabled = "Disabled"
	IndexStateIdle     = "Idle"
	IndexStateError    = "Error"
)
