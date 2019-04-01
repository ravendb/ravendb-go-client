package ravendb

type DocumentChangeTypes = string

const (
	DocumentChangeNone     = "None"
	DocumentChangePut      = "Put"
	DocumentChangeDelete   = "Delete"
	DocumentChangeConflict = "Conflict"
	DocumentChangeCommon   = "Common"
	DocumentChangeCounter  = "Counter"
)
