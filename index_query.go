package ravendb

import (
	"math"
	"time"
)

type IndexQuery struct {

	// from IndexQueryBase<T>
	_pageSize                     int
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

// from IndexQuery
func NewIndexQuery(query string) *IndexQuery {
	return &IndexQuery{
		query:     query,
		_pageSize: math.MaxInt32,
	}
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

func (q *IndexQuery) isDisableCaching() bool {
	return q.disableCaching
}

func (q *IndexQuery) setDisableCaching(disableCaching bool) {
	q.disableCaching = disableCaching
}

func (q *IndexQuery) getQueryHash() string {
	hasher := NewQueryHashCalculator()
	hasher.write(q.getQuery())
	hasher.write(q.isWaitForNonStaleResults())
	hasher.write(q.isSkipDuplicateChecking())
	//TBD 4.1 hasher.write(isShowTimings());
	//TBD 4.1 hasher.write(isExplainScores());
	n := int64(q.getWaitForNonStaleResultsTimeout())
	hasher.write(n)
	hasher.write(q.getStart())
	hasher.write(q.getPageSize())
	hasher.write(q.getQueryParameters())
	return hasher.getHash()
}

func (q *IndexQuery) String() string {
	return q.query
}
