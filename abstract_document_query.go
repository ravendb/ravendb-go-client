package ravendb

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"
)

// Note: Java's IAbstractDocumentQuery is abstractDocumentQuery

type abstractDocumentQuery struct {
	aliasToGroupByFieldName map[string]string
	defaultOperator         QueryOperator

	// Note: rootTypes is not used in Go because we only have one ID property

	negate             bool
	indexName          string
	collectionName     string
	currentClauseDepth int
	queryRaw           string
	queryParameters    Parameters

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

	start       int
	conventions *DocumentConventions

	timeout time.Duration

	theWaitForNonStaleResults bool

	includes []string

	queryStats *QueryStatistics

	disableEntitiesTracking bool

	disableCaching bool

	isInMoreLikeThis bool

	// Go doesn't allow comparing functions so to remove we use index returned
	// by add() function. We maintain stable index by never shrinking
	// callback arrays. We assume there is no high churn of adding/removing
	// callbacks
	beforeQueryExecutedCallback []func(*IndexQuery)
	afterQueryExecutedCallback  []func(*QueryResult)
	afterStreamExecutedCallback []func(map[string]interface{})

	queryOperation *queryOperation

	// to allow "fluid" API, in many methods instead of returning an error, we
	// remember it here and retrun in GetResults()
	err error
}

func (q *abstractDocumentQuery) isDistinct() bool {
	if len(q.selectTokens) == 0 {
		return false
	}
	_, ok := q.selectTokens[0].(*distinctToken)
	return ok
}

func (q *abstractDocumentQuery) getSession() *InMemoryDocumentSessionOperations {
	return q.theSession
}

func (q *abstractDocumentQuery) isDynamicMapReduce() bool {
	return len(q.groupByTokens) > 0
}

func getQueryDefaultTimeout() time.Duration {
	return time.Second * 15
}

// at this point we assume all
func newAbstractDocumentQuery(opts *DocumentQueryOptions) *abstractDocumentQuery {
	res := &abstractDocumentQuery{
		defaultOperator:         QueryOperatorAnd,
		isGroupBy:               opts.isGroupBy,
		indexName:               opts.IndexName,
		collectionName:          opts.CollectionName,
		declareToken:            opts.declareToken,
		loadTokens:              opts.loadTokens,
		theSession:              opts.session,
		aliasToGroupByFieldName: make(map[string]string),
		queryParameters:         make(map[string]interface{}),
		queryStats:              NewQueryStatistics(),
		queryRaw:                opts.rawQuery,
	}

	if opts.session == nil {
		res.err = newIllegalArgumentError("session must be provided")
		return res
	}

	if res.queryRaw == "" {
		if opts.IndexName == "" && opts.CollectionName == "" {
			res.err = newIllegalArgumentError("Either indexName or collectionName must be specified")
			return res
		}
		res.fromToken = createFromToken(opts.IndexName, opts.CollectionName, opts.fromAlias)
	}

	f := func(queryResult *QueryResult) {
		res.updateStatsAndHighlightings(queryResult)
	}
	res.addAfterQueryExecutedListener(f)
	if opts.session == nil {
		res.conventions = NewDocumentConventions()
	} else {
		res.conventions = opts.session.GetConventions()
	}
	return res
}

func (q *abstractDocumentQuery) usingDefaultOperator(operator QueryOperator) error {
	if len(q.whereTokens) > 0 {
		return newIllegalStateError("Default operator can only be set before any where clause is added.")
	}

	q.defaultOperator = operator
	return nil
}

func (q *abstractDocumentQuery) waitForNonStaleResults(waitTimeout time.Duration) {
	q.theWaitForNonStaleResults = true
	if waitTimeout == 0 {
		waitTimeout = getQueryDefaultTimeout()
	}
	q.timeout = waitTimeout
}

func (q *abstractDocumentQuery) initializeQueryOperation() (*queryOperation, error) {
	indexQuery, err := q.GetIndexQuery()
	if err != nil {
		return nil, err
	}

	return newQueryOperation(q.theSession, q.indexName, indexQuery, q.fieldsToFetchToken, q.disableEntitiesTracking, false, false)
}

func (q *abstractDocumentQuery) GetIndexQuery() (*IndexQuery, error) {
	query, err := q.string()
	if err != nil {
		return nil, err
	}
	indexQuery := q.generateIndexQuery(query)
	q.invokeBeforeQueryExecuted(indexQuery)
	return indexQuery, nil
}

func (q *abstractDocumentQuery) getProjectionFields() []string {

	if q.fieldsToFetchToken != nil && q.fieldsToFetchToken.projections != nil {
		return q.fieldsToFetchToken.projections
	}
	return nil
}

func (q *abstractDocumentQuery) randomOrdering() error {
	if err := q.assertNoRawQuery(); err != nil {
		return err
	}

	q.noCaching()
	q.orderByTokens = append(q.orderByTokens, orderByTokenRandom)
	return nil
}

func (q *abstractDocumentQuery) randomOrderingWithSeed(seed string) error {
	err := q.assertNoRawQuery()
	if err != nil {
		return err
	}

	if stringIsBlank(seed) {
		err = q.randomOrdering()
		if err != nil {
			return err
		}
		return nil
	}

	q.noCaching()
	q.orderByTokens = append(q.orderByTokens, orderByTokenCreateRandom(seed))
	return nil
}

func (q *abstractDocumentQuery) addGroupByAlias(fieldName string, projectedName string) {
	q.aliasToGroupByFieldName[projectedName] = fieldName
}

func (q *abstractDocumentQuery) assertNoRawQuery() error {
	if q.queryRaw != "" {
		return newIllegalStateError("RawQuery was called, cannot modify this query by calling on operations that would modify the query (such as Where, Select, OrderBy, GroupBy, etc)")
	}
	return nil
}

func (q *abstractDocumentQuery) addParameter(name string, value interface{}) error {
	name = strings.TrimPrefix(name, "$")
	if _, ok := q.queryParameters[name]; ok {
		return newIllegalStateError("The parameter " + name + " was already added")
	}

	q.queryParameters[name] = value
	return nil
}

func (q *abstractDocumentQuery) groupBy(fieldName string, fieldNames ...string) error {
	var mapping []*GroupBy
	for _, x := range fieldNames {
		el := NewGroupByField(x)
		mapping = append(mapping, el)
	}
	return q.groupBy2(NewGroupByField(fieldName), mapping...)
}

// TODO: better name
func (q *abstractDocumentQuery) groupBy2(field *GroupBy, fields ...*GroupBy) error {
	// TODO: if q.fromToken is nil, needs to do this check in ToList()
	if q.fromToken != nil && !q.fromToken.isDynamic {
		return newIllegalStateError("groupBy only works with dynamic queries")
	}

	if err := q.assertNoRawQuery(); err != nil {
		return err
	}
	q.isGroupBy = true

	fieldName, err := q.ensureValidFieldName(field.Field, false)
	if err != nil {
		return err
	}

	q.groupByTokens = append(q.groupByTokens, createGroupByTokenWithMethod(fieldName, field.Method))

	if len(fields) == 0 {
		return nil
	}

	for _, item := range fields {
		fieldName, err = q.ensureValidFieldName(item.Field, false)
		if err != nil {
			return err
		}
		q.groupByTokens = append(q.groupByTokens, createGroupByTokenWithMethod(fieldName, item.Method))
	}
	return nil
}

func (q *abstractDocumentQuery) groupByKey(fieldName string, projectedName string) error {
	if err := q.assertNoRawQuery(); err != nil {
		return err
	}
	q.isGroupBy = true

	_, hasProjectedName := q.aliasToGroupByFieldName[projectedName]
	_, hasFieldName := q.aliasToGroupByFieldName[fieldName]

	if projectedName != "" && hasProjectedName {
		aliasedFieldName := q.aliasToGroupByFieldName[projectedName]
		if fieldName == "" || strings.EqualFold(fieldName, projectedName) {
			fieldName = aliasedFieldName
		}
	} else if fieldName != "" && hasFieldName {
		aliasedFieldName := q.aliasToGroupByFieldName[fieldName]
		fieldName = aliasedFieldName
	}

	q.selectTokens = append(q.selectTokens, createGroupByKeyToken(fieldName, projectedName))
	return nil
}

// projectedName is optional
func (q *abstractDocumentQuery) groupBySum(fieldName string, projectedName string) error {
	if err := q.assertNoRawQuery(); err != nil {
		return err
	}
	q.isGroupBy = true

	var err error
	fieldName, err = q.ensureValidFieldName(fieldName, false)
	if err != nil {
		return err
	}
	q.selectTokens = append(q.selectTokens, createGroupBySumToken(fieldName, projectedName))
	return nil
}

// projectedName is optional
func (q *abstractDocumentQuery) groupByCount(projectedName string) error {
	if err := q.assertNoRawQuery(); err != nil {
		return err
	}
	q.isGroupBy = true

	t := &groupByCountToken{
		fieldName: projectedName,
	}
	q.selectTokens = append(q.selectTokens, t)
	return nil
}

func (q *abstractDocumentQuery) whereTrue() error {
	tokensRef, err := q.getCurrentWhereTokensRef()
	if err != nil {
		return err
	}
	err = q.appendOperatorIfNeeded(tokensRef)
	if err != nil {
		return err
	}
	err = q.negateIfNeeded(tokensRef, "")
	if err != nil {
		return err
	}

	tokens := *tokensRef
	tokens = append(tokens, trueTokenInstance)
	*tokensRef = tokens
	return nil
}

func (q *abstractDocumentQuery) moreLikeThis() (*moreLikeThisScope, error) {
	err := q.appendOperatorIfNeeded(&q.whereTokens)
	if err != nil {
		return nil, err
	}

	token := newMoreLikeThisToken()
	q.whereTokens = append(q.whereTokens, token)

	q.isInMoreLikeThis = true
	add := func(o interface{}) string {
		return q.addQueryParameter(o)
	}
	onDispose := func() {
		q.isInMoreLikeThis = false
	}
	return newMoreLikeThisScope(token, add, onDispose), nil
}

func (q *abstractDocumentQuery) include(path string) {
	q.includes = append(q.includes, path)
}

// TODO: see if count can be int
func (q *abstractDocumentQuery) take(count *int) {
	q.pageSize = count
}

func (q *abstractDocumentQuery) skip(count int) {
	q.start = count
}

func (q *abstractDocumentQuery) whereLucene(fieldName string, whereClause string) error {
	var err error
	fieldName, err = q.ensureValidFieldName(fieldName, false)
	if err != nil {
		return err
	}
	tokensRef, err := q.getCurrentWhereTokensRef()
	if err != nil {
		return err
	}
	tokens := *tokensRef
	err = q.appendOperatorIfNeeded(tokensRef)
	if err != nil {
		return err
	}
	err = q.negateIfNeeded(tokensRef, fieldName)
	if err != nil {
		return err
	}

	whereToken := createWhereTokenWithOptions(whereOperatorLucene, fieldName, q.addQueryParameter(whereClause), nil)
	tokens = append(tokens, whereToken)
	*tokensRef = tokens
	return nil
}

func (q *abstractDocumentQuery) openSubclause() error {
	q.currentClauseDepth++

	tokensRef, err := q.getCurrentWhereTokensRef()
	if err != nil {
		return err
	}
	err = q.appendOperatorIfNeeded(tokensRef)
	if err != nil {
		return err
	}
	err = q.negateIfNeeded(tokensRef, "")
	if err != nil {
		return err
	}

	tokens := *tokensRef
	tokens = append(tokens, openSubclauseTokenInstance)
	*tokensRef = tokens
	return nil
}

func (q *abstractDocumentQuery) closeSubclause() error {
	q.currentClauseDepth--

	tokensRef, err := q.getCurrentWhereTokensRef()
	if err != nil {
		return err
	}
	tokens := *tokensRef
	tokens = append(tokens, closeSubclauseTokenInstance)
	*tokensRef = tokens
	return nil
}

func (q *abstractDocumentQuery) whereEquals(fieldName string, value interface{}) error {
	params := &whereParams{
		fieldName: fieldName,
		value:     value,
	}
	return q.whereEqualsWithParams(params)
}

func (q *abstractDocumentQuery) whereEqualsWithMethodCall(fieldName string, method MethodCall) error {
	return q.whereEquals(fieldName, method)
}

func (q *abstractDocumentQuery) whereEqualsWithParams(whereParams *whereParams) error {
	if q.negate {
		q.negate = false
		return q.whereNotEqualsWithParams(whereParams)
	}

	var err error
	whereParams.fieldName, err = q.ensureValidFieldName(whereParams.fieldName, whereParams.isNestedPath)
	if err != nil {
		return err
	}

	tokensRef, err := q.getCurrentWhereTokensRef()
	if err != nil {
		return err
	}
	err = q.appendOperatorIfNeeded(tokensRef)
	if err != nil {
		return err
	}

	if q.ifValueIsMethod(whereOperatorEquals, whereParams, tokensRef) {
		return nil
	}

	transformToEqualValue := q.transformValue(whereParams)
	addQueryParameter := q.addQueryParameter(transformToEqualValue)
	whereToken := createWhereTokenWithOptions(whereOperatorEquals, whereParams.fieldName, addQueryParameter, newWhereOptionsWithExact(whereParams.isExact))

	tokens := *tokensRef
	tokens = append(tokens, whereToken)
	*tokensRef = tokens
	return nil
}

func (q *abstractDocumentQuery) ifValueIsMethod(op whereOperator, whereParams *whereParams, tokensRef *[]queryToken) bool {
	if mc, ok := whereParams.value.(*CmpXchg); ok {
		n := len(mc.args)
		args := make([]string, n)
		for i := 0; i < n; i++ {
			args[i] = q.addQueryParameter(mc.args[i])
		}

		opts := newWhereOptionsWithMethod(MethodsTypeCmpXChg, args, mc.accessPath, whereParams.isExact)
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

func (q *abstractDocumentQuery) whereNotEquals(fieldName string, value interface{}) error {
	params := &whereParams{
		fieldName: fieldName,
		value:     value,
	}

	return q.whereNotEqualsWithParams(params)
}

func (q *abstractDocumentQuery) whereNotEqualsWithMethod(fieldName string, method MethodCall) error {
	return q.whereNotEquals(fieldName, method)
}

func (q *abstractDocumentQuery) whereNotEqualsWithParams(whereParams *whereParams) error {
	if q.negate {
		q.negate = false
		return q.whereEqualsWithParams(whereParams)
	}

	transformToEqualValue := q.transformValue(whereParams)

	tokensRef, err := q.getCurrentWhereTokensRef()
	if err != nil {
		return err
	}
	err = q.appendOperatorIfNeeded(tokensRef)
	if err != nil {
		return err
	}

	whereParams.fieldName, err = q.ensureValidFieldName(whereParams.fieldName, whereParams.isNestedPath)
	if err != nil {
		return err
	}

	if q.ifValueIsMethod(whereOperatorNotEquals, whereParams, tokensRef) {
		return nil
	}

	whereToken := createWhereTokenWithOptions(whereOperatorNotEquals, whereParams.fieldName, q.addQueryParameter(transformToEqualValue), newWhereOptionsWithExact(whereParams.isExact))
	tokens := *tokensRef
	tokens = append(tokens, whereToken)
	*tokensRef = tokens
	return nil
}

func (q *abstractDocumentQuery) negateNext() {
	q.negate = !q.negate
}

// mark last created token as exact. only applies to select number of tokens.
// it allows fluid APIs like .Where().Exact()
// will panic if last token wasn't of compatible type as that is considered
// invalid use of API and returning an error would break fluid API
func (q *abstractDocumentQuery) markLastTokenExact() error {
	tokensRef, err := q.getCurrentWhereTokensRef()
	if err != nil {
		return err
	}
	tokens := *tokensRef
	n := len(tokens)
	lastToken := tokens[n-1]
	switch tok := lastToken.(type) {
	case *whereToken:
		if tok.options == nil {
			tok.options = newWhereOptionsWithExact(true)
		} else {
			tok.options.exact = true
		}
	default:
		return newIllegalStateError("expected whereToken, got %T", lastToken)
	}

	*tokensRef = tokens
	return nil
}

func (q *abstractDocumentQuery) whereIn(fieldName string, values []interface{}) error {
	var err error
	fieldName, err = q.ensureValidFieldName(fieldName, false)
	if err != nil {
		return err
	}

	tokensRef, err := q.getCurrentWhereTokensRef()
	if err != nil {
		return err
	}
	err = q.appendOperatorIfNeeded(tokensRef)
	if err != nil {
		return err
	}
	err = q.negateIfNeeded(tokensRef, fieldName)
	if err != nil {
		return err
	}

	whereToken := createWhereToken(whereOperatorIn, fieldName, q.addQueryParameter(q.transformCollection(fieldName, abstractDocumentQueryUnpackCollection(values))))

	tokens := *tokensRef
	tokens = append(tokens, whereToken)
	*tokensRef = tokens
	return nil
}

func (q *abstractDocumentQuery) whereStartsWith(fieldName string, value interface{}) error {
	whereParams := &whereParams{
		fieldName:      fieldName,
		value:          value,
		allowWildcards: true,
	}

	transformToEqualValue := q.transformValue(whereParams)

	tokensRef, err := q.getCurrentWhereTokensRef()
	if err != nil {
		return err
	}
	err = q.appendOperatorIfNeeded(tokensRef)
	if err != nil {
		return err
	}

	whereParams.fieldName, err = q.ensureValidFieldName(whereParams.fieldName, whereParams.isNestedPath)
	if err != nil {
		return err
	}
	err = q.negateIfNeeded(tokensRef, whereParams.fieldName)
	if err != nil {
		return err
	}

	whereToken := createWhereToken(whereOperatorStartsWith, whereParams.fieldName, q.addQueryParameter(transformToEqualValue))

	tokens := *tokensRef
	tokens = append(tokens, whereToken)
	*tokensRef = tokens
	return nil
}

func (q *abstractDocumentQuery) whereEndsWith(fieldName string, value interface{}) error {
	whereParams := &whereParams{
		fieldName:      fieldName,
		value:          value,
		allowWildcards: true,
	}

	transformToEqualValue := q.transformValue(whereParams)

	tokensRef, err := q.getCurrentWhereTokensRef()
	if err != nil {
		return err
	}
	err = q.appendOperatorIfNeeded(tokensRef)
	if err != nil {
		return err
	}

	whereParams.fieldName, err = q.ensureValidFieldName(whereParams.fieldName, whereParams.isNestedPath)
	if err != nil {
		return err
	}
	err = q.negateIfNeeded(tokensRef, whereParams.fieldName)
	if err != nil {
		return err
	}

	whereToken := createWhereToken(whereOperatorEndsWith, whereParams.fieldName, q.addQueryParameter(transformToEqualValue))

	tokens := *tokensRef
	tokens = append(tokens, whereToken)
	*tokensRef = tokens
	return nil
}

func (q *abstractDocumentQuery) whereBetween(fieldName string, start interface{}, end interface{}) error {
	var err error
	fieldName, err = q.ensureValidFieldName(fieldName, false)
	if err != nil {
		return err
	}

	tokensRef, err := q.getCurrentWhereTokensRef()
	if err != nil {
		return err
	}
	err = q.appendOperatorIfNeeded(tokensRef)
	if err != nil {
		return err
	}
	err = q.negateIfNeeded(tokensRef, fieldName)
	if err != nil {
		return err
	}

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
	if end != nil {
		toParam = q.transformValueWithRange(endParams, true)
	}
	toParameterName := q.addQueryParameter(toParam)

	whereToken := createWhereTokenWithOptions(whereOperatorBetween, fieldName, "", newWhereOptionsWithFromTo(false, fromParameterName, toParameterName))

	tokens := *tokensRef
	tokens = append(tokens, whereToken)
	*tokensRef = tokens
	return nil
}

func (q *abstractDocumentQuery) whereGreaterThan(fieldName string, value interface{}) error {
	var err error
	fieldName, err = q.ensureValidFieldName(fieldName, false)
	if err != nil {
		return err
	}

	tokensRef, err := q.getCurrentWhereTokensRef()
	if err != nil {
		return err
	}
	err = q.appendOperatorIfNeeded(tokensRef)
	if err != nil {
		return err
	}
	err = q.negateIfNeeded(tokensRef, fieldName)
	if err != nil {
		return err
	}

	whereParams := &whereParams{
		value:     value,
		fieldName: fieldName,
	}

	paramValue := interface{}("*")
	if value != nil {
		paramValue = q.transformValueWithRange(whereParams, true)
	}
	parameter := q.addQueryParameter(paramValue)

	whereToken := createWhereTokenWithOptions(whereOperatorGreaterThan, fieldName, parameter, nil)

	tokens := *tokensRef
	tokens = append(tokens, whereToken)
	*tokensRef = tokens
	return nil
}

func (q *abstractDocumentQuery) whereGreaterThanOrEqual(fieldName string, value interface{}) error {
	var err error
	fieldName, err = q.ensureValidFieldName(fieldName, false)
	if err != nil {
		return err
	}

	tokensRef, err := q.getCurrentWhereTokensRef()
	if err != nil {
		return err
	}
	err = q.appendOperatorIfNeeded(tokensRef)
	if err != nil {
		return err
	}
	err = q.negateIfNeeded(tokensRef, fieldName)
	if err != nil {
		return err
	}

	whereParams := &whereParams{
		value:     value,
		fieldName: fieldName,
	}

	paramValue := interface{}("*")
	if value != nil {
		paramValue = q.transformValueWithRange(whereParams, true)
	}

	parameter := q.addQueryParameter(paramValue)

	whereToken := createWhereTokenWithOptions(whereOperatorGreaterThanOrEqual, fieldName, parameter, nil)

	tokens := *tokensRef
	tokens = append(tokens, whereToken)
	*tokensRef = tokens
	return nil
}

func (q *abstractDocumentQuery) whereLessThan(fieldName string, value interface{}) error {
	var err error
	fieldName, err = q.ensureValidFieldName(fieldName, false)
	if err != nil {
		return err
	}

	tokensRef, err := q.getCurrentWhereTokensRef()
	if err != nil {
		return err
	}
	err = q.appendOperatorIfNeeded(tokensRef)
	if err != nil {
		return err
	}
	err = q.negateIfNeeded(tokensRef, fieldName)
	if err != nil {
		return err
	}

	whereParams := &whereParams{
		value:     value,
		fieldName: fieldName,
	}

	paramValue := interface{}("NULL")
	if value != nil {
		paramValue = q.transformValueWithRange(whereParams, true)
	}
	parameter := q.addQueryParameter(paramValue)
	whereToken := createWhereTokenWithOptions(whereOperatorLessThan, fieldName, parameter, nil)

	tokens := *tokensRef
	tokens = append(tokens, whereToken)
	*tokensRef = tokens
	return nil
}

func (q *abstractDocumentQuery) whereLessThanOrEqual(fieldName string, value interface{}) error {
	tokensRef, err := q.getCurrentWhereTokensRef()
	if err != nil {
		return err
	}
	err = q.appendOperatorIfNeeded(tokensRef)
	if err != nil {
		return err
	}
	err = q.negateIfNeeded(tokensRef, fieldName)
	if err != nil {
		return err
	}

	whereParams := &whereParams{
		value:     value,
		fieldName: fieldName,
	}

	paramValue := interface{}("NULL")
	if value != nil {
		paramValue = q.transformValueWithRange(whereParams, true)
	}
	parameter := q.addQueryParameter(paramValue)
	whereToken := createWhereTokenWithOptions(whereOperatorLessThanOrEqual, fieldName, parameter, nil)

	tokens := *tokensRef
	tokens = append(tokens, whereToken)
	*tokensRef = tokens
	return nil
}

func (q *abstractDocumentQuery) whereRegex(fieldName string, pattern string) error {
	tokensRef, err := q.getCurrentWhereTokensRef()
	if err != nil {
		return err
	}
	err = q.appendOperatorIfNeeded(tokensRef)
	if err != nil {
		return err
	}
	err = q.negateIfNeeded(tokensRef, fieldName)
	if err != nil {
		return err
	}

	whereParams := &whereParams{
		value:     pattern,
		fieldName: fieldName,
	}

	parameter := q.addQueryParameter(q.transformValue(whereParams))

	whereToken := createWhereToken(whereOperatorRegex, fieldName, parameter)

	tokens := *tokensRef
	tokens = append(tokens, whereToken)
	*tokensRef = tokens
	return nil
}

func (q *abstractDocumentQuery) andAlso() error {
	tokensRef, err := q.getCurrentWhereTokensRef()
	if err != nil {
		return err
	}
	tokens := *tokensRef

	n := len(tokens)
	if n == 0 {
		return nil
	}

	lastToken := tokens[n-1]
	if _, ok := lastToken.(*queryOperatorToken); ok {
		return newIllegalStateError("Cannot add AND, previous token was already an operator token.")
	}

	tokens = append(tokens, queryOperatorTokenAnd)
	*tokensRef = tokens
	return nil
}

func (q *abstractDocumentQuery) orElse() error {
	tokensRef, err := q.getCurrentWhereTokensRef()
	if err != nil {
		return err
	}
	tokens := *tokensRef
	n := len(tokens)
	if n == 0 {
		return nil
	}

	lastToken := tokens[n-1]
	if _, ok := lastToken.(*queryOperatorToken); ok {
		return newIllegalStateError("Cannot add OR, previous token was already an operator token.")
	}

	tokens = append(tokens, queryOperatorTokenOr)
	*tokensRef = tokens
	return nil
}

func (q *abstractDocumentQuery) boost(boost float64) error {
	if boost == 1.0 {
		return nil
	}

	tokens, err := q.getCurrentWhereTokens()
	if err != nil {
		return err
	}
	n := len(tokens)
	if n == 0 {
		return newIllegalStateError("Missing where clause")
	}

	maybeWhereToken := tokens[n-1]
	whereToken, ok := maybeWhereToken.(*whereToken)
	if !ok {
		return newIllegalStateError("Missing where clause")
	}

	if boost <= 0.0 {
		return newIllegalArgumentError("Boost factor must be a positive number")
	}

	whereToken.options.boost = boost
	return nil
}

func (q *abstractDocumentQuery) fuzzy(fuzzy float64) error {
	tokens, err := q.getCurrentWhereTokens()
	if err != nil {
		return err
	}
	n := len(tokens)
	if n == 0 {
		return newIllegalStateError("Missing where clause")
	}

	maybeWhereToken := tokens[n-1]
	whereToken, ok := maybeWhereToken.(*whereToken)
	if !ok {
		return newIllegalStateError("Missing where clause")
	}

	if fuzzy < 0.0 || fuzzy > 1.0 {
		return newIllegalArgumentError("Fuzzy distance must be between 0.0 and 1.0")
	}

	whereToken.options.fuzzy = fuzzy
	return nil
}

func (q *abstractDocumentQuery) proximity(proximity int) error {
	tokens, err := q.getCurrentWhereTokens()
	if err != nil {
		return err
	}

	n := len(tokens)
	if n == 0 {
		return newIllegalStateError("Missing where clause")
	}

	maybeWhereToken := tokens[n-1]
	whereToken, ok := maybeWhereToken.(*whereToken)
	if !ok {
		return newIllegalStateError("Missing where clause")
	}

	if proximity < 1 {
		return newIllegalArgumentError("Proximity distance must be a positive number")
	}

	whereToken.options.proximity = proximity
	return nil
}

func (q *abstractDocumentQuery) orderBy(field string) error {
	return q.orderByWithOrdering(field, OrderingTypeString)
}

func (q *abstractDocumentQuery) orderByWithOrdering(field string, ordering OrderingType) error {
	if err := q.assertNoRawQuery(); err != nil {
		return err
	}
	f, err := q.ensureValidFieldName(field, false)
	if err != nil {
		return err
	}
	q.orderByTokens = append(q.orderByTokens, orderByTokenCreateAscending(f, ordering))
	return nil
}

func (q *abstractDocumentQuery) orderByDescending(field string) error {
	return q.orderByDescendingWithOrdering(field, OrderingTypeString)
}

func (q *abstractDocumentQuery) orderByDescendingWithOrdering(field string, ordering OrderingType) error {
	if err := q.assertNoRawQuery(); err != nil {
		return err
	}
	f, err := q.ensureValidFieldName(field, false)
	if err != nil {
		return err
	}
	q.orderByTokens = append(q.orderByTokens, orderByTokenCreateDescending(f, ordering))
	return nil
}

func (q *abstractDocumentQuery) orderByScore() error {
	if err := q.assertNoRawQuery(); err != nil {
		return err
	}

	q.orderByTokens = append(q.orderByTokens, orderByTokenScoreAscending)
	return nil
}

func (q *abstractDocumentQuery) orderByScoreDescending() error {
	if err := q.assertNoRawQuery(); err != nil {
		return err
	}
	q.orderByTokens = append(q.orderByTokens, orderByTokenScoreDescending)
	return nil
}

func (q *abstractDocumentQuery) statistics(stats **QueryStatistics) {
	*stats = q.queryStats
}

func (q *abstractDocumentQuery) invokeAfterQueryExecuted(result *QueryResult) {
	for _, cb := range q.afterQueryExecutedCallback {
		if cb != nil {
			cb(result)
		}
	}
}

func (q *abstractDocumentQuery) invokeBeforeQueryExecuted(query *IndexQuery) {
	for _, cb := range q.beforeQueryExecutedCallback {
		if cb != nil {
			cb(query)
		}
	}
}

func (q *abstractDocumentQuery) invokeAfterStreamExecuted(result map[string]interface{}) {
	for _, cb := range q.afterStreamExecutedCallback {
		if cb != nil {
			cb(result)
		}
	}
}

func (q *abstractDocumentQuery) generateIndexQuery(query string) *IndexQuery {
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

func (q *abstractDocumentQuery) search(fieldName string, searchTerms string) error {
	return q.searchWithOperator(fieldName, searchTerms, SearchOperatorOr)
}

func (q *abstractDocumentQuery) searchWithOperator(fieldName string, searchTerms string, operator SearchOperator) error {
	tokensRef, err := q.getCurrentWhereTokensRef()
	if err != nil {
		return err
	}
	err = q.appendOperatorIfNeeded(tokensRef)
	if err != nil {
		return err
	}

	fieldName, err = q.ensureValidFieldName(fieldName, false)
	if err != nil {
		return err
	}
	err = q.negateIfNeeded(tokensRef, fieldName)
	if err != nil {
		return err
	}

	whereToken := createWhereTokenWithOptions(whereOperatorSearch, fieldName, q.addQueryParameter(searchTerms), newWhereOptionsWithOperator(operator))

	tokens := *tokensRef
	tokens = append(tokens, whereToken)
	*tokensRef = tokens
	return nil
}

func (q *abstractDocumentQuery) string() (string, error) {
	if q.queryRaw != "" {
		return q.queryRaw, nil
	}

	if q.currentClauseDepth != 0 {
		return "", newIllegalStateError("A clause was not closed correctly within this query, current clause depth = %d", q.currentClauseDepth)
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

	return queryText.String(), nil
}

func (q *abstractDocumentQuery) buildInclude(queryText *strings.Builder) {
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

func (q *abstractDocumentQuery) intersect() error {

	tokensRef, err := q.getCurrentWhereTokensRef()
	if err != nil {
		return err
	}
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
			return nil
		}
	}

	return newIllegalStateError("Cannot add INTERSECT at this point.")
}

func (q *abstractDocumentQuery) whereExists(fieldName string) error {
	var err error
	fieldName, err = q.ensureValidFieldName(fieldName, false)
	if err != nil {
		return err
	}

	tokensRef, err := q.getCurrentWhereTokensRef()
	if err != nil {
		return err
	}
	err = q.appendOperatorIfNeeded(tokensRef)
	if err != nil {
		return err
	}
	err = q.negateIfNeeded(tokensRef, fieldName)
	if err != nil {
		return err
	}

	tokens := *tokensRef
	tokens = append(tokens, createWhereToken(whereOperatorExists, fieldName, ""))
	*tokensRef = tokens
	return nil
}

func (q *abstractDocumentQuery) containsAny(fieldName string, values []interface{}) error {
	var err error
	fieldName, err = q.ensureValidFieldName(fieldName, false)
	if err != nil {
		return err
	}

	tokensRef, err := q.getCurrentWhereTokensRef()
	if err != nil {
		return err
	}
	err = q.appendOperatorIfNeeded(tokensRef)
	if err != nil {
		return err
	}
	err = q.negateIfNeeded(tokensRef, fieldName)
	if err != nil {
		return err
	}

	array := q.transformCollection(fieldName, abstractDocumentQueryUnpackCollection(values))
	whereToken := createWhereTokenWithOptions(whereOperatorIn, fieldName, q.addQueryParameter(array), newWhereOptionsWithExact(false))

	tokens := *tokensRef
	tokens = append(tokens, whereToken)
	*tokensRef = tokens
	return nil
}

func (q *abstractDocumentQuery) containsAll(fieldName string, values []interface{}) error {
	var err error
	fieldName, err = q.ensureValidFieldName(fieldName, false)
	if err != nil {
		return err
	}

	tokensRef, err := q.getCurrentWhereTokensRef()
	if err != nil {
		return err
	}
	err = q.appendOperatorIfNeeded(tokensRef)
	if err != nil {
		return err
	}
	err = q.negateIfNeeded(tokensRef, fieldName)
	if err != nil {
		return err
	}

	array := q.transformCollection(fieldName, abstractDocumentQueryUnpackCollection(values))

	tokens := *tokensRef
	if len(array) == 0 {
		tokens = append(tokens, trueTokenInstance)
	} else {
		whereToken := createWhereToken(whereOperatorAllIn, fieldName, q.addQueryParameter(array))
		tokens = append(tokens, whereToken)
	}
	*tokensRef = tokens
	return nil
}

func (q *abstractDocumentQuery) distinct() error {
	if q.isDistinct() {
		return newIllegalStateError("The is already a distinct query")
	}

	if len(q.selectTokens) == 0 {
		q.selectTokens = []queryToken{distinctTokenInstance}
		return nil
	}
	q.selectTokens = append([]queryToken{distinctTokenInstance}, q.selectTokens...)
	return nil
}

func (q *abstractDocumentQuery) updateStatsAndHighlightings(queryResult *QueryResult) {
	q.queryStats.UpdateQueryStats(queryResult)
	//TBD 4.1 Highlightings.Update(queryResult);
}

func (q *abstractDocumentQuery) buildSelect(writer *strings.Builder) {
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

func (q *abstractDocumentQuery) buildFrom(writer *strings.Builder) {
	q.fromToken.writeTo(writer)
}

func (q *abstractDocumentQuery) buildDeclare(writer *strings.Builder) {
	if q.declareToken != nil {
		q.declareToken.writeTo(writer)
	}
}

func (q *abstractDocumentQuery) buildLoad(writer *strings.Builder) {
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

func (q *abstractDocumentQuery) buildWhere(writer *strings.Builder) {
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

func (q *abstractDocumentQuery) buildGroupBy(writer *strings.Builder) {
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

func (q *abstractDocumentQuery) buildOrderBy(writer *strings.Builder) {
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

func (q *abstractDocumentQuery) appendOperatorIfNeeded(tokensRef *[]queryToken) error {
	tokens := *tokensRef
	if err := q.assertNoRawQuery(); err != nil {
		return err
	}

	n := len(tokens)
	if len(tokens) == 0 {
		return nil
	}

	lastToken := tokens[n-1]
	_, isWhereToken := lastToken.(*whereToken)
	_, isCloseSubclauseToken := lastToken.(*closeSubclauseToken)
	if !isWhereToken && !isCloseSubclauseToken {
		return nil
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

	if lastWhere != nil && lastWhere.options.searchOperator != SearchOperatorUnset {
		token = queryOperatorTokenOr // default to OR operator after search if AND was not specified explicitly
	}

	tokens = append(tokens, token)
	*tokensRef = tokens
	return nil
}

func (q *abstractDocumentQuery) transformCollection(fieldName string, values []interface{}) []interface{} {
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

func (q *abstractDocumentQuery) negateIfNeeded(tokensRef *[]queryToken, fieldName string) error {
	var err error
	if !q.negate {
		return nil
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
			err = q.whereExists(fieldName)
			if err != nil {
				return err
			}
		} else {
			err = q.whereTrue()
			if err != nil {
				return err
			}
		}
		err = q.andAlso()
		if err != nil {
			return err
		}
	}

	tokens = append(tokens, negateTokenInstance)
	*tokensRef = tokens
	return nil
}

func abstractDocumentQueryUnpackCollection(items []interface{}) []interface{} {
	var results []interface{}

	for _, item := range items {
		if itemCollection, ok := item.([]interface{}); ok {
			els := abstractDocumentQueryUnpackCollection(itemCollection)
			results = append(results, els...)
		} else {
			results = append(results, item)
		}
	}

	return results
}

func (q *abstractDocumentQuery) ensureValidFieldName(fieldName string, isNestedPath bool) (string, error) {
	if q.theSession == nil || q.theSession.GetConventions() == nil || isNestedPath || q.isGroupBy {
		return queryFieldUtilEscapeIfNecessary(fieldName), nil
	}

	if fieldName == documentConventionsIdentityPropertyName {
		return IndexingFieldNameDocumentID, nil
	}

	return queryFieldUtilEscapeIfNecessary(fieldName), nil
}

func (q *abstractDocumentQuery) transformValue(whereParams *whereParams) interface{} {
	return q.transformValueWithRange(whereParams, false)
}

func (q *abstractDocumentQuery) transformValueWithRange(whereParams *whereParams, forRange bool) interface{} {
	if whereParams.value == nil {
		return nil
	}

	if "" == whereParams.value {
		return ""
	}

	var stringValueReference string
	if q.conventions.TryConvertValueForQuery(whereParams.fieldName, whereParams.value, forRange, &stringValueReference) {
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

func (q *abstractDocumentQuery) addQueryParameter(value interface{}) string {
	parameterName := "p" + strconv.Itoa(len(q.queryParameters))
	q.queryParameters[parameterName] = value
	return parameterName
}

func (q *abstractDocumentQuery) getCurrentWhereTokens() ([]queryToken, error) {
	if !q.isInMoreLikeThis {
		return q.whereTokens, nil
	}

	n := len(q.whereTokens)

	if n == 0 {
		return nil, newIllegalStateError("Cannot get moreLikeThisToken because there are no where token specified.")
	}

	lastToken := q.whereTokens[n-1]

	if moreLikeThisToken, ok := lastToken.(*moreLikeThisToken); ok {
		return moreLikeThisToken.whereTokens, nil
	}
	return nil, newIllegalStateError("Last token is not moreLikeThisToken")
}

func (q *abstractDocumentQuery) getCurrentWhereTokensRef() (*[]queryToken, error) {
	if !q.isInMoreLikeThis {
		return &q.whereTokens, nil
	}

	n := len(q.whereTokens)

	if n == 0 {
		return nil, newIllegalStateError("Cannot get moreLikeThisToken because there are no where token specified.")
	}

	lastToken := q.whereTokens[n-1]

	if moreLikeThisToken, ok := lastToken.(*moreLikeThisToken); ok {
		return &moreLikeThisToken.whereTokens, nil
	}
	return nil, newIllegalStateError("Last token is not moreLikeThisToken")
}

func (q *abstractDocumentQuery) updateFieldsToFetchToken(fieldsToFetch *fieldsToFetchToken) {
	q.fieldsToFetchToken = fieldsToFetch

	if len(q.selectTokens) == 0 {
		q.selectTokens = append(q.selectTokens, fieldsToFetch)
		return
	}

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

func getSourceAliasIfExists(clazz reflect.Type, queryData *queryData, fields []string) string {
	if len(fields) != 1 || fields[0] == "" {
		return ""
	}

	if clazz != reflect.TypeOf("") && !isPrimitiveOrWrapper(clazz) {
		return ""
	}
	indexOf := strings.Index(fields[0], ".")
	if indexOf == -1 {
		return ""
	}

	possibleAlias := fields[0][:indexOf]
	if queryData.fromAlias == possibleAlias {
		return possibleAlias
	}

	if len(queryData.loadTokens) == 0 {
		return ""
	}

	// TODO: is this the logic?
	for _, x := range queryData.loadTokens {
		if x.alias == possibleAlias {
			return possibleAlias
		}
	}
	return ""
}

func (q *abstractDocumentQuery) addBeforeQueryExecutedListener(action func(*IndexQuery)) int {
	q.beforeQueryExecutedCallback = append(q.beforeQueryExecutedCallback, action)
	return len(q.beforeQueryExecutedCallback) - 1
}

func (q *abstractDocumentQuery) removeBeforeQueryExecutedListener(idx int) {
	q.beforeQueryExecutedCallback[idx] = nil
}

func (q *abstractDocumentQuery) addAfterQueryExecutedListener(action func(*QueryResult)) int {
	q.afterQueryExecutedCallback = append(q.afterQueryExecutedCallback, action)
	return len(q.afterQueryExecutedCallback) - 1
}

func (q *abstractDocumentQuery) removeAfterQueryExecutedListener(idx int) {
	q.afterQueryExecutedCallback[idx] = nil
}

func (q *abstractDocumentQuery) addAfterStreamExecutedListener(action func(map[string]interface{})) int {
	q.afterStreamExecutedCallback = append(q.afterStreamExecutedCallback, action)
	return len(q.afterStreamExecutedCallback) - 1
}

func (q *abstractDocumentQuery) removeAfterStreamExecutedListener(idx int) {
	q.afterStreamExecutedCallback[idx] = nil
}

func (q *abstractDocumentQuery) noTracking() {
	q.disableEntitiesTracking = true
}

func (q *abstractDocumentQuery) noCaching() {
	q.disableCaching = true
}

func (q *abstractDocumentQuery) withinRadiusOf(fieldName string, radius float64, latitude float64, longitude float64, radiusUnits SpatialUnits, distErrorPercent float64) error {
	var err error
	fieldName, err = q.ensureValidFieldName(fieldName, false)
	if err != nil {
		return err
	}

	tokensRef, err := q.getCurrentWhereTokensRef()
	if err != nil {
		return err
	}
	err = q.appendOperatorIfNeeded(tokensRef)
	if err != nil {
		return err
	}
	err = q.negateIfNeeded(tokensRef, fieldName)
	if err != nil {
		return err
	}

	shape := ShapeTokenCircle(q.addQueryParameter(radius), q.addQueryParameter(latitude), q.addQueryParameter(longitude), radiusUnits)
	opts := newWhereOptionsWithTokenAndDistance(shape, distErrorPercent)
	whereToken := createWhereTokenWithOptions(whereOperatorSpatialWithin, fieldName, "", opts)

	tokens := *tokensRef
	tokens = append(tokens, whereToken)
	*tokensRef = tokens
	return nil
}

func (q *abstractDocumentQuery) spatial(fieldName string, shapeWkt string, relation SpatialRelation, distErrorPercent float64) error {
	var err error
	fieldName, err = q.ensureValidFieldName(fieldName, false)
	if err != nil {
		return err
	}

	tokensRef, err := q.getCurrentWhereTokensRef()
	if err != nil {
		return err
	}
	err = q.appendOperatorIfNeeded(tokensRef)
	if err != nil {
		return err
	}
	err = q.negateIfNeeded(tokensRef, fieldName)
	if err != nil {
		return err
	}

	wktToken := ShapeTokenWkt(q.addQueryParameter(shapeWkt))

	var whereOperator whereOperator
	switch relation {
	case SpatialRelationWithin:
		whereOperator = whereOperatorSpatialWithin
	case SpatialRelationContains:
		whereOperator = whereOperatorSpatialContains
	case SpatialRelationDisjoin:
		whereOperator = whereOperatorSpatialDisjoint
	case SpatialRelationIntersects:
		whereOperator = whereOperatorSpatialIntersects
	default:
		return newIllegalArgumentError("unknown relation %s", relation)
	}

	tokens := *tokensRef
	opts := newWhereOptionsWithTokenAndDistance(wktToken, distErrorPercent)
	tok := createWhereTokenWithOptions(whereOperator, fieldName, "", opts)
	tokens = append(tokens, tok)
	*tokensRef = tokens
	return nil
}

func (q *abstractDocumentQuery) spatial2(dynamicField DynamicSpatialField, criteria SpatialCriteria) error {
	if err := q.assertIsDynamicQuery(dynamicField, "spatial"); err != nil {
		return err
	}

	tokensRef, err := q.getCurrentWhereTokensRef()
	if err != nil {
		return err
	}
	err = q.appendOperatorIfNeeded(tokensRef)
	if err != nil {
		return err
	}
	err = q.negateIfNeeded(tokensRef, "")
	if err != nil {
		return err
	}

	ensure := func(fieldName string, isNestedPath bool) string {
		// TODO: propagate error
		s, _ := q.ensureValidFieldName(fieldName, isNestedPath)
		return s
	}
	add := func(value interface{}) string {
		return q.addQueryParameter(value)
	}
	tok := criteria.ToQueryToken(dynamicField.ToField(ensure), add)
	tokens := *tokensRef
	tokens = append(tokens, tok)
	*tokensRef = tokens
	return nil
}

func (q *abstractDocumentQuery) spatial3(fieldName string, criteria SpatialCriteria) error {
	var err error
	fieldName, err = q.ensureValidFieldName(fieldName, false)
	if err != nil {
		return err
	}

	tokensRef, err := q.getCurrentWhereTokensRef()
	if err != nil {
		return err
	}
	err = q.appendOperatorIfNeeded(tokensRef)
	if err != nil {
		return err
	}
	err = q.negateIfNeeded(tokensRef, fieldName)
	if err != nil {
		return err
	}

	tokens := *tokensRef
	add := func(value interface{}) string {
		return q.addQueryParameter(value)
	}
	tok := criteria.ToQueryToken(fieldName, add)
	tokens = append(tokens, tok)
	*tokensRef = tokens
	return nil
}

func (q *abstractDocumentQuery) orderByDistance(field DynamicSpatialField, latitude float64, longitude float64) error {
	if field == nil {
		return newIllegalArgumentError("Field cannot be null")
	}
	if err := q.assertIsDynamicQuery(field, "orderByDistance"); err != nil {
		return err
	}

	ensure := func(fieldName string, isNestedPath bool) string {
		// TODO: propagate error
		s, _ := q.ensureValidFieldName(fieldName, isNestedPath)
		return s
	}

	return q.orderByDistanceLatLong("'"+field.ToField(ensure)+"'", latitude, longitude)
}

func (q *abstractDocumentQuery) orderByDistanceLatLong(fieldName string, latitude float64, longitude float64) error {
	tok := orderByTokenCreateDistanceAscending(fieldName, q.addQueryParameter(latitude), q.addQueryParameter(longitude))
	q.orderByTokens = append(q.orderByTokens, tok)
	return nil
}

func (q *abstractDocumentQuery) orderByDistance2(field DynamicSpatialField, shapeWkt string) error {
	if field == nil {
		return newIllegalArgumentError("Field cannot be null")
	}
	err := q.assertIsDynamicQuery(field, "orderByDistance2")
	if err != nil {
		return err
	}

	ensure := func(fieldName string, isNestedPath bool) string {
		// TODO: propagate error
		s, _ := q.ensureValidFieldName(fieldName, isNestedPath)
		return s
	}
	return q.orderByDistance3("'"+field.ToField(ensure)+"'", shapeWkt)
}

func (q *abstractDocumentQuery) orderByDistance3(fieldName string, shapeWkt string) error {
	tok := orderByTokenCreateDistanceAscending2(fieldName, q.addQueryParameter(shapeWkt))
	q.orderByTokens = append(q.orderByTokens, tok)
	return nil
}

func (q *abstractDocumentQuery) orderByDistanceDescending(field DynamicSpatialField, latitude float64, longitude float64) error {
	if field == nil {
		return newIllegalArgumentError("Field cannot be null")
	}
	err := q.assertIsDynamicQuery(field, "orderByDistanceDescending")
	if err != nil {
		return err
	}
	ensure := func(fieldName string, isNestedPath bool) string {
		// TODO: propagate error
		s, _ := q.ensureValidFieldName(fieldName, isNestedPath)
		return s
	}
	return q.orderByDistanceDescendingLatLong("'"+field.ToField(ensure)+"'", latitude, longitude)
}

func (q *abstractDocumentQuery) orderByDistanceDescendingLatLong(fieldName string, latitude float64, longitude float64) error {
	tok := orderByTokenCreateDistanceDescending(fieldName, q.addQueryParameter(latitude), q.addQueryParameter(longitude))
	q.orderByTokens = append(q.orderByTokens, tok)
	return nil
}

func (q *abstractDocumentQuery) orderByDistanceDescending2(field DynamicSpatialField, shapeWkt string) error {
	if field == nil {
		return newIllegalArgumentError("Field cannot be null")
	}
	err := q.assertIsDynamicQuery(field, "orderByDistanceDescending2")
	if err != nil {
		return err
	}
	ensure := func(fieldName string, isNestedPath bool) string {
		// TODO: propagate error
		s, _ := q.ensureValidFieldName(fieldName, isNestedPath)
		return s
	}
	return q.orderByDistanceDescending3("'"+field.ToField(ensure)+"'", shapeWkt)
}

func (q *abstractDocumentQuery) orderByDistanceDescending3(fieldName string, shapeWkt string) error {
	tok := orderByTokenCreateDistanceDescending2(fieldName, q.addQueryParameter(shapeWkt))
	q.orderByTokens = append(q.orderByTokens, tok)
	return nil
}

func (q *abstractDocumentQuery) assertIsDynamicQuery(dynamicField DynamicSpatialField, methodName string) error {
	if q.fromToken != nil && !q.fromToken.isDynamic {
		f := func(s string, f bool) string {
			// TODO: propgate error
			s, _ = q.ensureValidFieldName(s, f)
			return s
		}
		fld := dynamicField.ToField(f)
		return newIllegalStateError("Cannot execute query method '" + methodName + "'. Field '" + fld + "' cannot be used when static index '" + q.fromToken.indexName + "' is queried. Dynamic spatial fields can only be used with dynamic queries, " + "for static index queries please use valid spatial fields defined in index definition.")
	}
	return nil
}

func (q *abstractDocumentQuery) initSync() error {
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
	q.theSession.onBeforeQueryInvoke(beforeQueryEventArgs)

	var err error
	q.queryOperation, err = q.initializeQueryOperation()
	if err != nil {
		return err
	}
	return q.executeActualQuery()
}

func (q *abstractDocumentQuery) executeActualQuery() error {
	{
		context := q.queryOperation.enterQueryContext()
		defer func() {
			_ = context.Close()
		}()

		command, err := q.queryOperation.createRequest()
		if err != nil {
			return err
		}
		if err = q.theSession.GetRequestExecutor().ExecuteCommand(command, q.theSession.sessionInfo); err != nil {
			return err
		}
		if err = q.queryOperation.setResult(command.Result); err != nil {
			return err
		}
	}
	q.invokeAfterQueryExecuted(q.queryOperation.currentQueryResults)
	return nil
}

// GetQueryResult returns results of a query
func (q *abstractDocumentQuery) getQueryResult() (*QueryResult, error) {
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
	return nil, newIllegalArgumentError("expected value of type *[]<type>, got %T", results)
}

// check if v is a valid argument to query GetResults().
// it must be map[string]*<type> where <type> is struct
func checkValidQueryResults(v interface{}, argName string) error {
	if v == nil {
		return newIllegalArgumentError("%s can't be nil", argName)
	}

	tp := reflect.TypeOf(v)
	if tp.Kind() != reflect.Ptr {
		typeGot := fmt.Sprintf("%T", v)
		return newIllegalArgumentError("%s can't be of type %s, must be *[]<type>", argName, typeGot)
	}
	if reflect.ValueOf(v).IsNil() {
		return newIllegalArgumentError("%s can't be a nil slice", argName)
	}
	return nil
}

// GetResults executes the query and sets results to returned values.
// results should be of type *[]<type>
func (q *abstractDocumentQuery) GetResults(results interface{}) error {
	if q.err != nil {
		return q.err
	}
	// Note: in Java it's called ToList
	q.err = checkValidQueryResults(results, "results")
	if q.err != nil {
		return q.err
	}
	return q.executeQueryOperation(results, 0)
}

// First runs a query and returns a first result.
func (q *abstractDocumentQuery) First(result interface{}) error {
	if q.err != nil {
		return q.err
	}
	if result == nil {
		return newIllegalArgumentError("result can't be nil")
	}

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
	if slice.Len() == 0 {
		return newIllegalStateError("Expectecd at least one result")
	}
	el := slice.Index(0)
	setInterfaceToValue(result, el.Interface())
	return nil
}

// Single runs a query that expects only a single result.
// If there is more than one result, it retuns IllegalStateError.
func (q *abstractDocumentQuery) Single(result interface{}) error {
	if q.err != nil {
		return q.err
	}
	if result == nil {
		return fmt.Errorf("result can't be nil")
	}

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
	if slice.Len() != 1 {
		return newIllegalStateError("Expected single result, got: %d", slice.Len())
	}
	el := slice.Index(0)
	setInterfaceToValue(result, el.Interface())
	return nil
}

func (q *abstractDocumentQuery) Count() (int, error) {
	if q.err != nil {
		return 0, q.err
	}
	{
		var tmp = 0
		q.take(&tmp)
	}
	queryResult, err := q.getQueryResult()
	if err != nil {
		return 0, err
	}
	return queryResult.TotalResults, nil
}

// Any returns true if query returns at least one result
// TODO: write tests
func (q *abstractDocumentQuery) Any() (bool, error) {
	if q.err != nil {
		return false, q.err
	}
	if q.isDistinct() {
		// for distinct it is cheaper to do count 1

		toTake := 1
		q.take(&toTake)

		err := q.initSync()
		if err != nil {
			return false, err
		}
		return q.queryOperation.currentQueryResults.TotalResults > 0, nil
	}

	{
		var tmp = 0
		q.take(&tmp)
	}
	queryResult, err := q.getQueryResult()
	if err != nil {
		return false, err
	}
	return queryResult.TotalResults > 0, nil
}

func (q *abstractDocumentQuery) executeQueryOperation(results interface{}, take int) error {
	if take != 0 && (q.pageSize == nil || *q.pageSize > take) {
		q.take(&take)
	}

	err := q.initSync()
	if err != nil {
		return err
	}

	return q.queryOperation.complete(results)
}

func (q *abstractDocumentQuery) aggregateBy(facet FacetBase) error {
	for _, token := range q.selectTokens {
		if _, ok := token.(*facetToken); ok {
			continue
		}

		return newIllegalStateError("Aggregation query can select only facets while it got %T token", token)
	}

	add := func(o interface{}) string {
		return q.addQueryParameter(o)
	}
	t, err := createFacetTokenWithFacetBase(facet, add)
	if err != nil {
		return err
	}
	q.selectTokens = append(q.selectTokens, t)
	return nil
}

func (q *abstractDocumentQuery) aggregateUsing(facetSetupDocumentID string) {
	q.selectTokens = append(q.selectTokens, createFacetToken(facetSetupDocumentID))
}

func (q *abstractDocumentQuery) Lazily(results interface{}, onEval func(interface{})) (*Lazy, error) {
	if q.err != nil {
		return nil, q.err
	}
	if q.queryOperation == nil {
		q.queryOperation, q.err = q.initializeQueryOperation()
		if q.err != nil {
			return nil, q.err
		}
	}

	lazyQueryOperation := NewLazyQueryOperation(results, q.theSession.GetConventions(), q.queryOperation, q.afterQueryExecutedCallback)
	return q.theSession.session.addLazyOperation(results, lazyQueryOperation, onEval), nil
}

// CountLazily returns a lazy operation that returns number of results in a query. It'll set *count to
// number of results after Lazy.GetResult() is called.
// results should be of type []<type> and is only provided so that we know this is a query for <type>
// TODO: figure out better API.
func (q *abstractDocumentQuery) CountLazily(results interface{}, count *int) (*Lazy, error) {
	if count == nil {
		return nil, newIllegalArgumentError("count can't be nil")
	}
	if q.queryOperation == nil {
		v := 0
		q.take(&v)
		var err error
		q.queryOperation, err = q.initializeQueryOperation()
		if err != nil {
			return nil, err
		}
	}

	lazyQueryOperation := NewLazyQueryOperation(results, q.theSession.GetConventions(), q.queryOperation, q.afterQueryExecutedCallback)
	return q.theSession.session.addLazyCountOperation(count, lazyQueryOperation), nil
}

// suggestUsing adds a query part for suggestions
func (q *abstractDocumentQuery) suggestUsing(suggestion SuggestionBase) error {
	if suggestion == nil {
		return newIllegalArgumentError("suggestion cannot be null")
	}

	if err := q.assertCanSuggest(); err != nil {
		return err
	}

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
		return newUnsupportedOperationError("Unknown type of suggestion: %T", suggestion)
	}
	q.selectTokens = append(q.selectTokens, token)
	return nil
}

func (q *abstractDocumentQuery) getOptionsParameterName(options *SuggestionOptions) string {
	optionsParameterName := ""
	if options != nil && options != SuggestionOptionsDefaultOptions {
		optionsParameterName = q.addQueryParameter(options)
	}

	return optionsParameterName
}

func (q *abstractDocumentQuery) assertCanSuggest() error {
	if len(q.whereTokens) > 0 {
		return newIllegalStateError("Cannot add suggest when WHERE statements are present.")
	}

	if len(q.selectTokens) > 0 {
		return newIllegalStateError("Cannot add suggest when SELECT statements are present.")
	}

	if len(q.orderByTokens) > 0 {
		return newIllegalStateError("Cannot add suggest when ORDER BY statements are present.")
	}
	return nil
}
