package ravendb

var _ FacetBase = &Facet{}

type Facet struct {
	FacetBaseCommon

	fieldName string
}

func (f *Facet) getFieldName() string {
	return f.fieldName
}

func (f *Facet) setFieldName(fieldName string) {
	f.fieldName = fieldName
}

func (f *Facet) toFacetToken(addQueryParameter func(Object) string) *FacetToken {
	return FacetToken_createWithFacet(f, addQueryParameter)
}
