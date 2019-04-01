package ravendb

type FieldIndexing string

const (
	FieldIndexingNo     = "No"
	FieldIndexingSearch = "Search"
	FieldIndexingExact  = "Exact"
	// Index the tokens produced by running the field's value through an Analyzer (same as Search),
	// store them in index and track term vector positions and offsets. This is mandatory when highlighting is used.
	FieldIndexingHighlighting = "Highlighting"
	FieldIndexingDefault      = "Default"
)
