package ravendb

// TermsQueryResult represents results of terms query
type TermsQueryResult struct {
	Terms      []string `json:"Terms"`
	ResultEtag int      `json:"ResultEtag"`
	IndexName  string   `json:"IndexName"`
}
