package ravendb

type FacetResult struct {
	Name                string
	Values              []*FacetValue
	RemainingTerms      []string
	RemainingTermsCount int
	RemainingHits       int
}

func NewFacetResult() *FacetResult {
	return &FacetResult{}
}

func (r *FacetResult) getName() string {
	return r.Name
}

func (r *FacetResult) setName(name string) {
	r.Name = name
}

func (r *FacetResult) getValues() []*FacetValue {
	return r.Values
}

func (r *FacetResult) setValues(values []*FacetValue) {
	r.Values = values
}

func (r *FacetResult) getRemainingTerms() []string {
	return r.RemainingTerms
}

func (r *FacetResult) setRemainingTerms(remainingTerms []string) {
	r.RemainingTerms = remainingTerms
}

func (r *FacetResult) getRemainingTermsCount() int {
	return r.RemainingTermsCount
}

func (r *FacetResult) setRemainingTermsCount(remainingTermsCount int) {
	r.RemainingTermsCount = remainingTermsCount
}

func (r *FacetResult) getRemainingHits() int {
	return r.RemainingHits
}

func (r *FacetResult) setRemainingHits(remainingHits int) {
	r.RemainingHits = remainingHits
}
