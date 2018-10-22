package ravendb

import (
	"time"
)

type IndexQuery struct {

	// from IndexQueryBase<T>
	pageSize                      int // if 0, not set
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
		query:    query,
		pageSize: 0,
	}
}

// from IndexQueryBase<T>

// TODO: only for tests? Could expose with build-tags only for testing
func (q *IndexQuery) GetQuery() string {
	return q.query
}

// TODO: only for tests? Could expose with build-tags only for testing
func (q *IndexQuery) GetQueryParameters() Parameters {
	return q.queryParameters
}

func (q *IndexQuery) GetQueryHash() string {
	hasher := NewQueryHashCalculator()
	hasher.write(q.query)
	hasher.write(q.waitForNonStaleResults)
	hasher.write(q.skipDuplicateChecking)
	//TBD 4.1 hasher.write(isShowTimings());
	//TBD 4.1 hasher.write(isExplainScores());
	n := int64(q.waitForNonStaleResultsTimeout)
	hasher.write(n)
	hasher.write(q.start)
	hasher.write(q.pageSize)
	hasher.write(q.queryParameters)
	return hasher.getHash()
}

func (q *IndexQuery) String() string {
	return q.query
}
