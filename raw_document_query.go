package ravendb

import (
	"reflect"
	"time"
)

type RawDocumentQuery struct {
	*AbstractDocumentQuery
}

func NewRawDocumentQueryOld(clazz reflect.Type, session *InMemoryDocumentSessionOperations, rawQuery string) *RawDocumentQuery {
	res := &RawDocumentQuery{}
	res.AbstractDocumentQuery = NewAbstractDocumentQueryOld(clazz, session, "", "", false, nil, nil, "")
	res.queryRaw = rawQuery
	return res
}

func NewRawDocumentQuery(session *InMemoryDocumentSessionOperations, rawQuery string) *RawDocumentQuery {
	res := &RawDocumentQuery{}
	res.AbstractDocumentQuery = NewAbstractDocumentQuery(session, "", "", false, nil, nil, "")
	res.queryRaw = rawQuery
	return res
}

func (q *RawDocumentQuery) Skip(count int) *IRawDocumentQuery {
	q._skip(count)
	return q
}

func (q *RawDocumentQuery) Take(count int) *IRawDocumentQuery {
	q._take(&count)
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
	q._noTracking()
	return q
}

func (q *RawDocumentQuery) NoCaching() *IRawDocumentQuery {
	q._noCaching()
	return q
}

func (q *RawDocumentQuery) UsingDefaultOperator(queryOperator QueryOperator) *IRawDocumentQuery {
	q._usingDefaultOperator(queryOperator)
	return q
}

func (q *RawDocumentQuery) Statistics(stats **QueryStatistics) *IRawDocumentQuery {
	q._statistics(stats)
	return q
}

func (q *RawDocumentQuery) RemoveAfterQueryExecutedListener(idx int) *IRawDocumentQuery {
	q._removeAfterQueryExecutedListener(idx)
	return q
}

func (q *RawDocumentQuery) AddAfterQueryExecutedListener(action func(*QueryResult)) *IRawDocumentQuery {
	q._addAfterQueryExecutedListener(action)
	return q
}

func (q *RawDocumentQuery) AddBeforeQueryExecutedListener(action func(*IndexQuery)) *IRawDocumentQuery {
	q._addBeforeQueryExecutedListener(action)
	return q
}

func (q *RawDocumentQuery) RemoveBeforeQueryExecutedListener(idx int) *IRawDocumentQuery {
	q._removeBeforeQueryExecutedListener(idx)
	return q
}

func (q *RawDocumentQuery) AddAfterStreamExecutedListener(action func(ObjectNode)) *IRawDocumentQuery {
	q._addAfterStreamExecutedListener(action)
	return q
}

func (q *RawDocumentQuery) RemoveAfterStreamExecutedListener(idx int) *IRawDocumentQuery {
	q._removeAfterStreamExecutedListener(idx)
	return q
}

func (q *RawDocumentQuery) AddParameter(name string, value Object) *IRawDocumentQuery {
	q._addParameter(name, value)
	return q
}
