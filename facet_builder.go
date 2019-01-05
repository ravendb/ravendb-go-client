package ravendb

import "strings"

var (
	_ IFacetBuilder    = &FacetBuilder{}
	_ IFacetOperations = &FacetBuilder{}
)

func isRqlKeyword(s string) bool {
	s = strings.ToLower(s)
	switch s {
	case "as", "select", "where", "load", "group", "order", "include", "update":
		return true
	}
	return false
}

type FacetBuilder struct {
	_range   *GenericRangeFacet
	_default *Facet
}

func NewFacetBuilder() *FacetBuilder {
	return &FacetBuilder{}
}

func (b *FacetBuilder) ByRanges(rng *RangeBuilder, ranges ...*RangeBuilder) IFacetOperations {
	if rng == nil {
		//throw new IllegalArgumentError("Range cannot be null")
		panic("Range cannot be null")
	}

	if b._range == nil {
		b._range = NewGenericRangeFacet(nil)
	}

	b._range.addRange(rng)

	for _, rng := range ranges {
		b._range.addRange(rng)
	}

	return b
}

func (b *FacetBuilder) ByField(fieldName string) IFacetOperations {
	if b._default == nil {
		b._default = NewFacet()
	}

	if isRqlKeyword(fieldName) {
		fieldName = "'" + fieldName + "'"
	}

	b._default.FieldName = fieldName

	return b
}

func (b *FacetBuilder) AllResults() IFacetOperations {
	if b._default == nil {
		b._default = NewFacet()
	}

	b._default.FieldName = ""
	return b
}

func (b *FacetBuilder) WithOptions(options *FacetOptions) IFacetOperations {
	b.getFacet().SetOptions(options)
	return b
}

func (b *FacetBuilder) WithDisplayName(displayName string) IFacetOperations {
	b.getFacet().SetDisplayFieldName(displayName)
	return b
}

func (b *FacetBuilder) SumOn(path string) IFacetOperations {
	b.getFacet().GetAggregations()[FacetAggregationSum] = path
	return b
}

func (b *FacetBuilder) MinOn(path string) IFacetOperations {
	b.getFacet().GetAggregations()[FacetAggregationMin] = path
	return b
}

func (b *FacetBuilder) MaxOn(path string) IFacetOperations {
	b.getFacet().GetAggregations()[FacetAggregationMax] = path
	return b
}

func (b *FacetBuilder) AverageOn(path string) IFacetOperations {
	b.getFacet().GetAggregations()[FacetAggregationAverage] = path
	return b
}

func (b *FacetBuilder) getFacet() FacetBase {
	if b._default != nil {
		return b._default
	}

	return b._range
}
