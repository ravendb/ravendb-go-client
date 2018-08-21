package ravendb

type TermsQueryResult struct {
	Terms      []string `json:"Terms"`
	ResultEtag int      `json:"ResultEtag"`
	IndexName  string   `json:"IndexName"`
}
