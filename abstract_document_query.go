package ravendb

import (
	"reflect"
	"strconv"
	"strings"
	"time"
)

// TODO: implement me
type AbstractDocumentQuery struct {
	clazz                    reflect.Type
	_aliasToGroupByFieldName map[string]string
	defaultOperator          QueryOperator

	rootTypes *TypeSet

	negate              bool
	indexName           string
	collectionName      string
	_currentClauseDepth int
	queryRaw            string
	queryParameters     Parameters

	isIntersect bool
	isGroupBy   bool

	theSession *InMemoryDocumentSessionOperations

	pageSize int // 0 is unset

	selectTokens       []QueryToken
	fromToken          *FromToken
	declareToken       *DeclareToken
	loadTokens         []*LoadToken
	fieldsToFetchToken *FieldsToFetchToken

	whereTokens   []QueryToken
	groupByTokens []QueryToken
	orderByTokens []QueryToken

	start        int
	_conventions *DocumentConventions

	timeout time.Duration

	theWaitForNonStaleResults bool

	includes *StringSet

	queryStats *QueryStatistics // TODO: queryStats = NewQueryStatistics

	disableEntitiesTracking bool

	disableCaching bool

	_isInMoreLikeThis bool

	beforeQueryExecutedCallback []func(*IndexQuery)
	afterQueryExecutedCallback  []func(*QueryResult)
	afterStreamExecutedCallback []func(ObjectNode)

	queryOperation *QueryOperation
}

func (q *AbstractDocumentQuery) getIndexName() string {
	return q.indexName
}

func (q *AbstractDocumentQuery) getCollectionName() string {
	return q.collectionName
}

func (q *AbstractDocumentQuery) isDistinct() bool {
	if len(q.selectTokens) == 0 {
		return false
	}
	_, ok := q.selectTokens[0].(*DistinctToken)
	return ok
}

func (q *AbstractDocumentQuery) getConventions() *DocumentConventions {
	return q._conventions
}

func (q *AbstractDocumentQuery) getSession() *InMemoryDocumentSessionOperations {
	return q.theSession
}

func (q *AbstractDocumentQuery) isDynamicMapReduce() bool {
	return len(q.groupByTokens) > 0
}

func AbstractDocumentQuery_getDefaultTimeout() time.Duration {
	return time.Second * 15
}

func NewAbstractDocumentQuery(clazz reflect.Type, session *InMemoryDocumentSessionOperations, indexName string, collectionName string, isGroupBy bool, declareToken *DeclareToken, loadTokens []*LoadToken, fromAlias string) *AbstractDocumentQuery {
	res := &AbstractDocumentQuery{
		clazz:                    clazz,
		rootTypes:                NewTypeSet(),
		defaultOperator:          QueryOperator_AND,
		isGroupBy:                isGroupBy,
		indexName:                indexName,
		collectionName:           collectionName,
		declareToken:             declareToken,
		loadTokens:               loadTokens,
		theSession:               session,
		_aliasToGroupByFieldName: make(map[string]string),
		queryParameters:          make(map[string]Object),
	}
	res.rootTypes.add(clazz)
	res.fromToken = FromToken_create(indexName, collectionName, fromAlias)
	f := func(queryResult *QueryResult) {
		res.updateStatsAndHighlightings(queryResult)
	}
	res._addAfterQueryExecutedListener(f)
	if session == nil {
		res._conventions = NewDocumentConventions()
	} else {
		res._conventions = session.getConventions()
	}
	return res
}

func (q *AbstractDocumentQuery) getQueryClass() reflect.Type {
	return q.clazz
}

func (q *AbstractDocumentQuery) _usingDefaultOperator(operator QueryOperator) {
	if len(q.whereTokens) > 0 {
		//throw new IllegalStateException("Default operator can only be set before any where clause is added.");
		panicIf(true, "Default operator can only be set before any where clause is added.")
	}

	q.defaultOperator = operator
}

func (q *AbstractDocumentQuery) _waitForNonStaleResults(waitTimeout time.Duration) {
	q.theWaitForNonStaleResults = true
	if waitTimeout == 0 {
		waitTimeout = AbstractDocumentQuery_getDefaultTimeout()
	}
	q.timeout = waitTimeout
}

func (q *AbstractDocumentQuery) initializeQueryOperation() *QueryOperation {
	indexQuery := q.getIndexQuery()

	return NewQueryOperation(q.theSession, q.indexName, indexQuery, q.fieldsToFetchToken, q.disableEntitiesTracking, false, false)
}

func (q *AbstractDocumentQuery) getIndexQuery() *IndexQuery {
	query := q.String()
	indexQuery := q.generateIndexQuery(query)
	q.invokeBeforeQueryExecuted(indexQuery)
	return indexQuery
}

func (q *AbstractDocumentQuery) getProjectionFields() []string {

	if q.fieldsToFetchToken != nil && q.fieldsToFetchToken.projections != nil {
		return q.fieldsToFetchToken.projections
	}
	return nil
}

func (q *AbstractDocumentQuery) _randomOrdering() {
	q.assertNoRawQuery()
	q.orderByTokens = append(q.orderByTokens, OrderByToken_random)
}

func (q *AbstractDocumentQuery) _randomOrderingWithSeed(seed string) {
	q.assertNoRawQuery()

	if StringUtils_isBlank(seed) {
		q._randomOrdering()
		return
	}

	q.orderByTokens = append(q.orderByTokens, OrderByToken_createRandom(seed))
}

func (q *AbstractDocumentQuery) addGroupByAlias(fieldName string, projectedName string) {
	q._aliasToGroupByFieldName[projectedName] = fieldName
}

func (q *AbstractDocumentQuery) assertNoRawQuery() {
	panicIf(q.queryRaw != "", "RawQuery was called, cannot modify this query by calling on operations that would modify the query (such as Where, Select, OrderBy, GroupBy, etc)")
}

func (q *AbstractDocumentQuery) _addParameter(name string, value Object) {
	name = strings.TrimPrefix(name, "$")
	if _, ok := q.queryParameters[name]; ok {
		// throw new IllegalStateException("The parameter " + name + " was already added");
		panicIf(true, "The parameter "+name+" was already added")
	}

	q.queryParameters[name] = value
}

func (q *AbstractDocumentQuery) _groupBy(fieldName string, fieldNames ...string) {
	var mapping []*GroupBy
	for _, x := range fieldNames {
		el := GroupBy_field(x)
		mapping = append(mapping, el)
	}
	q._groupBy2(GroupBy_field(fieldName), mapping...)
}

// TODO: better name
func (q *AbstractDocumentQuery) _groupBy2(field *GroupBy, fields ...*GroupBy) {
	if !q.fromToken.isDynamic() {
		//throw new IllegalStateException("groupBy only works with dynamic queries");
		panicIf(true, "groupBy only works with dynamic queries")
	}

	q.assertNoRawQuery()
	q.isGroupBy = true

	fieldName := q.ensureValidFieldName(field.getField(), false)

	q.groupByTokens = append(q.groupByTokens, GroupByToken_createWithMethod(fieldName, field.getMethod()))

	if len(fields) == 0 {
		return
	}

	for _, item := range fields {
		fieldName = q.ensureValidFieldName(item.getField(), false)
		q.groupByTokens = append(q.groupByTokens, GroupByToken_createWithMethod(fieldName, item.getMethod()))
	}
}

func (q *AbstractDocumentQuery) _groupByKey(fieldName string, projectedName string) {
	q.assertNoRawQuery()
	q.isGroupBy = true

	_, hasProjectedName := q._aliasToGroupByFieldName[projectedName]
	_, hasFieldName := q._aliasToGroupByFieldName[fieldName]

	if projectedName != "" && hasProjectedName {
		aliasedFieldName := q._aliasToGroupByFieldName[projectedName]
		if fieldName == "" || strings.EqualFold(fieldName, projectedName) {
			fieldName = aliasedFieldName
		}
	} else if fieldName != "" && hasFieldName {
		aliasedFieldName := q._aliasToGroupByFieldName[fieldName]
		fieldName = aliasedFieldName
	}

	q.selectTokens = append(q.selectTokens, GroupByKeyToken_create(fieldName, projectedName))
}

// projectedName is optional
func (q *AbstractDocumentQuery) _groupBySum(fieldName string, projectedName string) {
	q.assertNoRawQuery()
	q.isGroupBy = true

	fieldName = q.ensureValidFieldName(fieldName, false)
	q.selectTokens = append(q.selectTokens, GroupBySumToken_create(fieldName, projectedName))
}

// projectedName is optional
func (q *AbstractDocumentQuery) _groupByCount(projectedName string) {
	q.assertNoRawQuery()
	q.isGroupBy = true

	q.selectTokens = append(q.selectTokens, GroupByCountToken_create(projectedName))
}

func (q *AbstractDocumentQuery) _whereTrue() {
	tokensRef := q.getCurrentWhereTokensRef()
	q.appendOperatorIfNeeded(tokensRef)
	q.negateIfNeeded(tokensRef, "")

	tokens := *tokensRef
	tokens = append(tokens, TrueToken_INSTANCE)
	*tokensRef = tokens
}

func (q *AbstractDocumentQuery) _moreLikeThis() *MoreLikeThisScope {
	q.appendOperatorIfNeeded(&q.whereTokens)

	token := NewMoreLikeThisToken()
	q.whereTokens = append(q.whereTokens, token)

	q._isInMoreLikeThis = true
	funcAddQueryParameter := func(o Object) string {
		return q.addQueryParameter(o)
	}
	funcOnDispose := func() {
		q._isInMoreLikeThis = false
	}
	return NewMoreLikeThisScope(token, funcAddQueryParameter, funcOnDispose)
}

func (q *AbstractDocumentQuery) _include(path string) {
	q.includes.add(path)
}

func (q *AbstractDocumentQuery) _take(count int) {
	q.pageSize = count
}

func (q *AbstractDocumentQuery) _skip(count int) {
	q.start = count
}

func (q *AbstractDocumentQuery) _whereLucene(fieldName string, whereClause string, exact bool) {
	fieldName = q.ensureValidFieldName(fieldName, false)

	tokensRef := q.getCurrentWhereTokensRef()
	tokens := *tokensRef
	q.appendOperatorIfNeeded(tokensRef)
	q.negateIfNeeded(tokensRef, fieldName)

	var options *WhereOptions
	if exact {
		options = NewWhereOptionsWithExact(exact)
	}
	whereToken := WhereToken_createWithOptions(WhereOperator_LUCENE, fieldName, q.addQueryParameter(whereClause), options)
	tokens = append(tokens, whereToken)
	*tokensRef = tokens
}

func (q *AbstractDocumentQuery) _openSubclause() {
	q._currentClauseDepth++

	tokensRef := q.getCurrentWhereTokensRef()
	q.appendOperatorIfNeeded(tokensRef)
	q.negateIfNeeded(tokensRef, "")

	tokens := *tokensRef
	tokens = append(tokens, OpenSubclauseToken_INSTANCE)
	*tokensRef = tokens
}

func (q *AbstractDocumentQuery) _closeSubclause() {
	q._currentClauseDepth--

	tokensRef := q.getCurrentWhereTokensRef()
	tokens := *tokensRef
	tokens = append(tokens, CloseSubclauseToken_INSTANCE)
	*tokensRef = tokens
}

func (q *AbstractDocumentQuery) _whereEquals(fieldName string, value Object) {
	q._whereEqualsWithExact(fieldName, value, false)
}

func (q *AbstractDocumentQuery) _whereEqualsWithExact(fieldName string, value Object, exact bool) {
	params := NewWhereParams()
	params.setFieldName(fieldName)
	params.setValue(value)
	params.setExact(exact)
	q._whereEqualsWithParams(params)
}

func (q *AbstractDocumentQuery) _whereEqualsWithMethodCall(fieldName string, method MethodCall, exact bool) {
	q._whereEqualsWithExact(fieldName, method, exact)
}

func (q *AbstractDocumentQuery) _whereEqualsWithParams(whereParams *WhereParams) {
	if q.negate {
		q.negate = false
		q._whereNotEqualsWithParams(whereParams)
		return
	}

	whereParams.setFieldName(q.ensureValidFieldName(whereParams.getFieldName(), whereParams.isNestedPath()))

	tokensRef := q.getCurrentWhereTokensRef()
	q.appendOperatorIfNeeded(tokensRef)

	if q.ifValueIsMethod(WhereOperator_EQUALS, whereParams, tokensRef) {
		return
	}

	transformToEqualValue := q.transformValue(whereParams)
	addQueryParameter := q.addQueryParameter(transformToEqualValue)
	whereToken := WhereToken_createWithOptions(WhereOperator_EQUALS, whereParams.getFieldName(), addQueryParameter, NewWhereOptionsWithExact(whereParams.isExact()))

	tokens := *tokensRef
	tokens = append(tokens, whereToken)
	*tokensRef = tokens
}

func (q *AbstractDocumentQuery) ifValueIsMethod(op WhereOperator, whereParams *WhereParams, tokensRef *[]QueryToken) bool {
	if mc, ok := whereParams.getValue().(*CmpXchg); ok {
		n := len(mc.args)
		args := make([]string, n)
		for i := 0; i < n; i++ {
			args[i] = q.addQueryParameter(mc.args[i])
		}

		opts := NewWhereOptionsWithMethod(MethodsType_CMP_X_CHG, args, mc.accessPath, whereParams.isExact())
		token := WhereToken_createWithOptions(op, whereParams.getFieldName(), "", opts)

		tokens := *tokensRef
		tokens = append(tokens, token)
		*tokensRef = tokens
		return true
	} else {
		//throw new IllegalArgumentException("Unknown method " + type);
		panicIf(true, "Unknown method %T", whereParams.getValue())
	}

	return false
}

func (q *AbstractDocumentQuery) _whereNotEquals(fieldName string, value Object) {
	q._whereNotEqualsWithExact(fieldName, value, false)
}

func (q *AbstractDocumentQuery) _whereNotEqualsWithExact(fieldName string, value Object, exact bool) {
	params := NewWhereParams()
	params.setFieldName(fieldName)
	params.setValue(value)
	params.setExact(exact)

	q._whereNotEqualsWithParams(params)
}

func (q *AbstractDocumentQuery) _whereNotEqualsWithMethod(fieldName string, method MethodCall) {
	q._whereNotEquals(fieldName, method)
}

func (q *AbstractDocumentQuery) _whereNotEqualsWithMethodAndExact(fieldName string, method MethodCall, exact bool) {
	q._whereNotEqualsWithExact(fieldName, method, exact)
}

func (q *AbstractDocumentQuery) _whereNotEqualsWithParams(whereParams *WhereParams) {
	if q.negate {
		q.negate = false
		q._whereEqualsWithParams(whereParams)
		return
	}

	transformToEqualValue := q.transformValue(whereParams)

	tokensRef := q.getCurrentWhereTokensRef()
	q.appendOperatorIfNeeded(tokensRef)

	whereParams.setFieldName(q.ensureValidFieldName(whereParams.getFieldName(), whereParams.isNestedPath()))

	if q.ifValueIsMethod(WhereOperator_NOT_EQUALS, whereParams, tokensRef) {
		return
	}

	whereToken := WhereToken_createWithOptions(WhereOperator_NOT_EQUALS, whereParams.getFieldName(), q.addQueryParameter(transformToEqualValue), NewWhereOptionsWithExact(whereParams.isExact()))
	tokens := *tokensRef
	tokens = append(tokens, whereToken)
	*tokensRef = tokens
}

func (q *AbstractDocumentQuery) negateNext() {
	q.negate = !q.negate
}

func (q *AbstractDocumentQuery) _whereIn(fieldName string, values []Object) {
	q._whereInWithExact(fieldName, values, false)
}

func (q *AbstractDocumentQuery) _whereInWithExact(fieldName string, values []Object, exact bool) {
	fieldName = q.ensureValidFieldName(fieldName, false)

	tokensRef := q.getCurrentWhereTokensRef()
	q.appendOperatorIfNeeded(tokensRef)
	q.negateIfNeeded(tokensRef, fieldName)

	whereToken := WhereToken_create(WhereOperator_IN, fieldName, q.addQueryParameter(q.transformCollection(fieldName, AbstractDocumentQuery_unpackCollection(values))))

	tokens := *tokensRef
	tokens = append(tokens, whereToken)
	*tokensRef = tokens
}

func (q *AbstractDocumentQuery) _whereStartsWith(fieldName string, value Object) {
	whereParams := NewWhereParams()
	whereParams.setFieldName(fieldName)
	whereParams.setValue(value)
	whereParams.setAllowWildcards(true)

	transformToEqualValue := q.transformValue(whereParams)

	tokensRef := q.getCurrentWhereTokensRef()
	q.appendOperatorIfNeeded(tokensRef)

	whereParams.setFieldName(q.ensureValidFieldName(whereParams.getFieldName(), whereParams.isNestedPath()))
	q.negateIfNeeded(tokensRef, whereParams.getFieldName())

	whereToken := WhereToken_create(WhereOperator_STARTS_WITH, whereParams.getFieldName(), q.addQueryParameter(transformToEqualValue))

	tokens := *tokensRef
	tokens = append(tokens, whereToken)
	*tokensRef = tokens
}

func (q *AbstractDocumentQuery) _whereEndsWith(fieldName string, value Object) {
	whereParams := NewWhereParams()
	whereParams.setFieldName(fieldName)
	whereParams.setValue(value)
	whereParams.setAllowWildcards(true)

	transformToEqualValue := q.transformValue(whereParams)

	tokensRef := q.getCurrentWhereTokensRef()
	q.appendOperatorIfNeeded(tokensRef)

	whereParams.setFieldName(q.ensureValidFieldName(whereParams.getFieldName(), whereParams.isNestedPath()))
	q.negateIfNeeded(tokensRef, whereParams.getFieldName())

	whereToken := WhereToken_create(WhereOperator_ENDS_WITH, whereParams.getFieldName(), q.addQueryParameter(transformToEqualValue))

	tokens := *tokensRef
	tokens = append(tokens, whereToken)
	*tokensRef = tokens
}

func (q *AbstractDocumentQuery) _whereBetween(fieldName string, start Object, end Object) {
	q._whereBetweenWithExact(fieldName, start, end, false)
}

func (q *AbstractDocumentQuery) _whereBetweenWithExact(fieldName string, start Object, end Object, exact bool) {
	fieldName = q.ensureValidFieldName(fieldName, false)

	tokensRef := q.getCurrentWhereTokensRef()
	q.appendOperatorIfNeeded(tokensRef)
	q.negateIfNeeded(tokensRef, fieldName)

	startParams := NewWhereParams()
	startParams.setValue(start)
	startParams.setFieldName(fieldName)

	endParams := NewWhereParams()
	endParams.setValue(end)
	endParams.setFieldName(fieldName)

	fromParam := interface{}("*")
	if start == nil {
		fromParam = q.transformValueWithRange(startParams, true)
	}
	fromParameterName := q.addQueryParameter(fromParam)

	toParam := interface{}("NULL")
	// TODO: should this be end == nil? A bug in Java code?
	if start == nil {
		toParam = q.transformValueWithRange(endParams, true)
	}
	toParameterName := q.addQueryParameter(toParam)

	whereToken := WhereToken_createWithOptions(WhereOperator_BETWEEN, fieldName, "", NewWhereOptionsWithFromTo(exact, fromParameterName, toParameterName))

	tokens := *tokensRef
	tokens = append(tokens, whereToken)
	*tokensRef = tokens
}

func (q *AbstractDocumentQuery) _whereGreaterThan(fieldName string, value Object) {
	q._whereGreaterThanWithExact(fieldName, value, false)
}

func (q *AbstractDocumentQuery) _whereGreaterThanWithExact(fieldName string, value Object, exact bool) {
	fieldName = q.ensureValidFieldName(fieldName, false)

	tokensRef := q.getCurrentWhereTokensRef()
	q.appendOperatorIfNeeded(tokensRef)
	q.negateIfNeeded(tokensRef, fieldName)

	whereParams := NewWhereParams()
	whereParams.setValue(value)
	whereParams.setFieldName(fieldName)

	paramValue := interface{}("*")
	if value != nil {
		paramValue = q.transformValueWithRange(whereParams, true)
	}
	parameter := q.addQueryParameter(paramValue)

	whereToken := WhereToken_createWithOptions(WhereOperator_GREATER_THAN, fieldName, parameter, NewWhereOptionsWithExact(exact))

	tokens := *tokensRef
	tokens = append(tokens, whereToken)
	*tokensRef = tokens
}

func (q *AbstractDocumentQuery) _whereGreaterThanOrEqual(fieldName string, value Object) {
	q._whereGreaterThanOrEqualWithExact(fieldName, value, false)
}

func (q *AbstractDocumentQuery) _whereGreaterThanOrEqualWithExact(fieldName string, value Object, exact bool) {
	fieldName = q.ensureValidFieldName(fieldName, false)

	tokensRef := q.getCurrentWhereTokensRef()
	q.appendOperatorIfNeeded(tokensRef)
	q.negateIfNeeded(tokensRef, fieldName)

	whereParams := NewWhereParams()
	whereParams.setValue(value)
	whereParams.setFieldName(fieldName)

	paramValue := interface{}("*")
	if value != nil {
		paramValue = q.transformValueWithRange(whereParams, true)
	}

	parameter := q.addQueryParameter(paramValue)

	whereToken := WhereToken_createWithOptions(WhereOperator_GREATER_THAN_OR_EQUAL, fieldName, parameter, NewWhereOptionsWithExact(exact))

	tokens := *tokensRef
	tokens = append(tokens, whereToken)
	*tokensRef = tokens
}

func (q *AbstractDocumentQuery) _whereLessThan(fieldName string, value Object) {
	q._whereLessThanWithExact(fieldName, value, false)
}

func (q *AbstractDocumentQuery) _whereLessThanWithExact(fieldName string, value Object, exact bool) {
	fieldName = q.ensureValidFieldName(fieldName, false)

	tokensRef := q.getCurrentWhereTokensRef()
	q.appendOperatorIfNeeded(tokensRef)
	q.negateIfNeeded(tokensRef, fieldName)

	whereParams := NewWhereParams()
	whereParams.setValue(value)
	whereParams.setFieldName(fieldName)

	paramValue := interface{}("NULL")
	if value != nil {
		paramValue = q.transformValueWithRange(whereParams, true)
	}
	parameter := q.addQueryParameter(paramValue)
	whereToken := WhereToken_createWithOptions(WhereOperator_LESS_THAN, fieldName, parameter, NewWhereOptionsWithExact(exact))

	tokens := *tokensRef
	tokens = append(tokens, whereToken)
	*tokensRef = tokens
}

func (q *AbstractDocumentQuery) _whereLessThanOrEqual(fieldName string, value Object) {
	q._whereLessThanOrEqualWithExact(fieldName, value, false)
}

func (q *AbstractDocumentQuery) _whereLessThanOrEqualWithExact(fieldName string, value Object, exact bool) {
	tokensRef := q.getCurrentWhereTokensRef()
	q.appendOperatorIfNeeded(tokensRef)
	q.negateIfNeeded(tokensRef, fieldName)

	whereParams := NewWhereParams()
	whereParams.setValue(value)
	whereParams.setFieldName(fieldName)

	paramValue := interface{}("NULL")
	if value != nil {
		paramValue = q.transformValueWithRange(whereParams, true)
	}
	parameter := q.addQueryParameter(paramValue)
	whereToken := WhereToken_createWithOptions(WhereOperator_LESS_THAN_OR_EQUAL, fieldName, parameter, NewWhereOptionsWithExact(exact))

	tokens := *tokensRef
	tokens = append(tokens, whereToken)
	*tokensRef = tokens
}

func (q *AbstractDocumentQuery) _whereRegex(fieldName string, pattern string) {
	tokensRef := q.getCurrentWhereTokensRef()
	q.appendOperatorIfNeeded(tokensRef)
	q.negateIfNeeded(tokensRef, fieldName)

	whereParams := NewWhereParams()
	whereParams.setValue(pattern)
	whereParams.setFieldName(fieldName)

	parameter := q.addQueryParameter(q.transformValue(whereParams))

	whereToken := WhereToken_create(WhereOperator_REGEX, fieldName, parameter)

	tokens := *tokensRef
	tokens = append(tokens, whereToken)
	*tokensRef = tokens
}

func (q *AbstractDocumentQuery) _andAlso() {
	tokensRef := q.getCurrentWhereTokensRef()
	tokens := *tokensRef

	n := len(tokens)
	if n == 0 {
		return
	}

	lastToken := tokens[n-1]
	if _, ok := lastToken.(*QueryOperatorToken); ok {
		//throw new IllegalStateException("Cannot add AND, previous token was already an operator token.");
		panicIf(true, "Cannot add AND, previous token was already an operator token.")
	}

	tokens = append(tokens, QueryOperatorToken_AND)
	*tokensRef = tokens
}

func (q *AbstractDocumentQuery) _orElse() {
	tokensRef := q.getCurrentWhereTokensRef()
	tokens := *tokensRef
	n := len(tokens)
	if n == 0 {
		return
	}

	lastToken := tokens[n-1]
	if _, ok := lastToken.(*QueryOperatorToken); ok {
		//throw new IllegalStateException("Cannot add OR, previous token was already an operator token.");
		panicIf(true, "Cannot add OR, previous token was already an operator token.")
	}

	tokens = append(tokens, QueryOperatorToken_OR)
	*tokensRef = tokens
}

func (q *AbstractDocumentQuery) _boost(boost float64) {
	if boost == 1.0 {
		return
	}

	tokens := q.getCurrentWhereTokens()
	n := len(tokens)
	if n == 0 {
		//throw new IllegalStateException("Missing where clause");
		panicIf(true, "Missing where clause")
	}

	maybeWhereToken := tokens[n-1]
	whereToken, ok := maybeWhereToken.(*WhereToken)
	if !ok {
		//throw new IllegalStateException("Missing where clause");
		panicIf(true, "Missing where clause")
	}

	if boost <= 0.0 {
		//throw new IllegalArgumentException("Boost factor must be a positive number");
		panicIf(true, "Boost factor must be a positive number")
	}

	whereToken.getOptions().setBoost(boost)
}

func (q *AbstractDocumentQuery) _fuzzy(fuzzy float64) {
	tokens := q.getCurrentWhereTokens()
	n := len(tokens)
	if n == 0 {
		//throw new IllegalStateException("Missing where clause");
		panicIf(true, "Missing where clause")
	}

	maybeWhereToken := tokens[n-1]
	whereToken, ok := maybeWhereToken.(*WhereToken)
	if !ok {
		//throw new IllegalStateException("Missing where clause");
		panicIf(true, "Missing where clause")
	}

	if fuzzy < 0.0 || fuzzy > 1.0 {
		//throw new IllegalArgumentException("Fuzzy distance must be between 0.0 and 1.0");
		panicIf(true, "Fuzzy distance must be between 0.0 and 1.0")
	}

	whereToken.getOptions().setFuzzy(fuzzy)
}

func (q *AbstractDocumentQuery) _proximity(proximity int) {
	tokens := q.getCurrentWhereTokens()

	n := len(tokens)
	if n == 0 {
		//throw new IllegalStateException("Missing where clause");
		panicIf(true, "Missing where clause")
	}

	maybeWhereToken := tokens[n-1]
	whereToken, ok := maybeWhereToken.(*WhereToken)
	if !ok {
		//throw new IllegalStateException("Missing where clause");
		panicIf(true, "Missing where clause")
	}

	if proximity < 1 {
		//throw new IllegalArgumentException("Proximity distance must be a positive number");
		panicIf(true, "Proximity distance must be a positive number")
	}

	whereToken.getOptions().setProximity(proximity)
}

func (q *AbstractDocumentQuery) _orderBy(field string) {
	q._orderByWithOrdering(field, OrderingType_STRING)
}

func (q *AbstractDocumentQuery) _orderByWithOrdering(field string, ordering OrderingType) {
	q.assertNoRawQuery()
	f := q.ensureValidFieldName(field, false)
	q.orderByTokens = append(q.orderByTokens, OrderByToken_createAscending(f, ordering))
}

func (q *AbstractDocumentQuery) _orderByDescending(field string) {
	q._orderByDescendingWithOrdering(field, OrderingType_STRING)
}

func (q *AbstractDocumentQuery) _orderByDescendingWithOrdering(field string, ordering OrderingType) {
	q.assertNoRawQuery()
	f := q.ensureValidFieldName(field, false)
	q.orderByTokens = append(q.orderByTokens, OrderByToken_createDescending(f, ordering))
}

func (q *AbstractDocumentQuery) _orderByScore() {
	q.assertNoRawQuery()

	q.orderByTokens = append(q.orderByTokens, OrderByToken_scoreAscending)
}

func (q *AbstractDocumentQuery) _orderByScoreDescending() {
	q.assertNoRawQuery()
	q.orderByTokens = append(q.orderByTokens, OrderByToken_scoreDescending)
}

func (q *AbstractDocumentQuery) _statistics(stats **QueryStatistics) {
	*stats = q.queryStats
}

func (q *AbstractDocumentQuery) invokeAfterQueryExecuted(result *QueryResult) {
	for _, cb := range q.afterQueryExecutedCallback {
		cb(result)
	}
}

func (q *AbstractDocumentQuery) invokeBeforeQueryExecuted(query *IndexQuery) {
	for _, cb := range q.beforeQueryExecutedCallback {
		cb(query)
	}
}

func (q *AbstractDocumentQuery) invokeAfterStreamExecuted(result ObjectNode) {
	for _, cb := range q.afterStreamExecutedCallback {
		cb(result)
	}
}

func (q *AbstractDocumentQuery) generateIndexQuery(query string) *IndexQuery {
	indexQuery := NewIndexQuery("")
	indexQuery.setQuery(query)
	indexQuery.setStart(q.start)
	indexQuery.setWaitForNonStaleResults(q.theWaitForNonStaleResults)
	indexQuery.setWaitForNonStaleResultsTimeout(q.timeout)
	indexQuery.setQueryParameters(q.queryParameters)
	indexQuery.setDisableCaching(q.disableCaching)

	if q.pageSize != 0 {
		indexQuery.setPageSize(q.pageSize)
	}
	return indexQuery
}

func (q *AbstractDocumentQuery) _search(fieldName string, searchTerms string) {
	q._searchWithOperator(fieldName, searchTerms, SearchOperator_OR)
}

func (q *AbstractDocumentQuery) _searchWithOperator(fieldName string, searchTerms string, operator SearchOperator) {
	tokensRef := q.getCurrentWhereTokensRef()
	q.appendOperatorIfNeeded(tokensRef)

	fieldName = q.ensureValidFieldName(fieldName, false)
	q.negateIfNeeded(tokensRef, fieldName)

	whereToken := WhereToken_createWithOptions(WhereOperator_SEARCH, fieldName, q.addQueryParameter(searchTerms), NewWhereOptionsWithOperator(operator))

	tokens := *tokensRef
	tokens = append(tokens, whereToken)
	*tokensRef = tokens
}

func (q *AbstractDocumentQuery) String() string {
	if q.queryRaw != "" {
		return q.queryRaw
	}

	if q._currentClauseDepth != 0 {
		// throw new IllegalStateException("A clause was not closed correctly within this query, current clause depth = " + _currentClauseDepth);
		panicIf(true, "A clause was not closed correctly within this query, current clause depth = %d", q._currentClauseDepth)
	}

	queryText := NewStringBuilder()
	q.buildDeclare(queryText)
	q.buildFrom(queryText)
	q.buildGroupBy(queryText)
	q.buildWhere(queryText)
	q.buildOrderBy(queryText)

	q.buildLoad(queryText)
	q.buildSelect(queryText)
	q.buildInclude(queryText)

	return queryText.String()
}

func (q *AbstractDocumentQuery) buildInclude(queryText *StringBuilder) {
	if q.includes != nil && q.includes.Size() == 0 {
		return
	}

	queryText.append(" include ")
	for i, include := range q.includes.strings {
		if i > 0 {
			queryText.append(",")
		}

		requiredQuotes := false

		for _, ch := range include {
			if !Character_isLetterOrDigit(ch) && ch != '_' && ch != '.' {
				requiredQuotes = true
				break
			}
		}

		if requiredQuotes {
			s := strings.Replace(include, "'", "\\'", -1)
			queryText.append("'").append(s).append("'")
		} else {
			queryText.append(include)
		}
	}
}

func (q *AbstractDocumentQuery) _intersect() {
	tokensRef := q.getCurrentWhereTokensRef()
	tokens := *tokensRef
	n := len(tokens)
	if n > 0 {
		last := tokens[n-1]
		_, isWhere := last.(*WhereToken)
		_, isClose := last.(*CloseSubclauseToken)
		if isWhere || isClose {
			q.isIntersect = true

			tokens = append(tokens, IntersectMarkerToken_INSTANCE)
			*tokensRef = tokens
			return
		}
	}

	//throw new IllegalStateException("Cannot add INTERSECT at this point.");
	panicIf(true, "Cannot add INTERSECT at this point.")
}

func (q *AbstractDocumentQuery) _whereExists(fieldName string) {
	fieldName = q.ensureValidFieldName(fieldName, false)

	tokensRef := q.getCurrentWhereTokensRef()
	q.appendOperatorIfNeeded(tokensRef)
	q.negateIfNeeded(tokensRef, fieldName)

	tokens := *tokensRef
	tokens = append(tokens, WhereToken_create(WhereOperator_EXISTS, fieldName, ""))
	*tokensRef = tokens
}

func (q *AbstractDocumentQuery) _containsAny(fieldName string, values []Object) {
	fieldName = q.ensureValidFieldName(fieldName, false)

	tokensRef := q.getCurrentWhereTokensRef()
	q.appendOperatorIfNeeded(tokensRef)
	q.negateIfNeeded(tokensRef, fieldName)

	array := q.transformCollection(fieldName, AbstractDocumentQuery_unpackCollection(values))
	whereToken := WhereToken_createWithOptions(WhereOperator_IN, fieldName, q.addQueryParameter(array), NewWhereOptionsWithExact(false))

	tokens := *tokensRef
	tokens = append(tokens, whereToken)
	*tokensRef = tokens
}

func (q *AbstractDocumentQuery) _containsAll(fieldName string, values []Object) {
	fieldName = q.ensureValidFieldName(fieldName, false)

	tokensRef := q.getCurrentWhereTokensRef()
	q.appendOperatorIfNeeded(tokensRef)
	q.negateIfNeeded(tokensRef, fieldName)

	array := q.transformCollection(fieldName, AbstractDocumentQuery_unpackCollection(values))

	tokens := *tokensRef
	if len(array) == 0 {
		tokens = append(tokens, TrueToken_INSTANCE)
	} else {
		whereToken := WhereToken_create(WhereOperator_ALL_IN, fieldName, q.addQueryParameter(array))
		tokens = append(tokens, whereToken)
	}
	*tokensRef = tokens
}

func (q *AbstractDocumentQuery) _addRootType(clazz reflect.Type) {
	q.rootTypes.add(clazz)
}

func (q *AbstractDocumentQuery) _distinct() {
	panicIf(q.isDistinct(), "The is already a distinct query")
	//throw new IllegalStateException("The is already a distinct query");

	q.selectTokens = append(q.selectTokens, DistinctToken_INSTANCE)
}

func (q *AbstractDocumentQuery) updateStatsAndHighlightings(queryResult *QueryResult) {
	q.queryStats.updateQueryStats(queryResult)
	//TBD 4.1 Highlightings.Update(queryResult);
}

func (q *AbstractDocumentQuery) buildSelect(writer *StringBuilder) {
	if len(q.selectTokens) == 0 {
		return
	}

	writer.append(" select ")

	if len(q.selectTokens) == 1 {
		tok := q.selectTokens[0]
		if dtok, ok := tok.(*DistinctToken); ok {
			dtok.writeTo(writer)
			writer.append(" *")
			return
		}
	}

	for i, token := range q.selectTokens {
		if i > 0 {
			prevToken := q.selectTokens[i-1]
			if _, ok := prevToken.(*DistinctToken); !ok {
				writer.append(",")
			}
		}

		var prevToken QueryToken
		if i > 0 {
			prevToken = q.selectTokens[i-1]
		}
		DocumentQueryHelper_addSpaceIfNeeded(prevToken, token, writer)

		token.writeTo(writer)
	}
}

func (q *AbstractDocumentQuery) buildFrom(writer *StringBuilder) {
	q.fromToken.writeTo(writer)
}

func (q *AbstractDocumentQuery) buildDeclare(writer *StringBuilder) {
	if q.declareToken != nil {
		q.declareToken.writeTo(writer)
	}
}

func (q *AbstractDocumentQuery) buildLoad(writer *StringBuilder) {
	if len(q.loadTokens) == 0 {
		return
	}

	writer.append(" load ")

	for i, tok := range q.loadTokens {
		if i != 0 {
			writer.append(", ")
		}

		tok.writeTo(writer)
	}
}

func (q *AbstractDocumentQuery) buildWhere(writer *StringBuilder) {
	if len(q.whereTokens) == 0 {
		return
	}

	writer.append(" where ")

	if q.isIntersect {
		writer.append("intersect(")
	}

	for i, tok := range q.whereTokens {
		var prevToken QueryToken
		if i > 0 {
			prevToken = q.whereTokens[i-1]
		}
		DocumentQueryHelper_addSpaceIfNeeded(prevToken, tok, writer)
		tok.writeTo(writer)
	}

	if q.isIntersect {
		writer.append(") ")
	}
}

func (q *AbstractDocumentQuery) buildGroupBy(writer *StringBuilder) {
	if len(q.groupByTokens) == 0 {
		return
	}

	writer.append(" group by ")

	for i, token := range q.groupByTokens {
		if i > 0 {
			writer.append(", ")
		}
		token.writeTo(writer)
	}
}

func (q *AbstractDocumentQuery) buildOrderBy(writer *StringBuilder) {
	if len(q.orderByTokens) == 0 {
		return
	}

	writer.append(" order by ")

	for i, token := range q.orderByTokens {
		if i > 0 {
			writer.append(", ")
		}

		token.writeTo(writer)
	}
}

func (q *AbstractDocumentQuery) appendOperatorIfNeeded(tokensRef *[]QueryToken) {
	tokens := *tokensRef
	q.assertNoRawQuery()

	n := len(tokens)
	if len(tokens) == 0 {
		return
	}

	lastToken := tokens[n-1]
	_, isWhereToken := lastToken.(*WhereToken)
	_, isCloseSubclauseToken := lastToken.(*CloseSubclauseToken)
	if !isWhereToken && !isCloseSubclauseToken {
		return
	}

	var lastWhere *WhereToken

	for i := n - 1; i >= 0; i-- {
		tok := tokens[i]
		if maybeLastWhere, ok := tok.(*WhereToken); ok {
			lastWhere = maybeLastWhere
			break
		}
	}

	var token *QueryOperatorToken
	if q.defaultOperator == QueryOperator_AND {
		token = QueryOperatorToken_AND
	} else {
		token = QueryOperatorToken_OR
	}

	if lastWhere != nil && lastWhere.getOptions().getSearchOperator() != SearchOperator_UNSET {
		token = QueryOperatorToken_OR // default to OR operator after search if AND was not specified explicitly
	}

	tokens = append(tokens, token)
	*tokensRef = tokens
}

func (q *AbstractDocumentQuery) transformCollection(fieldName string, values []Object) []Object {
	var result []Object
	for _, value := range values {
		if collectionValue, ok := value.([]Object); ok {
			tmp := q.transformCollection(fieldName, collectionValue)
			result = append(result, tmp...)
		} else {
			nestedWhereParams := NewWhereParams()
			nestedWhereParams.setAllowWildcards(true)
			nestedWhereParams.setFieldName(fieldName)
			nestedWhereParams.setValue(value)
			tmp := q.transformValue(nestedWhereParams)
			result = append(result, tmp)
		}
	}
	return result
}

func (q *AbstractDocumentQuery) negateIfNeeded(tokensRef *[]QueryToken, fieldName string) {
	if !q.negate {
		return
	}

	q.negate = false

	tokens := *tokensRef

	n := len(tokens)
	isOpenSubclauseToken := false
	if n > 0 {
		_, isOpenSubclauseToken = tokens[n-1].(*OpenSubclauseToken)
	}
	if n == 0 || isOpenSubclauseToken {
		if fieldName != "" {
			q._whereExists(fieldName)
		} else {
			q._whereTrue()
		}
		q._andAlso()
	}

	tokens = append(tokens, NegateToken_INSTANCE)
	*tokensRef = tokens
}

func AbstractDocumentQuery_unpackCollection(items []Object) []Object {
	var results []Object

	for _, item := range items {
		if itemCollection, ok := item.([]Object); ok {
			els := AbstractDocumentQuery_unpackCollection(itemCollection)
			results = append(results, els...)
		} else {
			results = append(results, item)
		}
	}

	return results
}

func (q *AbstractDocumentQuery) ensureValidFieldName(fieldName string, isNestedPath bool) string {
	if q.theSession == nil || q.theSession.getConventions() == nil || isNestedPath || q.isGroupBy {
		return QueryFieldUtil_escapeIfNecessary(fieldName)
	}

	if fieldName == DocumentConventions_identityPropertyName {
		return Constants_Documents_Indexing_Fields_DOCUMENT_ID_FIELD_NAME
	}

	return QueryFieldUtil_escapeIfNecessary(fieldName)
}

func (q *AbstractDocumentQuery) transformValue(whereParams *WhereParams) Object {
	return q.transformValueWithRange(whereParams, false)
}

func (q *AbstractDocumentQuery) transformValueWithRange(whereParams *WhereParams, forRange bool) Object {
	if whereParams.getValue() == nil {
		return nil
	}

	if "" == whereParams.getValue() {
		return ""
	}

	var stringValueReference string
	if q._conventions.tryConvertValueForQuery(whereParams.getFieldName(), whereParams.getValue(), forRange, &stringValueReference) {
		return stringValueReference
	}

	// TODO: this could be a type switch
	val := whereParams.getValue()
	switch val.(type) {
	case time.Time, string, int, int32, int64, float32, float64, bool:
		return val
	case time.Duration:
		panicIf(true, "NYI")
		//return ((Duration) whereParams.getValue()).toNanos() / 100;
	}
	return whereParams.getValue()
}

func (q *AbstractDocumentQuery) addQueryParameter(value Object) string {
	parameterName := "p" + strconv.Itoa(len(q.queryParameters))
	q.queryParameters[parameterName] = value
	return parameterName
}

func (q *AbstractDocumentQuery) getCurrentWhereTokens() []QueryToken {
	if !q._isInMoreLikeThis {
		return q.whereTokens
	}

	n := len(q.whereTokens)

	if n == 0 {
		// throw new IllegalStateException("Cannot get MoreLikeThisToken because there are no where token specified.");
		panicIf(true, "Cannot get MoreLikeThisToken because there are no where token specified.")
	}

	lastToken := q.whereTokens[n-1]

	if moreLikeThisToken, ok := lastToken.(*MoreLikeThisToken); ok {
		return moreLikeThisToken.whereTokens
	} else {
		//throw new IllegalStateException("Last token is not MoreLikeThisToken");
		panicIf(true, "Last token is not MoreLikeThisToken")
	}
	return nil
}

func (q *AbstractDocumentQuery) getCurrentWhereTokensRef() *[]QueryToken {
	if !q._isInMoreLikeThis {
		return &q.whereTokens
	}

	n := len(q.whereTokens)

	if n == 0 {
		// throw new IllegalStateException("Cannot get MoreLikeThisToken because there are no where token specified.");
		panicIf(true, "Cannot get MoreLikeThisToken because there are no where token specified.")
	}

	lastToken := q.whereTokens[n-1]

	if moreLikeThisToken, ok := lastToken.(*MoreLikeThisToken); ok {
		return &moreLikeThisToken.whereTokens
	} else {
		//throw new IllegalStateException("Last token is not MoreLikeThisToken");
		panicIf(true, "Last token is not MoreLikeThisToken")
	}
	return nil
}

func (q *AbstractDocumentQuery) updateFieldsToFetchToken(fieldsToFetch *FieldsToFetchToken) {
	q.fieldsToFetchToken = fieldsToFetch

	if len(q.selectTokens) == 0 {
		q.selectTokens = append(q.selectTokens, fieldsToFetch)
	} else {
		for _, x := range q.selectTokens {
			if _, ok := x.(*FieldsToFetchToken); ok {
				for idx, tok := range q.selectTokens {
					if tok == x {
						q.selectTokens[idx] = fieldsToFetch
					}
				}
				return
			}
		}
		q.selectTokens = append(q.selectTokens, fieldsToFetch)
	}
}

func (q *AbstractDocumentQuery) getQueryOperation() *QueryOperation {
	return q.queryOperation
}

func (q *AbstractDocumentQuery) _addBeforeQueryExecutedListener(action func(*IndexQuery)) {
	q.beforeQueryExecutedCallback = append(q.beforeQueryExecutedCallback, action)
}

func (q *AbstractDocumentQuery) _removeBeforeQueryExecutedListener(action func(*IndexQuery)) {
	panicIf(true, "NYI")
	// TODO: implement me
	// beforeQueryExecutedCallback.remove(action)
}

func (q *AbstractDocumentQuery) _addAfterQueryExecutedListener(action func(*QueryResult)) {
	q.afterQueryExecutedCallback = append(q.afterQueryExecutedCallback, action)
}

func (q *AbstractDocumentQuery) _removeAfterQueryExecutedListener(action func(*QueryResult)) {
	panicIf(true, "NYI")
	// TODO: implement me
	// afterQueryExecutedCallback.remove(action)
}

func (q *AbstractDocumentQuery) _addAfterStreamExecutedListener(action func(ObjectNode)) {
	q.afterStreamExecutedCallback = append(q.afterStreamExecutedCallback, action)
}

func (q *AbstractDocumentQuery) _removeAfterStreamExecutedListener(action func(ObjectNode)) {
	panicIf(true, "NYI")
	// TODO: implement me
	// afterStreamExecutedCallback.remove(action)
}

func (q *AbstractDocumentQuery) _noTracking() {
	q.disableEntitiesTracking = true
}

func (q *AbstractDocumentQuery) _noCaching() {
	q.disableCaching = true
}

func (q *AbstractDocumentQuery) _withinRadiusOf(fieldName string, radius float64, latitude float64, longitude float64, radiusUnits SpatialUnits, distErrorPercent float64) {
	fieldName = q.ensureValidFieldName(fieldName, false)

	tokensRef := q.getCurrentWhereTokensRef()
	q.appendOperatorIfNeeded(tokensRef)
	q.negateIfNeeded(tokensRef, fieldName)

	shape := ShapeToken_circle(q.addQueryParameter(radius), q.addQueryParameter(latitude), q.addQueryParameter(longitude), radiusUnits)
	opts := NewWhereOptionsWithTokenAndDistance(shape, distErrorPercent)
	whereToken := WhereToken_createWithOptions(WhereOperator_SPATIAL_WITHIN, fieldName, "", opts)

	tokens := *tokensRef
	tokens = append(tokens, whereToken)
	*tokensRef = tokens
}

func (q *AbstractDocumentQuery) _spatial(fieldName string, shapeWkt string, relation SpatialRelation, distErrorPercent float64) {
	fieldName = q.ensureValidFieldName(fieldName, false)

	tokensRef := q.getCurrentWhereTokensRef()
	q.appendOperatorIfNeeded(tokensRef)
	q.negateIfNeeded(tokensRef, fieldName)

	wktToken := ShapeToken_wkt(q.addQueryParameter(shapeWkt))

	var whereOperator WhereOperator
	switch relation {
	case SpatialRelation_WITHIN:
		whereOperator = WhereOperator_SPATIAL_WITHIN
	case SpatialRelation_CONTAINS:
		whereOperator = WhereOperator_SPATIAL_CONTAINS
	case SpatialRelation_DISJOINT:
		whereOperator = WhereOperator_SPATIAL_DISJOINT
	case SpatialRelation_INTERSECTS:
		whereOperator = WhereOperator_SPATIAL_INTERSECTS
	default:
		//throw new IllegalArgumentException();
		panicIf(true, "unknown relation %s", relation)
	}

	tokens := *tokensRef
	opts := NewWhereOptionsWithTokenAndDistance(wktToken, distErrorPercent)
	tok := WhereToken_createWithOptions(whereOperator, fieldName, "", opts)
	tokens = append(tokens, tok)
	*tokensRef = tokens
}

/*

func (q *AbstractDocumentQuery) _spatial2(dynamicField DynamicSpatialField, criteria SpatialCriteria) {
	tokensRef := q.getCurrentWhereTokensRef()
	q.appendOperatorIfNeeded(tokensRef)
	q.negateIfNeeded(tokensRef, "")

		tok := criteria.toQueryToken(dynamicField.toField(this::ensureValidFieldName), this::addQueryParameter)
		tokens := *tokensRef
		tokens = append(tokens, tok)
		*tokensRef = tokens
}
*/

/*
     _spatial(string fieldName, SpatialCriteria criteria) {
       fieldName = ensureValidFieldName(fieldName, false);

       List<QueryToken> tokens = getCurrentWhereTokens();
       appendOperatorIfNeeded(tokens);
       negateIfNeeded(tokens, fieldName);

       tokens.add(criteria.toQueryToken(fieldName, this::addQueryParameter));
   }

     _orderByDistance(DynamicSpatialField field, float64 latitude, float64 longitude) {
       if (field == null) {
           throw new IllegalArgumentException("Field cannot be null");
       }
       _orderByDistance("'" + field.toField(this::ensureValidFieldName) + "'", latitude, longitude);
   }

     _orderByDistance(string fieldName, float64 latitude, float64 longitude) {
       orderByTokens.add(OrderByToken.createDistanceAscending(fieldName, addQueryParameter(latitude), addQueryParameter(longitude)));
   }

     _orderByDistance(DynamicSpatialField field, string shapeWkt) {
       if (field == null) {
           throw new IllegalArgumentException("Field cannot be null");
       }
       _orderByDistance("'" + field.toField(this::ensureValidFieldName) + "'", shapeWkt);
   }

     _orderByDistance(string fieldName, string shapeWkt) {
       orderByTokens.add(OrderByToken.createDistanceAscending(fieldName, addQueryParameter(shapeWkt)));
   }

     _orderByDistanceDescending(DynamicSpatialField field, float64 latitude, float64 longitude) {
       if (field == null) {
           throw new IllegalArgumentException("Field cannot be null");
       }
       _orderByDistanceDescending("'" + field.toField(this::ensureValidFieldName) + "'", latitude, longitude);
   }

     _orderByDistanceDescending(string fieldName, float64 latitude, float64 longitude) {
       orderByTokens.add(OrderByToken.createDistanceDescending(fieldName, addQueryParameter(latitude), addQueryParameter(longitude)));
   }

     _orderByDistanceDescending(DynamicSpatialField field, string shapeWkt) {
       if (field == null) {
           throw new IllegalArgumentException("Field cannot be null");
       }
       _orderByDistanceDescending("'" + field.toField(this::ensureValidFieldName) + "'", shapeWkt);
   }

     _orderByDistanceDescending(string fieldName, string shapeWkt) {
       orderByTokens.add(OrderByToken.createDistanceDescending(fieldName, addQueryParameter(shapeWkt)));
   }

   protected  initSync() {
       if (queryOperation != null) {
           return;
       }

       BeforeQueryEventArgs beforeQueryEventArgs = new BeforeQueryEventArgs(theSession, new DocumentQueryCustomizationDelegate(this));
       theSession.onBeforeQueryInvoke(beforeQueryEventArgs);

       queryOperation = initializeQueryOperation();
       executeActualQuery();
   }

     executeActualQuery() {
       try (CleanCloseable context = queryOperation.enterQueryContext()) {
           QueryCommand command = queryOperation.createRequest();
           theSession.getRequestExecutor().execute(command, theSession.sessionInfo);
           queryOperation.setResult(command.getResult());
       }
       invokeAfterQueryExecuted(queryOperation.getCurrentQueryResults());
   }

    Iterator<T> iterator() {
       return executeQueryOperation(null).iterator();
   }

    List<T> toList() {
       return EnumerableUtils.toList(iterator());
   }

    QueryResult getQueryResult() {
       initSync();

       return queryOperation.getCurrentQueryResults().createSnapshot();
   }

    T first() {
       Collection<T> result = executeQueryOperation(1);
       return result.isEmpty() ? null : result.stream().findFirst().get();
   }

    T firstOrDefault() {
       Collection<T> result = executeQueryOperation(1);
       return result.stream().findFirst().orElseGet(() -> Defaults.defaultValue(clazz));
   }

    T single() {
       Collection<T> result = executeQueryOperation(2);
       if (result.size() > 1) {
           throw new IllegalStateException("Expected single result, got: " + result.size());
       }
       return result.stream().findFirst().orElse(null);
   }

    T singleOrDefault() {
       Collection<T> result = executeQueryOperation(2);
       if (result.size() > 1) {
           throw new IllegalStateException("Expected single result, got: " + result.size());
       }
       if (result.isEmpty()) {
           return Defaults.defaultValue(clazz);
       }
       return result.stream().findFirst().get();
   }

    int count() {
       _take(0);
       QueryResult queryResult = getQueryResult();
       return queryResult.getTotalResults();
   }

    bool any() {
       if (isDistinct()) {
           // for distinct it is cheaper to do count 1
           return executeQueryOperation(1).iterator().hasNext();
       }

       _take(0);
       QueryResult queryResult = getQueryResult();
       return queryResult.getTotalResults() > 0;
   }

    Collection<T> executeQueryOperation(Integer take) {
       if (take != null && (pageSize == null || pageSize > take)) {
           _take(take);
       }

       initSync();

       return queryOperation.complete(clazz);
   }

     _aggregateBy(FacetBase facet) {
       for (QueryToken token : selectTokens) {
           if (token instanceof FacetToken) {
               continue;
           }

           throw new IllegalStateException("Aggregation query can select only facets while it got " + token.getClass().getSimpleName() + " token");
       }

       selectTokens.add(FacetToken.create(facet, this::addQueryParameter));
   }

     _aggregateUsing(string facetSetupDocumentId) {
       selectTokens.add(FacetToken.create(facetSetupDocumentId));
   }

    Lazy<List<T>> lazily() {
       return lazily(null);
   }

    Lazy<List<T>> lazily(Consumer<List<T>> onEval) {
       if (getQueryOperation() == null) {
           queryOperation = initializeQueryOperation();
       }

       LazyQueryOperation<T> lazyQueryOperation = new LazyQueryOperation<>(clazz, theSession.getConventions(), queryOperation, afterQueryExecutedCallback);
       return ((DocumentSession)theSession).addLazyOperation((Class<List<T>>) (Class<?>)List.class, lazyQueryOperation, onEval);
   }

    Lazy<Integer> countLazily() {
       if (queryOperation == null) {
           _take(0);
           queryOperation = initializeQueryOperation();
       }

       LazyQueryOperation<T> lazyQueryOperation = new LazyQueryOperation<T>(clazz, theSession.getConventions(), queryOperation, afterQueryExecutedCallback);
       return ((DocumentSession)theSession).addLazyCountOperation(lazyQueryOperation);
   }

     _suggestUsing(SuggestionBase suggestion) {
       if (suggestion == null) {
           throw new IllegalArgumentException("suggestion cannot be null");
       }

       assertCanSuggest();

       SuggestToken token;

       if (suggestion instanceof SuggestionWithTerm) {
           SuggestionWithTerm term = (SuggestionWithTerm) suggestion;
           token = SuggestToken.create(term.getField(), addQueryParameter(term.getTerm()), getOptionsParameterName(term.getOptions()));
       } else if (suggestion instanceof SuggestionWithTerms) {
           SuggestionWithTerms terms = (SuggestionWithTerms) suggestion;
           token = SuggestToken.create(terms.getField(), addQueryParameter(terms.getTerms()), getOptionsParameterName(terms.getOptions()));
       } else {
           throw new UnsupportedOperationException("Unknown type of suggestion: " + suggestion.getClass());
       }

       selectTokens.add(token);
   }
*/

func (q *AbstractDocumentQuery) getOptionsParameterName(options *SuggestionOptions) string {
	optionsParameterName := ""
	if options != nil && options != SuggestionOptions_defaultOptions {
		optionsParameterName = q.addQueryParameter(options)
	}

	return optionsParameterName
}

func (q *AbstractDocumentQuery) assertCanSuggest() {
	if len(q.whereTokens) > 0 {
		//throw new IllegalStateException("Cannot add suggest when WHERE statements are present.");
		panicIf(true, "Cannot add suggest when WHERE statements are present.")
	}

	if len(q.selectTokens) > 0 {
		//throw new IllegalStateException("Cannot add suggest when SELECT statements are present.");
		panicIf(true, "Cannot add suggest when SELECT statements are present.")
	}

	if len(q.orderByTokens) > 0 {
		//throw new IllegalStateException("Cannot add suggest when ORDER BY statements are present.");
		panicIf(true, "Cannot add suggest when ORDER BY statements are present.")
	}
}
