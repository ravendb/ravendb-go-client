package ravendb

type GroupByDocumentQuery struct {
	_query *DocumentQuery
}

func NewGroupByDocumentQuery(query *DocumentQuery) *GroupByDocumentQuery {
	return &GroupByDocumentQuery{
		_query: query,
	}
}

func (q *GroupByDocumentQuery) selectKey() *IGroupByDocumentQuery {
	return q.selectKeyWithNameAndProjectedName("", "")
}

func (q *GroupByDocumentQuery) selectKeyWithName(fieldName string) *IGroupByDocumentQuery {
	return q.selectKeyWithNameAndProjectedName(fieldName, "")
}

func (q *GroupByDocumentQuery) selectKeyWithNameAndProjectedName(fieldName string, projectedName string) *IGroupByDocumentQuery {
	q._query._groupByKey(fieldName, projectedName)
	return q
}

func (q *GroupByDocumentQuery) selectSum(field *GroupByField, fields ...*GroupByField) *IDocumentQuery {
	if field == nil {
		panic("Field cannot be null")
		//throw new IllegalArgumentException("Field cannot be null");
	}

	q._query._groupBySum(field.getFieldName(), field.getProjectedName())

	if len(fields) == 0 {
		return q._query
	}

	for _, f := range fields {
		q._query._groupBySum(f.getFieldName(), f.getProjectedName())
	}

	return q._query
}

func (q *GroupByDocumentQuery) selectCount() *IDocumentQuery {
	return q.selectCountWithName("count")
}

func (q *GroupByDocumentQuery) selectCountWithName(projectedName string) *IDocumentQuery {
	q._query._groupByCount(projectedName)
	return q._query
}
