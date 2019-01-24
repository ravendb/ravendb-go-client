package ravendb

// FacetResults represents results of faceted search
type FacetResult struct {
	Name                string
	Values              []*FacetValue
	RemainingTerms      []string
	RemainingTermsCount int
	RemainingHits       int
}
