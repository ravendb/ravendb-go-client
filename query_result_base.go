package ravendb

// QueryResultBase represents results of the query returned by the server
// Note: in Java results and includes are templated but
// in practice they are only ever set to:
// TResult = ArrayNode i.e. []map[string]interface{}
// TInclude = ObjectNode i.e. map[string]interface{}
type QueryResultBase struct {
	Results        []map[string]interface{} `json:"Results"`
	Includes       map[string]interface{}   `json:"Includes"`
	IncludedPaths  []string                 `json:"IncludedPaths"`
	IsStale        bool                     `json:"IsStale"`
	IndexTimestamp *Time                    `json:"IndexTimestamp"`
	IndexName      string                   `json:"IndexName"`
	ResultEtag     int64                    `json:"ResultEtag"`
	LastQueryTime  *Time                    `json:"LastQueryTime"`
}
