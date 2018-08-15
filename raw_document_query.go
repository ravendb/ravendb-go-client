package ravendb

import (
	"reflect"
	"time"
)

type RawDocumentQuery struct {
	*AbstractDocumentQuery
}

func NewRawDocumentQuery(clazz reflect.Type, session *InMemoryDocumentSessionOperations, rawQuery string) *RawDocumentQuery {
	res := &RawDocumentQuery{}
	res.AbstractDocumentQuery = NewAbstractDocumentQuery(clazz, session, "", "", false, nil, nil, "")
	res.queryRaw = rawQuery
	return res
}

func (q *RawDocumentQuery) skip(count int) *IRawDocumentQuery {
	q._skip(count)
	return q
}

func (q *RawDocumentQuery) take(count int) *IRawDocumentQuery {
	q._take(&count)
	return q
}

func (q *RawDocumentQuery) waitForNonStaleResults() *IRawDocumentQuery {
	q._waitForNonStaleResults(0)
	return q
}

func (q *RawDocumentQuery) waitForNonStaleResultsWithTimeout(waitTimeout time.Duration) *IRawDocumentQuery {
	q._waitForNonStaleResults(waitTimeout)
	return q
}

//TBD 4.1  IRawDocumentQuery<T> showTimings() {

func (q *RawDocumentQuery) noTracking() *IRawDocumentQuery {
	q._noTracking()
	return q
}

func (q *RawDocumentQuery) noCaching() *IRawDocumentQuery {
	q._noCaching()
	return q
}

func (q *RawDocumentQuery) usingDefaultOperator(queryOperator QueryOperator) *IRawDocumentQuery {
	q._usingDefaultOperator(queryOperator)
	return q
}

func (q *RawDocumentQuery) statistics(stats *QueryStatistics) *IRawDocumentQuery {
	q._statistics(stats)
	return q
}

func (q *RawDocumentQuery) removeAfterQueryExecutedListener(idx int) *IRawDocumentQuery {
	q._removeAfterQueryExecutedListener(idx)
	return q
}

func (q *RawDocumentQuery) addAfterQueryExecutedListener(action func(*QueryResult)) *IRawDocumentQuery {
	q._addAfterQueryExecutedListener(action)
	return q
}

func (q *RawDocumentQuery) addBeforeQueryExecutedListener(action func(*IndexQuery)) *IRawDocumentQuery {
	q._addBeforeQueryExecutedListener(action)
	return q
}

func (q *RawDocumentQuery) removeBeforeQueryExecutedListener(idx int) *IRawDocumentQuery {
	q._removeBeforeQueryExecutedListener(idx)
	return q
}

func (q *RawDocumentQuery) addAfterStreamExecutedListener(action func(ObjectNode)) *IRawDocumentQuery {
	q._addAfterStreamExecutedListener(action)
	return q
}

func (q *RawDocumentQuery) removeAfterStreamExecutedListener(idx int) *IRawDocumentQuery {
	q._removeAfterStreamExecutedListener(idx)
	return q
}

func (q *RawDocumentQuery) addParameter(name string, value Object) *IRawDocumentQuery {
	q._addParameter(name, value)
	return q
}
