package ravendb

var _ FacetBase = &Facet{}

type Facet struct {
	FacetBaseCommon

	FieldName string
}

func NewFacet() *Facet {
	return &Facet{
		FacetBaseCommon: NewFacetBaseCommon(),
	}
}

func (f *Facet) ToFacetToken(addQueryParameter func(interface{}) string) *facetToken {
	return createFacetTokenWithFacet(f, addQueryParameter)
}
