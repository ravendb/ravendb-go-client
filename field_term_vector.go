package ravendb

type FieldTermVector = string

const (
	FieldTermVectorNo                      = "No"
	FieldTermVectorYes                     = "Yes"
	FieldTermVectorWithPositions           = "WithPositions"
	FieldTermVectorWithOffsets             = "WithOffsets"
	FieldTermVectorWithPositionsAndOffsets = "WithPositionsAndOffsets"
)
