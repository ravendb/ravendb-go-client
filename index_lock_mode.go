package ravendb

type IndexLockMode = string

const (
	IndexLockMode_UNLOCK        = "UNLOCK"
	IndexLockMode_LOCKED_IGNORE = "LOCKED_IGNORE"
	IndexLockMode_LOCKED_ERROR  = "LOCKED_ERROR"
)
