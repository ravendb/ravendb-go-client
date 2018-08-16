package ravendb

type QueryData struct {
	fields           []string
	projections      []string
	fromAlias        string
	declareToken     *DeclareToken
	loadTokens       []*LoadToken
	isCustomFunction bool
}

func (d *QueryData) getFields() []string {
	return d.fields
}

func (d *QueryData) setFields(fields []string) {
	d.fields = fields
}

func (d *QueryData) getProjections() []string {
	return d.projections
}

func (d *QueryData) setProjections(projections []string) {
	d.projections = projections
}

func (d *QueryData) getFromAlias() string {
	return d.fromAlias
}

func (d *QueryData) setFromAlias(fromAlias string) {
	d.fromAlias = fromAlias
}

func (d *QueryData) getDeclareToken() *DeclareToken {
	return d.declareToken
}

func (d *QueryData) setDeclareToken(declareToken *DeclareToken) {
	d.declareToken = declareToken
}

func (d *QueryData) getLoadTokens() []*LoadToken {
	return d.loadTokens
}

func (d *QueryData) setLoadTokens(loadTokens []*LoadToken) {
	d.loadTokens = loadTokens
}

func (d *QueryData) IsCustomFunction() bool {
	return d.isCustomFunction
}

func (d *QueryData) setCustomFunction(customFunction bool) {
	d.isCustomFunction = customFunction
}

func NewQueryData(fields []string, projections []string) *QueryData {
	return NewQueryDataWithTokens(fields, projections, "", nil, nil, false)
}

func NewQueryDataWithTokens(fields []string, projections []string, fromAlias string, declareToken *DeclareToken, loadTokens []*LoadToken, isCustomFunction bool) *QueryData {
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
