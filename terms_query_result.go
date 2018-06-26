package ravendb

type TermsQueryResult struct {
	Terms      interface{} `json:"Terms"` // TODO: Set<String>
	ResultEtag int         `json:"ResultEtag"`
	IndexName  string      `json:"IndexName"`
}

func (r *TermsQueryResult) getTerms() []string {
	panicIf(true, "NYI")
	//return r.Terms
	return nil
}

/*
	public void setTerms(Set<String> terms) {
        this.terms = terms;
    }
*/

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
