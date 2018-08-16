package ravendb

type FieldTermVector = string

const (
	FieldTermVector_NO                         = "No"
	FieldTermVector_YES                        = "Yes"
	FieldTermVector_WITH_POSITIONS             = "WithPositions"
	FieldTermVector_WITH_OFFSETS               = "WithOffsets"
	FieldTermVector_WITH_POSITIONS_AND_OFFSETS = "WithPositionsAndOffsets"
)
