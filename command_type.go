package ravendb

// using type alias for easy json serialization
type CommandType = string

// Note: this is enum in Java but those are serialized to json as strings so
// making them strings is better in Go
const (
	CommandNone                  = "None"
	CommandPut                   = "PUT"
	CommandPatch                 = "PATCH"
	CommandDelete                = "DELETE"
	CommandAttachmentPut         = "AttachmentPUT"
	CommandAttachmentDelete      = "AttachmentDELETE"
	CommandAttachmentMove        = "AttachmentMOVE"
	CommandAttachmentCopy        = "AttachmentCOPY"
	CommandCompareExchangePut    = "CompareExchangePUT"
	CommandCompareExchangeDelete = "CompareExchangeDELETE"

	CommandCounters = "Counters"

	CommandClientAnyCommand    = "CLIENT_ANY_COMMAND"
	CommandClientNotAttachment = "CLIENT_MODIFY_DOCUMENT_COMMAND"
)

func parseCSharpValue(input string) CommandType {
	// TODO: this is not necessary but we could validate that
	// input is one of the valid values
	return input
}
