package ravendb

type GenericRangeFacet struct {
	FacetBaseCommon
	_parent FacetBase
	ranges  []*RangeBuilder
}

// parent is optional, can be nil
func NewGenericRangeFacet(parent FacetBase) *GenericRangeFacet {
	return &GenericRangeFacet{
		FacetBaseCommon: NewFacetBaseCommon(),
		_parent:         parent,
	}
}

func GenericRangeFacet_parse(rangeBuilder *RangeBuilder, addQueryParameter func(Object) string) string {
	return rangeBuilder.GetStringRepresentation(addQueryParameter)
}

func (f *GenericRangeFacet) getRanges() []*RangeBuilder {
	return f.ranges
}

func (f *GenericRangeFacet) setRanges(ranges []*RangeBuilder) {
	f.ranges = ranges
}

func (f *GenericRangeFacet) ToFacetToken(addQueryParameter func(Object) string) *FacetToken {
	if f._parent != nil {
		return f._parent.ToFacetToken(addQueryParameter)
	}

	return FacetToken_createWithGenericRangeFacet(f, addQueryParameter)
}

func (f *GenericRangeFacet) addRange(rng *RangeBuilder) {
	f.ranges = append(f.ranges, rng)
}
