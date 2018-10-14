package ravendb

// StartsWithArgs contains arguments for functions that return documents
// matching one or more options.
// StartsWith is required, other fields are optional.
type StartsWithArgs struct {
	StartsWith string // TODO: Prefix?
	Matches    string
	Start      int
	PageSize   int
	StartAfter string

	Exclude string
}
