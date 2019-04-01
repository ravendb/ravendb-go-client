package ravendb

// TODO: make it less alias'y
type CounterOperationType = string

const (
	CounterOperationType_NONE      = "NONE"
	CounterOperationType_INCREMENT = "INCREMENT"
	CounterOperationType_DELETE    = "DELETE"
	CounterOperationType_GET       = "GET"
	CounterOperationType_PUT       = "PUT"
)
