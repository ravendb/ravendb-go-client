package ravendb

var _ FacetBase = &RangeFacet{}

type RangeFacet struct {
	FacetBaseCommon

	_parent FacetBase

	ranges []string
}

// parent is optional (can be nil)
func NewRangeFacet(parent FacetBase) *RangeFacet {
	return &RangeFacet{
		FacetBaseCommon: NewFacetBaseCommon(),
		_parent:         parent,
	}
}

func (f *RangeFacet) getRanges() []string {
	return f.ranges
}

func (f *RangeFacet) setRanges(ranges []string) {
	f.ranges = ranges
}

func (f *RangeFacet) ToFacetToken(addQueryParameter func(interface{}) string) *facetToken {
	if f._parent != nil {
		return f._parent.ToFacetToken(addQueryParameter)
	}

	return createFacetTokenWithRangeFacet(f, addQueryParameter)
}
