package ravendb

import (
	"time"
)

type IRawDocumentQuery = RawDocumentQuery

type RawDocumentQuery struct {
	*AbstractDocumentQuery
}

func NewRawDocumentQuery(session *InMemoryDocumentSessionOperations, rawQuery string) *RawDocumentQuery {
	res := &RawDocumentQuery{}
	res.AbstractDocumentQuery = NewAbstractDocumentQuery(session, "", "", false, nil, nil, "")
	res.queryRaw = rawQuery
	return res
}

func (q *RawDocumentQuery) Skip(count int) *IRawDocumentQuery {
	q.skip(count)
	return q
}

func (q *RawDocumentQuery) Take(count int) *IRawDocumentQuery {
	q.take(&count)
	return q
}

func (q *RawDocumentQuery) WaitForNonStaleResults() *IRawDocumentQuery {
	q._waitForNonStaleResults(0)
	return q
}

func (q *RawDocumentQuery) WaitForNonStaleResultsWithTimeout(waitTimeout time.Duration) *IRawDocumentQuery {
	q._waitForNonStaleResults(waitTimeout)
	return q
}

//TBD 4.1  IRawDocumentQuery<T> showTimings() {

func (q *RawDocumentQuery) NoTracking() *IRawDocumentQuery {
	q.noTracking()
	return q
}

func (q *RawDocumentQuery) NoCaching() *IRawDocumentQuery {
	q.noCaching()
	return q
}

func (q *RawDocumentQuery) UsingDefaultOperator(queryOperator QueryOperator) *IRawDocumentQuery {
	q._usingDefaultOperator(queryOperator)
	return q
}

func (q *RawDocumentQuery) Statistics(stats **QueryStatistics) *IRawDocumentQuery {
	q.statistics(stats)
	return q
}

func (q *RawDocumentQuery) RemoveAfterQueryExecutedListener(idx int) *IRawDocumentQuery {
	q.removeAfterQueryExecutedListener(idx)
	return q
}

func (q *RawDocumentQuery) AddAfterQueryExecutedListener(action func(*QueryResult)) *IRawDocumentQuery {
	q.addAfterQueryExecutedListener(action)
	return q
}

func (q *RawDocumentQuery) AddBeforeQueryExecutedListener(action func(*IndexQuery)) *IRawDocumentQuery {
	q.addBeforeQueryExecutedListener(action)
	return q
}

func (q *RawDocumentQuery) RemoveBeforeQueryExecutedListener(idx int) *IRawDocumentQuery {
	q.removeBeforeQueryExecutedListener(idx)
	return q
}

func (q *RawDocumentQuery) AddAfterStreamExecutedListener(action func(map[string]interface{})) *IRawDocumentQuery {
	q.addAfterStreamExecutedListener(action)
	return q
}

func (q *RawDocumentQuery) RemoveAfterStreamExecutedListener(idx int) *IRawDocumentQuery {
	q.removeAfterStreamExecutedListener(idx)
	return q
}

func (q *RawDocumentQuery) AddParameter(name string, value interface{}) *IRawDocumentQuery {
	q.addParameter(name, value)
	return q
}
