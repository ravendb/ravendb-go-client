package ravendb

var _ FacetBase = &Facet{}

// Facet describes a search facet
type Facet struct {
	FacetBaseCommon

	FieldName string `json:"FieldName"`
}

// NewFacet returns a new Facet
func NewFacet() *Facet {
	return &Facet{
		FacetBaseCommon: NewFacetBaseCommon(),
	}
}

// ToFacetToken returns token for this facet
func (f *Facet) ToFacetToken(addQueryParameter func(interface{}) string) (*facetToken, error) {
	return createFacetTokenWithFacet(f, addQueryParameter), nil
}
