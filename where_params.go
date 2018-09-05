package ravendb

// WhereParams are parameters for the Where Equals call
type WhereParams struct {
	fieldName      string
	value          Object
	allowWildcards bool
	isNestedPath   bool
	isExact        bool
}
