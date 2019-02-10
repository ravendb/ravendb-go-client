package ravendb

import "strings"

var _ queryToken = &facetToken{}

type facetToken struct {
	facetSetupDocumentID string
	aggregateByFieldName string
	alias                string
	ranges               []string
	optionsParameterName string

	aggregations []*facetAggregationToken
}

func (t *facetToken) GetName() string {
	return firstNonEmptyString(t.alias, t.aggregateByFieldName)
}

func newFacetTokenWithID(facetSetupDocumentID string) *facetToken {
	return &facetToken{
		facetSetupDocumentID: facetSetupDocumentID,
	}
}

func newFacetTokenAll(aggregateByFieldName string, alias string, ranges []string, optionsParameterName string) *facetToken {
	return &facetToken{
		aggregateByFieldName: aggregateByFieldName,
		alias:                alias,
		ranges:               ranges,
		optionsParameterName: optionsParameterName,
	}
}

func (t *facetToken) writeTo(writer *strings.Builder) error {
	writer.WriteString("facet(")

	if t.facetSetupDocumentID != "" {
		writer.WriteString("id('")
		writer.WriteString(t.facetSetupDocumentID)
		writer.WriteString("'))")

		return nil
	}

	firstArgument := false

	if t.aggregateByFieldName != "" {
		writer.WriteString(t.aggregateByFieldName)
	} else if len(t.ranges) != 0 {
		firstInRange := true

		for _, rang := range t.ranges {
			if !firstInRange {
				writer.WriteString(", ")
			}

			firstInRange = false
			writer.WriteString(rang)
		}
	} else {
		firstArgument = true
	}

	for _, aggregation := range t.aggregations {
		if !firstArgument {
			writer.WriteString(", ")
		}
		firstArgument = false
		if err := aggregation.writeTo(writer); err != nil {
			return err
		}
	}

	if stringIsNotBlank(t.optionsParameterName) {
		writer.WriteString(", $")
		writer.WriteString(t.optionsParameterName)
	}

	writer.WriteString(")")

	if stringIsBlank(t.alias) || t.alias == t.aggregateByFieldName {
		return nil
	}

	writer.WriteString(" as ")
	writer.WriteString(t.alias)

	return nil
}

func createFacetToken(facetSetupDocumentID string) (*facetToken, error) {
	if stringIsBlank(facetSetupDocumentID) {
		return nil, newIllegalArgumentError("facetSetupDocumentID cannot be null")
	}

	return newFacetTokenWithID(facetSetupDocumentID), nil
}

func createFacetTokenWithFacet(facet *Facet, addQueryParameter func(interface{}) string) *facetToken {
	optionsParameterName := getOptionsParameterName(facet, addQueryParameter)
	token := newFacetTokenAll(facet.FieldName, facet.DisplayFieldName, nil, optionsParameterName)

	applyAggregations(facet, token)
	return token
}

func createFacetTokenWithRangeFacet(facet *RangeFacet, addQueryParameter func(interface{}) string) (*facetToken, error) {
	optionsParameterName := getOptionsParameterName(facet, addQueryParameter)

	token := newFacetTokenAll("", facet.DisplayFieldName, facet.Ranges, optionsParameterName)

	applyAggregations(facet, token)

	return token, nil
}

func createFacetTokenWithGenericRangeFacet(facet *GenericRangeFacet, addQueryParameter func(interface{}) string) (*facetToken, error) {
	optionsParameterName := getOptionsParameterName(facet, addQueryParameter)

	var ranges []string
	for _, rangeBuilder := range facet.Ranges {
		r, err := genericRangeFacetParse(rangeBuilder, addQueryParameter)
		if err != nil {
			return nil, err
		}
		ranges = append(ranges, r)
	}

	token := newFacetTokenAll("", facet.DisplayFieldName, ranges, optionsParameterName)

	applyAggregations(facet, token)
	return token, nil
}

func createFacetTokenWithFacetBase(facet FacetBase, addQueryParameter func(interface{}) string) (*facetToken, error) {
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

		token.aggregations = append(token.aggregations, aggregationToken)
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

func (t *facetAggregationToken) writeTo(writer *strings.Builder) error {
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
	return nil
}

func facetAggregationTokenMax(fieldName string) *facetAggregationToken {
	panicIf(stringIsBlank(fieldName), "FieldName can not be null")
	return newFacetAggregationToken(fieldName, FacetAggregationMax)
}

func facetAggregationTokenMin(fieldName string) *facetAggregationToken {
	panicIf(stringIsBlank(fieldName), "FieldName can not be null")
	return newFacetAggregationToken(fieldName, FacetAggregationMin)
}

func facetAggregationTokenAverage(fieldName string) *facetAggregationToken {
	panicIf(stringIsBlank(fieldName), "FieldName can not be null")
	return newFacetAggregationToken(fieldName, FacetAggregationAverage)
}

func facetAggregationTokenSum(fieldName string) *facetAggregationToken {
	panicIf(stringIsBlank(fieldName), "FieldName can not be null")
	return newFacetAggregationToken(fieldName, FacetAggregationSum)
}
