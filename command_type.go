package ravendb

// using type alias for easy json serialization
type CommandType = string

// Note: this is enum in Java but those are serialized to json as strings so
// making them strings is better in Go
const (
	//CommandNone                = "NONE"
	CommandPut                 = "PUT"
	CommandPatch               = "PATCH"
	CommandDelete              = "DELETE"
	CommandAttachmentPut       = "ATTACHMENT_PUT"
	CommandAttachmentDelete    = "ATTACHMENT_DELETE"
	CommandClientAnyCommand    = "CLIENT_ANY_COMMAND"
	CommandClientNotAttachment = "CLIENT_NOT_ATTACHMENT"
)
