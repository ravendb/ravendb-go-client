package ravendb

// Query represents arguments for query.
// You can provide either name of the collection to query
// or name of the index to query
// TODO: not a great name. Maybe replace with methods on DocumentQuery:
// InCollection() and InIndex()
type Query struct {
	Collection string
	IndexName  string
}
