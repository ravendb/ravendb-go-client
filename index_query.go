package ravendb

import "time"

// TODO: implement me
type IndexQuery struct {

	// from IndexQueryBase<T>
	_pageSize                     int // = Integer.MAX_VALUE;
	pageSizeSet                   bool
	query                         string
	queryParameters               Parameters
	start                         int
	waitForNonStaleResults        bool
	waitForNonStaleResultsTimeout time.Duration

	// from IndexQueryWithParameters
	skipDuplicateChecking bool

	// from IndexQuery
	disableCaching bool
}

// from IndexQueryBase<T>
func (q *IndexQuery) isPageSizeSet() bool {
	return q.pageSizeSet
}

func (q *IndexQuery) getQuery() string {
	return q.query
}

func (q *IndexQuery) setQuery(query string) {
	q.query = query
}

func (q *IndexQuery) getQueryParameters() Parameters {
	return q.queryParameters
}

func (q *IndexQuery) setQueryParameters(queryParameters Parameters) {
	q.queryParameters = queryParameters
}

func (q *IndexQuery) getStart() int {
	return q.start
}

func (q *IndexQuery) setStart(start int) {
	q.start = start
}

func (q *IndexQuery) getPageSize() int {
	return q._pageSize
}

func (q *IndexQuery) setPageSize(pageSize int) {
	q._pageSize = pageSize
	q.pageSizeSet = true
}

func (q *IndexQuery) isWaitForNonStaleResults() bool {
	return q.waitForNonStaleResults
}

func (q *IndexQuery) setWaitForNonStaleResults(waitForNonStaleResults bool) {
	q.waitForNonStaleResults = waitForNonStaleResults
}

func (q *IndexQuery) getWaitForNonStaleResultsTimeout() time.Duration {
	return q.waitForNonStaleResultsTimeout
}

func (q *IndexQuery) setWaitForNonStaleResultsTimeout(waitForNonStaleResultsTimeout time.Duration) {
	q.waitForNonStaleResultsTimeout = waitForNonStaleResultsTimeout
}

// from IndexQueryWithParameters
func (q *IndexQuery) isSkipDuplicateChecking() bool {
	return q.skipDuplicateChecking
}

func (q *IndexQuery) setSkipDuplicateChecking(skipDuplicateChecking bool) {
	q.skipDuplicateChecking = skipDuplicateChecking
}

// from IndexQuery
func NewIndexQuery(query string) *IndexQuery {
	return &IndexQuery{
		query: query,
	}
}

func (q *IndexQuery) isDisableCaching() bool {
	return q.disableCaching
}

func (q *IndexQuery) setDisableCaching(disableCaching bool) {
	q.disableCaching = disableCaching
}

func (q *IndexQuery) getQueryHash() string {
	// TODO: port me
	panicIf(true, "NYI")
	/*
		QueryHashCalculator hasher = new QueryHashCalculator();
			hasher.write(getQuery());
			hasher.write(isWaitForNonStaleResults());
			hasher.write(isSkipDuplicateChecking());
			//TBD 4.1 hasher.write(isShowTimings());
			//TBD 4.1 hasher.write(isExplainScores());
			hasher.write(Optional.ofNullable(getWaitForNonStaleResultsTimeout()).map(x -> x.toMillis()).orElse(0L));
			hasher.write(getStart());
			hasher.write(getPageSize());
			hasher.write(getQueryParameters());
			return hasher.getHash();
	*/
	return ""
}
