package ravendb

var _ FacetBase = &Facet{}

type Facet struct {
	FacetBaseCommon

	fieldName string
}

func NewFacet() *Facet {
	return &Facet{
		FacetBaseCommon: NewFacetBaseCommon(),
	}
}

func (f *Facet) GetFieldName() string {
	return f.fieldName
}

func (f *Facet) SetFieldName(fieldName string) {
	f.fieldName = fieldName
}

func (f *Facet) ToFacetToken(addQueryParameter func(Object) string) *facetToken {
	return createFacetTokenWithFacet(f, addQueryParameter)
}
