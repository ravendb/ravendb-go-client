package ravendb

// QueryResultBase represents results of the query returned by the server
// Note: in Java results and includes are templated but
// in practice they are only ever set to:
// TResult = ArrayNode
// TInclude = ObjectNode
type QueryResultBase struct {
	Results        ArrayNode   `json:"Results"`
	Includes       ObjectNode  `json:"Includes"`
	IncludedPaths  []string    `json:"IncludedPaths"`
	IsStale        bool        `json:"IsStale"`
	IndexTimestamp *ServerTime `json:"IndexTimestamp"`
	IndexName      string      `json:"IndexName"`
	ResultEtag     int64       `json:"ResultEtag"`
	LastQueryTime  *ServerTime `json:"LastQueryTime"`
}
