package ravendb

// TermsQueryResult represents results of terms query
type TermsQueryResult struct {
	Terms      []string `json:"Terms"`
	ResultEtag int64    `json:"ResultEtag"`
	IndexName  string   `json:"IndexName"`
}
