package ravendb

import (
	"time"
)

// Note: Java's RawDocumentQuery is RawDocumentQuery

type RawDocumentQuery struct {
	*abstractDocumentQuery
}

func (q *RawDocumentQuery) Skip(count int) *RawDocumentQuery {
	q.skip(count)
	return q
}

func (q *RawDocumentQuery) Take(count int) *RawDocumentQuery {
	q.take(&count)
	return q
}

func (q *RawDocumentQuery) WaitForNonStaleResults() *RawDocumentQuery {
	q.waitForNonStaleResults(0)
	return q
}

func (q *RawDocumentQuery) WaitForNonStaleResultsWithTimeout(waitTimeout time.Duration) *RawDocumentQuery {
	q.waitForNonStaleResults(waitTimeout)
	return q
}

//TBD 4.1  RawDocumentQuery<T> showTimings() {

func (q *RawDocumentQuery) NoTracking() *RawDocumentQuery {
	q.noTracking()
	return q
}

func (q *RawDocumentQuery) NoCaching() *RawDocumentQuery {
	q.noCaching()
	return q
}

func (q *RawDocumentQuery) UsingDefaultOperator(queryOperator QueryOperator) *RawDocumentQuery {
	q.usingDefaultOperator(queryOperator)
	return q
}

func (q *RawDocumentQuery) Statistics(stats **QueryStatistics) *RawDocumentQuery {
	q.statistics(stats)
	return q
}

func (q *RawDocumentQuery) RemoveAfterQueryExecutedListener(idx int) *RawDocumentQuery {
	q.removeAfterQueryExecutedListener(idx)
	return q
}

func (q *RawDocumentQuery) AddAfterQueryExecutedListener(action func(*QueryResult)) *RawDocumentQuery {
	q.addAfterQueryExecutedListener(action)
	return q
}

func (q *RawDocumentQuery) AddBeforeQueryExecutedListener(action func(*IndexQuery)) *RawDocumentQuery {
	q.addBeforeQueryExecutedListener(action)
	return q
}

func (q *RawDocumentQuery) RemoveBeforeQueryExecutedListener(idx int) *RawDocumentQuery {
	q.removeBeforeQueryExecutedListener(idx)
	return q
}

func (q *RawDocumentQuery) AddAfterStreamExecutedListener(action func(map[string]interface{})) *RawDocumentQuery {
	q.addAfterStreamExecutedListener(action)
	return q
}

func (q *RawDocumentQuery) RemoveAfterStreamExecutedListener(idx int) *RawDocumentQuery {
	q.removeAfterStreamExecutedListener(idx)
	return q
}

func (q *RawDocumentQuery) AddParameter(name string, value interface{}) *RawDocumentQuery {
	q.addParameter(name, value)
	return q
}
