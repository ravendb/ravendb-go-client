package ravendb

// TODO: only used in ConnectionString which is unused
type ConnectionStringType = string

const (
	ConnectionStringTypeNone  = "None"
	ConnectionStringTypeRaven = "Raven"
	ConnectionStringTypeSQL   = "Sql"
)
