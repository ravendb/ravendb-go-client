package ravendb

type QueryData struct {
	fields           []string
	projections      []string
	fromAlias        string
	declareToken     *declareToken
	loadTokens       []*loadToken
	isCustomFunction bool
}

func NewQueryData(fields []string, projections []string) *QueryData {
	return NewQueryDataWithTokens(fields, projections, "", nil, nil, false)
}

func NewQueryDataWithTokens(fields []string, projections []string, fromAlias string, declareToken *declareToken, loadTokens []*loadToken, isCustomFunction bool) *QueryData {
	return &QueryData{
		fields:           fields,
		projections:      projections,
		fromAlias:        fromAlias,
		declareToken:     declareToken,
		loadTokens:       loadTokens,
		isCustomFunction: isCustomFunction,
	}
}

func QueryData_customFunction(alias string, fn string) *QueryData {
	return NewQueryDataWithTokens([]string{fn}, nil, alias, nil, nil, false)
}
