package ravendb

var _ QueryToken = &FacetToken{}

type FacetToken struct {
	_facetSetupDocumentId string
	_aggregateByFieldName string
	_alias                string
	_ranges               []string
	_optionsParameterName string

	_aggregations []*FacetAggregationToken
}

func (t *FacetToken) getName() string {
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

func (t *FacetToken) writeTo(writer *StringBuilder) {
	writer.append("facet(")

	if t._facetSetupDocumentId != "" {
		writer.append("id('")
		writer.append(t._facetSetupDocumentId)
		writer.append("'))")

		return
	}

	firstArgument := false

	if t._aggregateByFieldName != "" {
		writer.append(t._aggregateByFieldName)
	} else if len(t._ranges) != 0 {
		firstInRange := true

		for _, rang := range t._ranges {
			if !firstInRange {
				writer.append(", ")
			}

			firstInRange = false
			writer.append(rang)
		}
	} else {
		firstArgument = true
	}

	for _, aggregation := range t._aggregations {
		if !firstArgument {
			writer.append(", ")
		}
		firstArgument = false
		aggregation.writeTo(writer)
	}

	if StringUtils_isNotBlank(t._optionsParameterName) {
		writer.append(", $")
		writer.append(t._optionsParameterName)
	}

	writer.append(")")

	if StringUtils_isBlank(t._alias) || t._alias == t._aggregateByFieldName {
		return
	}

	writer.append(" as ")
	writer.append(t._alias)
}

func FacetToken_create(facetSetupDocumentId string) *FacetToken {
	if StringUtils_isWhitespace(facetSetupDocumentId) {
		//throw new IllegalArgumentException("facetSetupDocumentId cannot be null");
		panicIf(true, "facetSetupDocumentId cannot be null")
	}

	return NewFacetTokenWithID(facetSetupDocumentId)
}

func FacetToken_createWithFacet(facet *Facet, addQueryParameter func(Object) string) *FacetToken {
	optionsParameterName := getOptionsParameterName(facet, addQueryParameter)
	token := NewFacetTokenAll(facet.getFieldName(), facet.getDisplayFieldName(), nil, optionsParameterName)

	applyAggregations(facet, token)
	return token
}

func FacetToken_createWithRangeFacet(facet *RangeFacet, addQueryParameter func(Object) string) *FacetToken {
	optionsParameterName := getOptionsParameterName(facet, addQueryParameter)

	token := NewFacetTokenAll("", facet.getDisplayFieldName(), facet.getRanges(), optionsParameterName)

	applyAggregations(facet, token)

	return token
}

func FacetToken_createWithGenericRangeFacet(facet *GenericRangeFacet, addQueryParameter func(Object) string) *FacetToken {
	optionsParameterName := getOptionsParameterName(facet, addQueryParameter)

	var ranges []string
	for _, rangeBuilder := range facet.getRanges() {
		ranges = append(ranges, GenericRangeFacet_parse(rangeBuilder, addQueryParameter))
	}

	token := NewFacetTokenAll("", facet.getDisplayFieldName(), ranges, optionsParameterName)

	applyAggregations(facet, token)
	return token
}

func FacetToken_createWithFacetBase(facet FacetBase, addQueryParameter func(Object) string) *FacetToken {
	// this is just a dispatcher
	return facet.toFacetToken(addQueryParameter)
}

func applyAggregations(facet FacetBase, token *FacetToken) {
	m := facet.getAggregations()

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
	if facet.getOptions() == nil || facet.getOptions() == FacetOptions_getDefaultOptions() {
		return ""
	}
	return addQueryParameter(facet.getOptions())
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

func (t *FacetAggregationToken) writeTo(writer *StringBuilder) {
	switch t._aggregation {
	case FacetAggregation_MAX:
		writer.append("max(")
		writer.append(t._fieldName)
		writer.append(")")
	case FacetAggregation_MIN:
		writer.append("min(")
		writer.append(t._fieldName)
		writer.append(")")
	case FacetAggregation_AVERAGE:
		writer.append("avg(")
		writer.append(t._fieldName)
		writer.append(")")
	case FacetAggregation_SUM:
		writer.append("sum(")
		writer.append(t._fieldName)
		writer.append(")")
	default:
		panicIf(true, "Invalid aggregation mode: %s", t._aggregation)
		//throw new IllegalArgumentException("Invalid aggregation mode: " + _aggregation);
	}
}

func FacetAggregationToken_max(fieldName string) *FacetAggregationToken {
	panicIf(StringUtils_isWhitespace(fieldName), "FieldName can not be null")
	return NewFacetAggregationToken(fieldName, FacetAggregation_MAX)
}

func FacetAggregationToken_min(fieldName string) *FacetAggregationToken {
	panicIf(StringUtils_isWhitespace(fieldName), "FieldName can not be null")
	return NewFacetAggregationToken(fieldName, FacetAggregation_MIN)
}

func FacetAggregationToken_average(fieldName string) *FacetAggregationToken {
	panicIf(StringUtils_isWhitespace(fieldName), "FieldName can not be null")
	return NewFacetAggregationToken(fieldName, FacetAggregation_AVERAGE)
}

func FacetAggregationToken_sum(fieldName string) *FacetAggregationToken {
	panicIf(StringUtils_isWhitespace(fieldName), "FieldName can not be null")
	return NewFacetAggregationToken(fieldName, FacetAggregation_SUM)
}
