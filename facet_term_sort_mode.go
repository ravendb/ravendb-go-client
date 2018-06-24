package ravendb

type FacetTermSortMode = string

const (
	FacetTermSortMode_VALUE_ASC  = "ValueAsc"
	FacetTermSortMode_VALUE_DESC = "ValueDesc"
	FacetTermSortMode_COUNT_ASC  = "CountAsc"
	FacetTermSortMode_COUNT_DESC = "CountDesc"
)
