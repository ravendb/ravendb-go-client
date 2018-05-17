package ravendb

type CommandType int

// TODO: change names to be Go like like CommandNone
const (
	NONE CommandType = iota
	PUT
	PATCH
	DELETE
	ATTACHMENT_PUT
	ATTACHMENT_DELETE

	CLIENT_ANY_COMMAND
	CLIENT_NOT_ATTACHMENT
)
