package ravendb

var _ FacetBase = &RangeFacet{}

// RangeFacet describes range facet
type RangeFacet struct {
	FacetBaseCommon

	_parent FacetBase

	Ranges []string
}

// NewRnageFacet returns new RangeFacet
// parent is optional (can be nil)
func NewRangeFacet(parent FacetBase) *RangeFacet {
	return &RangeFacet{
		FacetBaseCommon: NewFacetBaseCommon(),
		_parent:         parent,
	}
}

// ToFacetToken converts RangeFacet to a token
func (f *RangeFacet) ToFacetToken(addQueryParameter func(interface{}) string) *facetToken {
	if f._parent != nil {
		return f._parent.ToFacetToken(addQueryParameter)
	}

	return createFacetTokenWithRangeFacet(f, addQueryParameter)
}
