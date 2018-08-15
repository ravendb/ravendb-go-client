package ravendb

import (
	"reflect"
	"time"
)

type DocumentQuery struct {
	*AbstractDocumentQuery
}

func NewDocumentQuery(clazz reflect.Type, session *InMemoryDocumentSessionOperations, indexName string, collectionName string, isGroupBy bool) *DocumentQuery {
	return &DocumentQuery{
		AbstractDocumentQuery: NewAbstractDocumentQuery(clazz, session, indexName, collectionName, isGroupBy, nil, nil, ""),
	}
}

func NewDocumentQueryWithToken(clazz reflect.Type, session *InMemoryDocumentSessionOperations, indexName string, collectionName string, isGroupBy bool, declareToken *DeclareToken, loadTokens []*LoadToken, fromAlias string) *DocumentQuery {
	return &DocumentQuery{
		AbstractDocumentQuery: NewAbstractDocumentQuery(clazz, session, indexName, collectionName, isGroupBy, declareToken, loadTokens, fromAlias),
	}
}

func (q *DocumentQuery) SelectFields(projectionClass reflect.Type, fields ...string) *DocumentQuery {
	if len(fields) > 0 {
		queryData := NewQueryData(fields, fields)
		return q.SelectFieldsWithQueryData(projectionClass, queryData)
	}

	projections := getJSONStructFieldNames(projectionClass)
	newFields := projections // java re-does the same calculations
	return q.SelectFieldsWithQueryData(projectionClass, NewQueryData(newFields, projections))
}

func (q *DocumentQuery) SelectFieldsWithQueryData(projectionClass reflect.Type, queryData *QueryData) *DocumentQuery {
	return q.createDocumentQueryInternalWithQueryData(projectionClass, queryData)
}

func (q *DocumentQuery) Distinct() *DocumentQuery {
	q._distinct()
	return q
}

func (q *DocumentQuery) OrderByScore() *DocumentQuery {
	q._orderByScore()
	return q
}

func (q *DocumentQuery) OrderByScoreDescending() *DocumentQuery {
	q._orderByScoreDescending()
	return q
}

//TBD 4.1  IDocumentQuery<T> explainScores() {

func (q *DocumentQuery) WaitForNonStaleResults(waitTimeout time.Duration) *DocumentQuery {
	q._waitForNonStaleResults(waitTimeout)
	return q
}

func (q *DocumentQuery) AddParameter(name string, value Object) *IDocumentQuery {
	q._addParameter(name, value)
	return q
}

func (q *DocumentQuery) AddOrder(fieldName string, descending bool) *IDocumentQuery {
	return q.AddOrderWithOrdering(fieldName, descending, OrderingType_STRING)
}

func (q *DocumentQuery) AddOrderWithOrdering(fieldName string, descending bool, ordering OrderingType) *IDocumentQuery {
	if descending {
		q.OrderByDescendingWithOrdering(fieldName, ordering)
	} else {
		q.OrderByWithOrdering(fieldName, ordering)
	}
	return q
}

//TBD expr  IDocumentQuery<T> AddOrder<TValue>(Expression<Func<T, TValue>> propertySelector, bool descending, OrderingType ordering)

/*
 IDocumentQuery<T> addAfterQueryExecutedListener(Consumer<QueryResult> action) {
	_addAfterQueryExecutedListener(action);
	return this;
}


 IDocumentQuery<T> removeAfterQueryExecutedListener(Consumer<QueryResult> action) {
	_removeAfterQueryExecutedListener(action);
	return this;
}


 IDocumentQuery<T> addAfterStreamExecutedListener(Consumer<ObjectNode> action) {
	_addAfterStreamExecutedListener(action);
	return this;
}


 IDocumentQuery<T> removeAfterStreamExecutedListener(Consumer<ObjectNode> action) {
	_removeAfterStreamExecutedListener(action);
	return this;
}
*/

func (q *DocumentQuery) OpenSubclause() *IDocumentQuery {
	q._openSubclause()
	return q
}

func (q *DocumentQuery) CloseSubclause() *IDocumentQuery {
	q._closeSubclause()
	return q
}

func (q *DocumentQuery) Search(fieldName string, searchTerms string) *IDocumentQuery {
	q._search(fieldName, searchTerms)
	return q
}

func (q *DocumentQuery) SearchWithOperator(fieldName string, searchTerms string, operator SearchOperator) *IDocumentQuery {
	q._searchWithOperator(fieldName, searchTerms, operator)
	return q
}

//TBD expr  IDocumentQuery<T> Search<TValue>(Expression<Func<T, TValue>> propertySelector, string searchTerms, SearchOperator @operator)

func (q *DocumentQuery) Intersect() *IDocumentQuery {
	q._intersect()
	return q
}

func (q *DocumentQuery) ContainsAny(fieldName string, values []Object) *DocumentQuery {
	q._containsAny(fieldName, values)
	return q
}

//TBD expr  IDocumentQuery<T> ContainsAny<TValue>(Expression<Func<T, TValue>> propertySelector, IEnumerable<TValue> values)

func (q *DocumentQuery) ContainsAll(fieldName string, values []Object) *DocumentQuery {
	q._containsAll(fieldName, values)
	return q
}

//TBD expr  IDocumentQuery<T> ContainsAll<TValue>(Expression<Func<T, TValue>> propertySelector, IEnumerable<TValue> values)

func (q *DocumentQuery) Statistics(stats **QueryStatistics) *DocumentQuery {
	q._statistics(stats)
	return q
}

func (q *DocumentQuery) UsingDefaultOperator(queryOperator QueryOperator) *IDocumentQuery {
	q._usingDefaultOperator(queryOperator)
	return q
}

func (q *DocumentQuery) NoTracking() *IDocumentQuery {
	q._noTracking()
	return q
}

func (q *DocumentQuery) NoCaching() *IDocumentQuery {
	q._noCaching()
	return q
}

//TBD 4.1  IDocumentQuery<T> showTimings()

func (q *DocumentQuery) Include(path string) *IDocumentQuery {
	q._include(path)
	return q
}

//TBD expr IDocumentQuery<T> IDocumentQueryBase<T, IDocumentQuery<T>>.Include(Expression<Func<T, object>> path)

func (q *DocumentQuery) Not() *DocumentQuery {
	q.NegateNext()
	return q
}

func (q *DocumentQuery) Take(count int) *DocumentQuery {
	q._take(&count)
	return q
}

func (q *DocumentQuery) Skip(count int) *DocumentQuery {
	q._skip(count)
	return q
}

func (q *DocumentQuery) WhereLucene(fieldName string, whereClause string) *IDocumentQuery {
	q._whereLucene(fieldName, whereClause, false)
	return q
}

func (q *DocumentQuery) WhereLuceneWithExact(fieldName string, whereClause string, exact bool) *IDocumentQuery {
	q._whereLucene(fieldName, whereClause, exact)
	return q
}

func (q *DocumentQuery) WhereEquals(fieldName string, value Object) *DocumentQuery {
	q._whereEqualsWithExact(fieldName, value, false)
	return q
}

func (q *DocumentQuery) WhereEqualsWithExact(fieldName string, value Object, exact bool) *DocumentQuery {
	q._whereEqualsWithExact(fieldName, value, exact)
	return q
}

func (q *DocumentQuery) WhereEqualsWithMethodCall(fieldName string, method MethodCall, exact bool) *DocumentQuery {
	q._whereEqualsWithMethodCall(fieldName, method, exact)
	return q
}

//TBD expr IDocumentQuery<T> IDocumentQueryBase<T, IDocumentQuery<T>>.WhereEquals<TValue>(Expression<Func<T, TValue>> propertySelector, TValue value, bool exact)
//TBD expr IDocumentQuery<T> IFilterDocumentQueryBase<T, IDocumentQuery<T>>.WhereEquals<TValue>(Expression<Func<T, TValue>> propertySelector, MethodCall value, bool exact)

func (q *DocumentQuery) WhereEqualsWithParams(whereParams *WhereParams) *DocumentQuery {
	q._whereEqualsWithParams(whereParams)
	return q
}

func (q *DocumentQuery) WhereNotEquals(fieldName string, value Object) *DocumentQuery {
	q._whereNotEquals(fieldName, value)
	return q
}

func (q *DocumentQuery) WhereNotEqualsWithExact(fieldName string, value Object, exact bool) *DocumentQuery {
	q._whereNotEqualsWithExact(fieldName, value, exact)
	return q
}

func (q *DocumentQuery) _whereNotEqualsWithMethod(fieldName string, method MethodCall) *DocumentQuery {
	q._whereNotEqualsWithMethod(fieldName, method)
	return q
}

func (q *DocumentQuery) _whereNotEqualsWithMethodAndExact(fieldName string, method MethodCall, exact bool) *DocumentQuery {
	q._whereNotEqualsWithMethodAndExact(fieldName, method, exact)
	return q
}

//TBD expr IDocumentQuery<T> IDocumentQueryBase<T, IDocumentQuery<T>>.WhereNotEquals<TValue>(Expression<Func<T, TValue>> propertySelector, TValue value, bool exact)
//TBD expr IDocumentQuery<T> IFilterDocumentQueryBase<T, IDocumentQuery<T>>.WhereNotEquals<TValue>(Expression<Func<T, TValue>> propertySelector, MethodCall value, bool exact)

func (q *DocumentQuery) WhereNotEqualsWithParams(whereParams *WhereParams) *DocumentQuery {
	q._whereNotEqualsWithParams(whereParams)
	return q
}

func (q *DocumentQuery) WhereIn(fieldName string, values []Object) *DocumentQuery {
	return q.WhereInWithExact(fieldName, values, false)
}

func (q *DocumentQuery) WhereInWithExact(fieldName string, values []Object, exact bool) *DocumentQuery {
	q._whereInWithExact(fieldName, values, exact)
	return q
}

//TBD expr  IDocumentQuery<T> WhereIn<TValue>(Expression<Func<T, TValue>> propertySelector, IEnumerable<TValue> values, bool exact = false)

func (q *DocumentQuery) WhereStartsWith(fieldName string, value Object) *DocumentQuery {
	q._whereStartsWith(fieldName, value)
	return q
}

func (q *DocumentQuery) WhereEndsWith(fieldName string, value Object) *DocumentQuery {
	q._whereEndsWith(fieldName, value)
	return q
}

//TBD expr  IDocumentQuery<T> WhereEndsWith<TValue>(Expression<Func<T, TValue>> propertySelector, TValue value)

func (q *DocumentQuery) WhereBetween(fieldName string, start Object, end Object) *DocumentQuery {
	return q.WhereBetweenWithExact(fieldName, start, end, false)
}

func (q *DocumentQuery) WhereBetweenWithExact(fieldName string, start Object, end Object, exact bool) *DocumentQuery {
	q._whereBetweenWithExact(fieldName, start, end, exact)
	return q
}

//TBD expr  IDocumentQuery<T> WhereBetween<TValue>(Expression<Func<T, TValue>> propertySelector, TValue start, TValue end, bool exact = false)

func (q *DocumentQuery) WhereGreaterThan(fieldName string, value Object) *DocumentQuery {
	return q.WhereGreaterThanWithExact(fieldName, value, false)
}

func (q *DocumentQuery) WhereGreaterThanWithExact(fieldName string, value Object, exact bool) *DocumentQuery {
	q._whereGreaterThanWithExact(fieldName, value, exact)
	return q
}

func (q *DocumentQuery) WhereGreaterThanOrEqual(fieldName string, value Object) *DocumentQuery {
	return q.WhereGreaterThanOrEqualWithExact(fieldName, value, false)
}

func (q *DocumentQuery) WhereGreaterThanOrEqualWithExact(fieldName string, value Object, exact bool) *DocumentQuery {
	q._whereGreaterThanOrEqualWithExact(fieldName, value, exact)
	return q
}

//TBD expr  IDocumentQuery<T> WhereGreaterThan<TValue>(Expression<Func<T, TValue>> propertySelector, TValue value, bool exact = false)
//TBD expr  IDocumentQuery<T> WhereGreaterThanOrEqual<TValue>(Expression<Func<T, TValue>> propertySelector, TValue value, bool exact = false)

func (q *DocumentQuery) WhereLessThan(fieldName string, value Object) *DocumentQuery {
	return q.WhereLessThanWithExact(fieldName, value, false)
}

func (q *DocumentQuery) WhereLessThanWithExact(fieldName string, value Object, exact bool) *DocumentQuery {
	q._whereLessThanWithExact(fieldName, value, exact)
	return q
}

//TBD expr  IDocumentQuery<T> WhereLessThanOrEqual<TValue>(Expression<Func<T, TValue>> propertySelector, TValue value, bool exact = false)

func (q *DocumentQuery) WhereLessThanOrEqual(fieldName string, value Object) *DocumentQuery {
	return q.WhereLessThanOrEqualWithExact(fieldName, value, false)
}

func (q *DocumentQuery) WhereLessThanOrEqualWithExact(fieldName string, value Object, exact bool) *DocumentQuery {
	q._whereLessThanOrEqualWithExact(fieldName, value, exact)
	return q
}

//TBD expr  IDocumentQuery<T> WhereLessThanOrEqual<TValue>(Expression<Func<T, TValue>> propertySelector, TValue value, bool exact = false)
//TBD expr  IDocumentQuery<T> WhereExists<TValue>(Expression<Func<T, TValue>> propertySelector)

func (q *DocumentQuery) WhereExists(fieldName string) *DocumentQuery {
	q._whereExists(fieldName)
	return q
}

//TBD expr IDocumentQuery<T> IFilterDocumentQueryBase<T, IDocumentQuery<T>>.WhereRegex<TValue>(Expression<Func<T, TValue>> propertySelector, string pattern)

func (q *DocumentQuery) WhereRegex(fieldName string, pattern string) *DocumentQuery {
	q._whereRegex(fieldName, pattern)
	return q
}

func (q *DocumentQuery) AndAlso() *DocumentQuery {
	q._andAlso()
	return q
}

func (q *DocumentQuery) OrElse() *DocumentQuery {
	q._orElse()
	return q
}

func (q *DocumentQuery) Boost(boost float64) *DocumentQuery {
	q._boost(boost)
	return q
}

func (q *DocumentQuery) Fuzzy(fuzzy float64) *DocumentQuery {
	q._fuzzy(fuzzy)
	return q
}

func (q *DocumentQuery) Proximity(proximity int) *DocumentQuery {
	q._proximity(proximity)
	return q
}

func (q *DocumentQuery) RandomOrdering() *DocumentQuery {
	q._randomOrdering()
	return q
}

func (q *DocumentQuery) RandomOrderingWithSeed(seed string) *DocumentQuery {
	q._randomOrderingWithSeed(seed)
	return q
}

//TBD 4.1  IDocumentQuery<T> customSortUsing(string typeName, bool descending)

func (q *DocumentQuery) GroupBy(fieldName string, fieldNames ...string) *IGroupByDocumentQuery {
	q._groupBy(fieldName, fieldNames...)

	return NewGroupByDocumentQuery(q)
}

func (q *DocumentQuery) GroupBy2(field *GroupBy, fields ...*GroupBy) *IGroupByDocumentQuery {
	q._groupBy2(field, fields...)

	return NewGroupByDocumentQuery(q)
}

func (q *DocumentQuery) OfType(tResultClass reflect.Type) *IDocumentQuery {
	return q.createDocumentQueryInternal(tResultClass)
}

func (q *DocumentQuery) OrderBy(field string) *IDocumentQuery {
	return q.OrderByWithOrdering(field, OrderingType_STRING)
}

func (q *DocumentQuery) OrderByWithOrdering(field string, ordering OrderingType) *IDocumentQuery {
	q._orderByWithOrdering(field, ordering)
	return q
}

//TBD expr  IDocumentQuery<T> OrderBy<TValue>(params Expression<Func<T, TValue>>[] propertySelectors)

func (q *DocumentQuery) OrderByDescending(field string) *IDocumentQuery {
	return q.OrderByDescendingWithOrdering(field, OrderingType_STRING)
}

func (q *DocumentQuery) OrderByDescendingWithOrdering(field string, ordering OrderingType) *IDocumentQuery {
	q._orderByDescendingWithOrdering(field, ordering)
	return q
}

//TBD expr  IDocumentQuery<T> OrderByDescending<TValue>(params Expression<Func<T, TValue>>[] propertySelectors)

/*
 IDocumentQuery<T> addBeforeQueryExecutedListener(Consumer<IndexQuery> action) {
	_addBeforeQueryExecutedListener(action);
	return this;
}


 IDocumentQuery<T> removeBeforeQueryExecutedListener(Consumer<IndexQuery> action) {
	_removeBeforeQueryExecutedListener(action);
	return this;
}
*/

func (q *DocumentQuery) createDocumentQueryInternal(resultClass reflect.Type) *DocumentQuery {
	return q.createDocumentQueryInternalWithQueryData(resultClass, nil)
}

func (q *DocumentQuery) createDocumentQueryInternalWithQueryData(resultClass reflect.Type, queryData *QueryData) *DocumentQuery {

	var newFieldsToFetch *FieldsToFetchToken

	if queryData != nil && len(queryData.getFields()) > 0 {
		fields := queryData.getFields()

		identityProperty := q.GetConventions().getIdentityProperty(resultClass)

		if identityProperty != "" {
			// make a copy, just in case, because we might modify it
			fields = append([]string{}, fields...)

			for idx, p := range fields {
				if p == identityProperty {
					fields[idx] = Constants_Documents_Indexing_Fields_DOCUMENT_ID_FIELD_NAME
				}
			}
		}

		newFieldsToFetch = FieldsToFetchToken_create(fields, queryData.getProjections(), queryData.IsCustomFunction())
	}

	if newFieldsToFetch != nil {
		q.updateFieldsToFetchToken(newFieldsToFetch)
	}

	var declareToken *DeclareToken
	var loadTokens []*LoadToken
	var fromAlias string
	if queryData != nil {
		declareToken = queryData.getDeclareToken()
		loadTokens = queryData.getLoadTokens()
		fromAlias = queryData.getFromAlias()
	}
	query := NewDocumentQueryWithToken(resultClass,
		q.theSession,
		q.GetIndexName(),
		q.GetCollectionName(),
		q.isGroupBy,
		declareToken,
		loadTokens,
		fromAlias)

	query.queryRaw = q.queryRaw
	query.pageSize = q.pageSize
	query.selectTokens = q.selectTokens
	query.fieldsToFetchToken = q.fieldsToFetchToken
	query.whereTokens = q.whereTokens
	query.orderByTokens = q.orderByTokens
	query.groupByTokens = q.groupByTokens
	query.queryParameters = q.queryParameters
	query.start = q.start
	query.timeout = q.timeout
	query.queryStats = q.queryStats
	query.theWaitForNonStaleResults = q.theWaitForNonStaleResults
	query.negate = q.negate
	//noinspection unchecked
	query.includes = NewStringSetFromStrings(q.includes.Strings()...)
	query.rootTypes = NewTypeSetWithType(q.clazz)
	// TODO: should this be deep copy so that adding/removing in one
	// doesn't affect the other?
	query.beforeQueryExecutedCallback = q.beforeQueryExecutedCallback
	query.afterQueryExecutedCallback = q.afterQueryExecutedCallback
	query.afterStreamExecutedCallback = q.afterStreamExecutedCallback
	query.disableEntitiesTracking = q.disableEntitiesTracking
	query.disableCaching = q.disableCaching
	//TBD 4.1 ShowQueryTimings = ShowQueryTimings,
	//TBD 4.1 query.shouldExplainScores = shouldExplainScores;
	query.isIntersect = q.isIntersect
	query.defaultOperator = q.defaultOperator

	return query
}

// TODO: rename to aggregateByBuilder and aggregateByFacet => aggregateBy
func (q *DocumentQuery) AggregateBy(builder func(IFacetBuilder)) *IAggregationDocumentQuery {
	ff := NewFacetBuilder()
	builder(ff)

	return q.AggregateByFacet(ff.getFacet())
}

func (q *DocumentQuery) AggregateByFacet(facet FacetBase) *IAggregationDocumentQuery {
	q._aggregateBy(facet)

	return NewAggregationDocumentQuery(q)
}

func (q *DocumentQuery) AggregateByFacets(facets ...*Facet) *IAggregationDocumentQuery {
	for _, facet := range facets {
		q._aggregateBy(facet)
	}

	return NewAggregationDocumentQuery(q)
}

func (q *DocumentQuery) AggregateUsing(facetSetupDocumentId string) *IAggregationDocumentQuery {
	q._aggregateUsing(facetSetupDocumentId)

	return NewAggregationDocumentQuery(q)
}

//TBD 4.1 IDocumentQuery<T> IDocumentQueryBase<T, IDocumentQuery<T>>.Highlight(string fieldName, int fragmentLength, int fragmentCount, string fragmentsField)
//TBD 4.1 IDocumentQuery<T> IDocumentQueryBase<T, IDocumentQuery<T>>.Highlight(string fieldName, int fragmentLength, int fragmentCount, out FieldHighlightings highlightings)
//TBD 4.1 IDocumentQuery<T> IDocumentQueryBase<T, IDocumentQuery<T>>.Highlight(string fieldName,string fieldKeyName, int fragmentLength,int fragmentCount,out FieldHighlightings highlightings)
//TBD 4.1 IDocumentQuery<T> IDocumentQueryBase<T, IDocumentQuery<T>>.Highlight<TValue>(Expression<Func<T, TValue>> propertySelector, int fragmentLength, int fragmentCount, Expression<Func<T, IEnumerable>> fragmentsPropertySelector)
//TBD 4.1 IDocumentQuery<T> IDocumentQueryBase<T, IDocumentQuery<T>>.Highlight<TValue>(Expression<Func<T, TValue>> propertySelector, int fragmentLength, int fragmentCount, out FieldHighlightings fieldHighlightings)
//TBD 4.1 IDocumentQuery<T> IDocumentQueryBase<T, IDocumentQuery<T>>.Highlight<TValue>(Expression<Func<T, TValue>> propertySelector, Expression<Func<T, TValue>> keyPropertySelector, int fragmentLength, int fragmentCount, out FieldHighlightings fieldHighlightings)
//TBD 4.1 IDocumentQuery<T> IDocumentQueryBase<T, IDocumentQuery<T>>.SetHighlighterTags(string preTag, string postTag)
//TBD 4.1 IDocumentQuery<T> IDocumentQueryBase<T, IDocumentQuery<T>>.SetHighlighterTags(string[] preTags, string[] postTags)
//TBD expr  IDocumentQuery<T> Spatial(Expression<Func<T, object>> path, Func<SpatialCriteriaFactory, SpatialCriteria> clause)

func (q *DocumentQuery) Spatial3(fieldName string, clause func(*SpatialCriteriaFactory) SpatialCriteria) *IDocumentQuery {
	criteria := clause(SpatialCriteriaFactory_INSTANCE)
	q._spatial3(fieldName, criteria)
	return q
}

func (q *DocumentQuery) Spatial2(field DynamicSpatialField, clause func(*SpatialCriteriaFactory) SpatialCriteria) *IDocumentQuery {
	criteria := clause(SpatialCriteriaFactory_INSTANCE)
	q._spatial2(field, criteria)
	return q
}

//TBD expr  IDocumentQuery<T> Spatial(Func<SpatialDynamicFieldFactory<T>, DynamicSpatialField> field, Func<SpatialCriteriaFactory, SpatialCriteria> clause)
//TBD expr IDocumentQuery<T> IDocumentQueryBase<T, IDocumentQuery<T>>.WithinRadiusOf<TValue>(Expression<Func<T, TValue>> propertySelector, float64 radius, float64 latitude, float64 longitude, SpatialUnits? radiusUnits, float64 distanceErrorPct)

func (q *DocumentQuery) WithinRadiusOf(fieldName string, radius float64, latitude float64, longitude float64) *IDocumentQuery {
	q._withinRadiusOf(fieldName, radius, latitude, longitude, "", Constants_Documents_Indexing_Spatial_DEFAULT_DISTANCE_ERROR_PCT)
	return q
}

func (q *DocumentQuery) WithinRadiusOfWithUnits(fieldName string, radius float64, latitude float64, longitude float64, radiusUnits SpatialUnits) *IDocumentQuery {
	q._withinRadiusOf(fieldName, radius, latitude, longitude, radiusUnits, Constants_Documents_Indexing_Spatial_DEFAULT_DISTANCE_ERROR_PCT)
	return q
}

func (q *DocumentQuery) WithinRadiusOfWithUnitsAndError(fieldName string, radius float64, latitude float64, longitude float64, radiusUnits SpatialUnits, distanceErrorPct float64) *IDocumentQuery {
	q._withinRadiusOf(fieldName, radius, latitude, longitude, radiusUnits, distanceErrorPct)
	return q
}

//TBD expr IDocumentQuery<T> IDocumentQueryBase<T, IDocumentQuery<T>>.RelatesToShape<TValue>(Expression<Func<T, TValue>> propertySelector, string shapeWkt, SpatialRelation relation, float64 distanceErrorPct)

func (q *DocumentQuery) RelatesToShape(fieldName string, shapeWkt string, relation SpatialRelation) *IDocumentQuery {
	return q.RelatesToShapeWithError(fieldName, shapeWkt, relation, Constants_Documents_Indexing_Spatial_DEFAULT_DISTANCE_ERROR_PCT)
}

func (q *DocumentQuery) RelatesToShapeWithError(fieldName string, shapeWkt string, relation SpatialRelation, distanceErrorPct float64) *IDocumentQuery {
	q._spatial(fieldName, shapeWkt, relation, distanceErrorPct)
	return q
}

func (q *DocumentQuery) OrderByDistance(field DynamicSpatialField, latitude float64, longitude float64) *IDocumentQuery {
	q._orderByDistance(field, latitude, longitude)
	return q
}

//TBD expr IDocumentQuery<T> IDocumentQueryBase<T, IDocumentQuery<T>>.OrderByDistance(Func<DynamicSpatialFieldFactory<T>, DynamicSpatialField> field, float64 latitude, float64 longitude)

func (q *DocumentQuery) OrderByDistance2(field DynamicSpatialField, shapeWkt string) *IDocumentQuery {
	q._orderByDistance2(field, shapeWkt)
	return q
}

//TBD expr IDocumentQuery<T> IDocumentQueryBase<T, IDocumentQuery<T>>.OrderByDistance(Func<DynamicSpatialFieldFactory<T>, DynamicSpatialField> field, string shapeWkt)

//TBD expr IDocumentQuery<T> IDocumentQueryBase<T, IDocumentQuery<T>>.OrderByDistance<TValue>(Expression<Func<T, TValue>> propertySelector, float64 latitude, float64 longitude)

func (q *DocumentQuery) OrderByDistanceLatLong(fieldName string, latitude float64, longitude float64) *IDocumentQuery {
	q._orderByDistanceLatLong(fieldName, latitude, longitude)
	return q
}

//TBD expr IDocumentQuery<T> IDocumentQueryBase<T, IDocumentQuery<T>>.OrderByDistance<TValue>(Expression<Func<T, TValue>> propertySelector, string shapeWkt)

func (q *DocumentQuery) OrderByDistance3(fieldName string, shapeWkt string) *IDocumentQuery {
	q._orderByDistance3(fieldName, shapeWkt)
	return q
}

func (q *DocumentQuery) OrderByDistanceDescending(field DynamicSpatialField, latitude float64, longitude float64) *IDocumentQuery {
	q._orderByDistanceDescending(field, latitude, longitude)
	return q
}

//TBD expr IDocumentQuery<T> IDocumentQueryBase<T, IDocumentQuery<T>>.OrderByDistanceDescending(Func<DynamicSpatialFieldFactory<T>, DynamicSpatialField> field, float64 latitude, float64 longitude)

func (q *DocumentQuery) OrderByDistanceDescending2(field DynamicSpatialField, shapeWkt string) *IDocumentQuery {
	q._orderByDistanceDescending2(field, shapeWkt)
	return q
}

//TBD expr IDocumentQuery<T> IDocumentQueryBase<T, IDocumentQuery<T>>.OrderByDistanceDescending(Func<DynamicSpatialFieldFactory<T>, DynamicSpatialField> field, string shapeWkt)

//TBD expr IDocumentQuery<T> IDocumentQueryBase<T, IDocumentQuery<T>>.OrderByDistanceDescending<TValue>(Expression<Func<T, TValue>> propertySelector, float64 latitude, float64 longitude)

func (q *DocumentQuery) orderByDistanceDescendingLatLong(fieldName string, latitude float64, longitude float64) *IDocumentQuery {
	q._orderByDistanceDescendingLatLong(fieldName, latitude, longitude)
	return q
}

//TBD expr IDocumentQuery<T> IDocumentQueryBase<T, IDocumentQuery<T>>.OrderByDistanceDescending<TValue>(Expression<Func<T, TValue>> propertySelector, string shapeWkt)

func (q *DocumentQuery) OrderByDistanceDescending3(fieldName string, shapeWkt string) *IDocumentQuery {
	q._orderByDistanceDescending3(fieldName, shapeWkt)
	return q
}

func (q *DocumentQuery) MoreLikeThis(moreLikeThis MoreLikeThisBase) *DocumentQuery {
	mlt := q._moreLikeThis()
	defer mlt.Close()

	mlt.withOptions(moreLikeThis.getOptions())

	if mltud, ok := moreLikeThis.(*MoreLikeThisUsingDocument); ok {
		mlt.withDocument(mltud.getDocumentJson())

	}

	return q
}

func (q *DocumentQuery) MoreLikeThisWithBuilder(builder func(IMoreLikeThisBuilderForDocumentQuery)) *DocumentQuery {
	f := NewMoreLikeThisBuilder()
	builder(f)

	moreLikeThis := q._moreLikeThis()
	defer moreLikeThis.Close()

	moreLikeThis.withOptions(f.getMoreLikeThis().getOptions())

	tmp := f.getMoreLikeThis()
	if mlt, ok := tmp.(*MoreLikeThisUsingDocument); ok {
		moreLikeThis.withDocument(mlt.getDocumentJson())
	} else if mlt, ok := tmp.(*MoreLikeThisUsingDocumentForDocumentQuery); ok {
		mlt.getForDocumentQuery()(q)
	}

	return q
}

/*
 ISuggestionDocumentQuery<T> suggestUsing(SuggestionBase suggestion) {
	_suggestUsing(suggestion);
	return new SuggestionDocumentQuery<>(this);
}

 ISuggestionDocumentQuery<T> suggestUsing(Consumer<ISuggestionBuilder<T>> builder) {
	SuggestionBuilder<T> f = new SuggestionBuilder<>();
	builder.accept(f);

	suggestUsing(f.getSuggestion());
	return new SuggestionDocumentQuery<>(this);
}
*/
