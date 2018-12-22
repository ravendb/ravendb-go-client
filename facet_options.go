package ravendb

var (
	// DefaultFacetOptions are default facet options
	DefaultFacetOptions = &FacetOptions{}
)

// FacetOptions describes options for facet
// TODO: is it json-encoded and therefore should be json-annotated?
type FacetOptions struct {
	TermSortMode          FacetTermSortMode
	IncludeRemainingTerms bool
	Start                 int
	PageSize              int
}
