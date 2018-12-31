package ravendb

type StringDistanceTypes = string

const (
	StringDistanceNone        = "None"
	StringDistanceDefault     = "Default"
	StringDistanceLevenshtein = "Levenshtein"
	StringDistanceJaroWinkler = "JaroWinkler"
	StringDistanceNGram       = "NGram"
)
