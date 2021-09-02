package ravendb

// FacetSetup describes new facet setup
type FacetSetup struct {
	ID     string
	Facets []*Facet `json:"Facets,omitempty"`
	// Note: omitempty here is important. If we Send 'null',
	// (as opposed to Java's '[]') the server will error out
	// when parsing query referencing this setup
	RangeFacets []*RangeFacet `json:"RangeFacets,omitempty"`
}
