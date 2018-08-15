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

func (r *FacetResult) GetName() string {
	return r.Name
}

func (r *FacetResult) SetName(name string) {
	r.Name = name
}

func (r *FacetResult) GetValues() []*FacetValue {
	return r.Values
}

func (r *FacetResult) SetValues(values []*FacetValue) {
	r.Values = values
}

func (r *FacetResult) GetRemainingTerms() []string {
	return r.RemainingTerms
}

func (r *FacetResult) SetRemainingTerms(remainingTerms []string) {
	r.RemainingTerms = remainingTerms
}

func (r *FacetResult) GetRemainingTermsCount() int {
	return r.RemainingTermsCount
}

func (r *FacetResult) SetRemainingTermsCount(remainingTermsCount int) {
	r.RemainingTermsCount = remainingTermsCount
}

func (r *FacetResult) GetRemainingHits() int {
	return r.RemainingHits
}

func (r *FacetResult) SetRemainingHits(remainingHits int) {
	r.RemainingHits = remainingHits
}
