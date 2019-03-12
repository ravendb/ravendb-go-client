package ravendb

// QueryData represents
type QueryData struct {
	// Fields lists fields to be selected from queried document
	Fields []string
	// Projections lists fields in the result entity
	Projections []string

	// TODO: should those be exposed as well?
	fromAlias        string
	declareToken     *declareToken
	loadTokens       []*loadToken
	isCustomFunction bool
}
