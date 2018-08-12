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

func (b *FacetBuilder) byRanges(rng *RangeBuilder, ranges ...*RangeBuilder) IFacetOperations {
	if rng == nil {
		//throw new IllegalArgumentException("Range cannot be null")
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

func (b *FacetBuilder) byField(fieldName string) IFacetOperations {
	if b._default == nil {
		b._default = NewFacet()
	}

	if isRqlKeyword(fieldName) {
		fieldName = "'" + fieldName + "'"
	}

	b._default.setFieldName(fieldName)

	return b
}

func (b *FacetBuilder) allResults() IFacetOperations {
	if b._default == nil {
		b._default = NewFacet()
	}

	b._default.setFieldName("")
	return b
}

func (b *FacetBuilder) withOptions(options *FacetOptions) IFacetOperations {
	b.getFacet().setOptions(options)
	return b
}

func (b *FacetBuilder) withDisplayName(displayName string) IFacetOperations {
	b.getFacet().setDisplayFieldName(displayName)
	return b
}

func (b *FacetBuilder) sumOn(path string) IFacetOperations {
	b.getFacet().getAggregations()[FacetAggregation_SUM] = path
	return b
}

func (b *FacetBuilder) minOn(path string) IFacetOperations {
	b.getFacet().getAggregations()[FacetAggregation_MIN] = path
	return b
}

func (b *FacetBuilder) maxOn(path string) IFacetOperations {
	b.getFacet().getAggregations()[FacetAggregation_MAX] = path
	return b
}

func (b *FacetBuilder) averageOn(path string) IFacetOperations {
	b.getFacet().getAggregations()[FacetAggregation_AVERAGE] = path
	return b
}

func (b *FacetBuilder) getFacet() FacetBase {
	if b._default != nil {
		return b._default
	}

	return b._range
}
