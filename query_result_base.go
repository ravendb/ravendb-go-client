package ravendb

// Note: in Java results and includes are templated but
// in practice they only ever are set to:
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

func (r *QueryResultBase) getResults() ArrayNode {
	return r.Results
}

func (r *QueryResultBase) setResults(results ArrayNode) {
	r.Results = results
}

func (r *QueryResultBase) getIncludes() ObjectNode {
	return r.Includes
}

func (r *QueryResultBase) setIncludes(includes ObjectNode) {
	r.Includes = includes
}

func (r *QueryResultBase) getIncludedPaths() []string {
	return r.IncludedPaths
}

func (r *QueryResultBase) setIncludedPaths(includedPaths []string) {
	r.IncludedPaths = includedPaths
}

func (r *QueryResultBase) isStale() bool {
	return r.IsStale
}

func (r *QueryResultBase) setStale(stale bool) {
	r.IsStale = stale
}

func (r *QueryResultBase) getIndexTimestamp() *ServerTime {
	return r.IndexTimestamp
}

func (r *QueryResultBase) setIndexTimestamp(indexTimestamp *ServerTime) {
	r.IndexTimestamp = indexTimestamp
}

func (r *QueryResultBase) getIndexName() string {
	return r.IndexName
}

func (r *QueryResultBase) setIndexName(indexName string) {
	r.IndexName = indexName
}

func (r *QueryResultBase) getResultEtag() int64 {
	return r.ResultEtag
}

func (r *QueryResultBase) setResultEtag(resultEtag int64) {
	r.ResultEtag = resultEtag
}

func (r *QueryResultBase) getLastQueryTime() *ServerTime {
	return r.LastQueryTime
}

func (r *QueryResultBase) setLastQueryTime(lastQueryTime *ServerTime) {
	r.LastQueryTime = lastQueryTime
}
