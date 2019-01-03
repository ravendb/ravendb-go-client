package ravendb

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"
	"unicode"
)

// Note: IAbstractDocumentQuery is AbstractDocumentQuery

// AbstractDocumentQuery is a base class for describing a query
type AbstractDocumentQuery struct {
	clazz                    reflect.Type
	_aliasToGroupByFieldName map[string]string
	defaultOperator          QueryOperator

	// Note: rootTypes is not used in Go because we only have one ID property

	negate              bool
	indexName           string
	collectionName      string
	_currentClauseDepth int
	queryRaw            string
	queryParameters     Parameters

	isIntersect bool
	isGroupBy   bool

	theSession *InMemoryDocumentSessionOperations

	pageSize *int

	selectTokens       []queryToken
	fromToken          *fromToken
	declareToken       *declareToken
	loadTokens         []*loadToken
	fieldsToFetchToken *fieldsToFetchToken

	whereTokens   []queryToken
	groupByTokens []queryToken
	orderByTokens []queryToken

	start        int
	_conventions *DocumentConventions

	timeout time.Duration

	theWaitForNonStaleResults bool

	includes []string

	queryStats *QueryStatistics

	disableEntitiesTracking bool

	disableCaching bool

	_isInMoreLikeThis bool

	// Go doesn't allow comparing functions so to remove we use index returned
	// by add() function. We maintain stable index by never shrinking
	// callback arrays. We assume there is no high churn of adding/removing
	// callbacks
	beforeQueryExecutedCallback []func(*IndexQuery)
	afterQueryExecutedCallback  []func(*QueryResult)
	afterStreamExecutedCallback []func(ObjectNode)

	queryOperation *QueryOperation

	// SelectFields logic has to be delayed until ToList
	// because only then we know the type of the result
	selectFieldsArgs *QueryData
}

func (q *AbstractDocumentQuery) isDistinct() bool {
	if len(q.selectTokens) == 0 {
		return false
	}
	_, ok := q.selectTokens[0].(*distinctToken)
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

func getQueryDefaultTimeout() time.Duration {
	return time.Second * 15
}

func NewAbstractDocumentQueryOld(clazz reflect.Type, session *InMemoryDocumentSessionOperations, indexName string, collectionName string, isGroupBy bool, declareToken *declareToken, loadTokens []*loadToken, fromAlias string) *AbstractDocumentQuery {
	res := &AbstractDocumentQuery{
		clazz:                    clazz,
		defaultOperator:          QueryOperatorAnd,
		isGroupBy:                isGroupBy,
		indexName:                indexName,
		collectionName:           collectionName,
		declareToken:             declareToken,
		loadTokens:               loadTokens,
		theSession:               session,
		_aliasToGroupByFieldName: make(map[string]string),
		queryParameters:          make(map[string]interface{}),
		queryStats:               NewQueryStatistics(),
	}
	res.fromToken = createFromToken(indexName, collectionName, fromAlias)
	f := func(queryResult *QueryResult) {
		res.UpdateStatsAndHighlightings(queryResult)
	}
	res._addAfterQueryExecutedListener(f)
	if session == nil {
		res._conventions = NewDocumentConventions()
	} else {
		res._conventions = session.GetConventions()
	}
	return res
}

// NewAbstractDocumentQuery returns new AbstractDocumentQuery
func NewAbstractDocumentQuery(session *InMemoryDocumentSessionOperations, indexName string, collectionName string, isGroupBy bool, declareToken *declareToken, loadTokens []*loadToken, fromAlias string) *AbstractDocumentQuery {
	res := &AbstractDocumentQuery{
		defaultOperator:          QueryOperatorAnd,
		isGroupBy:                isGroupBy,
		indexName:                indexName,
		collectionName:           collectionName,
		declareToken:             declareToken,
		loadTokens:               loadTokens,
		theSession:               session,
		_aliasToGroupByFieldName: make(map[string]string),
		queryParameters:          make(map[string]interface{}),
		queryStats:               NewQueryStatistics(),
	}
	// if those are not provided, we delay creating fromToken
	// until ToList()
	if indexName != "" || collectionName != "" || fromAlias != "" {
		res.fromToken = createFromToken(indexName, collectionName, fromAlias)
	}
	f := func(queryResult *QueryResult) {
		res.UpdateStatsAndHighlightings(queryResult)
	}
	res._addAfterQueryExecutedListener(f)
	if session == nil {
		res._conventions = NewDocumentConventions()
	} else {
		res._conventions = session.GetConventions()
	}
	return res
}

func (q *AbstractDocumentQuery) _usingDefaultOperator(operator QueryOperator) {
	if len(q.whereTokens) > 0 {
		//throw new IllegalStateError("Default operator can only be set before any where clause is added.");
		panicIf(true, "Default operator can only be set before any where clause is added.")
	}

	q.defaultOperator = operator
}

func (q *AbstractDocumentQuery) _waitForNonStaleResults(waitTimeout time.Duration) {
	q.theWaitForNonStaleResults = true
	if waitTimeout == 0 {
		waitTimeout = getQueryDefaultTimeout()
	}
	q.timeout = waitTimeout
}

func (q *AbstractDocumentQuery) initializeQueryOperation() *QueryOperation {
	indexQuery := q.GetIndexQuery()

	return NewQueryOperation(q.theSession, q.indexName, indexQuery, q.fieldsToFetchToken, q.disableEntitiesTracking, false, false)
}

func (q *AbstractDocumentQuery) GetIndexQuery() *IndexQuery {
	query := q.String()
	indexQuery := q.GenerateIndexQuery(query)
	q.InvokeBeforeQueryExecuted(indexQuery)
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
	q.orderByTokens = append(q.orderByTokens, orderByTokenRandom)
}

func (q *AbstractDocumentQuery) _randomOrderingWithSeed(seed string) {
	q.assertNoRawQuery()

	if stringIsBlank(seed) {
		q._randomOrdering()
		return
	}

	q.orderByTokens = append(q.orderByTokens, orderByTokenCreateRandom(seed))
}

func (q *AbstractDocumentQuery) AddGroupByAlias(fieldName string, projectedName string) {
	q._aliasToGroupByFieldName[projectedName] = fieldName
}

func (q *AbstractDocumentQuery) assertNoRawQuery() {
	panicIf(q.queryRaw != "", "RawQuery was called, cannot modify this query by calling on operations that would modify the query (such as Where, Select, OrderBy, GroupBy, etc)")
}

func (q *AbstractDocumentQuery) _addParameter(name string, value interface{}) {
	name = strings.TrimPrefix(name, "$")
	if _, ok := q.queryParameters[name]; ok {
		// throw new IllegalStateError("The parameter " + name + " was already added");
		panicIf(true, "The parameter "+name+" was already added")
	}

	q.queryParameters[name] = value
}

func (q *AbstractDocumentQuery) _groupBy(fieldName string, fieldNames ...string) {
	var mapping []*GroupBy
	for _, x := range fieldNames {
		el := NewGroupByField(x)
		mapping = append(mapping, el)
	}
	q._groupBy2(NewGroupByField(fieldName), mapping...)
}

// TODO: better name
func (q *AbstractDocumentQuery) _groupBy2(field *GroupBy, fields ...*GroupBy) {
	// TODO: if q.fromToken is nil, needs to do this check in ToList()
	if q.fromToken != nil && !q.fromToken.isDynamic {
		//throw new IllegalStateError("groupBy only works with dynamic queries");
		panicIf(true, "groupBy only works with dynamic queries")
	}

	q.assertNoRawQuery()
	q.isGroupBy = true

	fieldName := q.ensureValidFieldName(field.Field, false)

	q.groupByTokens = append(q.groupByTokens, GroupByToken_createWithMethod(fieldName, field.Method))

	if len(fields) == 0 {
		return
	}

	for _, item := range fields {
		fieldName = q.ensureValidFieldName(item.Field, false)
		q.groupByTokens = append(q.groupByTokens, GroupByToken_createWithMethod(fieldName, item.Method))
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

	t := &groupByCountToken{
		fieldName: projectedName,
	}
	q.selectTokens = append(q.selectTokens, t)
}

func (q *AbstractDocumentQuery) _whereTrue() {
	tokensRef := q.getCurrentWhereTokensRef()
	q.appendOperatorIfNeeded(tokensRef)
	q.negateIfNeeded(tokensRef, "")

	tokens := *tokensRef
	tokens = append(tokens, trueTokenInstance)
	*tokensRef = tokens
}

func (q *AbstractDocumentQuery) _moreLikeThis() *MoreLikeThisScope {
	q.appendOperatorIfNeeded(&q.whereTokens)

	token := newMoreLikeThisToken()
	q.whereTokens = append(q.whereTokens, token)

	q._isInMoreLikeThis = true
	add := func(o interface{}) string {
		return q.addQueryParameter(o)
	}
	onDispose := func() {
		q._isInMoreLikeThis = false
	}
	return NewMoreLikeThisScope(token, add, onDispose)
}

func (q *AbstractDocumentQuery) _include(path string) {
	q.includes = append(q.includes, path)
}

// TODO: see if count can be int
func (q *AbstractDocumentQuery) _take(count *int) {
	q.pageSize = count
}

func (q *AbstractDocumentQuery) _skip(count int) {
	q.start = count
}

func (q *AbstractDocumentQuery) _whereLucene(fieldName string, whereClause string) {
	fieldName = q.ensureValidFieldName(fieldName, false)

	tokensRef := q.getCurrentWhereTokensRef()
	tokens := *tokensRef
	q.appendOperatorIfNeeded(tokensRef)
	q.negateIfNeeded(tokensRef, fieldName)

	whereToken := createWhereTokenWithOptions(WhereOperator_LUCENE, fieldName, q.addQueryParameter(whereClause), nil)
	tokens = append(tokens, whereToken)
	*tokensRef = tokens
}

func (q *AbstractDocumentQuery) _openSubclause() {
	q._currentClauseDepth++

	tokensRef := q.getCurrentWhereTokensRef()
	q.appendOperatorIfNeeded(tokensRef)
	q.negateIfNeeded(tokensRef, "")

	tokens := *tokensRef
	tokens = append(tokens, openSubclauseTokenInstance)
	*tokensRef = tokens
}

func (q *AbstractDocumentQuery) _closeSubclause() {
	q._currentClauseDepth--

	tokensRef := q.getCurrentWhereTokensRef()
	tokens := *tokensRef
	tokens = append(tokens, closeSubclauseTokenInstance)
	*tokensRef = tokens
}

func (q *AbstractDocumentQuery) _whereEquals(fieldName string, value interface{}) {
	params := &whereParams{
		fieldName: fieldName,
		value:     value,
	}
	q._whereEqualsWithParams(params)
}

func (q *AbstractDocumentQuery) _whereEqualsWithMethodCall(fieldName string, method MethodCall) {
	q._whereEquals(fieldName, method)
}

func (q *AbstractDocumentQuery) _whereEqualsWithParams(whereParams *whereParams) {
	if q.negate {
		q.negate = false
		q._whereNotEqualsWithParams(whereParams)
		return
	}

	whereParams.fieldName = q.ensureValidFieldName(whereParams.fieldName, whereParams.isNestedPath)

	tokensRef := q.getCurrentWhereTokensRef()
	q.appendOperatorIfNeeded(tokensRef)

	if q.ifValueIsMethod(WhereOperator_EQUALS, whereParams, tokensRef) {
		return
	}

	transformToEqualValue := q.transformValue(whereParams)
	addQueryParameter := q.addQueryParameter(transformToEqualValue)
	whereToken := createWhereTokenWithOptions(WhereOperator_EQUALS, whereParams.fieldName, addQueryParameter, NewWhereOptionsWithExact(whereParams.isExact))

	tokens := *tokensRef
	tokens = append(tokens, whereToken)
	*tokensRef = tokens
}

func (q *AbstractDocumentQuery) ifValueIsMethod(op WhereOperator, whereParams *whereParams, tokensRef *[]queryToken) bool {
	if mc, ok := whereParams.value.(*CmpXchg); ok {
		n := len(mc.args)
		args := make([]string, n)
		for i := 0; i < n; i++ {
			args[i] = q.addQueryParameter(mc.args[i])
		}

		opts := NewWhereOptionsWithMethod(MethodsTypeCmpXChg, args, mc.accessPath, whereParams.isExact)
		token := createWhereTokenWithOptions(op, whereParams.fieldName, "", opts)

		tokens := *tokensRef
		tokens = append(tokens, token)
		*tokensRef = tokens
		return true
	}

	// add more if there are more types that "derive" from MethodCall
	// (by embedding MethodCallData)

	return false
}

func (q *AbstractDocumentQuery) _whereNotEquals(fieldName string, value interface{}) {
	params := &whereParams{
		fieldName: fieldName,
		value:     value,
	}

	q._whereNotEqualsWithParams(params)
}

func (q *AbstractDocumentQuery) _whereNotEqualsWithMethod(fieldName string, method MethodCall) {
	q._whereNotEquals(fieldName, method)
}

func (q *AbstractDocumentQuery) _whereNotEqualsWithParams(whereParams *whereParams) {
	if q.negate {
		q.negate = false
		q._whereEqualsWithParams(whereParams)
		return
	}

	transformToEqualValue := q.transformValue(whereParams)

	tokensRef := q.getCurrentWhereTokensRef()
	q.appendOperatorIfNeeded(tokensRef)

	whereParams.fieldName = q.ensureValidFieldName(whereParams.fieldName, whereParams.isNestedPath)

	if q.ifValueIsMethod(WhereOperator_NOT_EQUALS, whereParams, tokensRef) {
		return
	}

	whereToken := createWhereTokenWithOptions(WhereOperator_NOT_EQUALS, whereParams.fieldName, q.addQueryParameter(transformToEqualValue), NewWhereOptionsWithExact(whereParams.isExact))
	tokens := *tokensRef
	tokens = append(tokens, whereToken)
	*tokensRef = tokens
}

func (q *AbstractDocumentQuery) NegateNext() {
	q.negate = !q.negate
}

// mark last created token as exact. only applies to select number of tokens.
// it allows fluid APIs like .Where().Exact()
// will panic if last token wasn't of compatible type as that is considered
// invalid use of API and returning an error would break fluid API
func (q *AbstractDocumentQuery) markLastTokenExact() {
	tokensRef := q.getCurrentWhereTokensRef()
	tokens := *tokensRef
	n := len(tokens)
	lastToken := tokens[n-1]
	switch tok := lastToken.(type) {
	case *whereToken:
		if tok.options == nil {
			tok.options = NewWhereOptionsWithExact(true)
		} else {
			tok.options.exact = true
		}
	default:
		panicIf(true, "expected whereToken, got %T", lastToken)
	}

	*tokensRef = tokens
}

func (q *AbstractDocumentQuery) _whereIn(fieldName string, values []interface{}) {
	fieldName = q.ensureValidFieldName(fieldName, false)

	tokensRef := q.getCurrentWhereTokensRef()
	q.appendOperatorIfNeeded(tokensRef)
	q.negateIfNeeded(tokensRef, fieldName)

	whereToken := createWhereToken(WhereOperator_IN, fieldName, q.addQueryParameter(q.transformCollection(fieldName, AbstractDocumentQuery_unpackCollection(values))))

	tokens := *tokensRef
	tokens = append(tokens, whereToken)
	*tokensRef = tokens
}

func (q *AbstractDocumentQuery) _whereStartsWith(fieldName string, value interface{}) {
	whereParams := &whereParams{
		fieldName:      fieldName,
		value:          value,
		allowWildcards: true,
	}

	transformToEqualValue := q.transformValue(whereParams)

	tokensRef := q.getCurrentWhereTokensRef()
	q.appendOperatorIfNeeded(tokensRef)

	whereParams.fieldName = q.ensureValidFieldName(whereParams.fieldName, whereParams.isNestedPath)
	q.negateIfNeeded(tokensRef, whereParams.fieldName)

	whereToken := createWhereToken(WhereOperator_STARTS_WITH, whereParams.fieldName, q.addQueryParameter(transformToEqualValue))

	tokens := *tokensRef
	tokens = append(tokens, whereToken)
	*tokensRef = tokens
}

func (q *AbstractDocumentQuery) _whereEndsWith(fieldName string, value interface{}) {
	whereParams := &whereParams{
		fieldName:      fieldName,
		value:          value,
		allowWildcards: true,
	}

	transformToEqualValue := q.transformValue(whereParams)

	tokensRef := q.getCurrentWhereTokensRef()
	q.appendOperatorIfNeeded(tokensRef)

	whereParams.fieldName = q.ensureValidFieldName(whereParams.fieldName, whereParams.isNestedPath)
	q.negateIfNeeded(tokensRef, whereParams.fieldName)

	whereToken := createWhereToken(WhereOperator_ENDS_WITH, whereParams.fieldName, q.addQueryParameter(transformToEqualValue))

	tokens := *tokensRef
	tokens = append(tokens, whereToken)
	*tokensRef = tokens
}

func (q *AbstractDocumentQuery) _whereBetween(fieldName string, start interface{}, end interface{}) {
	fieldName = q.ensureValidFieldName(fieldName, false)

	tokensRef := q.getCurrentWhereTokensRef()
	q.appendOperatorIfNeeded(tokensRef)
	q.negateIfNeeded(tokensRef, fieldName)

	startParams := &whereParams{
		value:     start,
		fieldName: fieldName,
	}

	endParams := &whereParams{
		value:     end,
		fieldName: fieldName,
	}

	fromParam := interface{}("*")
	if start != nil {
		fromParam = q.transformValueWithRange(startParams, true)
	}
	fromParameterName := q.addQueryParameter(fromParam)

	toParam := interface{}("NULL")
	// TODO: should this be end == nil? A bug in Java code?
	if start != nil {
		toParam = q.transformValueWithRange(endParams, true)
	}
	toParameterName := q.addQueryParameter(toParam)

	whereToken := createWhereTokenWithOptions(WhereOperator_BETWEEN, fieldName, "", NewWhereOptionsWithFromTo(false, fromParameterName, toParameterName))

	tokens := *tokensRef
	tokens = append(tokens, whereToken)
	*tokensRef = tokens
}

func (q *AbstractDocumentQuery) _whereGreaterThan(fieldName string, value interface{}) {
	fieldName = q.ensureValidFieldName(fieldName, false)

	tokensRef := q.getCurrentWhereTokensRef()
	q.appendOperatorIfNeeded(tokensRef)
	q.negateIfNeeded(tokensRef, fieldName)

	whereParams := &whereParams{
		value:     value,
		fieldName: fieldName,
	}

	paramValue := interface{}("*")
	if value != nil {
		paramValue = q.transformValueWithRange(whereParams, true)
	}
	parameter := q.addQueryParameter(paramValue)

	whereToken := createWhereTokenWithOptions(WhereOperator_GREATER_THAN, fieldName, parameter, nil)

	tokens := *tokensRef
	tokens = append(tokens, whereToken)
	*tokensRef = tokens
}

func (q *AbstractDocumentQuery) _whereGreaterThanOrEqual(fieldName string, value interface{}) {
	fieldName = q.ensureValidFieldName(fieldName, false)

	tokensRef := q.getCurrentWhereTokensRef()
	q.appendOperatorIfNeeded(tokensRef)
	q.negateIfNeeded(tokensRef, fieldName)

	whereParams := &whereParams{
		value:     value,
		fieldName: fieldName,
	}

	paramValue := interface{}("*")
	if value != nil {
		paramValue = q.transformValueWithRange(whereParams, true)
	}

	parameter := q.addQueryParameter(paramValue)

	whereToken := createWhereTokenWithOptions(WhereOperator_GREATER_THAN_OR_EQUAL, fieldName, parameter, nil)

	tokens := *tokensRef
	tokens = append(tokens, whereToken)
	*tokensRef = tokens
}

func (q *AbstractDocumentQuery) _whereLessThan(fieldName string, value interface{}) {
	fieldName = q.ensureValidFieldName(fieldName, false)

	tokensRef := q.getCurrentWhereTokensRef()
	q.appendOperatorIfNeeded(tokensRef)
	q.negateIfNeeded(tokensRef, fieldName)

	whereParams := &whereParams{
		value:     value,
		fieldName: fieldName,
	}

	paramValue := interface{}("NULL")
	if value != nil {
		paramValue = q.transformValueWithRange(whereParams, true)
	}
	parameter := q.addQueryParameter(paramValue)
	whereToken := createWhereTokenWithOptions(WhereOperator_LESS_THAN, fieldName, parameter, nil)

	tokens := *tokensRef
	tokens = append(tokens, whereToken)
	*tokensRef = tokens
}

func (q *AbstractDocumentQuery) _whereLessThanOrEqual(fieldName string, value interface{}) {
	tokensRef := q.getCurrentWhereTokensRef()
	q.appendOperatorIfNeeded(tokensRef)
	q.negateIfNeeded(tokensRef, fieldName)

	whereParams := &whereParams{
		value:     value,
		fieldName: fieldName,
	}

	paramValue := interface{}("NULL")
	if value != nil {
		paramValue = q.transformValueWithRange(whereParams, true)
	}
	parameter := q.addQueryParameter(paramValue)
	whereToken := createWhereTokenWithOptions(WhereOperator_LESS_THAN_OR_EQUAL, fieldName, parameter, nil)

	tokens := *tokensRef
	tokens = append(tokens, whereToken)
	*tokensRef = tokens
}

func (q *AbstractDocumentQuery) _whereRegex(fieldName string, pattern string) {
	tokensRef := q.getCurrentWhereTokensRef()
	q.appendOperatorIfNeeded(tokensRef)
	q.negateIfNeeded(tokensRef, fieldName)

	whereParams := &whereParams{
		value:     pattern,
		fieldName: fieldName,
	}

	parameter := q.addQueryParameter(q.transformValue(whereParams))

	whereToken := createWhereToken(WhereOperator_REGEX, fieldName, parameter)

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
	if _, ok := lastToken.(*queryOperatorToken); ok {
		//throw new IllegalStateError("Cannot add AND, previous token was already an operator token.");
		panicIf(true, "Cannot add AND, previous token was already an operator token.")
	}

	tokens = append(tokens, queryOperatorTokenAnd)
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
	if _, ok := lastToken.(*queryOperatorToken); ok {
		//throw new IllegalStateError("Cannot add OR, previous token was already an operator token.");
		panicIf(true, "Cannot add OR, previous token was already an operator token.")
	}

	tokens = append(tokens, queryOperatorTokenOr)
	*tokensRef = tokens
}

func (q *AbstractDocumentQuery) _boost(boost float64) {
	if boost == 1.0 {
		return
	}

	tokens := q.getCurrentWhereTokens()
	n := len(tokens)
	if n == 0 {
		//throw new IllegalStateError("Missing where clause");
		panicIf(true, "Missing where clause")
	}

	maybeWhereToken := tokens[n-1]
	whereToken, ok := maybeWhereToken.(*whereToken)
	if !ok {
		//throw new IllegalStateError("Missing where clause");
		panicIf(true, "Missing where clause")
	}

	if boost <= 0.0 {
		//throw new IllegalArgumentError("Boost factor must be a positive number");
		panicIf(true, "Boost factor must be a positive number")
	}

	whereToken.options.boost = boost
}

func (q *AbstractDocumentQuery) _fuzzy(fuzzy float64) {
	tokens := q.getCurrentWhereTokens()
	n := len(tokens)
	if n == 0 {
		//throw new IllegalStateError("Missing where clause");
		panicIf(true, "Missing where clause")
	}

	maybeWhereToken := tokens[n-1]
	whereToken, ok := maybeWhereToken.(*whereToken)
	if !ok {
		//throw new IllegalStateError("Missing where clause");
		panicIf(true, "Missing where clause")
	}

	if fuzzy < 0.0 || fuzzy > 1.0 {
		//throw new IllegalArgumentError("Fuzzy distance must be between 0.0 and 1.0");
		panicIf(true, "Fuzzy distance must be between 0.0 and 1.0")
	}

	whereToken.options.fuzzy = fuzzy
}

func (q *AbstractDocumentQuery) _proximity(proximity int) {
	tokens := q.getCurrentWhereTokens()

	n := len(tokens)
	if n == 0 {
		//throw new IllegalStateError("Missing where clause");
		panicIf(true, "Missing where clause")
	}

	maybeWhereToken := tokens[n-1]
	whereToken, ok := maybeWhereToken.(*whereToken)
	if !ok {
		//throw new IllegalStateError("Missing where clause");
		panicIf(true, "Missing where clause")
	}

	if proximity < 1 {
		//throw new IllegalArgumentError("Proximity distance must be a positive number");
		panicIf(true, "Proximity distance must be a positive number")
	}

	whereToken.options.proximity = proximity
}

func (q *AbstractDocumentQuery) _orderBy(field string) {
	q._orderByWithOrdering(field, OrderingTypeString)
}

func (q *AbstractDocumentQuery) _orderByWithOrdering(field string, ordering OrderingType) {
	q.assertNoRawQuery()
	f := q.ensureValidFieldName(field, false)
	q.orderByTokens = append(q.orderByTokens, orderByTokenCreateAscending(f, ordering))
}

func (q *AbstractDocumentQuery) _orderByDescending(field string) {
	q._orderByDescendingWithOrdering(field, OrderingTypeString)
}

func (q *AbstractDocumentQuery) _orderByDescendingWithOrdering(field string, ordering OrderingType) {
	q.assertNoRawQuery()
	f := q.ensureValidFieldName(field, false)
	q.orderByTokens = append(q.orderByTokens, orderByTokenCreateDescending(f, ordering))
}

func (q *AbstractDocumentQuery) _orderByScore() {
	q.assertNoRawQuery()

	q.orderByTokens = append(q.orderByTokens, orderByTokenScoreAscending)
}

func (q *AbstractDocumentQuery) _orderByScoreDescending() {
	q.assertNoRawQuery()
	q.orderByTokens = append(q.orderByTokens, orderByTokenScoreDescending)
}

func (q *AbstractDocumentQuery) _statistics(stats **QueryStatistics) {
	*stats = q.queryStats
}

func (q *AbstractDocumentQuery) InvokeAfterQueryExecuted(result *QueryResult) {
	for _, cb := range q.afterQueryExecutedCallback {
		if cb != nil {
			cb(result)
		}
	}
}

func (q *AbstractDocumentQuery) InvokeBeforeQueryExecuted(query *IndexQuery) {
	for _, cb := range q.beforeQueryExecutedCallback {
		if cb != nil {
			cb(query)
		}
	}
}

func (q *AbstractDocumentQuery) InvokeAfterStreamExecuted(result ObjectNode) {
	for _, cb := range q.afterStreamExecutedCallback {
		if cb != nil {
			cb(result)
		}
	}
}

func (q *AbstractDocumentQuery) GenerateIndexQuery(query string) *IndexQuery {
	indexQuery := NewIndexQuery("")
	indexQuery.query = query
	indexQuery.start = q.start
	indexQuery.waitForNonStaleResults = q.theWaitForNonStaleResults
	indexQuery.waitForNonStaleResultsTimeout = q.timeout
	indexQuery.queryParameters = q.queryParameters
	indexQuery.disableCaching = q.disableCaching

	if q.pageSize != nil {
		indexQuery.pageSize = *q.pageSize
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

	whereToken := createWhereTokenWithOptions(WhereOperator_SEARCH, fieldName, q.addQueryParameter(searchTerms), NewWhereOptionsWithOperator(operator))

	tokens := *tokensRef
	tokens = append(tokens, whereToken)
	*tokensRef = tokens
}

func (q *AbstractDocumentQuery) String() string {
	if q.queryRaw != "" {
		return q.queryRaw
	}

	if q._currentClauseDepth != 0 {
		// throw new IllegalStateError("A clause was not closed correctly within this query, current clause depth = " + _currentClauseDepth);
		panicIf(true, "A clause was not closed correctly within this query, current clause depth = %d", q._currentClauseDepth)
	}

	queryText := &strings.Builder{}
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

func (q *AbstractDocumentQuery) buildInclude(queryText *strings.Builder) {
	if len(q.includes) == 0 {
		return
	}

	q.includes = stringArrayRemoveDuplicates(q.includes)
	queryText.WriteString(" include ")
	for i, include := range q.includes {
		if i > 0 {
			queryText.WriteString(",")
		}

		requiredQuotes := false

		for _, ch := range include {
			if !isLetterOrDigit(ch) && ch != '_' && ch != '.' {
				requiredQuotes = true
				break
			}
		}

		if requiredQuotes {
			s := strings.Replace(include, "'", "\\'", -1)
			queryText.WriteString("'")
			queryText.WriteString(s)
			queryText.WriteString("'")
		} else {
			queryText.WriteString(include)
		}
	}
}

func (q *AbstractDocumentQuery) _intersect() {
	tokensRef := q.getCurrentWhereTokensRef()
	tokens := *tokensRef
	n := len(tokens)
	if n > 0 {
		last := tokens[n-1]
		_, isWhere := last.(*whereToken)
		_, isClose := last.(*closeSubclauseToken)
		if isWhere || isClose {
			q.isIntersect = true

			tokens = append(tokens, intersectMarkerTokenInstance)
			*tokensRef = tokens
			return
		}
	}

	//throw new IllegalStateError("Cannot add INTERSECT at this point.");
	panicIf(true, "Cannot add INTERSECT at this point.")
}

func (q *AbstractDocumentQuery) _whereExists(fieldName string) {
	fieldName = q.ensureValidFieldName(fieldName, false)

	tokensRef := q.getCurrentWhereTokensRef()
	q.appendOperatorIfNeeded(tokensRef)
	q.negateIfNeeded(tokensRef, fieldName)

	tokens := *tokensRef
	tokens = append(tokens, createWhereToken(WhereOperator_EXISTS, fieldName, ""))
	*tokensRef = tokens
}

func (q *AbstractDocumentQuery) _containsAny(fieldName string, values []interface{}) {
	fieldName = q.ensureValidFieldName(fieldName, false)

	tokensRef := q.getCurrentWhereTokensRef()
	q.appendOperatorIfNeeded(tokensRef)
	q.negateIfNeeded(tokensRef, fieldName)

	array := q.transformCollection(fieldName, AbstractDocumentQuery_unpackCollection(values))
	whereToken := createWhereTokenWithOptions(WhereOperator_IN, fieldName, q.addQueryParameter(array), NewWhereOptionsWithExact(false))

	tokens := *tokensRef
	tokens = append(tokens, whereToken)
	*tokensRef = tokens
}

func (q *AbstractDocumentQuery) _containsAll(fieldName string, values []interface{}) {
	fieldName = q.ensureValidFieldName(fieldName, false)

	tokensRef := q.getCurrentWhereTokensRef()
	q.appendOperatorIfNeeded(tokensRef)
	q.negateIfNeeded(tokensRef, fieldName)

	array := q.transformCollection(fieldName, AbstractDocumentQuery_unpackCollection(values))

	tokens := *tokensRef
	if len(array) == 0 {
		tokens = append(tokens, trueTokenInstance)
	} else {
		whereToken := createWhereToken(WhereOperator_ALL_IN, fieldName, q.addQueryParameter(array))
		tokens = append(tokens, whereToken)
	}
	*tokensRef = tokens
}

func (q *AbstractDocumentQuery) _distinct() {
	panicIf(q.isDistinct(), "The is already a distinct query")
	//throw new IllegalStateError("The is already a distinct query");

	if len(q.selectTokens) == 0 {
		q.selectTokens = []queryToken{distinctTokenInstance}
		return
	}
	q.selectTokens = append([]queryToken{distinctTokenInstance}, q.selectTokens...)
}

func (q *AbstractDocumentQuery) UpdateStatsAndHighlightings(queryResult *QueryResult) {
	q.queryStats.UpdateQueryStats(queryResult)
	//TBD 4.1 Highlightings.Update(queryResult);
}

func (q *AbstractDocumentQuery) buildSelect(writer *strings.Builder) {
	if len(q.selectTokens) == 0 {
		return
	}

	writer.WriteString(" select ")

	if len(q.selectTokens) == 1 {
		tok := q.selectTokens[0]
		if dtok, ok := tok.(*distinctToken); ok {
			dtok.writeTo(writer)
			writer.WriteString(" *")
			return
		}
	}

	for i, token := range q.selectTokens {
		if i > 0 {
			prevToken := q.selectTokens[i-1]
			if _, ok := prevToken.(*distinctToken); !ok {
				writer.WriteString(",")
			}
		}

		var prevToken queryToken
		if i > 0 {
			prevToken = q.selectTokens[i-1]
		}
		documentQueryHelperAddSpaceIfNeeded(prevToken, token, writer)

		token.writeTo(writer)
	}
}

func (q *AbstractDocumentQuery) buildFrom(writer *strings.Builder) {
	q.fromToken.writeTo(writer)
}

func (q *AbstractDocumentQuery) buildDeclare(writer *strings.Builder) {
	if q.declareToken != nil {
		q.declareToken.writeTo(writer)
	}
}

func (q *AbstractDocumentQuery) buildLoad(writer *strings.Builder) {
	if len(q.loadTokens) == 0 {
		return
	}

	writer.WriteString(" load ")

	for i, tok := range q.loadTokens {
		if i != 0 {
			writer.WriteString(", ")
		}

		tok.writeTo(writer)
	}
}

func (q *AbstractDocumentQuery) buildWhere(writer *strings.Builder) {
	if len(q.whereTokens) == 0 {
		return
	}

	writer.WriteString(" where ")

	if q.isIntersect {
		writer.WriteString("intersect(")
	}

	for i, tok := range q.whereTokens {
		var prevToken queryToken
		if i > 0 {
			prevToken = q.whereTokens[i-1]
		}
		documentQueryHelperAddSpaceIfNeeded(prevToken, tok, writer)
		tok.writeTo(writer)
	}

	if q.isIntersect {
		writer.WriteString(") ")
	}
}

func (q *AbstractDocumentQuery) buildGroupBy(writer *strings.Builder) {
	if len(q.groupByTokens) == 0 {
		return
	}

	writer.WriteString(" group by ")

	for i, token := range q.groupByTokens {
		if i > 0 {
			writer.WriteString(", ")
		}
		token.writeTo(writer)
	}
}

func (q *AbstractDocumentQuery) buildOrderBy(writer *strings.Builder) {
	if len(q.orderByTokens) == 0 {
		return
	}

	writer.WriteString(" order by ")

	for i, token := range q.orderByTokens {
		if i > 0 {
			writer.WriteString(", ")
		}

		token.writeTo(writer)
	}
}

func (q *AbstractDocumentQuery) appendOperatorIfNeeded(tokensRef *[]queryToken) {
	tokens := *tokensRef
	q.assertNoRawQuery()

	n := len(tokens)
	if len(tokens) == 0 {
		return
	}

	lastToken := tokens[n-1]
	_, isWhereToken := lastToken.(*whereToken)
	_, isCloseSubclauseToken := lastToken.(*closeSubclauseToken)
	if !isWhereToken && !isCloseSubclauseToken {
		return
	}

	var lastWhere *whereToken

	for i := n - 1; i >= 0; i-- {
		tok := tokens[i]
		if maybeLastWhere, ok := tok.(*whereToken); ok {
			lastWhere = maybeLastWhere
			break
		}
	}

	var token *queryOperatorToken
	if q.defaultOperator == QueryOperatorAnd {
		token = queryOperatorTokenAnd
	} else {
		token = queryOperatorTokenOr
	}

	if lastWhere != nil && lastWhere.options.searchOperator != SearchOperator_UNSET {
		token = queryOperatorTokenOr // default to OR operator after search if AND was not specified explicitly
	}

	tokens = append(tokens, token)
	*tokensRef = tokens
}

func (q *AbstractDocumentQuery) transformCollection(fieldName string, values []interface{}) []interface{} {
	var result []interface{}
	for _, value := range values {
		if collectionValue, ok := value.([]interface{}); ok {
			tmp := q.transformCollection(fieldName, collectionValue)
			result = append(result, tmp...)
		} else {
			nestedWhereParams := &whereParams{
				allowWildcards: true,
				fieldName:      fieldName,
				value:          value,
			}
			tmp := q.transformValue(nestedWhereParams)
			result = append(result, tmp)
		}
	}
	return result
}

func (q *AbstractDocumentQuery) negateIfNeeded(tokensRef *[]queryToken, fieldName string) {
	if !q.negate {
		return
	}

	q.negate = false

	tokens := *tokensRef

	n := len(tokens)
	isOpenSubclauseToken := false
	if n > 0 {
		_, isOpenSubclauseToken = tokens[n-1].(*openSubclauseToken)
	}
	if n == 0 || isOpenSubclauseToken {
		if fieldName != "" {
			q._whereExists(fieldName)
		} else {
			q._whereTrue()
		}
		q._andAlso()
	}

	tokens = append(tokens, negateTokenInstance)
	*tokensRef = tokens
}

func AbstractDocumentQuery_unpackCollection(items []interface{}) []interface{} {
	var results []interface{}

	for _, item := range items {
		if itemCollection, ok := item.([]interface{}); ok {
			els := AbstractDocumentQuery_unpackCollection(itemCollection)
			results = append(results, els...)
		} else {
			results = append(results, item)
		}
	}

	return results
}

func assertValidFieldName(fieldName string) {
	// TODO: for now all names are valid.
	// The code below checks
	if true {
		return
	}
	// in Go only public fields can be serialized so check that first
	// letter is uppercase
	if len(fieldName) == 0 {
		return
	}
	for i, c := range fieldName {
		if i > 0 {
			return
		}
		isUpper := unicode.IsUpper(c)
		panicIf(!isUpper, "field '%s' is not public (doesn't start with uppercase letter)", fieldName)
	}
}

func (q *AbstractDocumentQuery) ensureValidFieldName(fieldName string, isNestedPath bool) string {
	assertValidFieldName(fieldName)
	if q.theSession == nil || q.theSession.GetConventions() == nil || isNestedPath || q.isGroupBy {
		return QueryFieldUtil_escapeIfNecessary(fieldName)
	}

	if fieldName == documentConventionsIdentityPropertyName {
		return IndexingFieldNameDocumentID
	}

	return QueryFieldUtil_escapeIfNecessary(fieldName)
}

func (q *AbstractDocumentQuery) transformValue(whereParams *whereParams) interface{} {
	return q.transformValueWithRange(whereParams, false)
}

func (q *AbstractDocumentQuery) transformValueWithRange(whereParams *whereParams, forRange bool) interface{} {
	if whereParams.value == nil {
		return nil
	}

	if "" == whereParams.value {
		return ""
	}

	var stringValueReference string
	if q._conventions.TryConvertValueForQuery(whereParams.fieldName, whereParams.value, forRange, &stringValueReference) {
		return stringValueReference
	}

	val := whereParams.value
	switch v := val.(type) {
	case time.Time, string, int, int32, int64, float32, float64, bool:
		return val
	case time.Duration:
		n := int64(v/time.Nanosecond) / 100
		return n
	}
	return whereParams.value
}

func (q *AbstractDocumentQuery) addQueryParameter(value interface{}) string {
	parameterName := "p" + strconv.Itoa(len(q.queryParameters))
	q.queryParameters[parameterName] = value
	return parameterName
}

func (q *AbstractDocumentQuery) getCurrentWhereTokens() []queryToken {
	if !q._isInMoreLikeThis {
		return q.whereTokens
	}

	n := len(q.whereTokens)

	if n == 0 {
		// throw new IllegalStateError("Cannot get moreLikeThisToken because there are no where token specified.");
		panicIf(true, "Cannot get moreLikeThisToken because there are no where token specified.")
	}

	lastToken := q.whereTokens[n-1]

	if moreLikeThisToken, ok := lastToken.(*moreLikeThisToken); ok {
		return moreLikeThisToken.whereTokens
	} else {
		//throw new IllegalStateError("Last token is not moreLikeThisToken");
		panicIf(true, "Last token is not moreLikeThisToken")
	}
	return nil
}

func (q *AbstractDocumentQuery) getCurrentWhereTokensRef() *[]queryToken {
	if !q._isInMoreLikeThis {
		return &q.whereTokens
	}

	n := len(q.whereTokens)

	if n == 0 {
		// throw new IllegalStateError("Cannot get moreLikeThisToken because there are no where token specified.");
		panicIf(true, "Cannot get moreLikeThisToken because there are no where token specified.")
	}

	lastToken := q.whereTokens[n-1]

	if moreLikeThisToken, ok := lastToken.(*moreLikeThisToken); ok {
		return &moreLikeThisToken.whereTokens
	} else {
		//throw new IllegalStateError("Last token is not moreLikeThisToken");
		panicIf(true, "Last token is not moreLikeThisToken")
	}
	return nil
}

func (q *AbstractDocumentQuery) updateFieldsToFetchToken(fieldsToFetch *fieldsToFetchToken) {
	q.fieldsToFetchToken = fieldsToFetch

	if len(q.selectTokens) == 0 {
		q.selectTokens = append(q.selectTokens, fieldsToFetch)
	} else {
		for _, x := range q.selectTokens {
			if _, ok := x.(*fieldsToFetchToken); ok {
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

func (q *AbstractDocumentQuery) _addBeforeQueryExecutedListener(action func(*IndexQuery)) int {
	q.beforeQueryExecutedCallback = append(q.beforeQueryExecutedCallback, action)
	return len(q.beforeQueryExecutedCallback) - 1
}

func (q *AbstractDocumentQuery) _removeBeforeQueryExecutedListener(idx int) {
	q.beforeQueryExecutedCallback[idx] = nil
}

func (q *AbstractDocumentQuery) _addAfterQueryExecutedListener(action func(*QueryResult)) int {
	q.afterQueryExecutedCallback = append(q.afterQueryExecutedCallback, action)
	return len(q.afterQueryExecutedCallback) - 1
}

func (q *AbstractDocumentQuery) _removeAfterQueryExecutedListener(idx int) {
	q.afterQueryExecutedCallback[idx] = nil
}

func (q *AbstractDocumentQuery) _addAfterStreamExecutedListener(action func(ObjectNode)) int {
	q.afterStreamExecutedCallback = append(q.afterStreamExecutedCallback, action)
	return len(q.afterStreamExecutedCallback) - 1
}

func (q *AbstractDocumentQuery) _removeAfterStreamExecutedListener(idx int) {
	q.afterStreamExecutedCallback[idx] = nil
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
	whereToken := createWhereTokenWithOptions(WhereOperator_SPATIAL_WITHIN, fieldName, "", opts)

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
	case SpatialRelationWithin:
		whereOperator = WhereOperator_SPATIAL_WITHIN
	case SpatialRelationContains:
		whereOperator = WhereOperator_SPATIAL_CONTAINS
	case SpatialRelationDisjoin:
		whereOperator = WhereOperator_SPATIAL_DISJOINT
	case SpatialRelationIntersects:
		whereOperator = WhereOperator_SPATIAL_INTERSECTS
	default:
		//throw new IllegalArgumentError();
		panicIf(true, "unknown relation %s", relation)
	}

	tokens := *tokensRef
	opts := NewWhereOptionsWithTokenAndDistance(wktToken, distErrorPercent)
	tok := createWhereTokenWithOptions(whereOperator, fieldName, "", opts)
	tokens = append(tokens, tok)
	*tokensRef = tokens
}

func (q *AbstractDocumentQuery) _spatial2(dynamicField DynamicSpatialField, criteria SpatialCriteria) {
	tokensRef := q.getCurrentWhereTokensRef()
	q.appendOperatorIfNeeded(tokensRef)
	q.negateIfNeeded(tokensRef, "")

	ensure := func(fieldName string, isNestedPath bool) string {
		return q.ensureValidFieldName(fieldName, isNestedPath)
	}
	add := func(value interface{}) string {
		return q.addQueryParameter(value)
	}
	tok := criteria.ToQueryToken(dynamicField.ToField(ensure), add)
	tokens := *tokensRef
	tokens = append(tokens, tok)
	*tokensRef = tokens
}

func (q *AbstractDocumentQuery) _spatial3(fieldName string, criteria SpatialCriteria) {
	fieldName = q.ensureValidFieldName(fieldName, false)

	tokensRef := q.getCurrentWhereTokensRef()
	q.appendOperatorIfNeeded(tokensRef)
	q.negateIfNeeded(tokensRef, fieldName)

	tokens := *tokensRef
	add := func(value interface{}) string {
		return q.addQueryParameter(value)
	}
	tok := criteria.ToQueryToken(fieldName, add)
	tokens = append(tokens, tok)
	*tokensRef = tokens
}

func (q *AbstractDocumentQuery) _orderByDistance(field DynamicSpatialField, latitude float64, longitude float64) {
	if field == nil {
		//throw new IllegalArgumentError("Field cannot be null");
		panicIf(true, "Field cannot be null")
	}
	ensure := func(fieldName string, isNestedPath bool) string {
		return q.ensureValidFieldName(fieldName, isNestedPath)
	}

	q._orderByDistanceLatLong("'"+field.ToField(ensure)+"'", latitude, longitude)
}

func (q *AbstractDocumentQuery) _orderByDistanceLatLong(fieldName string, latitude float64, longitude float64) {
	tok := orderByTokenCreateDistanceAscending(fieldName, q.addQueryParameter(latitude), q.addQueryParameter(longitude))
	q.orderByTokens = append(q.orderByTokens, tok)
}

func (q *AbstractDocumentQuery) _orderByDistance2(field DynamicSpatialField, shapeWkt string) {
	if field == nil {
		//throw new IllegalArgumentError("Field cannot be null");
		panicIf(true, "Field cannot be null")
	}
	ensure := func(fieldName string, isNestedPath bool) string {
		return q.ensureValidFieldName(fieldName, isNestedPath)
	}
	q._orderByDistance3("'"+field.ToField(ensure)+"'", shapeWkt)
}

func (q *AbstractDocumentQuery) _orderByDistance3(fieldName string, shapeWkt string) {
	tok := orderByTokenCreateDistanceAscending2(fieldName, q.addQueryParameter(shapeWkt))
	q.orderByTokens = append(q.orderByTokens, tok)
}

func (q *AbstractDocumentQuery) _orderByDistanceDescending(field DynamicSpatialField, latitude float64, longitude float64) {
	if field == nil {
		//throw new IllegalArgumentError("Field cannot be null");
		panicIf(true, "Field cannot be null")
	}
	ensure := func(fieldName string, isNestedPath bool) string {
		return q.ensureValidFieldName(fieldName, isNestedPath)
	}
	q._orderByDistanceDescendingLatLong("'"+field.ToField(ensure)+"'", latitude, longitude)
}

func (q *AbstractDocumentQuery) _orderByDistanceDescendingLatLong(fieldName string, latitude float64, longitude float64) {
	tok := orderByTokenCreateDistanceDescending(fieldName, q.addQueryParameter(latitude), q.addQueryParameter(longitude))
	q.orderByTokens = append(q.orderByTokens, tok)
}

func (q *AbstractDocumentQuery) _orderByDistanceDescending2(field DynamicSpatialField, shapeWkt string) {
	if field == nil {
		//throw new IllegalArgumentError("Field cannot be null");
		panicIf(true, "Field cannot be null")
	}
	ensure := func(fieldName string, isNestedPath bool) string {
		return q.ensureValidFieldName(fieldName, isNestedPath)
	}
	q._orderByDistanceDescending3("'"+field.ToField(ensure)+"'", shapeWkt)
}

func (q *AbstractDocumentQuery) _orderByDistanceDescending3(fieldName string, shapeWkt string) {
	tok := orderByTokenCreateDistanceDescending2(fieldName, q.addQueryParameter(shapeWkt))
	q.orderByTokens = append(q.orderByTokens, tok)
}

func (q *AbstractDocumentQuery) initSync() error {
	if q.queryOperation != nil {
		return nil
	}

	delegate := &DocumentQueryCustomization{
		query: q,
	}
	beforeQueryEventArgs := &BeforeQueryEventArgs{
		Session:            q.theSession,
		QueryCustomization: delegate,
	}
	q.theSession.OnBeforeQueryInvoke(beforeQueryEventArgs)

	q.queryOperation = q.initializeQueryOperation()
	return q.executeActualQuery()
}

func (q *AbstractDocumentQuery) executeActualQuery() error {
	{
		context := q.queryOperation.enterQueryContext()
		command := q.queryOperation.CreateRequest()
		err := q.theSession.GetRequestExecutor().ExecuteCommandWithSessionInfo(command, q.theSession.sessionInfo)
		q.queryOperation.setResult(command.Result)
		context.Close()
		if err != nil {
			return err
		}
	}
	q.InvokeAfterQueryExecuted(q.queryOperation.currentQueryResults)
	return nil
}

// GetQueryResult returns results of a query
func (q *AbstractDocumentQuery) GetQueryResult() (*QueryResult, error) {
	err := q.initSync()
	if err != nil {
		return nil, err
	}

	return q.queryOperation.currentQueryResults.createSnapshot(), nil
}

// given *[]<type> returns <type>
func getTypeFromQueryResults(results interface{}) (reflect.Type, error) {
	rt := reflect.TypeOf(results)
	if (rt.Kind() == reflect.Ptr) && (rt.Elem() != nil) && (rt.Elem().Kind() == reflect.Slice) {
		return rt.Elem().Elem(), nil
	}
	return nil, fmt.Errorf("expected value of type *[]<type>, got %T", results)
}

func (q *AbstractDocumentQuery) setClazzFromResult(result interface{}) {
	if q.clazz != nil {
		return
	}

	var err error
	// it's possible to search in index without needing to know the type
	// see moreLikeThisCanGetResultsUsingTermVectorsLazy test
	if q.fromToken != nil {
		q.clazz, err = getTypeFromQueryResults(result)
		if err != nil {
			panic(err.Error())
		}
		return
	}

	// query was created without providing the type to query
	panicIf(q.indexName != "", "q.clazz is not set but q.indexName is")
	panicIf(q.collectionName != "", "q.clazz is not set but q.collectionName is")
	panicIf(q.fromToken != nil, "q.clazz is not set but q.fromToken is")
	tp := reflect.TypeOf(result)
	if tp2, ok := isPtrPtrStruct(tp); ok {
		q.clazz = tp2
	} else if tp2, ok := isPtrSlicePtrStruct(tp); ok {
		q.clazz = tp2
	} else {
		panicIf(true, "expected result to be **struct or *[]*struct, got: %T", result)
	}
	s := q.theSession
	indexName, collectionName := s.processQueryParameters(q.clazz, "", "", s.GetConventions())
	q.fromToken = createFromToken(indexName, collectionName, "")
}

// GetResults executes the query and sets results to returned values.
// results should be of type *[]<type>
// TODO: name it Execute() instead?
// Note: ToList in java
func (q *AbstractDocumentQuery) GetResults(results interface{}) error {
	if results == nil {
		return fmt.Errorf("results can't be nil")
	}

	// delayed SelectFields logic
	if q.selectFieldsArgs != nil {
		hadClazz := (q.clazz != nil)
		// query was created without providing the type to query
		projectionClass, err := getTypeFromQueryResults(results)
		if err != nil {
			return err
		}
		dq := q.createDocumentQueryInternal(projectionClass, q.selectFieldsArgs)
		q = dq.AbstractDocumentQuery
		panicIf(q.clazz != projectionClass, "q.clazz != projectionClass")
		if !hadClazz {
			s := q.theSession
			q.indexName, q.collectionName = s.processQueryParameters(q.clazz, q.indexName, q.collectionName, s.GetConventions())
			q.fromToken = createFromToken(q.indexName, q.collectionName, "")
		}
	}

	if q.clazz == nil {
		// query was created without providing the type to query
		var err error
		q.clazz, err = getTypeFromQueryResults(results)
		if err != nil {
			return err
		}
		s := q.theSession
		q.indexName, q.collectionName = s.processQueryParameters(q.clazz, q.indexName, q.collectionName, s.GetConventions())
		q.fromToken = createFromToken(q.indexName, q.collectionName, "")
	}

	return q.executeQueryOperation(results, 0)
}

// First runs a query and returns a first result.
func (q *AbstractDocumentQuery) First(result interface{}) error {
	if result == nil {
		return fmt.Errorf("result can't be nil")
	}
	q.setClazzFromResult(result)

	tp := reflect.TypeOf(result)
	// **struct => *struct
	if tp.Kind() == reflect.Ptr && tp.Elem().Kind() == reflect.Ptr {
		tp = tp.Elem()
	}
	// create a pointer to a slice. executeQueryOperation creates the actual slice
	sliceType := reflect.SliceOf(tp)
	slicePtr := reflect.New(sliceType)
	err := q.executeQueryOperation(slicePtr.Interface(), 1)
	if err != nil {
		return err
	}
	slice := slicePtr.Elem()
	if slice.Len() < 1 {
		return nil
	}
	el := slice.Index(0)
	setInterfaceToValue(result, el.Interface())
	return nil
}

// Single runs a query that expects only a single result.
// If there is more than one result, it retuns IllegalStateError.
func (q *AbstractDocumentQuery) Single(result interface{}) error {
	if result == nil {
		return fmt.Errorf("result can't be nil")
	}
	q.setClazzFromResult(result)

	tp := reflect.TypeOf(result)
	// **struct => *struct
	if tp.Kind() == reflect.Ptr && tp.Elem().Kind() == reflect.Ptr {
		tp = tp.Elem()
	}
	// create a pointer to a slice. executeQueryOperation creates the actual slice
	sliceType := reflect.SliceOf(tp)
	slicePtr := reflect.New(sliceType)
	err := q.executeQueryOperation(slicePtr.Interface(), 2)
	if err != nil {
		return err
	}
	slice := slicePtr.Elem()
	if slice.Len() > 1 {
		return newIllegalStateError("Expected single result, got: %d", slice.Len())
	}
	el := slice.Index(0)
	setInterfaceToValue(result, el.Interface())
	return nil
}

func (q *AbstractDocumentQuery) Count() (int, error) {
	{
		var tmp = 0
		q._take(&tmp)
	}
	queryResult, err := q.GetQueryResult()
	if err != nil {
		return 0, err
	}
	return queryResult.TotalResults, nil
}

// Any returns true if query returns at least one result
// TODO: write tests
func (q *AbstractDocumentQuery) Any() (bool, error) {
	if q.isDistinct() {
		// for distinct it is cheaper to do count 1

		tp := q.clazz
		// **struct => *struct
		if tp.Kind() == reflect.Ptr && tp.Elem().Kind() == reflect.Ptr {
			tp = tp.Elem()
		}
		// create a pointer to a slice. executeQueryOperation creates the actual slice
		sliceType := reflect.SliceOf(tp)
		slicePtr := reflect.New(sliceType)
		err := q.executeQueryOperation(slicePtr.Interface(), 1)
		if err != nil {
			return false, err
		}
		slice := slicePtr.Elem()
		return slice.Len() > 0, nil
	}

	{
		var tmp = 0
		q._take(&tmp)
	}
	queryResult, err := q.GetQueryResult()
	if err != nil {
		return false, err
	}
	return queryResult.TotalResults > 0, nil
}

func (q *AbstractDocumentQuery) executeQueryOperation(results interface{}, take int) error {
	if take != 0 && (q.pageSize == nil || *q.pageSize > take) {
		q._take(&take)
	}

	err := q.initSync()
	if err != nil {
		return err
	}

	return q.queryOperation.complete(results)
}

func (q *AbstractDocumentQuery) _aggregateBy(facet FacetBase) {
	for _, token := range q.selectTokens {
		if _, ok := token.(*facetToken); ok {
			continue
		}

		//throw new IllegalStateError("Aggregation query can select only facets while it got " + token.getClass().getSimpleName() + " token");
		panicIf(true, "Aggregation query can select only facets while it got %T token", token)
	}

	add := func(o interface{}) string {
		return q.addQueryParameter(o)
	}
	q.selectTokens = append(q.selectTokens, createFacetTokenWithFacetBase(facet, add))
}

func (q *AbstractDocumentQuery) _aggregateUsing(facetSetupDocumentID string) {
	q.selectTokens = append(q.selectTokens, createFacetToken(facetSetupDocumentID))
}

func (q *AbstractDocumentQuery) Lazily(results interface{}, onEval func(interface{})) *Lazy {
	q.setClazzFromResult(results)
	if q.queryOperation == nil {
		q.queryOperation = q.initializeQueryOperation()
	}

	lazyQueryOperation := NewLazyQueryOperation(results, q.theSession.GetConventions(), q.queryOperation, q.afterQueryExecutedCallback)
	return q.theSession.session.addLazyOperation(results, lazyQueryOperation, onEval)
}

func (q *AbstractDocumentQuery) CountLazily(results interface{}, count *int) *Lazy {
	if q.queryOperation == nil {
		v := 0
		q._take(&v)
		q.queryOperation = q.initializeQueryOperation()
	}

	lazyQueryOperation := NewLazyQueryOperation(results, q.theSession.GetConventions(), q.queryOperation, q.afterQueryExecutedCallback)
	return q.theSession.session.addLazyCountOperation(count, lazyQueryOperation)
}

// SuggestUsing adds a query part for suggestions
func (q *AbstractDocumentQuery) _suggestUsing(suggestion SuggestionBase) {
	if suggestion == nil {
		panic(newIllegalArgumentError("suggestion cannot be null"))
		// throw new IllegalArgumentError("suggestion cannot be null");
	}

	q.assertCanSuggest()

	var token *suggestToken

	if term, ok := suggestion.(*SuggestionWithTerm); ok {
		token = &suggestToken{
			fieldName:            term.Field,
			termParameterName:    q.addQueryParameter(term.Term),
			optionsParameterName: q.getOptionsParameterName(term.Options),
		}
	} else if terms, ok := suggestion.(*SuggestionWithTerms); ok {
		token = &suggestToken{
			fieldName:            terms.Field,
			termParameterName:    q.addQueryParameter(terms.Terms),
			optionsParameterName: q.getOptionsParameterName(terms.Options),
		}
	} else {
		// throw new UnsupportedOperationError("Unknown type of suggestion: " + suggestion.getClass());
		panic(newUnsupportedOperationError("Unknown type of suggestion: %T", suggestion))
	}
	q.selectTokens = append(q.selectTokens, token)
}

func (q *AbstractDocumentQuery) getOptionsParameterName(options *SuggestionOptions) string {
	optionsParameterName := ""
	if options != nil && options != SuggestionOptionsDefaultOptions {
		optionsParameterName = q.addQueryParameter(options)
	}

	return optionsParameterName
}

func (q *AbstractDocumentQuery) assertCanSuggest() {
	if len(q.whereTokens) > 0 {
		//throw new IllegalStateError("Cannot add suggest when WHERE statements are present.");
		panicIf(true, "Cannot add suggest when WHERE statements are present.")
	}

	if len(q.selectTokens) > 0 {
		//throw new IllegalStateError("Cannot add suggest when SELECT statements are present.");
		panicIf(true, "Cannot add suggest when SELECT statements are present.")
	}

	if len(q.orderByTokens) > 0 {
		//throw new IllegalStateError("Cannot add suggest when ORDER BY statements are present.");
		panicIf(true, "Cannot add suggest when ORDER BY statements are present.")
	}
}
