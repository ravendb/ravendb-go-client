package ravendb

type FacetResult struct {
	Name                string
	Values              []*FacetValue
	RemainingTerms      []string
	RemainingTermsCount int
	RemainingHits       int
}
