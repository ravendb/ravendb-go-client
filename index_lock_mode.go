package ravendb

type IndexLockMode = string

const (
	IndexLockModeUnlock       = "Unlock"
	IndexLockModeLockedIgnore = "LockedIgnore"
	IndexLockModeLockedError  = "LockedError"
)
