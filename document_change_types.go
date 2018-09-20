package ravendb

type DocumentChangeTypes = string

const (
	DocumentChangeTypes_NONE     = "None"
	DocumentChangeTypes_PUT      = "Put"
	DocumentChangeTypes_DELETE   = "Delete"
	DocumentChangeTypes_CONFLICT = "Conflict"
	DocumentChangeTypes_COMMON   = "Common"
)
