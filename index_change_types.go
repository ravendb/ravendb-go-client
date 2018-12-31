package ravendb

type IndexChangeTypes = string

const (
	IndexChangeNone                   = "None"
	IndexChangeBatchCompleted         = "BatchCompleted"
	IndexChangeIndexAdded             = "IndexAdded"
	IndexChangeIndexRemoved           = "IndexRemoved"
	IndexChangeIndexDemotedToIdle     = "IndexDemotedToIdle"
	IndexChangeIndexPromotedFromIdle  = "IndexPromotedFromIdle"
	IndexChangeIndexDemotedToDisabled = "IndexDemotedToDisabled"
	IndexChangeIndexMarkedAsErrored   = "IndexMarkedAsErrored"
	IndexChangeSideBySideReplace      = "SideBySideReplace"
	IndexChangeRenamed                = "Renamed"
	IndexChangeIndexPaused            = "IndexPaused"
	IndexChangeLockModeChanged        = "LockModeChanged"
	IndexChangePriorityChanged        = "PriorityChanged"
)
