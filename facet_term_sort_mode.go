package ravendb

// FacetTermSortMode describes sort mode for a facet
type FacetTermSortMode = string

const (
	FacetTermSortByValueAsc  = "ValueAsc"
	FacetTermSortByValueDesc = "ValueDesc"
	FacetTermSortByCountAsc  = "CountAsc"
	FacetTermSortByCountDesc = "CountDesc"
)
