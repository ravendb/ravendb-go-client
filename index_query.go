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
func (q *IndexQuery) IsPageSizeSet() bool {
	return q.pageSizeSet
}

func (q *IndexQuery) GetQuery() string {
	return q.query
}

func (q *IndexQuery) SetQuery(query string) {
	q.query = query
}

func (q *IndexQuery) GetQueryParameters() Parameters {
	return q.queryParameters
}

func (q *IndexQuery) SetQueryParameters(queryParameters Parameters) {
	q.queryParameters = queryParameters
}

func (q *IndexQuery) GetStart() int {
	return q.start
}

func (q *IndexQuery) SetStart(start int) {
	q.start = start
}

func (q *IndexQuery) GetPageSize() int {
	return q._pageSize
}

func (q *IndexQuery) SetPageSize(pageSize int) {
	q._pageSize = pageSize
	q.pageSizeSet = true
}

func (q *IndexQuery) IsWaitForNonStaleResults() bool {
	return q.waitForNonStaleResults
}

func (q *IndexQuery) SetWaitForNonStaleResults(waitForNonStaleResults bool) {
	q.waitForNonStaleResults = waitForNonStaleResults
}

func (q *IndexQuery) GetWaitForNonStaleResultsTimeout() time.Duration {
	return q.waitForNonStaleResultsTimeout
}

func (q *IndexQuery) SetWaitForNonStaleResultsTimeout(waitForNonStaleResultsTimeout time.Duration) {
	q.waitForNonStaleResultsTimeout = waitForNonStaleResultsTimeout
}

// from IndexQueryWithParameters
func (q *IndexQuery) IsSkipDuplicateChecking() bool {
	return q.skipDuplicateChecking
}

func (q *IndexQuery) SetSkipDuplicateChecking(skipDuplicateChecking bool) {
	q.skipDuplicateChecking = skipDuplicateChecking
}

func (q *IndexQuery) IsDisableCaching() bool {
	return q.disableCaching
}

func (q *IndexQuery) SetDisableCaching(disableCaching bool) {
	q.disableCaching = disableCaching
}

func (q *IndexQuery) GetQueryHash() string {
	hasher := NewQueryHashCalculator()
	hasher.write(q.GetQuery())
	hasher.write(q.IsWaitForNonStaleResults())
	hasher.write(q.IsSkipDuplicateChecking())
	//TBD 4.1 hasher.write(isShowTimings());
	//TBD 4.1 hasher.write(isExplainScores());
	n := int64(q.GetWaitForNonStaleResultsTimeout())
	hasher.write(n)
	hasher.write(q.GetStart())
	hasher.write(q.GetPageSize())
	hasher.write(q.GetQueryParameters())
	return hasher.getHash()
}

func (q *IndexQuery) String() string {
	return q.query
}
