package ravendb

type TermsQueryResult struct {
	Terms      *StringSet `json:"Terms"` // TODO: Set<String>
	ResultEtag int        `json:"ResultEtag"`
	IndexName  string     `json:"IndexName"`
}

func (r *TermsQueryResult) getTerms() []string {
	return r.Terms.strings
}

func (r *TermsQueryResult) setTerms(terms *StringSet) {
	r.Terms = terms
}

func (r *TermsQueryResult) getResultEtag() int {
	return r.ResultEtag
}

func (r *TermsQueryResult) setResultEtag(resultEtag int) {
	r.ResultEtag = resultEtag
}

func (r *TermsQueryResult) getIndexName() string {
	return r.IndexName
}

func (r *TermsQueryResult) setIndexName(indexName string) {
	r.IndexName = indexName
}
