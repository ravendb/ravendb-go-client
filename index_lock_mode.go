package ravendb

type IndexLockMode = string

const (
	IndexLockMode_UNLOCK        = "Unlock"
	IndexLockMode_LOCKED_IGNORE = "LockedIgnore"
	IndexLockMode_LOCKED_ERROR  = "LockedError"
)
