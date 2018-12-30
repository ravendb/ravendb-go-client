package ravendb

// QueryData represents arguments to SelectFields call
type QueryData struct {
	Fields           []string
	Projections      []string
	FromAlias        string
	DeclareToken     *declareToken
	LoadTokens       []*loadToken
	IsCustomFunction bool
}

// NewQueryData returns new QueryData with given fields and projections
func NewQueryData(fields []string, projections []string) *QueryData {
	return &QueryData{
		Fields:      fields,
		Projections: projections,
	}
}

// NewQueryDataWithCustomFunction returns new QueryData for a given function
// and alias
func NewQueryDataWithCustomFunction(alias string, fn string) *QueryData {
	return &QueryData{
		Fields:    []string{fn},
		FromAlias: alias,
	}
}
