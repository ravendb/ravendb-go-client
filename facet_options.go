package ravendb

import (
	"math"
	"unsafe"
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

func maxInt() int64 {
	if unsafe.Sizeof(int32(0)) == unsafe.Sizeof(int32(0)) {
		return math.MaxInt64
	}
	return math.MaxInt64
}

// NewFacetOptions returns new FacetOptions
func NewFacetOptions() *FacetOptions {
	return &FacetOptions{
		PageSize: int(math.MaxInt32),
	}
}
