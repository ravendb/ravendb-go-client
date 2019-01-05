package ravendb

import "strings"

var _ queryToken = &facetToken{}

type facetToken struct {
	_facetSetupDocumentID string
	_aggregateByFieldName string
	_alias                string
	_ranges               []string
	_optionsParameterName string

	_aggregations []*facetAggregationToken
}

func (t *facetToken) GetName() string {
	return firstNonEmptyString(t._alias, t._aggregateByFieldName)
}

func NewFacetTokenWithID(facetSetupDocumentID string) *facetToken {
	return &facetToken{
		_facetSetupDocumentID: facetSetupDocumentID,
	}
}

func NewFacetTokenAll(aggregateByFieldName string, alias string, ranges []string, optionsParameterName string) *facetToken {
	return &facetToken{
		_aggregateByFieldName: aggregateByFieldName,
		_alias:                alias,
		_ranges:               ranges,
		_optionsParameterName: optionsParameterName,
	}
}

func (t *facetToken) writeTo(writer *strings.Builder) {
	writer.WriteString("facet(")

	if t._facetSetupDocumentID != "" {
		writer.WriteString("id('")
		writer.WriteString(t._facetSetupDocumentID)
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
		aggregation.writeTo(writer)
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

func createFacetToken(facetSetupDocumentID string) *facetToken {
	if stringIsWhitespace(facetSetupDocumentID) {
		//throw new IllegalArgumentError("facetSetupDocumentID cannot be null");
		panicIf(true, "facetSetupDocumentID cannot be null")
	}

	return NewFacetTokenWithID(facetSetupDocumentID)
}

func createFacetTokenWithFacet(facet *Facet, addQueryParameter func(interface{}) string) *facetToken {
	optionsParameterName := getOptionsParameterName(facet, addQueryParameter)
	token := NewFacetTokenAll(facet.FieldName, facet.GetDisplayFieldName(), nil, optionsParameterName)

	applyAggregations(facet, token)
	return token
}

func createFacetTokenWithRangeFacet(facet *RangeFacet, addQueryParameter func(interface{}) string) *facetToken {
	optionsParameterName := getOptionsParameterName(facet, addQueryParameter)

	token := NewFacetTokenAll("", facet.GetDisplayFieldName(), facet.Ranges, optionsParameterName)

	applyAggregations(facet, token)

	return token
}

func createFacetTokenWithGenericRangeFacet(facet *GenericRangeFacet, addQueryParameter func(interface{}) string) *facetToken {
	optionsParameterName := getOptionsParameterName(facet, addQueryParameter)

	var ranges []string
	for _, rangeBuilder := range facet.Ranges {
		ranges = append(ranges, GenericRangeFacetParse(rangeBuilder, addQueryParameter))
	}

	token := NewFacetTokenAll("", facet.GetDisplayFieldName(), ranges, optionsParameterName)

	applyAggregations(facet, token)
	return token
}

func createFacetTokenWithFacetBase(facet FacetBase, addQueryParameter func(interface{}) string) *facetToken {
	// this is just a dispatcher
	return facet.ToFacetToken(addQueryParameter)
}

func applyAggregations(facet FacetBase, token *facetToken) {
	m := facet.GetAggregations()

	for key, value := range m {
		var aggregationToken *facetAggregationToken
		switch key {
		case FacetAggregationMax:
			aggregationToken = facetAggregationTokenMax(value)
		case FacetAggregationMin:
			aggregationToken = facetAggregationTokenMin(value)
		case FacetAggregationAverage:
			aggregationToken = facetAggregationTokenAverage(value)
		case FacetAggregationSum:
			aggregationToken = facetAggregationTokenSum(value)
		default:
			panic("Unsupported aggregation method: " + key)
			//throw new NotImplementedError("Unsupported aggregation method: " + aggregation.getKey());
		}

		token._aggregations = append(token._aggregations, aggregationToken)
	}
}

func getOptionsParameterName(facet FacetBase, addQueryParameter func(interface{}) string) string {
	if facet.GetOptions() == nil || facet.GetOptions() == DefaultFacetOptions {
		return ""
	}
	return addQueryParameter(facet.GetOptions())
}

var _ queryToken = &facetAggregationToken{}

type facetAggregationToken struct {
	_fieldName   string
	_aggregation FacetAggregation
}

func newFacetAggregationToken(fieldName string, aggregation FacetAggregation) *facetAggregationToken {
	return &facetAggregationToken{
		_fieldName:   fieldName,
		_aggregation: aggregation,
	}
}

func (t *facetAggregationToken) writeTo(writer *strings.Builder) {
	switch t._aggregation {
	case FacetAggregationMax:
		writer.WriteString("max(")
		writer.WriteString(t._fieldName)
		writer.WriteString(")")
	case FacetAggregationMin:
		writer.WriteString("min(")
		writer.WriteString(t._fieldName)
		writer.WriteString(")")
	case FacetAggregationAverage:
		writer.WriteString("avg(")
		writer.WriteString(t._fieldName)
		writer.WriteString(")")
	case FacetAggregationSum:
		writer.WriteString("sum(")
		writer.WriteString(t._fieldName)
		writer.WriteString(")")
	default:
		panicIf(true, "Invalid aggregation mode: %s", t._aggregation)
		//throw new IllegalArgumentError("Invalid aggregation mode: " + _aggregation);
	}
}

func facetAggregationTokenMax(fieldName string) *facetAggregationToken {
	panicIf(stringIsWhitespace(fieldName), "FieldName can not be null")
	return newFacetAggregationToken(fieldName, FacetAggregationMax)
}

func facetAggregationTokenMin(fieldName string) *facetAggregationToken {
	panicIf(stringIsWhitespace(fieldName), "FieldName can not be null")
	return newFacetAggregationToken(fieldName, FacetAggregationMin)
}

func facetAggregationTokenAverage(fieldName string) *facetAggregationToken {
	panicIf(stringIsWhitespace(fieldName), "FieldName can not be null")
	return newFacetAggregationToken(fieldName, FacetAggregationAverage)
}

func facetAggregationTokenSum(fieldName string) *facetAggregationToken {
	panicIf(stringIsWhitespace(fieldName), "FieldName can not be null")
	return newFacetAggregationToken(fieldName, FacetAggregationSum)
}
