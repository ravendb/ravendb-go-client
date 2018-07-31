package ravendb

// WhereParams are parameters for the Where Equals call
type WhereParams struct {
	fieldName      string
	value          Object
	allowWildcards bool
	nestedPath     bool
	exact          bool
}

func NewWhereParams() *WhereParams {
	return &WhereParams{
		nestedPath:     false,
		allowWildcards: false,
	}
}

func (p *WhereParams) getFieldName() string {
	return p.fieldName
}

func (p *WhereParams) setFieldName(fieldName string) {
	assertValidFieldName(fieldName)
	p.fieldName = fieldName
}

func (p *WhereParams) getValue() Object {
	return p.value
}

func (p *WhereParams) setValue(value Object) {
	p.value = value
}

func (p *WhereParams) isAllowWildcards() bool {
	return p.allowWildcards
}

func (p *WhereParams) setAllowWildcards(allowWildcards bool) {
	p.allowWildcards = allowWildcards
}

func (p *WhereParams) isNestedPath() bool {
	return p.nestedPath
}

func (p *WhereParams) setNestedPath(nestedPath bool) {
	p.nestedPath = nestedPath
}

func (p *WhereParams) isExact() bool {
	return p.exact
}

func (p *WhereParams) setExact(exact bool) {
	p.exact = exact
}
