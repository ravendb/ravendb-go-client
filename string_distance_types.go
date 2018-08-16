package ravendb

type StringDistanceTypes = string

const (
	StringDistanceTypes_NONE         = "None"
	StringDistanceTypes_DEFAULT      = "Default"
	StringDistanceTypes_LEVENSHTEIN  = "Levenshtein"
	StringDistanceTypes_JARO_WINKLER = "JaroWinkler"
	StringDistanceTypes_N_GRAM       = "NGram"
)
