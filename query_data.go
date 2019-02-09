package ravendb

type queryData struct {
	fields           []string
	projections      []string
	fromAlias        string
	declareToken     *declareToken
	loadTokens       []*loadToken
	isCustomFunction bool
}
