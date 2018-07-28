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

func (o *FacetOptions) getTermSortMode() FacetTermSortMode {
	return o.termSortMode
}

func (o *FacetOptions) setTermSortMode(termSortMode FacetTermSortMode) {
	o.termSortMode = termSortMode
}

func (o *FacetOptions) isIncludeRemainingTerms() bool {
	return o.includeRemainingTerms
}

func (o *FacetOptions) setIncludeRemainingTerms(includeRemainingTerms bool) {
	o.includeRemainingTerms = includeRemainingTerms
}

func (o *FacetOptions) getStart() int {
	return o.start
}

func (o *FacetOptions) setStart(start int) {
	o.start = start
}

func (o *FacetOptions) getPageSize() int {
	return o.pageSize
}

func (o *FacetOptions) setPageSize(pageSize int) {
	o.pageSize = pageSize
}
