package ravendb

// using type alias for easy json serialization
type CommandType = string

// TODO: change names to be Go like like CommandNone
// Note: this is enum in Java but those are serialized to json as strings so
// making them strings is better in Go
const (
	CommandType_NONE              = "NONE"
	CommandType_PUT               = "PUT"
	CommandType_PATCH             = "PATCH"
	CommandType_DELETE            = "DELETE"
	CommandType_ATTACHMENT_PUT    = "ATTACHMENT_PUT"
	CommandType_ATTACHMENT_DELETE = "ATTACHMENT_DELETE"

	CommandType_CLIENT_ANY_COMMAND    = "CLIENT_ANY_COMMAND"
	CommandType_CLIENT_NOT_ATTACHMENT = "CLIENT_NOT_ATTACHMENT"
)
