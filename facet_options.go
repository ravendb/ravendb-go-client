package ravendb

import (
	"math"
)

var (
	// DefaultFacetOptions are default facet options
	DefaultFacetOptions = &FacetOptions{}
)

// FacetOptions describes options for facet
type FacetOptions struct {
	TermSortMode          FacetTermSortMode `json:"TermSortMode"`
	IncludeRemainingTerms bool              `json:"IncludeRemainingTerms"`
	Start                 int               `json:"Start"`
	PageSize              int               `json:"PageSize"`
}

// NewFacetOptions returns new FacetOptions
func NewFacetOptions() *FacetOptions {
	return &FacetOptions{
		PageSize: int(math.MaxInt32),
	}
}
