package ravendb

type DocumentChangeTypes = string

const (
	DocumentChangeTypes_NONE     = "NONE"
	DocumentChangeTypes_PUT      = "PUT"
	DocumentChangeTypes_DELETE   = "DELETE"
	DocumentChangeTypes_CONFLICT = "CONFLICT"
	DocumentChangeTypes_COMMON   = "COMMON"
)
