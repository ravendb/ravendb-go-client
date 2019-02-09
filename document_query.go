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

func newDocumentQuery(opts *DocumentQueryOptions) (*DocumentQuery, error) {
	var err error
	if opts.session == nil {
		return nil, newIllegalArgumentError("session must be provided")
	}
	opts.IndexName, opts.CollectionName, err = processQueryParameters(opts.Type, opts.IndexName, opts.CollectionName, opts.conventions)
	if err != nil {
		return nil, err
	}
	aq, err := newAbstractDocumentQuery(opts)
	if err != nil {
		return nil, err
	}
	return &DocumentQuery{
		abstractDocumentQuery: aq,
	}, nil
}

// SelectFields limits the returned values to one or more fields of the queried type.
func (q *DocumentQuery) SelectFields(projectionType reflect.Type, fieldsIn ...string) (*DocumentQuery, error) {
	// TODO: add SelectFieldsWithProjection(projectionType reflect.Type, fields []string, projections []string)
	var fields []string
	if len(fieldsIn) == 0 {
		fields = FieldsFor(projectionType)
		if len(fields) == 0 {
			return nil, newIllegalArgumentError("type %T has no exported fields to select", projectionType)
		}
	} else {
		fields = fieldsIn
	}

	queryData := &queryData{
		fields:      fields,
		projections: fields,
	}
	return q.createDocumentQueryInternal(projectionType, queryData)
}

// Distinct marks query as distinct
func (q *DocumentQuery) Distinct() *DocumentQuery {
	q.distinct()
	return q
}

// OrderByScore orders results of the query by score
func (q *DocumentQuery) OrderByScore() *DocumentQuery {
	q.orderByScore()
	return q
}

// OrderByScoreDescending orders results of the query by score
// in descending order
func (q *DocumentQuery) OrderByScoreDescending() *DocumentQuery {
	q.orderByScoreDescending()
	return q
}

//TBD 4.1  IDocumentQuery<T> explainScores() {

func (q *DocumentQuery) WaitForNonStaleResults(waitTimeout time.Duration) *DocumentQuery {
	q.waitForNonStaleResults(waitTimeout)
	return q
}

func (q *DocumentQuery) AddParameter(name string, value interface{}) *DocumentQuery {
	q.addParameter(name, value)
	return q
}

func (q *DocumentQuery) AddOrder(fieldName string, descending bool) *DocumentQuery {
	return q.AddOrderWithOrdering(fieldName, descending, OrderingTypeString)
}

func (q *DocumentQuery) AddOrderWithOrdering(fieldName string, descending bool, ordering OrderingType) *DocumentQuery {
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
	q.openSubclause()
	return q
}

func (q *DocumentQuery) CloseSubclause() *DocumentQuery {
	q.closeSubclause()
	return q
}

func (q *DocumentQuery) Search(fieldName string, searchTerms string) *DocumentQuery {
	q.search(fieldName, searchTerms)
	return q
}

func (q *DocumentQuery) SearchWithOperator(fieldName string, searchTerms string, operator SearchOperator) *DocumentQuery {
	q.searchWithOperator(fieldName, searchTerms, operator)
	return q
}

//TBD expr  IDocumentQuery<T> Search<TValue>(Expression<Func<T, TValue>> propertySelector, string searchTerms, SearchOperator @operator)

func (q *DocumentQuery) Intersect() *DocumentQuery {
	q.intersect()
	return q
}

func (q *DocumentQuery) ContainsAny(fieldName string, values []interface{}) *DocumentQuery {
	q.containsAny(fieldName, values)
	return q
}

//TBD expr  IDocumentQuery<T> ContainsAny<TValue>(Expression<Func<T, TValue>> propertySelector, IEnumerable<TValue> values)

func (q *DocumentQuery) ContainsAll(fieldName string, values []interface{}) *DocumentQuery {
	q.containsAll(fieldName, values)
	return q
}

//TBD expr  IDocumentQuery<T> ContainsAll<TValue>(Expression<Func<T, TValue>> propertySelector, IEnumerable<TValue> values)

func (q *DocumentQuery) Statistics(stats **QueryStatistics) *DocumentQuery {
	q.statistics(stats)
	return q
}

func (q *DocumentQuery) UsingDefaultOperator(queryOperator QueryOperator) *DocumentQuery {
	q.usingDefaultOperator(queryOperator)
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
	q.take(&count)
	return q
}

func (q *DocumentQuery) Skip(count int) *DocumentQuery {
	q.skip(count)
	return q
}

func (q *DocumentQuery) WhereLucene(fieldName string, whereClause string) *DocumentQuery {
	q.whereLucene(fieldName, whereClause)
	return q
}

func (q *DocumentQuery) WhereEquals(fieldName string, value interface{}) *DocumentQuery {
	q.whereEquals(fieldName, value)
	return q
}

// Exact marks previous Where statement (e.g. WhereEquals or WhereLucene) as exact
func (q *DocumentQuery) Exact() *DocumentQuery {
	q.markLastTokenExact()
	return q
}

func (q *DocumentQuery) WhereEqualsWithMethodCall(fieldName string, method MethodCall) *DocumentQuery {
	q.whereEqualsWithMethodCall(fieldName, method)
	return q
}

//TBD expr IDocumentQuery<T> IDocumentQueryBase<T, IDocumentQuery<T>>.WhereEquals<TValue>(Expression<Func<T, TValue>> propertySelector, TValue value, bool exact)
//TBD expr IDocumentQuery<T> IFilterDocumentQueryBase<T, IDocumentQuery<T>>.WhereEquals<TValue>(Expression<Func<T, TValue>> propertySelector, MethodCall value, bool exact)

func (q *DocumentQuery) WhereEqualsWithParams(whereParams *whereParams) *DocumentQuery {
	q.whereEqualsWithParams(whereParams)
	return q
}

func (q *DocumentQuery) WhereNotEquals(fieldName string, value interface{}) *DocumentQuery {
	q.whereNotEquals(fieldName, value)
	return q
}

func (q *DocumentQuery) WhereNotEqualsWithMethod(fieldName string, method MethodCall) *DocumentQuery {
	q.whereNotEqualsWithMethod(fieldName, method)
	return q
}

//TBD expr IDocumentQuery<T> IDocumentQueryBase<T, IDocumentQuery<T>>.WhereNotEquals<TValue>(Expression<Func<T, TValue>> propertySelector, TValue value, bool exact)
//TBD expr IDocumentQuery<T> IFilterDocumentQueryBase<T, IDocumentQuery<T>>.WhereNotEquals<TValue>(Expression<Func<T, TValue>> propertySelector, MethodCall value, bool exact)

func (q *DocumentQuery) WhereNotEqualsWithParams(whereParams *whereParams) *DocumentQuery {
	q.whereNotEqualsWithParams(whereParams)
	return q
}

func (q *DocumentQuery) WhereIn(fieldName string, values []interface{}) *DocumentQuery {
	q.whereIn(fieldName, values)
	return q
}

//TBD expr  IDocumentQuery<T> WhereIn<TValue>(Expression<Func<T, TValue>> propertySelector, IEnumerable<TValue> values, bool exact = false)

func (q *DocumentQuery) WhereStartsWith(fieldName string, value interface{}) *DocumentQuery {
	q.whereStartsWith(fieldName, value)
	return q
}

func (q *DocumentQuery) WhereEndsWith(fieldName string, value interface{}) *DocumentQuery {
	q.whereEndsWith(fieldName, value)
	return q
}

//TBD expr  IDocumentQuery<T> WhereEndsWith<TValue>(Expression<Func<T, TValue>> propertySelector, TValue value)

func (q *DocumentQuery) WhereBetween(fieldName string, start interface{}, end interface{}) *DocumentQuery {
	q.whereBetween(fieldName, start, end)
	return q
}

//TBD expr  IDocumentQuery<T> WhereBetween<TValue>(Expression<Func<T, TValue>> propertySelector, TValue start, TValue end, bool exact = false)

func (q *DocumentQuery) WhereGreaterThan(fieldName string, value interface{}) *DocumentQuery {
	q.whereGreaterThan(fieldName, value)
	return q
}

func (q *DocumentQuery) WhereGreaterThanOrEqual(fieldName string, value interface{}) *DocumentQuery {
	q.whereGreaterThanOrEqual(fieldName, value)
	return q
}

//TBD expr  IDocumentQuery<T> WhereGreaterThan<TValue>(Expression<Func<T, TValue>> propertySelector, TValue value, bool exact = false)
//TBD expr  IDocumentQuery<T> WhereGreaterThanOrEqual<TValue>(Expression<Func<T, TValue>> propertySelector, TValue value, bool exact = false)

func (q *DocumentQuery) WhereLessThan(fieldName string, value interface{}) *DocumentQuery {
	q.whereLessThan(fieldName, value)
	return q
}

//TBD expr  IDocumentQuery<T> WhereLessThanOrEqual<TValue>(Expression<Func<T, TValue>> propertySelector, TValue value, bool exact = false)

func (q *DocumentQuery) WhereLessThanOrEqual(fieldName string, value interface{}) *DocumentQuery {
	q.whereLessThanOrEqual(fieldName, value)
	return q
}

//TBD expr  IDocumentQuery<T> WhereLessThanOrEqual<TValue>(Expression<Func<T, TValue>> propertySelector, TValue value, bool exact = false)
//TBD expr  IDocumentQuery<T> WhereExists<TValue>(Expression<Func<T, TValue>> propertySelector)

func (q *DocumentQuery) WhereExists(fieldName string) *DocumentQuery {
	q.whereExists(fieldName)
	return q
}

//TBD expr IDocumentQuery<T> IFilterDocumentQueryBase<T, IDocumentQuery<T>>.WhereRegex<TValue>(Expression<Func<T, TValue>> propertySelector, string pattern)

func (q *DocumentQuery) WhereRegex(fieldName string, pattern string) *DocumentQuery {
	q.whereRegex(fieldName, pattern)
	return q
}

func (q *DocumentQuery) AndAlso() *DocumentQuery {
	q.andAlso()
	return q
}

func (q *DocumentQuery) OrElse() *DocumentQuery {
	q.orElse()
	return q
}

func (q *DocumentQuery) Boost(boost float64) *DocumentQuery {
	q.boost(boost)
	return q
}

func (q *DocumentQuery) Fuzzy(fuzzy float64) *DocumentQuery {
	q.fuzzy(fuzzy)
	return q
}

func (q *DocumentQuery) Proximity(proximity int) *DocumentQuery {
	q.proximity(proximity)
	return q
}

func (q *DocumentQuery) RandomOrdering() *DocumentQuery {
	q.randomOrdering()
	return q
}

func (q *DocumentQuery) RandomOrderingWithSeed(seed string) *DocumentQuery {
	q.randomOrderingWithSeed(seed)
	return q
}

//TBD 4.1  IDocumentQuery<T> customSortUsing(string typeName, bool descending)

// GroupBy makes a query grouped by fields
func (q *DocumentQuery) GroupBy(fieldName string, fieldNames ...string) *IGroupByDocumentQuery {
	q.groupBy(fieldName, fieldNames...)

	return NewGroupByDocumentQuery(q)
}

func (q *DocumentQuery) GroupBy2(field *GroupBy, fields ...*GroupBy) *IGroupByDocumentQuery {
	q.groupBy2(field, fields...)

	return NewGroupByDocumentQuery(q)
}

// OrderBy makes a query ordered by a given field
func (q *DocumentQuery) OrderBy(field string) *DocumentQuery {
	return q.OrderByWithOrdering(field, OrderingTypeString)
}

func (q *DocumentQuery) OrderByWithOrdering(field string, ordering OrderingType) *DocumentQuery {
	q.orderByWithOrdering(field, ordering)
	return q
}

//TBD expr  IDocumentQuery<T> OrderBy<TValue>(params Expression<Func<T, TValue>>[] propertySelectors)

func (q *DocumentQuery) OrderByDescending(field string) *DocumentQuery {
	return q.OrderByDescendingWithOrdering(field, OrderingTypeString)
}

func (q *DocumentQuery) OrderByDescendingWithOrdering(field string, ordering OrderingType) *DocumentQuery {
	q.orderByDescendingWithOrdering(field, ordering)
	return q
}

//TBD expr  IDocumentQuery<T> OrderByDescending<TValue>(params Expression<Func<T, TValue>>[] propertySelectors)

/*
 IDocumentQuery<T> AddBeforeQueryExecutedListener(Consumer<IndexQuery> action) {
	addBeforeQueryExecutedListener(action);
	return this;
}


 IDocumentQuery<T> RemoveBeforeQueryExecutedListener(Consumer<IndexQuery> action) {
	removeBeforeQueryExecutedListener(action);
	return this;
}
*/

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
	query, err := newDocumentQuery(opts)
	if err != nil {
		return nil, err
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

func (q *DocumentQuery) AggregateByFacet(facet FacetBase) (*AggregationDocumentQuery, error) {
	if err := q.aggregateBy(facet); err != nil {
		return nil, err
	}

	return NewAggregationDocumentQuery(q), nil
}

func (q *DocumentQuery) AggregateByFacets(facets ...*Facet) (*AggregationDocumentQuery, error) {
	for _, facet := range facets {
		if err := q.aggregateBy(facet); err != nil {
			return nil, err
		}
	}

	return NewAggregationDocumentQuery(q), nil
}

func (q *DocumentQuery) AggregateUsing(facetSetupDocumentID string) *AggregationDocumentQuery {
	q.aggregateUsing(facetSetupDocumentID)

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

func (q *DocumentQuery) Spatial3(fieldName string, clause func(*SpatialCriteriaFactory) SpatialCriteria) *DocumentQuery {
	criteria := clause(spatialCriteriaFactoryInstance)
	q.spatial3(fieldName, criteria)
	return q
}

func (q *DocumentQuery) Spatial2(field DynamicSpatialField, clause func(*SpatialCriteriaFactory) SpatialCriteria) *DocumentQuery {
	criteria := clause(spatialCriteriaFactoryInstance)
	q.spatial2(field, criteria)
	return q
}

//TBD expr  IDocumentQuery<T> Spatial(Func<SpatialDynamicFieldFactory<T>, DynamicSpatialField> field, Func<SpatialCriteriaFactory, SpatialCriteria> clause)
//TBD expr IDocumentQuery<T> IDocumentQueryBase<T, IDocumentQuery<T>>.WithinRadiusOf<TValue>(Expression<Func<T, TValue>> propertySelector, float64 radius, float64 latitude, float64 longitude, SpatialUnits? radiusUnits, float64 distanceErrorPct)

func (q *DocumentQuery) WithinRadiusOf(fieldName string, radius float64, latitude float64, longitude float64) *DocumentQuery {
	q.withinRadiusOf(fieldName, radius, latitude, longitude, "", IndexingSpatialDefaultDistnaceErrorPct)
	return q
}

func (q *DocumentQuery) WithinRadiusOfWithUnits(fieldName string, radius float64, latitude float64, longitude float64, radiusUnits SpatialUnits) *DocumentQuery {
	q.withinRadiusOf(fieldName, radius, latitude, longitude, radiusUnits, IndexingSpatialDefaultDistnaceErrorPct)
	return q
}

func (q *DocumentQuery) WithinRadiusOfWithUnitsAndError(fieldName string, radius float64, latitude float64, longitude float64, radiusUnits SpatialUnits, distanceErrorPct float64) *DocumentQuery {
	q.withinRadiusOf(fieldName, radius, latitude, longitude, radiusUnits, distanceErrorPct)
	return q
}

//TBD expr IDocumentQuery<T> IDocumentQueryBase<T, IDocumentQuery<T>>.RelatesToShape<TValue>(Expression<Func<T, TValue>> propertySelector, string shapeWkt, SpatialRelation relation, float64 distanceErrorPct)

func (q *DocumentQuery) RelatesToShape(fieldName string, shapeWkt string, relation SpatialRelation) *DocumentQuery {
	return q.RelatesToShapeWithError(fieldName, shapeWkt, relation, IndexingSpatialDefaultDistnaceErrorPct)
}

func (q *DocumentQuery) RelatesToShapeWithError(fieldName string, shapeWkt string, relation SpatialRelation, distanceErrorPct float64) *DocumentQuery {
	q.spatial(fieldName, shapeWkt, relation, distanceErrorPct)
	return q
}

func (q *DocumentQuery) OrderByDistance(field DynamicSpatialField, latitude float64, longitude float64) *DocumentQuery {
	q.orderByDistance(field, latitude, longitude)
	return q
}

//TBD expr IDocumentQuery<T> IDocumentQueryBase<T, IDocumentQuery<T>>.OrderByDistance(Func<DynamicSpatialFieldFactory<T>, DynamicSpatialField> field, float64 latitude, float64 longitude)

func (q *DocumentQuery) OrderByDistance2(field DynamicSpatialField, shapeWkt string) *DocumentQuery {
	q.orderByDistance2(field, shapeWkt)
	return q
}

//TBD expr IDocumentQuery<T> IDocumentQueryBase<T, IDocumentQuery<T>>.OrderByDistance(Func<DynamicSpatialFieldFactory<T>, DynamicSpatialField> field, string shapeWkt)

//TBD expr IDocumentQuery<T> IDocumentQueryBase<T, IDocumentQuery<T>>.OrderByDistance<TValue>(Expression<Func<T, TValue>> propertySelector, float64 latitude, float64 longitude)

func (q *DocumentQuery) OrderByDistanceLatLong(fieldName string, latitude float64, longitude float64) *DocumentQuery {
	q.orderByDistanceLatLong(fieldName, latitude, longitude)
	return q
}

//TBD expr IDocumentQuery<T> IDocumentQueryBase<T, IDocumentQuery<T>>.OrderByDistance<TValue>(Expression<Func<T, TValue>> propertySelector, string shapeWkt)

func (q *DocumentQuery) OrderByDistance3(fieldName string, shapeWkt string) *DocumentQuery {
	q.orderByDistance3(fieldName, shapeWkt)
	return q
}

func (q *DocumentQuery) OrderByDistanceDescending(field DynamicSpatialField, latitude float64, longitude float64) *DocumentQuery {
	q.orderByDistanceDescending(field, latitude, longitude)
	return q
}

//TBD expr IDocumentQuery<T> IDocumentQueryBase<T, IDocumentQuery<T>>.OrderByDistanceDescending(Func<DynamicSpatialFieldFactory<T>, DynamicSpatialField> field, float64 latitude, float64 longitude)

func (q *DocumentQuery) OrderByDistanceDescending2(field DynamicSpatialField, shapeWkt string) *DocumentQuery {
	q.orderByDistanceDescending2(field, shapeWkt)
	return q
}

//TBD expr IDocumentQuery<T> IDocumentQueryBase<T, IDocumentQuery<T>>.OrderByDistanceDescending(Func<DynamicSpatialFieldFactory<T>, DynamicSpatialField> field, string shapeWkt)

//TBD expr IDocumentQuery<T> IDocumentQueryBase<T, IDocumentQuery<T>>.OrderByDistanceDescending<TValue>(Expression<Func<T, TValue>> propertySelector, float64 latitude, float64 longitude)

func (q *DocumentQuery) OrderByDistanceDescendingLatLong(fieldName string, latitude float64, longitude float64) *DocumentQuery {
	q.orderByDistanceDescendingLatLong(fieldName, latitude, longitude)
	return q
}

//TBD expr IDocumentQuery<T> IDocumentQueryBase<T, IDocumentQuery<T>>.OrderByDistanceDescending<TValue>(Expression<Func<T, TValue>> propertySelector, string shapeWkt)

func (q *DocumentQuery) OrderByDistanceDescending3(fieldName string, shapeWkt string) *DocumentQuery {
	q.orderByDistanceDescending3(fieldName, shapeWkt)
	return q
}

func (q *DocumentQuery) MoreLikeThis(moreLikeThis MoreLikeThisBase) (*DocumentQuery, error) {
	mlt, err := q.moreLikeThis()
	if err != nil {
		return nil, err
	}
	defer mlt.Close()

	mlt.withOptions(moreLikeThis.GetOptions())

	if mltud, ok := moreLikeThis.(*MoreLikeThisUsingDocument); ok {
		mlt.withDocument(mltud.documentJSON)

	}

	return q, nil
}

func (q *DocumentQuery) MoreLikeThisWithBuilder(builder func(IMoreLikeThisBuilderForDocumentQuery)) (*DocumentQuery, error) {
	f := NewMoreLikeThisBuilder()
	builder(f)

	moreLikeThis, err := q.moreLikeThis()
	if err != nil {
		return nil, err
	}

	moreLikeThis.withOptions(f.GetMoreLikeThis().GetOptions())

	tmp := f.GetMoreLikeThis()
	if mlt, ok := tmp.(*MoreLikeThisUsingDocument); ok {
		moreLikeThis.withDocument(mlt.documentJSON)
	} else if mlt, ok := tmp.(*MoreLikeThisUsingDocumentForDocumentQuery); ok {
		mlt.GetForDocumentQuery()(q)
	}
	moreLikeThis.Close()

	return q, nil
}

func (q *DocumentQuery) SuggestUsing(suggestion SuggestionBase) (*SuggestionDocumentQuery, error) {
	if err := q.suggestUsing(suggestion); err != nil {
		return nil, err
	}
	return NewSuggestionDocumentQuery(q), nil
}
