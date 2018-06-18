package ravendb

type FacetAggregation = string

const (
	FacetAggregation_NONE    = "None"
	FacetAggregation_MAX     = "Max"
	FacetAggregation_MIN     = "Min"
	FacetAggregation_AVERAGE = "Average"
	FacetAggregation_SUM     = "Sum"
)
