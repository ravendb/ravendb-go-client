package ravendb

// GroupByDocumentQuery represents a "group by" query
type GroupByDocumentQuery struct {
	query *DocumentQuery
	err   error
}

func newGroupByDocumentQuery(query *DocumentQuery) *GroupByDocumentQuery {
	return &GroupByDocumentQuery{
		query: query,
	}
}

func (q *GroupByDocumentQuery) SelectKey() *GroupByDocumentQuery {
	if q.err != nil {
		return q
	}
	return q.SelectKeyWithNameAndProjectedName("", "")
}

func (q *GroupByDocumentQuery) SelectKeyWithName(fieldName string) *GroupByDocumentQuery {
	if q.err != nil {
		return q
	}
	return q.SelectKeyWithNameAndProjectedName(fieldName, "")
}

func (q *GroupByDocumentQuery) SelectKeyWithNameAndProjectedName(fieldName string, projectedName string) *GroupByDocumentQuery {
	if q.err != nil {
		return q
	}
	q.err = q.query.groupByKey(fieldName, projectedName)
	return q
}

func (q *GroupByDocumentQuery) SelectSum(field *GroupByField, fields ...*GroupByField) *DocumentQuery {
	if q.err != nil {
		q.query.err = q.err
		return q.query
	}

	if field == nil {
		q.err = newIllegalArgumentError("Field cannot be null")
	}

	q.err = q.query.groupBySum(field.FieldName, field.ProjectedName)
	if q.err != nil {
		q.query.err = q.err
		return q.query
	}

	if len(fields) == 0 {
		return q.query
	}

	for _, f := range fields {
		q.err = q.query.groupBySum(f.FieldName, f.ProjectedName)
		if q.err != nil {
			q.query.err = q.err
			break
		}
	}

	return q.query
}

func (q *GroupByDocumentQuery) SelectCount() *DocumentQuery {
	return q.SelectCountWithName("count")
}

func (q *GroupByDocumentQuery) SelectCountWithName(projectedName string) *DocumentQuery {
	if q.err != nil {
		q.query.err = q.err
		return q.query
	}
	q.err = q.query.groupByCount(projectedName)
	if q.err != nil {
		q.query.err = q.err
	}
	return q.query
}
