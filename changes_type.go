package ravendb

type ChangesType = string

const (
	ChangesType_DOCUMENT  = "DOCUMENT"
	ChangesType_INDEX     = "INDEX"
	ChangesType_OPERATION = "OPERATION"
)
