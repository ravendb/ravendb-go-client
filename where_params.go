package ravendb

// whereParams are parameters for the Where Equals call
type whereParams struct {
	fieldName      string
	value          Object
	allowWildcards bool
	isNestedPath   bool
	isExact        bool
}
