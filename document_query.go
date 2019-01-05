package ravendb

import (
	"reflect"
	"time"
)

// TODO: remove those interfaces
type IDocumentQueryBase = DocumentQuery
type IDocumentQueryBaseSingle = DocumentQuery
type IDocumentQuery = DocumentQuery
type IFilterDocumentQueryBase = DocumentQuery

// DocumentQuery describes a query
type DocumentQuery struct {
	*AbstractDocumentQuery
}

// NewDocumentQuery returns new DocumentQuery
func NewDocumentQuery(session *InMemoryDocumentSessionOperations, indexName string, collectionName string, isGroupBy bool) *DocumentQuery {
	return &DocumentQuery{
		AbstractDocumentQuery: NewAbstractDocumentQuery(session, indexName, collectionName, isGroupBy, nil, nil, ""),
	}
}

func NewDocumentQueryOld(clazz reflect.Type, session *InMemoryDocumentSessionOperations, indexName string, collectionName string, isGroupBy bool) *DocumentQuery {
	return &DocumentQuery{
		AbstractDocumentQuery: NewAbstractDocumentQueryOld(clazz, session, indexName, collectionName, isGroupBy, nil, nil, ""),
	}
}

func NewDocumentQueryType(clazz reflect.Type, session *InMemoryDocumentSessionOperations, indexName string, collectionName string, isGroupBy bool) *DocumentQuery {
	return &DocumentQuery{
		AbstractDocumentQuery: NewAbstractDocumentQueryOld(clazz, session, indexName, collectionName, isGroupBy, nil, nil, ""),
	}
}

func NewDocumentQueryWithTokenOld(clazz reflect.Type, session *InMemoryDocumentSessionOperations, indexName string, collectionName string, isGroupBy bool, declareToken *declareToken, loadTokens []*loadToken, fromAlias string) *DocumentQuery {
	return &DocumentQuery{
		AbstractDocumentQuery: NewAbstractDocumentQueryOld(clazz, session, indexName, collectionName, isGroupBy, declareToken, loadTokens, fromAlias),
	}
}

// SelectFields limits the returned values to one or more fields of the queried type.
// To select fields for the whole type, do:
// fields := ravendb.FieldsFor(&MyType{})
// q = q.SelectFields(fields...)
func (q *DocumentQuery) SelectFields(fields ...string) *DocumentQuery {
	// Note: we delay executing the logic until ToList because only then
	// we know the type of the result
	panicIf(len(fields) == 0, "must provide at least one field")
	for _, field := range fields {
		panicIf(field == "", "field cannot be empty string")
	}
	q.selectFieldsArgs = &QueryData{
		Fields:      fields,
		Projections: fields,
	}
	return q
}

/*
TODO: should expose this version?
func (q *DocumentQuery) SelectFieldsWithQueryData(queryData *QueryData) *DocumentQuery {
	q.selectFieldsArgs = queryData
	return q
}
*/

// Distinct marks query as distinct
func (q *DocumentQuery) Distinct() *DocumentQuery {
	q._distinct()
	return q
}

// OrderByScore orders results of the query by score
func (q *DocumentQuery) OrderByScore() *DocumentQuery {
	q._orderByScore()
	return q
}

// OrderByScoreDescending orders results of the query by score
// in descending order
func (q *DocumentQuery) OrderByScoreDescending() *DocumentQuery {
	q._orderByScoreDescending()
	return q
}

//TBD 4.1  IDocumentQuery<T> explainScores() {

func (q *DocumentQuery) WaitForNonStaleResults(waitTimeout time.Duration) *DocumentQuery {
	q._waitForNonStaleResults(waitTimeout)
	return q
}

func (q *DocumentQuery) AddParameter(name string, value interface{}) *IDocumentQuery {
	q._addParameter(name, value)
	return q
}

func (q *DocumentQuery) AddOrder(fieldName string, descending bool) *IDocumentQuery {
	return q.AddOrderWithOrdering(fieldName, descending, OrderingTypeString)
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
 IDocumentQuery<T> AddAfterQueryExecutedListener(Consumer<QueryResult> action) {
	_addAfterQueryExecutedListener(action);
	return this;
}


 IDocumentQuery<T> RemoveAfterQueryExecutedListener(Consumer<QueryResult> action) {
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

func (q *DocumentQuery) ContainsAny(fieldName string, values []interface{}) *DocumentQuery {
	q._containsAny(fieldName, values)
	return q
}

//TBD expr  IDocumentQuery<T> ContainsAny<TValue>(Expression<Func<T, TValue>> propertySelector, IEnumerable<TValue> values)

func (q *DocumentQuery) ContainsAll(fieldName string, values []interface{}) *DocumentQuery {
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
	q._whereLucene(fieldName, whereClause)
	return q
}

func (q *DocumentQuery) WhereEquals(fieldName string, value interface{}) *DocumentQuery {
	q._whereEquals(fieldName, value)
	return q
}

// Exact marks previous Where statement (e.g. WhereEquals or WhereLucene) as exact
func (q *DocumentQuery) Exact() *DocumentQuery {
	q.markLastTokenExact()
	return q
}

func (q *DocumentQuery) WhereEqualsWithMethodCall(fieldName string, method MethodCall) *DocumentQuery {
	q._whereEqualsWithMethodCall(fieldName, method)
	return q
}

//TBD expr IDocumentQuery<T> IDocumentQueryBase<T, IDocumentQuery<T>>.WhereEquals<TValue>(Expression<Func<T, TValue>> propertySelector, TValue value, bool exact)
//TBD expr IDocumentQuery<T> IFilterDocumentQueryBase<T, IDocumentQuery<T>>.WhereEquals<TValue>(Expression<Func<T, TValue>> propertySelector, MethodCall value, bool exact)

func (q *DocumentQuery) WhereEqualsWithParams(whereParams *whereParams) *DocumentQuery {
	q._whereEqualsWithParams(whereParams)
	return q
}

func (q *DocumentQuery) WhereNotEquals(fieldName string, value interface{}) *DocumentQuery {
	q._whereNotEquals(fieldName, value)
	return q
}

func (q *DocumentQuery) WhereNotEqualsWithMethod(fieldName string, method MethodCall) *DocumentQuery {
	q._whereNotEqualsWithMethod(fieldName, method)
	return q
}

//TBD expr IDocumentQuery<T> IDocumentQueryBase<T, IDocumentQuery<T>>.WhereNotEquals<TValue>(Expression<Func<T, TValue>> propertySelector, TValue value, bool exact)
//TBD expr IDocumentQuery<T> IFilterDocumentQueryBase<T, IDocumentQuery<T>>.WhereNotEquals<TValue>(Expression<Func<T, TValue>> propertySelector, MethodCall value, bool exact)

func (q *DocumentQuery) WhereNotEqualsWithParams(whereParams *whereParams) *DocumentQuery {
	q._whereNotEqualsWithParams(whereParams)
	return q
}

func (q *DocumentQuery) WhereIn(fieldName string, values []interface{}) *DocumentQuery {
	q._whereIn(fieldName, values)
	return q
}

//TBD expr  IDocumentQuery<T> WhereIn<TValue>(Expression<Func<T, TValue>> propertySelector, IEnumerable<TValue> values, bool exact = false)

func (q *DocumentQuery) WhereStartsWith(fieldName string, value interface{}) *DocumentQuery {
	q._whereStartsWith(fieldName, value)
	return q
}

func (q *DocumentQuery) WhereEndsWith(fieldName string, value interface{}) *DocumentQuery {
	q._whereEndsWith(fieldName, value)
	return q
}

//TBD expr  IDocumentQuery<T> WhereEndsWith<TValue>(Expression<Func<T, TValue>> propertySelector, TValue value)

func (q *DocumentQuery) WhereBetween(fieldName string, start interface{}, end interface{}) *DocumentQuery {
	q._whereBetween(fieldName, start, end)
	return q
}

//TBD expr  IDocumentQuery<T> WhereBetween<TValue>(Expression<Func<T, TValue>> propertySelector, TValue start, TValue end, bool exact = false)

func (q *DocumentQuery) WhereGreaterThan(fieldName string, value interface{}) *DocumentQuery {
	q._whereGreaterThan(fieldName, value)
	return q
}

func (q *DocumentQuery) WhereGreaterThanOrEqual(fieldName string, value interface{}) *DocumentQuery {
	q._whereGreaterThanOrEqual(fieldName, value)
	return q
}

//TBD expr  IDocumentQuery<T> WhereGreaterThan<TValue>(Expression<Func<T, TValue>> propertySelector, TValue value, bool exact = false)
//TBD expr  IDocumentQuery<T> WhereGreaterThanOrEqual<TValue>(Expression<Func<T, TValue>> propertySelector, TValue value, bool exact = false)

func (q *DocumentQuery) WhereLessThan(fieldName string, value interface{}) *DocumentQuery {
	q._whereLessThan(fieldName, value)
	return q
}

//TBD expr  IDocumentQuery<T> WhereLessThanOrEqual<TValue>(Expression<Func<T, TValue>> propertySelector, TValue value, bool exact = false)

func (q *DocumentQuery) WhereLessThanOrEqual(fieldName string, value interface{}) *DocumentQuery {
	q._whereLessThanOrEqual(fieldName, value)
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

// GroupBy makes a query grouped by fields
func (q *DocumentQuery) GroupBy(fieldName string, fieldNames ...string) *IGroupByDocumentQuery {
	q._groupBy(fieldName, fieldNames...)

	return NewGroupByDocumentQuery(q)
}

func (q *DocumentQuery) GroupBy2(field *GroupBy, fields ...*GroupBy) *IGroupByDocumentQuery {
	q._groupBy2(field, fields...)

	return NewGroupByDocumentQuery(q)
}

// OrderBy makes a query ordered by a given field
func (q *DocumentQuery) OrderBy(field string) *IDocumentQuery {
	return q.OrderByWithOrdering(field, OrderingTypeString)
}

func (q *DocumentQuery) OrderByWithOrdering(field string, ordering OrderingType) *IDocumentQuery {
	q._orderByWithOrdering(field, ordering)
	return q
}

//TBD expr  IDocumentQuery<T> OrderBy<TValue>(params Expression<Func<T, TValue>>[] propertySelectors)

func (q *DocumentQuery) OrderByDescending(field string) *IDocumentQuery {
	return q.OrderByDescendingWithOrdering(field, OrderingTypeString)
}

func (q *DocumentQuery) OrderByDescendingWithOrdering(field string, ordering OrderingType) *IDocumentQuery {
	q._orderByDescendingWithOrdering(field, ordering)
	return q
}

//TBD expr  IDocumentQuery<T> OrderByDescending<TValue>(params Expression<Func<T, TValue>>[] propertySelectors)

/*
 IDocumentQuery<T> AddBeforeQueryExecutedListener(Consumer<IndexQuery> action) {
	_addBeforeQueryExecutedListener(action);
	return this;
}


 IDocumentQuery<T> RemoveBeforeQueryExecutedListener(Consumer<IndexQuery> action) {
	_removeBeforeQueryExecutedListener(action);
	return this;
}
*/

// Note: had to move it down to AbstractDocumentQuery
func (q *AbstractDocumentQuery) createDocumentQueryInternal(resultClass reflect.Type, queryData *QueryData) *DocumentQuery {

	var newFieldsToFetch *fieldsToFetchToken

	if queryData != nil && len(queryData.Fields) > 0 {
		fields := queryData.Fields

		identityProperty := q.getConventions().GetIdentityProperty(resultClass)

		if identityProperty != "" {
			// make a copy, just in case, because we might modify it
			fields = append([]string{}, fields...)

			for idx, p := range fields {
				if p == identityProperty {
					fields[idx] = IndexingFieldNameDocumentID
				}
			}
		}

		newFieldsToFetch = createFieldsToFetchToken(fields, queryData.Projections, queryData.IsCustomFunction)
	}

	if newFieldsToFetch != nil {
		q.updateFieldsToFetchToken(newFieldsToFetch)
	}

	var declareToken *declareToken
	var loadTokens []*loadToken
	var fromAlias string
	if queryData != nil {
		declareToken = queryData.DeclareToken
		loadTokens = queryData.LoadTokens
		fromAlias = queryData.FromAlias
	}
	query := NewDocumentQueryWithTokenOld(resultClass,
		q.theSession,
		q.indexName,
		q.collectionName,
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
	query.includes = stringArrayCopy(q.includes)
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
func (q *DocumentQuery) AggregateBy(builder func(IFacetBuilder)) *AggregationDocumentQuery {
	ff := NewFacetBuilder()
	builder(ff)

	return q.AggregateByFacet(ff.getFacet())
}

func (q *DocumentQuery) AggregateByFacet(facet FacetBase) *AggregationDocumentQuery {
	q._aggregateBy(facet)

	return NewAggregationDocumentQuery(q)
}

func (q *DocumentQuery) AggregateByFacets(facets ...*Facet) *AggregationDocumentQuery {
	for _, facet := range facets {
		q._aggregateBy(facet)
	}

	return NewAggregationDocumentQuery(q)
}

func (q *DocumentQuery) AggregateUsing(facetSetupDocumentID string) *AggregationDocumentQuery {
	q._aggregateUsing(facetSetupDocumentID)

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
	criteria := clause(spatialCriteriaFactoryInstance)
	q._spatial3(fieldName, criteria)
	return q
}

func (q *DocumentQuery) Spatial2(field DynamicSpatialField, clause func(*SpatialCriteriaFactory) SpatialCriteria) *IDocumentQuery {
	criteria := clause(spatialCriteriaFactoryInstance)
	q._spatial2(field, criteria)
	return q
}

//TBD expr  IDocumentQuery<T> Spatial(Func<SpatialDynamicFieldFactory<T>, DynamicSpatialField> field, Func<SpatialCriteriaFactory, SpatialCriteria> clause)
//TBD expr IDocumentQuery<T> IDocumentQueryBase<T, IDocumentQuery<T>>.WithinRadiusOf<TValue>(Expression<Func<T, TValue>> propertySelector, float64 radius, float64 latitude, float64 longitude, SpatialUnits? radiusUnits, float64 distanceErrorPct)

func (q *DocumentQuery) WithinRadiusOf(fieldName string, radius float64, latitude float64, longitude float64) *IDocumentQuery {
	q._withinRadiusOf(fieldName, radius, latitude, longitude, "", IndexingSpatialDefaultDistnaceErrorPct)
	return q
}

func (q *DocumentQuery) WithinRadiusOfWithUnits(fieldName string, radius float64, latitude float64, longitude float64, radiusUnits SpatialUnits) *IDocumentQuery {
	q._withinRadiusOf(fieldName, radius, latitude, longitude, radiusUnits, IndexingSpatialDefaultDistnaceErrorPct)
	return q
}

func (q *DocumentQuery) WithinRadiusOfWithUnitsAndError(fieldName string, radius float64, latitude float64, longitude float64, radiusUnits SpatialUnits, distanceErrorPct float64) *IDocumentQuery {
	q._withinRadiusOf(fieldName, radius, latitude, longitude, radiusUnits, distanceErrorPct)
	return q
}

//TBD expr IDocumentQuery<T> IDocumentQueryBase<T, IDocumentQuery<T>>.RelatesToShape<TValue>(Expression<Func<T, TValue>> propertySelector, string shapeWkt, SpatialRelation relation, float64 distanceErrorPct)

func (q *DocumentQuery) RelatesToShape(fieldName string, shapeWkt string, relation SpatialRelation) *IDocumentQuery {
	return q.RelatesToShapeWithError(fieldName, shapeWkt, relation, IndexingSpatialDefaultDistnaceErrorPct)
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

func (q *DocumentQuery) OrderByDistanceDescendingLatLong(fieldName string, latitude float64, longitude float64) *IDocumentQuery {
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

	mlt.WithOptions(moreLikeThis.GetOptions())

	if mltud, ok := moreLikeThis.(*MoreLikeThisUsingDocument); ok {
		mlt.withDocument(mltud.documentJSON)

	}

	return q
}

func (q *DocumentQuery) MoreLikeThisWithBuilder(builder func(IMoreLikeThisBuilderForDocumentQuery)) *DocumentQuery {
	f := NewMoreLikeThisBuilder()
	builder(f)

	moreLikeThis := q._moreLikeThis()

	moreLikeThis.WithOptions(f.GetMoreLikeThis().GetOptions())

	tmp := f.GetMoreLikeThis()
	if mlt, ok := tmp.(*MoreLikeThisUsingDocument); ok {
		moreLikeThis.withDocument(mlt.documentJSON)
	} else if mlt, ok := tmp.(*MoreLikeThisUsingDocumentForDocumentQuery); ok {
		mlt.GetForDocumentQuery()(q)
	}
	moreLikeThis.Close()

	return q
}

func (q *DocumentQuery) SuggestUsing(suggestion SuggestionBase) *ISuggestionDocumentQuery {
	q._suggestUsing(suggestion)
	return NewSuggestionDocumentQuery(q)
}

func (q *DocumentQuery) SuggestUsingBuilder(builder func(ISuggestionBuilder)) *ISuggestionDocumentQuery {
	f := NewSuggestionBuilder()
	builder(f)

	q.SuggestUsing(f.getSuggestion())
	return NewSuggestionDocumentQuery(q)
}
