package ravendb

var (
	facetOptionsDefault = &FacetOptions{}
)

// FacetOptions describes options for facet
type FacetOptions struct {
	//termSortMode          FacetTermSortMode // TODO: why unused?
	includeRemainingTerms bool
	start                 int
	pageSize              int
}

func getDefaultFacetOptions() *FacetOptions {
	return facetOptionsDefault
}
