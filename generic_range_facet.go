package ravendb

// GenericRangeFacet represents generic range facet
type GenericRangeFacet struct {
	FacetBaseCommon
	_parent FacetBase
	Ranges  []*RangeBuilder
}

// NewGenericRangeFacet returns new GenericRangeFacet
// parent is optional, can be nil
func NewGenericRangeFacet(parent FacetBase) *GenericRangeFacet {
	return &GenericRangeFacet{
		FacetBaseCommon: NewFacetBaseCommon(),
		_parent:         parent,
	}
}

// GenericRangeFacetParse parses generic range facet
func GenericRangeFacetParse(rangeBuilder *RangeBuilder, addQueryParameter func(interface{}) string) string {
	return rangeBuilder.GetStringRepresentation(addQueryParameter)
}

// ToFacetToken returns facetToken from GenericRangeFacet
func (f *GenericRangeFacet) ToFacetToken(addQueryParameter func(interface{}) string) *facetToken {
	if f._parent != nil {
		return f._parent.ToFacetToken(addQueryParameter)
	}

	return createFacetTokenWithGenericRangeFacet(f, addQueryParameter)
}

func (f *GenericRangeFacet) addRange(rng *RangeBuilder) {
	f.Ranges = append(f.Ranges, rng)
}
