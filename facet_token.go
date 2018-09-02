package ravendb

import "strings"

var _ QueryToken = &FacetToken{}

type FacetToken struct {
	_facetSetupDocumentId string
	_aggregateByFieldName string
	_alias                string
	_ranges               []string
	_optionsParameterName string

	_aggregations []*FacetAggregationToken
}

func (t *FacetToken) GetName() string {
	return firstNonEmptyString(t._alias, t._aggregateByFieldName)
}

func NewFacetToken() *FacetToken {
	return &FacetToken{}
}

func NewFacetTokenWithID(facetSetupDocumentId string) *FacetToken {
	return &FacetToken{
		_facetSetupDocumentId: facetSetupDocumentId,
	}
}

func NewFacetTokenAll(aggregateByFieldName string, alias string, ranges []string, optionsParameterName string) *FacetToken {
	return &FacetToken{
		_aggregateByFieldName: aggregateByFieldName,
		_alias:                alias,
		_ranges:               ranges,
		_optionsParameterName: optionsParameterName,
	}
}

func (t *FacetToken) WriteTo(writer *strings.Builder) {
	writer.WriteString("facet(")

	if t._facetSetupDocumentId != "" {
		writer.WriteString("id('")
		writer.WriteString(t._facetSetupDocumentId)
		writer.WriteString("'))")

		return
	}

	firstArgument := false

	if t._aggregateByFieldName != "" {
		writer.WriteString(t._aggregateByFieldName)
	} else if len(t._ranges) != 0 {
		firstInRange := true

		for _, rang := range t._ranges {
			if !firstInRange {
				writer.WriteString(", ")
			}

			firstInRange = false
			writer.WriteString(rang)
		}
	} else {
		firstArgument = true
	}

	for _, aggregation := range t._aggregations {
		if !firstArgument {
			writer.WriteString(", ")
		}
		firstArgument = false
		aggregation.WriteTo(writer)
	}

	if stringIsNotBlank(t._optionsParameterName) {
		writer.WriteString(", $")
		writer.WriteString(t._optionsParameterName)
	}

	writer.WriteString(")")

	if stringIsBlank(t._alias) || t._alias == t._aggregateByFieldName {
		return
	}

	writer.WriteString(" as ")
	writer.WriteString(t._alias)
}

func FacetToken_create(facetSetupDocumentId string) *FacetToken {
	if stringIsWhitespace(facetSetupDocumentId) {
		//throw new IllegalArgumentException("facetSetupDocumentId cannot be null");
		panicIf(true, "facetSetupDocumentId cannot be null")
	}

	return NewFacetTokenWithID(facetSetupDocumentId)
}

func FacetToken_createWithFacet(facet *Facet, addQueryParameter func(Object) string) *FacetToken {
	optionsParameterName := getOptionsParameterName(facet, addQueryParameter)
	token := NewFacetTokenAll(facet.GetFieldName(), facet.GetDisplayFieldName(), nil, optionsParameterName)

	applyAggregations(facet, token)
	return token
}

func FacetToken_createWithRangeFacet(facet *RangeFacet, addQueryParameter func(Object) string) *FacetToken {
	optionsParameterName := getOptionsParameterName(facet, addQueryParameter)

	token := NewFacetTokenAll("", facet.GetDisplayFieldName(), facet.getRanges(), optionsParameterName)

	applyAggregations(facet, token)

	return token
}

func FacetToken_createWithGenericRangeFacet(facet *GenericRangeFacet, addQueryParameter func(Object) string) *FacetToken {
	optionsParameterName := getOptionsParameterName(facet, addQueryParameter)

	var ranges []string
	for _, rangeBuilder := range facet.getRanges() {
		ranges = append(ranges, GenericRangeFacet_parse(rangeBuilder, addQueryParameter))
	}

	token := NewFacetTokenAll("", facet.GetDisplayFieldName(), ranges, optionsParameterName)

	applyAggregations(facet, token)
	return token
}

func FacetToken_createWithFacetBase(facet FacetBase, addQueryParameter func(Object) string) *FacetToken {
	// this is just a dispatcher
	return facet.ToFacetToken(addQueryParameter)
}

func applyAggregations(facet FacetBase, token *FacetToken) {
	m := facet.GetAggregations()

	for key, value := range m {
		var aggregationToken *FacetAggregationToken
		switch key {
		case FacetAggregation_MAX:
			aggregationToken = FacetAggregationToken_max(value)
		case FacetAggregation_MIN:
			aggregationToken = FacetAggregationToken_min(value)
		case FacetAggregation_AVERAGE:
			aggregationToken = FacetAggregationToken_average(value)
		case FacetAggregation_SUM:
			aggregationToken = FacetAggregationToken_sum(value)
		default:
			panic("Unsupported aggregation method: " + key)
			//throw new NotImplementedException("Unsupported aggregation method: " + aggregation.getKey());
		}

		token._aggregations = append(token._aggregations, aggregationToken)
	}
}

func getOptionsParameterName(facet FacetBase, addQueryParameter func(Object) string) string {
	if facet.GetOptions() == nil || facet.GetOptions() == FacetOptions_getDefaultOptions() {
		return ""
	}
	return addQueryParameter(facet.GetOptions())
}

type FacetAggregationToken struct {
	_fieldName   string
	_aggregation FacetAggregation
}

func NewFacetAggregationToken(fieldName string, aggregation FacetAggregation) *FacetAggregationToken {
	return &FacetAggregationToken{
		_fieldName:   fieldName,
		_aggregation: aggregation,
	}
}

func (t *FacetAggregationToken) WriteTo(writer *strings.Builder) {
	switch t._aggregation {
	case FacetAggregation_MAX:
		writer.WriteString("max(")
		writer.WriteString(t._fieldName)
		writer.WriteString(")")
	case FacetAggregation_MIN:
		writer.WriteString("min(")
		writer.WriteString(t._fieldName)
		writer.WriteString(")")
	case FacetAggregation_AVERAGE:
		writer.WriteString("avg(")
		writer.WriteString(t._fieldName)
		writer.WriteString(")")
	case FacetAggregation_SUM:
		writer.WriteString("sum(")
		writer.WriteString(t._fieldName)
		writer.WriteString(")")
	default:
		panicIf(true, "Invalid aggregation mode: %s", t._aggregation)
		//throw new IllegalArgumentException("Invalid aggregation mode: " + _aggregation);
	}
}

func FacetAggregationToken_max(fieldName string) *FacetAggregationToken {
	panicIf(stringIsWhitespace(fieldName), "FieldName can not be null")
	return NewFacetAggregationToken(fieldName, FacetAggregation_MAX)
}

func FacetAggregationToken_min(fieldName string) *FacetAggregationToken {
	panicIf(stringIsWhitespace(fieldName), "FieldName can not be null")
	return NewFacetAggregationToken(fieldName, FacetAggregation_MIN)
}

func FacetAggregationToken_average(fieldName string) *FacetAggregationToken {
	panicIf(stringIsWhitespace(fieldName), "FieldName can not be null")
	return NewFacetAggregationToken(fieldName, FacetAggregation_AVERAGE)
}

func FacetAggregationToken_sum(fieldName string) *FacetAggregationToken {
	panicIf(stringIsWhitespace(fieldName), "FieldName can not be null")
	return NewFacetAggregationToken(fieldName, FacetAggregation_SUM)
}
