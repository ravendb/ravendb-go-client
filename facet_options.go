package ravendb

var (
	FacetOptions_defaultOptions = &FacetOptions{}
)

type FacetOptions struct {
	termSortMode          FacetTermSortMode
	includeRemainingTerms bool
	start                 int
	pageSize              int
}

func FacetOptions_getDefaultOptions() *FacetOptions {
	return FacetOptions_defaultOptions
}
