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

func (q *DocumentQuery) selectFields(projectionClass reflect.Type, fields ...string) *DocumentQuery {
	if len(fields) > 0 {
		queryData := NewQueryData(fields, fields)
		return q.selectFieldsWithQueryData(projectionClass, queryData)
	}

	panic("NYI")
	// TODO: implement me
	/*
		PropertyDescriptor[] propertyDescriptors = Introspector.getBeanInfo(projectionClass).getPropertyDescriptors();

		string[] projections = Arrays.stream(propertyDescriptors)
				.map(x -> x.getName())
				.toArray(string[]::new);

		string[] fields = Arrays.stream(propertyDescriptors)
				.map(x -> x.getName())
				.toArray(string[]::new);

		return selectFields(projectionClass, new QueryData(fields, projections));
	*/
	return nil
}

func (q *DocumentQuery) selectFieldsWithQueryData(projectionClass reflect.Type, queryData *QueryData) *DocumentQuery {
	return q.createDocumentQueryInternalWithQueryData(projectionClass, queryData)
}

func (q *DocumentQuery) distinct() *DocumentQuery {
	q._distinct()
	return q
}

func (q *DocumentQuery) orderByScore() *DocumentQuery {
	q._orderByScore()
	return q
}

func (q *DocumentQuery) orderByScoreDescending() *DocumentQuery {
	q._orderByScoreDescending()
	return q
}

//TBD 4.1  IDocumentQuery<T> explainScores() {

func (q *DocumentQuery) waitForNonStaleResults(waitTimeout time.Duration) *DocumentQuery {
	q._waitForNonStaleResults(waitTimeout)
	return q
}

/*
 IDocumentQuery<T> addParameter(string name, Object value) {
	_addParameter(name, value);
	return this;
}


 IDocumentQuery<T> addOrder(string fieldName, bool descending) {
	return addOrder(fieldName, descending, OrderingType.STRING);
}


 IDocumentQuery<T> addOrder(string fieldName, bool descending, OrderingType ordering) {
	if (descending) {
		orderByDescending(fieldName, ordering);
	} else {
		orderBy(fieldName, ordering);
	}
	return this;
}

//TBD expr  IDocumentQuery<T> AddOrder<TValue>(Expression<Func<T, TValue>> propertySelector, bool descending, OrderingType ordering)



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

 IDocumentQuery<T> openSubclause() {
	_openSubclause();
	return this;
}


 IDocumentQuery<T> closeSubclause() {
	_closeSubclause();
	return this;
}


 IDocumentQuery<T> search(string fieldName, string searchTerms) {
	_search(fieldName, searchTerms);
	return this;
}


 IDocumentQuery<T> search(string fieldName, string searchTerms, SearchOperator operator) {
	_search(fieldName, searchTerms, operator);
	return this;
}

//TBD expr  IDocumentQuery<T> Search<TValue>(Expression<Func<T, TValue>> propertySelector, string searchTerms, SearchOperator @operator)


 IDocumentQuery<T> intersect() {
	_intersect();
	return this;
}
*/

func (q *DocumentQuery) containsAny(fieldName string, values []Object) *DocumentQuery {
	q._containsAny(fieldName, values)
	return q
}

//TBD expr  IDocumentQuery<T> ContainsAny<TValue>(Expression<Func<T, TValue>> propertySelector, IEnumerable<TValue> values)

func (q *DocumentQuery) containsAll(fieldName string, values []Object) *DocumentQuery {
	q._containsAll(fieldName, values)
	return q
}

//TBD expr  IDocumentQuery<T> ContainsAll<TValue>(Expression<Func<T, TValue>> propertySelector, IEnumerable<TValue> values)

func (q *DocumentQuery) statistics(stats **QueryStatistics) *DocumentQuery {
	q._statistics(stats)
	return q
}

/*



 IDocumentQuery<T> usingDefaultOperator(QueryOperator queryOperator) {
	_usingDefaultOperator(queryOperator);
	return this;
}


 IDocumentQuery<T> noTracking() {
	_noTracking();
	return this;
}


 IDocumentQuery<T> noCaching() {
	_noCaching();
	return this;
}

//TBD 4.1  IDocumentQuery<T> showTimings()


 IDocumentQuery<T> include(string path) {
	_include(path);
	return this;
}
//TBD expr IDocumentQuery<T> IDocumentQueryBase<T, IDocumentQuery<T>>.Include(Expression<Func<T, object>> path)


 IDocumentQuery<T> not() {
	negateNext();
	return this;
}


 IDocumentQuery<T> take(int count) {
	_take(count);
	return this;
}

 IDocumentQuery<T> skip(int count) {
	_skip(count);
	return this;
}


 IDocumentQuery<T> whereLucene(string fieldName, string whereClause) {
	_whereLucene(fieldName, whereClause, false);
	return this;
}


 IDocumentQuery<T> whereLucene(string fieldName, string whereClause, bool exact) {
	_whereLucene(fieldName, whereClause, exact);
	return this;
}
*/

func (q *DocumentQuery) whereEquals(fieldName string, value Object) *DocumentQuery {
	q._whereEqualsWithExact(fieldName, value, false)
	return q
}

func (q *DocumentQuery) whereEqualsWithExact(fieldName string, value Object, exact bool) *DocumentQuery {
	q._whereEqualsWithExact(fieldName, value, exact)
	return q
}

func (q *DocumentQuery) whereEqualsWithMethodCall(fieldName string, method MethodCall, exact bool) *DocumentQuery {
	q._whereEqualsWithMethodCall(fieldName, method, exact)
	return q
}

//TBD expr IDocumentQuery<T> IDocumentQueryBase<T, IDocumentQuery<T>>.WhereEquals<TValue>(Expression<Func<T, TValue>> propertySelector, TValue value, bool exact)
//TBD expr IDocumentQuery<T> IFilterDocumentQueryBase<T, IDocumentQuery<T>>.WhereEquals<TValue>(Expression<Func<T, TValue>> propertySelector, MethodCall value, bool exact)

func (q *DocumentQuery) whereEqualsWithParams(whereParams *WhereParams) *DocumentQuery {
	q._whereEqualsWithParams(whereParams)
	return q
}

func (q *DocumentQuery) whereNotEquals(fieldName string, value Object) *DocumentQuery {
	q._whereNotEquals(fieldName, value)
	return q
}

func (q *DocumentQuery) whereNotEqualsWithExact(fieldName string, value Object, exact bool) *DocumentQuery {
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

func (q *DocumentQuery) whereNotEqualsWithParams(whereParams *WhereParams) *DocumentQuery {
	q._whereNotEqualsWithParams(whereParams)
	return q
}

/*
 IDocumentQuery<T> whereIn(string fieldName, Collection<Object> values) {
	return whereIn(fieldName, values, false);
}


 IDocumentQuery<T> whereIn(string fieldName, Collection<Object> values, bool exact) {
	_whereIn(fieldName, values, exact);
	return this;
}

//TBD expr  IDocumentQuery<T> WhereIn<TValue>(Expression<Func<T, TValue>> propertySelector, IEnumerable<TValue> values, bool exact = false)


 IDocumentQuery<T> whereStartsWith(string fieldName, Object value) {
	_whereStartsWith(fieldName, value);
	return this;
}


 IDocumentQuery<T> whereEndsWith(string fieldName, Object value) {
	_whereEndsWith(fieldName, value);
	return this;
}

//TBD expr  IDocumentQuery<T> WhereEndsWith<TValue>(Expression<Func<T, TValue>> propertySelector, TValue value)


 IDocumentQuery<T> whereBetween(string fieldName, Object start, Object end) {
	return whereBetween(fieldName, start, end, false);
}


 IDocumentQuery<T> whereBetween(string fieldName, Object start, Object end, bool exact) {
	_whereBetween(fieldName, start, end, exact);
	return this;
}

//TBD expr  IDocumentQuery<T> WhereBetween<TValue>(Expression<Func<T, TValue>> propertySelector, TValue start, TValue end, bool exact = false)

*/

func (q *DocumentQuery) whereGreaterThan(fieldName string, value Object) *DocumentQuery {
	return q.whereGreaterThanWithExact(fieldName, value, false)
}

func (q *DocumentQuery) whereGreaterThanWithExact(fieldName string, value Object, exact bool) *DocumentQuery {
	q._whereGreaterThanWithExact(fieldName, value, exact)
	return q
}

func (q *DocumentQuery) whereGreaterThanOrEqual(fieldName string, value Object) *DocumentQuery {
	return q.whereGreaterThanOrEqualWithExact(fieldName, value, false)
}

func (q *DocumentQuery) whereGreaterThanOrEqualWithExact(fieldName string, value Object, exact bool) *DocumentQuery {
	q._whereGreaterThanOrEqualWithExact(fieldName, value, exact)
	return q
}

//TBD expr  IDocumentQuery<T> WhereGreaterThan<TValue>(Expression<Func<T, TValue>> propertySelector, TValue value, bool exact = false)
//TBD expr  IDocumentQuery<T> WhereGreaterThanOrEqual<TValue>(Expression<Func<T, TValue>> propertySelector, TValue value, bool exact = false)

func (q *DocumentQuery) whereLessThan(fieldName string, value Object) *DocumentQuery {
	return q.whereLessThanWithExact(fieldName, value, false)
}

func (q *DocumentQuery) whereLessThanWithExact(fieldName string, value Object, exact bool) *DocumentQuery {
	q._whereLessThanWithExact(fieldName, value, exact)
	return q
}

//TBD expr  IDocumentQuery<T> WhereLessThanOrEqual<TValue>(Expression<Func<T, TValue>> propertySelector, TValue value, bool exact = false)

func (q *DocumentQuery) whereLessThanOrEqual(fieldName string, value Object) *DocumentQuery {
	return q.whereLessThanOrEqualWithExact(fieldName, value, false)
}

func (q *DocumentQuery) whereLessThanOrEqualWithExact(fieldName string, value Object, exact bool) *DocumentQuery {
	q._whereLessThanOrEqualWithExact(fieldName, value, exact)
	return q
}

//TBD expr  IDocumentQuery<T> WhereLessThanOrEqual<TValue>(Expression<Func<T, TValue>> propertySelector, TValue value, bool exact = false)
//TBD expr  IDocumentQuery<T> WhereExists<TValue>(Expression<Func<T, TValue>> propertySelector)

/*
 IDocumentQuery<T> whereExists(string fieldName) {
	_whereExists(fieldName);
	return this;
}

//TBD expr IDocumentQuery<T> IFilterDocumentQueryBase<T, IDocumentQuery<T>>.WhereRegex<TValue>(Expression<Func<T, TValue>> propertySelector, string pattern)

 IDocumentQuery<T> whereRegex(string fieldName, string pattern) {
	_whereRegex(fieldName, pattern);
	return this;
}

 IDocumentQuery<T> andAlso() {
	_andAlso();
	return this;
}


 IDocumentQuery<T> orElse() {
	_orElse();
	return this;
}


 IDocumentQuery<T> boost(double boost) {
	_boost(boost);
	return this;
}


 IDocumentQuery<T> fuzzy(double fuzzy) {
	_fuzzy(fuzzy);
	return this;
}


 IDocumentQuery<T> proximity(int proximity) {
	_proximity(proximity);
	return this;
}


 IDocumentQuery<T> randomOrdering() {
	_randomOrdering();
	return this;
}


 IDocumentQuery<T> randomOrdering(string seed) {
	_randomOrdering(seed);
	return this;
}

//TBD 4.1  IDocumentQuery<T> customSortUsing(string typeName, bool descending)


 IGroupByDocumentQuery<T> groupBy(string fieldName, string... fieldNames) {
	_groupBy(fieldName, fieldNames);

	return new GroupByDocumentQuery<>(this);
}


 IGroupByDocumentQuery<T> groupBy(GroupBy field, GroupBy... fields) {
	_groupBy(field, fields);

	return new GroupByDocumentQuery<>(this);
}


 <TResult> IDocumentQuery<TResult> ofType(Class<TResult> tResultClass) {
	return createDocumentQueryInternal(tResultClass);
}

 IDocumentQuery<T> orderBy(string field) {
	return orderBy(field, OrderingType.STRING);
}

 IDocumentQuery<T> orderBy(string field, OrderingType ordering) {
	_orderBy(field, ordering);
	return this;
}

//TBD expr  IDocumentQuery<T> OrderBy<TValue>(params Expression<Func<T, TValue>>[] propertySelectors)

 IDocumentQuery<T> orderByDescending(string field) {
	return orderByDescending(field, OrderingType.STRING);
}

 IDocumentQuery<T> orderByDescending(string field, OrderingType ordering) {
	_orderByDescending(field, ordering);
	return this;
}

//TBD expr  IDocumentQuery<T> OrderByDescending<TValue>(params Expression<Func<T, TValue>>[] propertySelectors)


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

		identityProperty := q.getConventions().getIdentityProperty(resultClass)

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
		q.getIndexName(),
		q.getCollectionName(),
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

/*
 IAggregationDocumentQuery<T> aggregateBy(Consumer<IFacetBuilder<T>> builder) {
	FacetBuilder ff = new FacetBuilder<>();
	builder.accept(ff);

	return aggregateBy(ff.getFacet());
}


 IAggregationDocumentQuery<T> aggregateBy(FacetBase facet) {
	_aggregateBy(facet);

	return new AggregationDocumentQuery<T>(this);
}


 IAggregationDocumentQuery<T> aggregateBy(Facet... facets) {
	for (Facet facet : facets) {
		_aggregateBy(facet);
	}

	return new AggregationDocumentQuery<T>(this);
}


 IAggregationDocumentQuery<T> aggregateUsing(string facetSetupDocumentId) {
	_aggregateUsing(facetSetupDocumentId);

	return new AggregationDocumentQuery<T>(this);
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


 IDocumentQuery<T> spatial(string fieldName, Function<SpatialCriteriaFactory, SpatialCriteria> clause) {
	SpatialCriteria criteria = clause.apply(SpatialCriteriaFactory.INSTANCE);
	_spatial(fieldName, criteria);
	return this;
}


 IDocumentQuery<T> spatial(DynamicSpatialField field, Function<SpatialCriteriaFactory, SpatialCriteria> clause) {
	SpatialCriteria criteria = clause.apply(SpatialCriteriaFactory.INSTANCE);
	_spatial(field, criteria);
	return this;
}

//TBD expr  IDocumentQuery<T> Spatial(Func<SpatialDynamicFieldFactory<T>, DynamicSpatialField> field, Func<SpatialCriteriaFactory, SpatialCriteria> clause)
//TBD expr IDocumentQuery<T> IDocumentQueryBase<T, IDocumentQuery<T>>.WithinRadiusOf<TValue>(Expression<Func<T, TValue>> propertySelector, double radius, double latitude, double longitude, SpatialUnits? radiusUnits, double distanceErrorPct)


 IDocumentQuery<T> withinRadiusOf(string fieldName, double radius, double latitude, double longitude) {
	return withinRadiusOf(fieldName, radius, latitude, longitude, nil, Constants.Documents.Indexing.Spatial.DEFAULT_DISTANCE_ERROR_PCT);
}


 IDocumentQuery<T> withinRadiusOf(string fieldName, double radius, double latitude, double longitude, SpatialUnits radiusUnits) {
	return withinRadiusOf(fieldName, radius, latitude, longitude, radiusUnits, Constants.Documents.Indexing.Spatial.DEFAULT_DISTANCE_ERROR_PCT);
}


 IDocumentQuery<T> withinRadiusOf(string fieldName, double radius, double latitude, double longitude, SpatialUnits radiusUnits, double distanceErrorPct) {
	_withinRadiusOf(fieldName, radius, latitude, longitude, radiusUnits, distanceErrorPct);
	return this;
}

//TBD expr IDocumentQuery<T> IDocumentQueryBase<T, IDocumentQuery<T>>.RelatesToShape<TValue>(Expression<Func<T, TValue>> propertySelector, string shapeWkt, SpatialRelation relation, double distanceErrorPct)


 IDocumentQuery<T> relatesToShape(string fieldName, string shapeWkt, SpatialRelation relation) {
	return relatesToShape(fieldName, shapeWkt, relation, Constants.Documents.Indexing.Spatial.DEFAULT_DISTANCE_ERROR_PCT);
}


 IDocumentQuery<T> relatesToShape(string fieldName, string shapeWkt, SpatialRelation relation, double distanceErrorPct) {
	_spatial(fieldName, shapeWkt, relation, distanceErrorPct);
	return this;
}


 IDocumentQuery<T> orderByDistance(DynamicSpatialField field, double latitude, double longitude) {
	_orderByDistance(field, latitude, longitude);
	return this;
}

//TBD expr IDocumentQuery<T> IDocumentQueryBase<T, IDocumentQuery<T>>.OrderByDistance(Func<DynamicSpatialFieldFactory<T>, DynamicSpatialField> field, double latitude, double longitude)


 IDocumentQuery<T> orderByDistance(DynamicSpatialField field, string shapeWkt) {
	_orderByDistance(field, shapeWkt);
	return this;
}

//TBD expr IDocumentQuery<T> IDocumentQueryBase<T, IDocumentQuery<T>>.OrderByDistance(Func<DynamicSpatialFieldFactory<T>, DynamicSpatialField> field, string shapeWkt)


//TBD expr IDocumentQuery<T> IDocumentQueryBase<T, IDocumentQuery<T>>.OrderByDistance<TValue>(Expression<Func<T, TValue>> propertySelector, double latitude, double longitude)


 IDocumentQuery<T> orderByDistance(string fieldName, double latitude, double longitude) {
	_orderByDistance(fieldName, latitude, longitude);
	return this;
}

//TBD expr IDocumentQuery<T> IDocumentQueryBase<T, IDocumentQuery<T>>.OrderByDistance<TValue>(Expression<Func<T, TValue>> propertySelector, string shapeWkt)


 IDocumentQuery<T> orderByDistance(string fieldName, string shapeWkt) {
	_orderByDistance(fieldName, shapeWkt);
	return this;
}


 IDocumentQuery<T> orderByDistanceDescending(DynamicSpatialField field, double latitude, double longitude) {
	_orderByDistanceDescending(field, latitude, longitude);
	return this;
}

//TBD expr IDocumentQuery<T> IDocumentQueryBase<T, IDocumentQuery<T>>.OrderByDistanceDescending(Func<DynamicSpatialFieldFactory<T>, DynamicSpatialField> field, double latitude, double longitude)


 IDocumentQuery<T> orderByDistanceDescending(DynamicSpatialField field, string shapeWkt) {
	_orderByDistanceDescending(field, shapeWkt);
	return this;
}

//TBD expr IDocumentQuery<T> IDocumentQueryBase<T, IDocumentQuery<T>>.OrderByDistanceDescending(Func<DynamicSpatialFieldFactory<T>, DynamicSpatialField> field, string shapeWkt)

//TBD expr IDocumentQuery<T> IDocumentQueryBase<T, IDocumentQuery<T>>.OrderByDistanceDescending<TValue>(Expression<Func<T, TValue>> propertySelector, double latitude, double longitude)


 IDocumentQuery<T> orderByDistanceDescending(string fieldName, double latitude, double longitude) {
	_orderByDistanceDescending(fieldName, latitude, longitude);
	return this;
}

//TBD expr IDocumentQuery<T> IDocumentQueryBase<T, IDocumentQuery<T>>.OrderByDistanceDescending<TValue>(Expression<Func<T, TValue>> propertySelector, string shapeWkt)


 IDocumentQuery<T> orderByDistanceDescending(string fieldName, string shapeWkt) {
	_orderByDistanceDescending(fieldName, shapeWkt);
	return this;
}


 IDocumentQuery<T> moreLikeThis(MoreLikeThisBase moreLikeThis) {
	try (MoreLikeThisScope mlt = _moreLikeThis()) {
		mlt.withOptions(moreLikeThis.getOptions());

		if (moreLikeThis instanceof MoreLikeThisUsingDocument) {
			mlt.withDocument(((MoreLikeThisUsingDocument) moreLikeThis).getDocumentJson());
		}
	}

	return this;
}


 IDocumentQuery<T> moreLikeThis(Consumer<IMoreLikeThisBuilderForDocumentQuery<T>> builder) {
	MoreLikeThisBuilder<T> f = new MoreLikeThisBuilder<>();
	builder.accept(f);

	try (MoreLikeThisScope moreLikeThis = _moreLikeThis()) {
		moreLikeThis.withOptions(f.getMoreLikeThis().getOptions());

		if (f.getMoreLikeThis() instanceof MoreLikeThisUsingDocument) {
			moreLikeThis.withDocument(((MoreLikeThisUsingDocument) f.getMoreLikeThis()).getDocumentJson());
		} else if (f.getMoreLikeThis() instanceof MoreLikeThisUsingDocumentForDocumentQuery) {
			((MoreLikeThisUsingDocumentForDocumentQuery) f.getMoreLikeThis()).getForDocumentQuery().accept(this);
		}
	}

	return this;
}


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
