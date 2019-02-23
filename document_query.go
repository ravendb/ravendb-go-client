package ravendb

import (
	"reflect"
	"time"
)

// Note: Java's IDocumentQueryBase is DocumentQuery
// Note: Java's IDocumentQueryBaseSingle is DocumentQuery
// Note: Java's IDocumentQuery is DocumentQuery
// Note: Java's IFilterDocumentQueryBase is DocumentQuery

// DocumentQuery describes a query
type DocumentQuery struct {
	*abstractDocumentQuery
}

// DocumentQueryOptions describes options for creating a query
type DocumentQueryOptions struct {
	// CollectionName and Type are mutually exclusive
	// if Collection is empty string we'll derive name of the collection
	// from Type
	CollectionName string
	Type           reflect.Type

	// name of the index used for search query
	// if set, CollectionName and Type should not be set
	IndexName string

	IsMapReduce bool

	conventions *DocumentConventions
	// rawQuery is mutually exclusive with IndexName and CollectionName/Type
	rawQuery string

	session      *InMemoryDocumentSessionOperations
	isGroupBy    bool
	declareToken *declareToken
	loadTokens   []*loadToken
	fromAlias    string
}

func newDocumentQuery(opts *DocumentQueryOptions) *DocumentQuery {

	var err error
	opts.IndexName, opts.CollectionName, err = processQueryParameters(opts.Type, opts.IndexName, opts.CollectionName, opts.conventions)
	aq := newAbstractDocumentQuery(opts)
	if err != nil {
		aq.err = err
	}
	return &DocumentQuery{
		abstractDocumentQuery: aq,
	}
}

// SelectFields limits the returned values to one or more fields of the queried type.
func (q *DocumentQuery) SelectFields(projectionType reflect.Type, fieldsIn ...string) *DocumentQuery {
	if q.err != nil {
		return q
	}
	// TODO: add SelectFieldsWithProjection(projectionType reflect.Type, fields []string, projections []string)
	var fields []string
	if len(fieldsIn) == 0 {
		fields = FieldsFor(projectionType)
		if len(fields) == 0 {
			q.err = newIllegalArgumentError("type %T has no exported fields to select", projectionType)
			return q
		}
	} else {
		fields = fieldsIn
	}

	queryData := &queryData{
		fields:      fields,
		projections: fields,
	}
	res, err := q.createDocumentQueryInternal(projectionType, queryData)
	if err != nil {
		res.err = err
	}
	return res
}

// Distinct marks query as distinct
func (q *DocumentQuery) Distinct() *DocumentQuery {
	if q.err != nil {
		return q
	}
	q.err = q.distinct()
	return q
}

// OrderByScore orders results of the query by score
func (q *DocumentQuery) OrderByScore() *DocumentQuery {
	if q.err != nil {
		return q
	}
	q.err = q.orderByScore()
	return q
}

// OrderByScoreDescending orders results of the query by score
// in descending order
func (q *DocumentQuery) OrderByScoreDescending() *DocumentQuery {
	if q.err != nil {
		return q
	}
	q.err = q.orderByScoreDescending()
	return q
}

//TBD 4.1  IDocumentQuery<T> explainScores() {

func (q *DocumentQuery) WaitForNonStaleResults(waitTimeout time.Duration) *DocumentQuery {
	if q.err != nil {
		return q
	}
	q.waitForNonStaleResults(waitTimeout)
	return q
}

func (q *DocumentQuery) AddParameter(name string, value interface{}) *DocumentQuery {
	if q.err != nil {
		return q
	}
	q.err = q.addParameter(name, value)
	return q
}

func (q *DocumentQuery) AddOrder(fieldName string, descending bool) *DocumentQuery {
	if q.err != nil {
		return q
	}
	return q.AddOrderWithOrdering(fieldName, descending, OrderingTypeString)
}

func (q *DocumentQuery) AddOrderWithOrdering(fieldName string, descending bool, ordering OrderingType) *DocumentQuery {
	if q.err != nil {
		return q
	}
	if descending {
		return q.OrderByDescendingWithOrdering(fieldName, ordering)
	}
	return q.OrderByWithOrdering(fieldName, ordering)
}

//TBD expr  IDocumentQuery<T> AddOrder<TValue>(Expression<Func<T, TValue>> propertySelector, bool descending, OrderingType ordering)

/*
 IDocumentQuery<T> AddAfterQueryExecutedListener(Consumer<QueryResult> action) {
	addAfterQueryExecutedListener(action);
	return this;
}


 IDocumentQuery<T> RemoveAfterQueryExecutedListener(Consumer<QueryResult> action) {
	removeAfterQueryExecutedListener(action);
	return this;
}


 IDocumentQuery<T> addAfterStreamExecutedListener(Consumer<ObjectNode> action) {
	addAfterStreamExecutedListener(action);
	return this;
}


 IDocumentQuery<T> removeAfterStreamExecutedListener(Consumer<ObjectNode> action) {
	removeAfterStreamExecutedListener(action);
	return this;
}
*/

func (q *DocumentQuery) OpenSubclause() *DocumentQuery {
	if q.err != nil {
		return q
	}
	q.err = q.openSubclause()
	return q
}

func (q *DocumentQuery) CloseSubclause() *DocumentQuery {
	if q.err != nil {
		return q
	}
	q.err = q.closeSubclause()
	return q
}

func (q *DocumentQuery) Search(fieldName string, searchTerms string) *DocumentQuery {
	if q.err != nil {
		return q
	}
	q.err = q.search(fieldName, searchTerms)
	return q
}

func (q *DocumentQuery) SearchWithOperator(fieldName string, searchTerms string, operator SearchOperator) *DocumentQuery {
	if q.err != nil {
		return q
	}
	q.err = q.searchWithOperator(fieldName, searchTerms, operator)
	return q
}

//TBD expr  IDocumentQuery<T> Search<TValue>(Expression<Func<T, TValue>> propertySelector, string searchTerms, SearchOperator @operator)

func (q *DocumentQuery) Intersect() *DocumentQuery {
	if q.err != nil {
		return q
	}
	q.err = q.intersect()
	return q
}

func (q *DocumentQuery) ContainsAny(fieldName string, values []interface{}) *DocumentQuery {
	if q.err != nil {
		return q
	}
	q.err = q.containsAny(fieldName, values)
	return q
}

//TBD expr  IDocumentQuery<T> ContainsAny<TValue>(Expression<Func<T, TValue>> propertySelector, IEnumerable<TValue> values)

func (q *DocumentQuery) ContainsAll(fieldName string, values []interface{}) *DocumentQuery {
	if q.err != nil {
		return q
	}
	q.err = q.containsAll(fieldName, values)
	return q
}

//TBD expr  IDocumentQuery<T> ContainsAll<TValue>(Expression<Func<T, TValue>> propertySelector, IEnumerable<TValue> values)

func (q *DocumentQuery) Statistics(stats **QueryStatistics) *DocumentQuery {
	q.statistics(stats)
	return q
}

func (q *DocumentQuery) UsingDefaultOperator(queryOperator QueryOperator) *DocumentQuery {
	if q.err != nil {
		return q
	}
	q.err = q.usingDefaultOperator(queryOperator)
	return q
}

func (q *DocumentQuery) NoTracking() *DocumentQuery {
	q.noTracking()
	return q
}

func (q *DocumentQuery) NoCaching() *DocumentQuery {
	q.noCaching()
	return q
}

//TBD 4.1  IDocumentQuery<T> showTimings()

func (q *DocumentQuery) Include(path string) *DocumentQuery {
	q.include(path)
	return q
}

//TBD expr IDocumentQuery<T> IDocumentQueryBase<T, IDocumentQuery<T>>.Include(Expression<Func<T, object>> path)

func (q *DocumentQuery) Not() *DocumentQuery {
	q.negateNext()
	return q
}

func (q *DocumentQuery) Take(count int) *DocumentQuery {
	q.take(count)
	return q
}

func (q *DocumentQuery) Skip(count int) *DocumentQuery {
	q.skip(count)
	return q
}

func (q *DocumentQuery) WhereLucene(fieldName string, whereClause string) *DocumentQuery {
	if q.err != nil {
		return q
	}
	q.err = q.whereLucene(fieldName, whereClause)
	return q
}

func (q *DocumentQuery) WhereEquals(fieldName string, value interface{}) *DocumentQuery {
	if q.err != nil {
		return q
	}
	q.err = q.whereEquals(fieldName, value)
	return q
}

// Exact marks previous Where statement (e.g. WhereEquals or WhereLucene) as exact
func (q *DocumentQuery) Exact() *DocumentQuery {
	if q.err != nil {
		return q
	}
	q.err = q.markLastTokenExact()
	return q
}

func (q *DocumentQuery) WhereEqualsWithMethodCall(fieldName string, method MethodCall) *DocumentQuery {
	if q.err != nil {
		return q
	}
	q.err = q.whereEqualsWithMethodCall(fieldName, method)
	return q
}

//TBD expr IDocumentQuery<T> IDocumentQueryBase<T, IDocumentQuery<T>>.WhereEquals<TValue>(Expression<Func<T, TValue>> propertySelector, TValue value, bool exact)
//TBD expr IDocumentQuery<T> IFilterDocumentQueryBase<T, IDocumentQuery<T>>.WhereEquals<TValue>(Expression<Func<T, TValue>> propertySelector, MethodCall value, bool exact)

func (q *DocumentQuery) WhereEqualsWithParams(whereParams *whereParams) *DocumentQuery {
	if q.err != nil {
		return q
	}
	q.err = q.whereEqualsWithParams(whereParams)
	return q
}

func (q *DocumentQuery) WhereNotEquals(fieldName string, value interface{}) *DocumentQuery {
	if q.err != nil {
		return q
	}
	q.err = q.whereNotEquals(fieldName, value)
	return q
}

func (q *DocumentQuery) WhereNotEqualsWithMethod(fieldName string, method MethodCall) *DocumentQuery {
	if q.err != nil {
		return q
	}
	q.err = q.whereNotEqualsWithMethod(fieldName, method)
	return q
}

//TBD expr IDocumentQuery<T> IDocumentQueryBase<T, IDocumentQuery<T>>.WhereNotEquals<TValue>(Expression<Func<T, TValue>> propertySelector, TValue value, bool exact)
//TBD expr IDocumentQuery<T> IFilterDocumentQueryBase<T, IDocumentQuery<T>>.WhereNotEquals<TValue>(Expression<Func<T, TValue>> propertySelector, MethodCall value, bool exact)

func (q *DocumentQuery) WhereNotEqualsWithParams(whereParams *whereParams) *DocumentQuery {
	if q.err != nil {
		return q
	}
	q.err = q.whereNotEqualsWithParams(whereParams)
	return q
}

func (q *DocumentQuery) WhereIn(fieldName string, values []interface{}) *DocumentQuery {
	if q.err != nil {
		return q
	}
	q.err = q.whereIn(fieldName, values)
	return q
}

//TBD expr  IDocumentQuery<T> WhereIn<TValue>(Expression<Func<T, TValue>> propertySelector, IEnumerable<TValue> values, bool exact = false)

func (q *DocumentQuery) WhereStartsWith(fieldName string, value interface{}) *DocumentQuery {
	if q.err != nil {
		return q
	}
	q.err = q.whereStartsWith(fieldName, value)
	return q
}

func (q *DocumentQuery) WhereEndsWith(fieldName string, value interface{}) *DocumentQuery {
	if q.err != nil {
		return q
	}
	q.err = q.whereEndsWith(fieldName, value)
	return q
}

//TBD expr  IDocumentQuery<T> WhereEndsWith<TValue>(Expression<Func<T, TValue>> propertySelector, TValue value)

func (q *DocumentQuery) WhereBetween(fieldName string, start interface{}, end interface{}) *DocumentQuery {
	if q.err != nil {
		return q
	}
	q.err = q.whereBetween(fieldName, start, end)
	return q
}

//TBD expr  IDocumentQuery<T> WhereBetween<TValue>(Expression<Func<T, TValue>> propertySelector, TValue start, TValue end, bool exact = false)

func (q *DocumentQuery) WhereGreaterThan(fieldName string, value interface{}) *DocumentQuery {
	if q.err != nil {
		return q
	}
	q.err = q.whereGreaterThan(fieldName, value)
	return q
}

func (q *DocumentQuery) WhereGreaterThanOrEqual(fieldName string, value interface{}) *DocumentQuery {
	if q.err != nil {
		return q
	}
	q.err = q.whereGreaterThanOrEqual(fieldName, value)
	return q
}

//TBD expr  IDocumentQuery<T> WhereGreaterThan<TValue>(Expression<Func<T, TValue>> propertySelector, TValue value, bool exact = false)
//TBD expr  IDocumentQuery<T> WhereGreaterThanOrEqual<TValue>(Expression<Func<T, TValue>> propertySelector, TValue value, bool exact = false)

func (q *DocumentQuery) WhereLessThan(fieldName string, value interface{}) *DocumentQuery {
	if q.err != nil {
		return q
	}
	q.err = q.whereLessThan(fieldName, value)
	return q
}

//TBD expr  IDocumentQuery<T> WhereLessThanOrEqual<TValue>(Expression<Func<T, TValue>> propertySelector, TValue value, bool exact = false)

func (q *DocumentQuery) WhereLessThanOrEqual(fieldName string, value interface{}) *DocumentQuery {
	if q.err != nil {
		return q
	}
	q.err = q.whereLessThanOrEqual(fieldName, value)
	return q
}

//TBD expr  IDocumentQuery<T> WhereLessThanOrEqual<TValue>(Expression<Func<T, TValue>> propertySelector, TValue value, bool exact = false)
//TBD expr  IDocumentQuery<T> WhereExists<TValue>(Expression<Func<T, TValue>> propertySelector)

func (q *DocumentQuery) WhereExists(fieldName string) *DocumentQuery {
	if q.err != nil {
		return q
	}
	q.err = q.whereExists(fieldName)
	return q
}

//TBD expr IDocumentQuery<T> IFilterDocumentQueryBase<T, IDocumentQuery<T>>.WhereRegex<TValue>(Expression<Func<T, TValue>> propertySelector, string pattern)

func (q *DocumentQuery) WhereRegex(fieldName string, pattern string) *DocumentQuery {
	if q.err != nil {
		return q
	}
	q.err = q.whereRegex(fieldName, pattern)
	return q
}

func (q *DocumentQuery) AndAlso() *DocumentQuery {
	if q.err != nil {
		return q
	}
	q.err = q.andAlso()
	return q
}

func (q *DocumentQuery) OrElse() *DocumentQuery {
	if q.err != nil {
		return q
	}
	q.err = q.orElse()
	return q
}

func (q *DocumentQuery) Boost(boost float64) *DocumentQuery {
	if q.err != nil {
		return q
	}
	q.err = q.boost(boost)
	return q
}

func (q *DocumentQuery) Fuzzy(fuzzy float64) *DocumentQuery {
	if q.err != nil {
		return q
	}
	q.err = q.fuzzy(fuzzy)
	return q
}

func (q *DocumentQuery) Proximity(proximity int) *DocumentQuery {
	if q.err != nil {
		return q
	}
	q.err = q.proximity(proximity)
	return q
}

func (q *DocumentQuery) RandomOrdering() *DocumentQuery {
	if q.err != nil {
		return q
	}
	q.err = q.randomOrdering()
	return q
}

func (q *DocumentQuery) RandomOrderingWithSeed(seed string) *DocumentQuery {
	if q.err != nil {
		return q
	}
	q.err = q.randomOrderingWithSeed(seed)
	return q
}

//TBD 4.1  IDocumentQuery<T> customSortUsing(string typeName, bool descending)

// GroupBy makes a query grouped by fields
func (q *DocumentQuery) GroupBy(fieldName string, fieldNames ...string) *GroupByDocumentQuery {
	res := newGroupByDocumentQuery(q)
	if q.err == nil {
		q.err = q.groupBy(fieldName, fieldNames...)
	}

	res.err = q.err
	return res
}

// GroupBy makes a query grouped by fields and also allows specifying method
// of grouping for each field
func (q *DocumentQuery) GroupByFieldWithMethod(field *GroupBy, fields ...*GroupBy) *GroupByDocumentQuery {
	res := newGroupByDocumentQuery(q)
	if q.err == nil {
		q.err = q.q2.GroupByFieldWithMethod(ravendb.NewGroupByArray("lines[].product"))(field, fields...)
	}
	res.err = q.err
	return res
}

// OrderBy makes a query ordered by a given field
func (q *DocumentQuery) OrderBy(field string) *DocumentQuery {
	return q.OrderByWithOrdering(field, OrderingTypeString)
}

func (q *DocumentQuery) OrderByWithOrdering(field string, ordering OrderingType) *DocumentQuery {
	if q.err != nil {
		return q
	}
	q.err = q.orderByWithOrdering(field, ordering)
	return q
}

//TBD expr  IDocumentQuery<T> OrderBy<TValue>(params Expression<Func<T, TValue>>[] propertySelectors)

func (q *DocumentQuery) OrderByDescending(field string) *DocumentQuery {
	return q.OrderByDescendingWithOrdering(field, OrderingTypeString)
}

func (q *DocumentQuery) OrderByDescendingWithOrdering(field string, ordering OrderingType) *DocumentQuery {
	if q.err != nil {
		return q
	}
	q.err = q.orderByDescendingWithOrdering(field, ordering)
	return q
}

//TBD expr  IDocumentQuery<T> OrderByDescending<TValue>(params Expression<Func<T, TValue>>[] propertySelectors)

func (q *DocumentQuery) AddBeforeQueryExecutedListener(action func(*IndexQuery)) int {
	return q.addBeforeQueryExecutedListener(action)
}

func (q *DocumentQuery) RemoveBeforeQueryExecutedListener(idx int) *DocumentQuery {
	q.removeBeforeQueryExecutedListener(idx)
	return q
}

// Note: had to move it down to abstractDocumentQuery
func (q *abstractDocumentQuery) createDocumentQueryInternal(resultClass reflect.Type, queryData *queryData) (*DocumentQuery, error) {

	var newFieldsToFetch *fieldsToFetchToken

	if queryData != nil && len(queryData.fields) > 0 {
		fields := queryData.fields

		identityProperty := q.conventions.GetIdentityProperty(resultClass)

		if identityProperty != "" {
			// make a copy, just in case, because we might modify it
			fields = append([]string{}, fields...)

			for idx, p := range fields {
				if p == identityProperty {
					fields[idx] = IndexingFieldNameDocumentID
				}
			}
		}

		sourceAliasReference := getSourceAliasIfExists(resultClass, queryData, fields)
		newFieldsToFetch = createFieldsToFetchToken(fields, queryData.projections, queryData.isCustomFunction, sourceAliasReference)
	}

	if newFieldsToFetch != nil {
		q.updateFieldsToFetchToken(newFieldsToFetch)
	}

	var declareToken *declareToken
	var loadTokens []*loadToken
	var fromAlias string
	if queryData != nil {
		declareToken = queryData.declareToken
		loadTokens = queryData.loadTokens
		fromAlias = queryData.fromAlias
	}

	opts := &DocumentQueryOptions{
		Type:           resultClass,
		session:        q.theSession,
		IndexName:      q.indexName,
		CollectionName: q.collectionName,
		isGroupBy:      q.isGroupBy,
		declareToken:   declareToken,
		loadTokens:     loadTokens,
		fromAlias:      fromAlias,
	}
	query := newDocumentQuery(opts)
	if query.err != nil {
		return nil, query.err
	}

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

	return query, nil
}

func (q *DocumentQuery) AggregateByFacet(facet FacetBase) *AggregationDocumentQuery {
	res := newAggregationDocumentQuery(q)
	if q.err != nil {
		return res
	}

	res.err = q.aggregateBy(facet)
	return res
}

func (q *DocumentQuery) AggregateByFacets(facets ...*Facet) *AggregationDocumentQuery {
	res := newAggregationDocumentQuery(q)
	if q.err != nil {
		return res
	}

	for _, facet := range facets {
		if res.err = q.aggregateBy(facet); res.err != nil {
			return res
		}
	}
	return res
}

func (q *DocumentQuery) AggregateUsing(facetSetupDocumentID string) *AggregationDocumentQuery {
	res := newAggregationDocumentQuery(q)
	if q.err != nil {
		return res
	}
	res.err = q.aggregateUsing(facetSetupDocumentID)
	return res
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

func (q *DocumentQuery) Spatial3(fieldName string, clause func(*SpatialCriteriaFactory) SpatialCriteria) *DocumentQuery {
	if q.err != nil {
		return q
	}
	criteria := clause(spatialCriteriaFactoryInstance)
	q.err = q.spatial3(fieldName, criteria)
	return q
}

func (q *DocumentQuery) Spatial2(field DynamicSpatialField, clause func(*SpatialCriteriaFactory) SpatialCriteria) *DocumentQuery {
	if q.err != nil {
		return q
	}
	criteria := clause(spatialCriteriaFactoryInstance)
	q.err = q.spatial2(field, criteria)
	return q
}

//TBD expr  IDocumentQuery<T> Spatial(Func<SpatialDynamicFieldFactory<T>, DynamicSpatialField> field, Func<SpatialCriteriaFactory, SpatialCriteria> clause)
//TBD expr IDocumentQuery<T> IDocumentQueryBase<T, IDocumentQuery<T>>.WithinRadiusOf<TValue>(Expression<Func<T, TValue>> propertySelector, float64 radius, float64 latitude, float64 longitude, SpatialUnits? radiusUnits, float64 distanceErrorPct)

func (q *DocumentQuery) WithinRadiusOf(fieldName string, radius float64, latitude float64, longitude float64) *DocumentQuery {
	if q.err != nil {
		return q
	}
	q.err = q.withinRadiusOf(fieldName, radius, latitude, longitude, "", IndexingSpatialDefaultDistnaceErrorPct)
	return q
}

func (q *DocumentQuery) WithinRadiusOfWithUnits(fieldName string, radius float64, latitude float64, longitude float64, radiusUnits SpatialUnits) *DocumentQuery {
	if q.err != nil {
		return q
	}
	q.err = q.withinRadiusOf(fieldName, radius, latitude, longitude, radiusUnits, IndexingSpatialDefaultDistnaceErrorPct)
	return q
}

func (q *DocumentQuery) WithinRadiusOfWithUnitsAndError(fieldName string, radius float64, latitude float64, longitude float64, radiusUnits SpatialUnits, distanceErrorPct float64) *DocumentQuery {
	if q.err != nil {
		return q
	}
	q.err = q.withinRadiusOf(fieldName, radius, latitude, longitude, radiusUnits, distanceErrorPct)
	return q
}

//TBD expr IDocumentQuery<T> IDocumentQueryBase<T, IDocumentQuery<T>>.RelatesToShape<TValue>(Expression<Func<T, TValue>> propertySelector, string shapeWkt, SpatialRelation relation, float64 distanceErrorPct)

func (q *DocumentQuery) RelatesToShape(fieldName string, shapeWkt string, relation SpatialRelation) *DocumentQuery {
	return q.RelatesToShapeWithError(fieldName, shapeWkt, relation, IndexingSpatialDefaultDistnaceErrorPct)
}

func (q *DocumentQuery) RelatesToShapeWithError(fieldName string, shapeWkt string, relation SpatialRelation, distanceErrorPct float64) *DocumentQuery {
	if q.err != nil {
		return q
	}
	q.err = q.spatial(fieldName, shapeWkt, relation, distanceErrorPct)
	return q
}

func (q *DocumentQuery) OrderByDistance(field DynamicSpatialField, latitude float64, longitude float64) *DocumentQuery {
	if q.err != nil {
		return q
	}
	q.err = q.orderByDistance(field, latitude, longitude)
	return q
}

//TBD expr IDocumentQuery<T> IDocumentQueryBase<T, IDocumentQuery<T>>.OrderByDistance(Func<DynamicSpatialFieldFactory<T>, DynamicSpatialField> field, float64 latitude, float64 longitude)

func (q *DocumentQuery) OrderByDistance2(field DynamicSpatialField, shapeWkt string) *DocumentQuery {
	if q.err != nil {
		return q
	}
	q.err = q.orderByDistance2(field, shapeWkt)
	return q
}

//TBD expr IDocumentQuery<T> IDocumentQueryBase<T, IDocumentQuery<T>>.OrderByDistance(Func<DynamicSpatialFieldFactory<T>, DynamicSpatialField> field, string shapeWkt)

//TBD expr IDocumentQuery<T> IDocumentQueryBase<T, IDocumentQuery<T>>.OrderByDistance<TValue>(Expression<Func<T, TValue>> propertySelector, float64 latitude, float64 longitude)

func (q *DocumentQuery) OrderByDistanceLatLong(fieldName string, latitude float64, longitude float64) *DocumentQuery {
	if q.err != nil {
		return q
	}
	q.err = q.orderByDistanceLatLong(fieldName, latitude, longitude)
	return q
}

//TBD expr IDocumentQuery<T> IDocumentQueryBase<T, IDocumentQuery<T>>.OrderByDistance<TValue>(Expression<Func<T, TValue>> propertySelector, string shapeWkt)

func (q *DocumentQuery) OrderByDistance3(fieldName string, shapeWkt string) *DocumentQuery {
	if q.err != nil {
		return q
	}
	q.err = q.orderByDistance3(fieldName, shapeWkt)
	return q
}

func (q *DocumentQuery) OrderByDistanceDescending(field DynamicSpatialField, latitude float64, longitude float64) *DocumentQuery {
	if q.err != nil {
		return q
	}
	q.err = q.orderByDistanceDescending(field, latitude, longitude)
	return q
}

//TBD expr IDocumentQuery<T> IDocumentQueryBase<T, IDocumentQuery<T>>.OrderByDistanceDescending(Func<DynamicSpatialFieldFactory<T>, DynamicSpatialField> field, float64 latitude, float64 longitude)

func (q *DocumentQuery) OrderByDistanceDescending2(field DynamicSpatialField, shapeWkt string) *DocumentQuery {
	if q.err != nil {
		return q
	}
	q.err = q.orderByDistanceDescending2(field, shapeWkt)
	return q
}

//TBD expr IDocumentQuery<T> IDocumentQueryBase<T, IDocumentQuery<T>>.OrderByDistanceDescending(Func<DynamicSpatialFieldFactory<T>, DynamicSpatialField> field, string shapeWkt)

//TBD expr IDocumentQuery<T> IDocumentQueryBase<T, IDocumentQuery<T>>.OrderByDistanceDescending<TValue>(Expression<Func<T, TValue>> propertySelector, float64 latitude, float64 longitude)

func (q *DocumentQuery) OrderByDistanceDescendingLatLong(fieldName string, latitude float64, longitude float64) *DocumentQuery {
	if q.err != nil {
		return q
	}
	q.err = q.orderByDistanceDescendingLatLong(fieldName, latitude, longitude)
	return q
}

//TBD expr IDocumentQuery<T> IDocumentQueryBase<T, IDocumentQuery<T>>.OrderByDistanceDescending<TValue>(Expression<Func<T, TValue>> propertySelector, string shapeWkt)

func (q *DocumentQuery) OrderByDistanceDescending3(fieldName string, shapeWkt string) *DocumentQuery {
	if q.err != nil {
		return q
	}
	q.err = q.orderByDistanceDescending3(fieldName, shapeWkt)
	return q
}

func (q *DocumentQuery) MoreLikeThis(moreLikeThis MoreLikeThisBase) *DocumentQuery {
	if q.err != nil {
		return q
	}
	mlt, err := q.moreLikeThis()
	if err != nil {
		q.err = err
		return q
	}
	defer mlt.Close()

	mlt.withOptions(moreLikeThis.GetOptions())

	if mltud, ok := moreLikeThis.(*MoreLikeThisUsingDocument); ok {
		mlt.withDocument(mltud.documentJSON)
	}

	return q
}

func (q *DocumentQuery) MoreLikeThisWithBuilder(builder func(IMoreLikeThisBuilderForDocumentQuery)) *DocumentQuery {
	if q.err != nil {
		return q
	}
	f := NewMoreLikeThisBuilder()
	builder(f)

	moreLikeThis, err := q.moreLikeThis()
	if err != nil {
		q.err = err
		return q
	}

	moreLikeThis.withOptions(f.GetMoreLikeThis().GetOptions())

	tmp := f.GetMoreLikeThis()
	if mlt, ok := tmp.(*MoreLikeThisUsingDocument); ok {
		moreLikeThis.withDocument(mlt.documentJSON)
	} else if mlt, ok := tmp.(*MoreLikeThisUsingDocumentForDocumentQuery); ok {
		mlt.GetForDocumentQuery()(q)
	}
	moreLikeThis.Close()

	return q
}

func (q *DocumentQuery) SuggestUsing(suggestion SuggestionBase) *SuggestionDocumentQuery {
	res := newSuggestionDocumentQuery(q)
	if q.err != nil {
		res.err = q.err
		return res
	}

	if q.err = q.suggestUsing(suggestion); q.err != nil {
		res.err = q.err
		return res
	}
	return res
}
