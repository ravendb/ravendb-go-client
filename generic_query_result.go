package ravendb

// GenericQueryResult represents query results
type GenericQueryResult struct {
	queryResultBase
	TotalResults   int                            `json:"TotalResults"`
	SkippedResults int                            `json:"SkippedResults"`
	Highlightings  map[string]map[string][]string `json:"Highlightings"`
	Explanations   map[string][]string            `json:"Explanations"`
	DurationInMs   int64                          `json:"DurationInMs"`
	ResultSize     int64                          `json:"ResultSize"`
}
