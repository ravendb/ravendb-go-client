package ravendb

type IndexChangeTypes = string

const (
	IndexChangeTypes_NONE                      = "None"
	IndexChangeTypes_BATCH_COMPLETED           = "BatchCompleted"
	IndexChangeTypes_INDEX_ADDED               = "IndexAdded"
	IndexChangeTypes_INDEX_REMOVED             = "IndexRemoved"
	IndexChangeTypes_INDEX_DEMOTED_TO_IDLE     = "IndexDemotedToIdle"
	IndexChangeTypes_INDEX_PROMOTED_FROM_IDLE  = "IndexPromotedFromIdle"
	IndexChangeTypes_INDEX_DEMOTED_TO_DISABLED = "IndexDemotedToDisabled"
	IndexChangeTypes_INDEX_MARKED_AS_ERRORED   = "IndexMarkedAsErrored"
	IndexChangeTypes_SIDE_BY_SIDE_REPLACE      = "SideBySideReplace"
	IndexChangeTypes_RENAMED                   = "Renamed"
	IndexChangeTypes_INDEX_PAUSED              = "IndexPaused"
	IndexChangeTypes_LOCK_MODE_CHANGED         = "LockModeChanged"
	IndexChangeTypes_PRIORITY_CHANGED          = "PriorityChanged"
)
