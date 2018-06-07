package ravendb

type IndexChangeTypes = string

const (
	IndexChangeTypes_NONE                      = "NONE"
	IndexChangeTypes_BATCH_COMPLETED           = "BATCH_COMPLETED"
	IndexChangeTypes_INDEX_ADDED               = "INDEX_ADDED"
	IndexChangeTypes_INDEX_REMOVED             = "INDEX_REMOVED"
	IndexChangeTypes_INDEX_DEMOTED_TO_IDLE     = "INDEX_DEMOTED_TO_IDLE"
	IndexChangeTypes_INDEX_PROMOTED_FROM_IDLE  = "INDEX_PROMOTED_FROM_IDLE"
	IndexChangeTypes_INDEX_DEMOTED_TO_DISABLED = "INDEX_DEMOTED_TO_DISABLED"
	IndexChangeTypes_INDEX_MARKED_AS_ERRORED   = "INDEX_MARKED_AS_ERRORED"
	IndexChangeTypes_SIDE_BY_SIDE_REPLACE      = "SIDE_BY_SIDE_REPLACE"
	IndexChangeTypes_RENAMED                   = "RENAMED"
	IndexChangeTypes_INDEX_PAUSED              = "INDEX_PAUSED"
	IndexChangeTypes_LOCK_MODE_CHANGED         = "LOCK_MODE_CHANGED"
	IndexChangeTypes_PRIORITY_CHANGED          = "PRIORITY_CHANGED"
)
