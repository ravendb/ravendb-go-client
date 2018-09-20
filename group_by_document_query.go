package ravendb

type IGroupByDocumentQuery = GroupByDocumentQuery

type GroupByDocumentQuery struct {
	_query *DocumentQuery
}

func NewGroupByDocumentQuery(query *DocumentQuery) *GroupByDocumentQuery {
	return &GroupByDocumentQuery{
		_query: query,
	}
}

func (q *GroupByDocumentQuery) SelectKey() *IGroupByDocumentQuery {
	return q.SelectKeyWithNameAndProjectedName("", "")
}

func (q *GroupByDocumentQuery) SelectKeyWithName(fieldName string) *IGroupByDocumentQuery {
	return q.SelectKeyWithNameAndProjectedName(fieldName, "")
}

func (q *GroupByDocumentQuery) SelectKeyWithNameAndProjectedName(fieldName string, projectedName string) *IGroupByDocumentQuery {
	q._query._groupByKey(fieldName, projectedName)
	return q
}

func (q *GroupByDocumentQuery) SelectSum(field *GroupByField, fields ...*GroupByField) *IDocumentQuery {
	if field == nil {
		panic("Field cannot be null")
		//throw new IllegalArgumentException("Field cannot be null");
	}

	q._query._groupBySum(field.FieldName, field.ProjectedName)

	if len(fields) == 0 {
		return q._query
	}

	for _, f := range fields {
		q._query._groupBySum(f.FieldName, f.ProjectedName)
	}

	return q._query
}

func (q *GroupByDocumentQuery) SelectCount() *IDocumentQuery {
	return q.SelectCountWithName("count")
}

func (q *GroupByDocumentQuery) SelectCountWithName(projectedName string) *IDocumentQuery {
	q._query._groupByCount(projectedName)
	return q._query
}
