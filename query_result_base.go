package ravendb

import "time"

// Note: in Java results and includes are templated but
// in practice they only ever are set to:
// TResult = ArrayNode
// TInclude = ObjectNode
// TODO: is time.Time *ServerTime ?
// TODO: json annotate?
type QueryResultBase struct {
	results        ArrayNode
	includes       ObjectNode
	includedPaths  []string
	_isStale       bool
	indexTimestamp time.Time
	indexName      string
	resultEtag     int64
	lastQueryTime  time.Time
}

func (r *QueryResultBase) getResults() ArrayNode {
	return r.results
}

func (r *QueryResultBase) setResults(results ArrayNode) {
	r.results = results
}

func (r *QueryResultBase) getIncludes() ObjectNode {
	return r.includes
}

func (r *QueryResultBase) setIncludes(includes ObjectNode) {
	r.includes = includes
}

func (r *QueryResultBase) getIncludedPaths() []string {
	return r.includedPaths
}

func (r *QueryResultBase) setIncludedPaths(includedPaths []string) {
	r.includedPaths = includedPaths
}

func (r *QueryResultBase) isStale() bool {
	return r._isStale
}

func (r *QueryResultBase) setStale(stale bool) {
	r._isStale = stale
}

func (r *QueryResultBase) getIndexTimestamp() time.Time {
	return r.indexTimestamp
}

func (r *QueryResultBase) setIndexTimestamp(indexTimestamp time.Time) {
	r.indexTimestamp = indexTimestamp
}

func (r *QueryResultBase) getIndexName() string {
	return r.indexName
}

func (r *QueryResultBase) setIndexName(indexName string) {
	r.indexName = indexName
}

func (r *QueryResultBase) getResultEtag() int64 {
	return r.resultEtag
}

func (r *QueryResultBase) setResultEtag(resultEtag int64) {
	r.resultEtag = resultEtag
}

func (r *QueryResultBase) getLastQueryTime() time.Time {
	return r.lastQueryTime
}

func (r *QueryResultBase) setLastQueryTime(lastQueryTime time.Time) {
	r.lastQueryTime = lastQueryTime
}
