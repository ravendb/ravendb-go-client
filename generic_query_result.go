package ravendb

// GenericQueryResult represents query results
type GenericQueryResult struct {
	queryResultBase
	TotalResults   int `json:"TotalResults"`
	SkippedResults int `json:"SkippedResults"`
	//TBD 4.1  map[string]map[string]List<String>>> highlightings
	DurationInMs      int64              `json:"DurationInMs"`
	ScoreExplanations map[string]string  `json:"ScoreExplanation"`
	TimingsInMs       map[string]float64 `json:"TimingsInMs"`
	ResultSize        int64              `json:"ResultSize"`
}
