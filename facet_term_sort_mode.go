package ravendb

// FacetTermSortMode describes sort mode for a facet
type FacetTermSortMode = string

const (
	FacetTermSortModeValueAsc  = "ValueAsc"
	FacetTermSortModeValueDesc = "ValueDesc"
	FacetTermSortModeCountAsc  = "CountAsc"
	FacetTermSortModeCountDesc = "CountDesc"
)
