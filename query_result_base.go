package ravendb

// queryResultBase represents results of the query returned by the server
// Note: in Java results and includes are templated but
// in practice they are only ever set to:
// TResult = ArrayNode i.e. []map[string]interface{}
// TInclude = ObjectNode i.e. map[string]interface{}
type queryResultBase struct {
	Results              []map[string]interface{} `json:"Results"`
	Includes             map[string]interface{}   `json:"Includes"`
	CounterIncludes      map[string]interface{}   `json:"CounterIncludes"`
	IncludedCounterNames map[string][]string      `json:"IncludedCounterNames"`
	IncludedPaths        []string                 `json:"IncludedPaths"`
	IsStale              bool                     `json:"IsStale"`
	IndexTimestamp       *Time                    `json:"IndexTimestamp"`
	IndexName            string                   `json:"IndexName"`
	ResultEtag           int64                    `json:"ResultEtag"`
	LastQueryTime        *Time                    `json:"LastQueryTime"`
	NodeTag              string                   `json:"NodeTag"`
	Timings              *QueryTimings            `json:"Timings"`
}
