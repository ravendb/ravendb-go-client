package ravendb

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

/*
public class FacetToken extends QueryToken {

    public static FacetToken create(string facetSetupDocumentId) {
        if (stringUtils.isWhitespace(facetSetupDocumentId)) {
            throw new IllegalArgumentException("facetSetupDocumentId cannot be null");
        }

        return new FacetToken(facetSetupDocumentId);
    }

    public static FacetToken create(Facet facet, Function<Object, string> addQueryParameter) {
        string optionsParameterName = getOptionsParameterName(facet, addQueryParameter);
        FacetToken token = new FacetToken(facet.getFieldName(), facet.getDisplayFieldName(), null, optionsParameterName);

        applyAggregations(facet, token);

        return token;
    }

    public static FacetToken create(RangeFacet facet, Function<Object, string> addQueryParameter) {
        string optionsParameterName = getOptionsParameterName(facet, addQueryParameter);

        FacetToken token = new FacetToken(null, facet.getDisplayFieldName(), facet.getRanges(), optionsParameterName);

        applyAggregations(facet, token);

        return token;
    }

    public static FacetToken create(GenericRangeFacet facet, Function<Object, string> addQueryParameter) {
        string optionsParameterName = getOptionsParameterName(facet, addQueryParameter);

        []string ranges = new ArrayList<>();
        for (RangeBuilder<?> rangeBuilder : facet.getRanges()) {
            ranges.add(GenericRangeFacet.parse(rangeBuilder, addQueryParameter));
        }

        FacetToken token = new FacetToken(null, facet.getDisplayFieldName(), ranges, optionsParameterName);

        applyAggregations(facet, token);
        return token;
    }

    public static FacetToken create(FacetBase facet, Function<Object, string> addQueryParameter) {
        // this is just a dispatcher
        return facet.toFacetToken(addQueryParameter);
    }


    private static void applyAggregations(FacetBase facet, FacetToken token) {
        for (Map.Entry<FacetAggregation, string> aggregation : facet.getAggregations().entrySet()) {
            FacetAggregationToken aggregationToken;
            switch (aggregation.getKey()) {
                case MAX:
                    aggregationToken = FacetAggregationToken.max(aggregation.getValue());
                    break;
                case MIN:
                    aggregationToken = FacetAggregationToken.min(aggregation.getValue());
                    break;
                case AVERAGE:
                    aggregationToken = FacetAggregationToken.average(aggregation.getValue());
                    break;
                case SUM:
                    aggregationToken = FacetAggregationToken.sum(aggregation.getValue());
                    break;
                default :
                    throw new NotImplementedException("Unsupported aggregation method: " + aggregation.getKey());
            }

            token._aggregations.add(aggregationToken);
        }
    }

    private static string getOptionsParameterName(FacetBase facet, Function<Object, string> addQueryParameter) {
        return facet.getOptions() != null && facet.getOptions() != FacetOptions.getDefaultOptions() ? addQueryParameter.apply(facet.getOptions()) : null;
    }
}
*/

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
