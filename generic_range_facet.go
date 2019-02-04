package ravendb

// GenericRangeFacet represents generic range facet
type GenericRangeFacet struct {
	FacetBaseCommon
	parent FacetBase
	Ranges []*RangeBuilder
}

// NewGenericRangeFacet returns new GenericRangeFacet
// parent is optional, can be nil
func NewGenericRangeFacet(parent FacetBase) *GenericRangeFacet {
	return &GenericRangeFacet{
		FacetBaseCommon: NewFacetBaseCommon(),
		parent:          parent,
	}
}

// GenericRangeFacetParse parses generic range facet
func genericRangeFacetParse(rangeBuilder *RangeBuilder, addQueryParameter func(interface{}) string) (string, error) {
	return rangeBuilder.GetStringRepresentation(addQueryParameter)
}

// ToFacetToken returns facetToken from GenericRangeFacet
func (f *GenericRangeFacet) ToFacetToken(addQueryParameter func(interface{}) string) (*facetToken, error) {
	if f.parent != nil {
		return f.parent.ToFacetToken(addQueryParameter)
	}

	return createFacetTokenWithGenericRangeFacet(f, addQueryParameter)
}

func (f *GenericRangeFacet) addRange(rng *RangeBuilder) {
	f.Ranges = append(f.Ranges, rng)
}
