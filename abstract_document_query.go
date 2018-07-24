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

/*
   @Override
    List<string> getProjectionFields() {
       return fieldsToFetchToken != null && fieldsToFetchToken.projections != null ? Arrays.asList(fieldsToFetchToken.projections) : Collections.emptyList();
   }
*/

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

/*
     _whereLucene(string fieldName, string whereClause, bool exact) {
       fieldName = ensureValidFieldName(fieldName, false);

       List<QueryToken> tokens = getCurrentWhereTokens();
       appendOperatorIfNeeded(tokens);
       negateIfNeeded(tokens, fieldName);

       WhereToken.WhereOptions options = exact ? new WhereToken.WhereOptions(exact) : null;
       WhereToken whereToken = WhereToken.create(WhereOperator.LUCENE, fieldName, addQueryParameter(whereClause), options);
       tokens.add(whereToken);
   }

   @Override
     _openSubclause() {
       _currentClauseDepth++;

       List<QueryToken> tokens = getCurrentWhereTokens();
       appendOperatorIfNeeded(tokens);
       negateIfNeeded(tokens, null);

       tokens.add(OpenSubclauseToken.INSTANCE);
   }

   @Override
     _closeSubclause() {
       _currentClauseDepth--;

       List<QueryToken> tokens = getCurrentWhereTokens();
       tokens.add(CloseSubclauseToken.INSTANCE);
   }

   @Override
     _whereEquals(string fieldName, Object value) {
       _whereEquals(fieldName, value, false);
   }

   @Override
     _whereEquals(string fieldName, Object value, bool exact) {
       WhereParams params = new WhereParams();
       params.setFieldName(fieldName);
       params.setValue(value);
       params.setExact(exact);
       _whereEquals(params);
   }

   @Override
     _whereEquals(string fieldName, MethodCall method) {
       _whereEquals(fieldName, method, false);
   }

   @Override
     _whereEquals(string fieldName, MethodCall method, bool exact) {
       _whereEquals(fieldName, (Object) method, exact);
   }

   @SuppressWarnings("unchecked")
     _whereEquals(WhereParams whereParams) {
       if (negate) {
           negate = false;
           _whereNotEquals(whereParams);
           return;
       }

       whereParams.setFieldName(ensureValidFieldName(whereParams.getFieldName(), whereParams.isNestedPath()));

       List<QueryToken> tokens = getCurrentWhereTokens();
       appendOperatorIfNeeded(tokens);

       if (ifValueIsMethod(WhereOperator.EQUALS, whereParams, tokens)) {
           return;
       }

       Object transformToEqualValue = transformValue(whereParams);
       string addQueryParameter = addQueryParameter(transformToEqualValue);
       WhereToken whereToken = WhereToken.create(WhereOperator.EQUALS, whereParams.getFieldName(), addQueryParameter, new WhereToken.WhereOptions(whereParams.isExact()));
       tokens.add(whereToken);
   }

    bool ifValueIsMethod(WhereOperator op, WhereParams whereParams, List<QueryToken> tokens) {
       if (whereParams.getValue() instanceof MethodCall) {
           MethodCall mc = (MethodCall) whereParams.getValue();

           string[] args = new string[mc.args.length];
           for (int i = 0; i < mc.args.length; i++) {
               args[i] = addQueryParameter(mc.args[i]);
           }

           WhereToken token;
           Class<? extends MethodCall> type = mc.getClass();
           if (CmpXchg.class.equals(type)) {
               token = WhereToken.create(op, whereParams.getFieldName(), null, new WhereToken.WhereOptions(WhereToken.MethodsType.CMP_X_CHG, args, mc.accessPath, whereParams.isExact()));
           } else {
               throw new IllegalArgumentException("Unknown method " + type);
           }

           tokens.add(token);
           return true;
       }

       return false;
   }

     _whereNotEquals(string fieldName, Object value) {
       _whereNotEquals(fieldName, value, false);
   }

     _whereNotEquals(string fieldName, Object value, bool exact) {
       WhereParams params = new WhereParams();
       params.setFieldName(fieldName);
       params.setValue(value);
       params.setExact(exact);

       _whereNotEquals(params);
   }

   @Override
     _whereNotEquals(string fieldName, MethodCall method) {
       _whereNotEquals(fieldName, (Object) method);
   }

   @Override
     _whereNotEquals(string fieldName, MethodCall method, bool exact) {
       _whereNotEquals(fieldName, (Object) method, exact);
   }

   @SuppressWarnings("unchecked")
     _whereNotEquals(WhereParams whereParams) {
       if (negate) {
           negate = false;
           _whereEquals(whereParams);
           return;
       }

       Object transformToEqualValue = transformValue(whereParams);

       List<QueryToken> tokens = getCurrentWhereTokens();
       appendOperatorIfNeeded(tokens);

       whereParams.setFieldName(ensureValidFieldName(whereParams.getFieldName(), whereParams.isNestedPath()));

       if (ifValueIsMethod(WhereOperator.NOT_EQUALS, whereParams, tokens)) {
           return;
       }

       WhereToken whereToken = WhereToken.create(WhereOperator.NOT_EQUALS, whereParams.getFieldName(), addQueryParameter(transformToEqualValue), new WhereToken.WhereOptions(whereParams.isExact()));
       tokens.add(whereToken);
   }

     negateNext() {
       negate = !negate;
   }

   @Override
     _whereIn(string fieldName, Collection<Object> values) {
       _whereIn(fieldName, values, false);
   }

   @Override
     _whereIn(string fieldName, Collection<Object> values, bool exact) {
       fieldName = ensureValidFieldName(fieldName, false);

       List<QueryToken> tokens = getCurrentWhereTokens();
       appendOperatorIfNeeded(tokens);
       negateIfNeeded(tokens, fieldName);

       WhereToken whereToken = WhereToken.create(WhereOperator.IN, fieldName, addQueryParameter(transformCollection(fieldName, unpackCollection(values))));
       tokens.add(whereToken);
   }

   @SuppressWarnings("unchecked")
     _whereStartsWith(string fieldName, Object value) {
       WhereParams whereParams = new WhereParams();
       whereParams.setFieldName(fieldName);
       whereParams.setValue(value);
       whereParams.setAllowWildcards(true);

       Object transformToEqualValue = transformValue(whereParams);

       List<QueryToken> tokens = getCurrentWhereTokens();
       appendOperatorIfNeeded(tokens);

       whereParams.setFieldName(ensureValidFieldName(whereParams.getFieldName(), whereParams.isNestedPath()));
       negateIfNeeded(tokens, whereParams.getFieldName());

       WhereToken whereToken = WhereToken.create(WhereOperator.STARTS_WITH, whereParams.getFieldName(), addQueryParameter(transformToEqualValue));
       tokens.add(whereToken);
   }

   @SuppressWarnings("unchecked")
     _whereEndsWith(string fieldName, Object value) {
       WhereParams whereParams = new WhereParams();
       whereParams.setFieldName(fieldName);
       whereParams.setValue(value);
       whereParams.setAllowWildcards(true);

       Object transformToEqualValue = transformValue(whereParams);

       List<QueryToken> tokens = getCurrentWhereTokens();
       appendOperatorIfNeeded(tokens);

       whereParams.setFieldName(ensureValidFieldName(whereParams.getFieldName(), whereParams.isNestedPath()));
       negateIfNeeded(tokens, whereParams.getFieldName());

       WhereToken whereToken = WhereToken.create(WhereOperator.ENDS_WITH, whereParams.getFieldName(), addQueryParameter(transformToEqualValue));
       tokens.add(whereToken);
   }

   @Override
     _whereBetween(string fieldName, Object start, Object end) {
       _whereBetween(fieldName, start, end, false);
   }

   @Override
     _whereBetween(string fieldName, Object start, Object end, bool exact) {
       fieldName = ensureValidFieldName(fieldName, false);

       List<QueryToken> tokens = getCurrentWhereTokens();
       appendOperatorIfNeeded(tokens);
       negateIfNeeded(tokens, fieldName);

       WhereParams startParams = new WhereParams();
       startParams.setValue(start);
       startParams.setFieldName(fieldName);

       WhereParams endParams = new WhereParams();
       endParams.setValue(end);
       endParams.setFieldName(fieldName);

       string fromParameterName = addQueryParameter(start == null ? "*" : transformValue(startParams, true));
       string toParameterName = addQueryParameter(start == null ? "NULL" : transformValue(endParams, true));

       WhereToken whereToken = WhereToken.create(WhereOperator.BETWEEN, fieldName, null, new WhereToken.WhereOptions(exact, fromParameterName, toParameterName));
       tokens.add(whereToken);
   }

     _whereGreaterThan(string fieldName, Object value) {
       _whereGreaterThan(fieldName, value, false);
   }

     _whereGreaterThan(string fieldName, Object value, bool exact) {
       fieldName = ensureValidFieldName(fieldName, false);

       List<QueryToken> tokens = getCurrentWhereTokens();
       appendOperatorIfNeeded(tokens);
       negateIfNeeded(tokens, fieldName);
       WhereParams whereParams = new WhereParams();
       whereParams.setValue(value);
       whereParams.setFieldName(fieldName);

       string parameter = addQueryParameter(value == null ? "*" : transformValue(whereParams, true));
       WhereToken whereToken = WhereToken.create(WhereOperator.GREATER_THAN, fieldName, parameter, new WhereToken.WhereOptions(exact));
       tokens.add(whereToken);
   }

     _whereGreaterThanOrEqual(string fieldName, Object value) {
       _whereGreaterThanOrEqual(fieldName, value, false);
   }

     _whereGreaterThanOrEqual(string fieldName, Object value, bool exact) {
       fieldName = ensureValidFieldName(fieldName, false);

       List<QueryToken> tokens = getCurrentWhereTokens();
       appendOperatorIfNeeded(tokens);
       negateIfNeeded(tokens, fieldName);
       WhereParams whereParams = new WhereParams();
       whereParams.setValue(value);
       whereParams.setFieldName(fieldName);

       string parameter = addQueryParameter(value == null ? "*" : transformValue(whereParams, true));
       WhereToken whereToken = WhereToken.create(WhereOperator.GREATER_THAN_OR_EQUAL, fieldName, parameter, new WhereToken.WhereOptions(exact));
       tokens.add(whereToken);
   }

     _whereLessThan(string fieldName, Object value) {
       _whereLessThan(fieldName, value, false);
   }

     _whereLessThan(string fieldName, Object value, bool exact) {
       fieldName = ensureValidFieldName(fieldName, false);

       List<QueryToken> tokens = getCurrentWhereTokens();
       appendOperatorIfNeeded(tokens);
       negateIfNeeded(tokens, fieldName);

       WhereParams whereParams = new WhereParams();
       whereParams.setValue(value);
       whereParams.setFieldName(fieldName);

       string parameter = addQueryParameter(value == null ? "NULL" : transformValue(whereParams, true));
       WhereToken whereToken = WhereToken.create(WhereOperator.LESS_THAN, fieldName, parameter, new WhereToken.WhereOptions(exact));
       tokens.add(whereToken);
   }

     _whereLessThanOrEqual(string fieldName, Object value) {
       _whereLessThanOrEqual(fieldName, value, false);
   }

     _whereLessThanOrEqual(string fieldName, Object value, bool exact) {
       List<QueryToken> tokens = getCurrentWhereTokens();
       appendOperatorIfNeeded(tokens);
       negateIfNeeded(tokens, fieldName);

       WhereParams whereParams = new WhereParams();
       whereParams.setValue(value);
       whereParams.setFieldName(fieldName);

       string parameter = addQueryParameter(value == null ? "NULL" : transformValue(whereParams, true));
       WhereToken whereToken = WhereToken.create(WhereOperator.LESS_THAN_OR_EQUAL, fieldName, parameter, new WhereToken.WhereOptions(exact));
       tokens.add(whereToken);
   }

   @Override
     _whereRegex(string fieldName, string pattern) {
       List<QueryToken> tokens = getCurrentWhereTokens();
       appendOperatorIfNeeded(tokens);
       negateIfNeeded(tokens, fieldName);

       WhereParams whereParams = new WhereParams();
       whereParams.setValue(pattern);
       whereParams.setFieldName(fieldName);

       string parameter = addQueryParameter(transformValue(whereParams));

       WhereToken whereToken = WhereToken.create(WhereOperator.REGEX, fieldName, parameter);
       tokens.add(whereToken);
   }
*/

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

/*
   @Override
     _boost(double boost) {
       if (boost == 1.0) {
           return;
       }

       List<QueryToken> tokens = getCurrentWhereTokens();
       if (tokens.isEmpty()) {
           throw new IllegalStateException("Missing where clause");
       }

       QueryToken whereToken = tokens.get(tokens.size() - 1);
       if (!(whereToken instanceof WhereToken)) {
           throw new IllegalStateException("Missing where clause");
       }

       if (boost <= 0.0) {
           throw new IllegalArgumentException("Boost factor must be a positive number");
       }

       ((WhereToken) whereToken).getOptions().setBoost(boost);
   }

   @Override
     _fuzzy(double fuzzy) {
       List<QueryToken> tokens = getCurrentWhereTokens();
       if (tokens.isEmpty()) {
           throw new IllegalStateException("Missing where clause");
       }

       QueryToken whereToken = tokens.get(tokens.size() - 1);
       if (!(whereToken instanceof WhereToken)) {
           throw new IllegalStateException("Missing where clause");
       }

       if (fuzzy < 0.0 || fuzzy > 1.0) {
           throw new IllegalArgumentException("Fuzzy distance must be between 0.0 and 1.0");
       }

       ((WhereToken) whereToken).getOptions().setFuzzy(fuzzy);
   }

   @Override
     _proximity(int proximity) {
       List<QueryToken> tokens = getCurrentWhereTokens();
       if (tokens.isEmpty()) {
           throw new IllegalStateException("Missing where clause");
       }

       QueryToken whereToken = tokens.get(tokens.size() - 1);
       if (!(whereToken instanceof WhereToken)) {
           throw new IllegalStateException("Missing where clause");
       }

       if (proximity < 1) {
           throw new IllegalArgumentException("Proximity distance must be a positive number");
       }

       ((WhereToken) whereToken).getOptions().setProximity(proximity);
   }

     _orderBy(string field) {
       _orderBy(field, OrderingType.string);
   }

     _orderBy(string field, OrderingType ordering) {
       assertNoRawQuery();
       string f = ensureValidFieldName(field, false);
       orderByTokens.add(OrderByToken.createAscending(f, ordering));
   }

     _orderByDescending(string field) {
       _orderByDescending(field, OrderingType.string);
   }

     _orderByDescending(string field, OrderingType ordering) {
       assertNoRawQuery();
       string f = ensureValidFieldName(field, false);
       orderByTokens.add(OrderByToken.createDescending(f, ordering));
   }

     _orderByScore() {
       assertNoRawQuery();

       orderByTokens.add(OrderByToken.scoreAscending);
   }
*/

func (q *AbstractDocumentQuery) _orderByScoreDescending() {
	q.assertNoRawQuery()
	q.orderByTokens = append(q.orderByTokens, OrderByToken_scoreDescending)
}

func (q *AbstractDocumentQuery) _statistics(stats **QueryStatistics) {
	*stats = q.queryStats
}

func (q *AbstractDocumentQuery) invokeAfterQueryExecuted(result *QueryResult) {
	panicIf(true, "NYI")
	// TODO:
	// EventHelper.invoke(afterQueryExecutedCallback, result);
}

func (q *AbstractDocumentQuery) invokeBeforeQueryExecuted(query *IndexQuery) {
	panicIf(true, "NYI")
	// TODO:
	// EventHelper.invoke(beforeQueryExecutedCallback, query)
}

func (q *AbstractDocumentQuery) invokeAfterStreamExecuted(result ObjectNode) {
	panicIf(true, "NYI")
	// TODO:
	// EventHelper.invoke(afterStreamExecutedCallback, result)
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

/*
   @Override
     _search(string fieldName, string searchTerms) {
       _search(fieldName, searchTerms, SearchOperator.OR);
   }

   @Override
     _search(string fieldName, string searchTerms, SearchOperator operator) {
       List<QueryToken> tokens = getCurrentWhereTokens();
       appendOperatorIfNeeded(tokens);

       fieldName = ensureValidFieldName(fieldName, false);
       negateIfNeeded(tokens, fieldName);

       WhereToken whereToken = WhereToken.create(WhereOperator.SEARCH, fieldName, addQueryParameter(searchTerms), new WhereToken.WhereOptions(operator));
       tokens.add(whereToken);
   }
*/

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

/*
   @Override
     _intersect() {
       List<QueryToken> tokens = getCurrentWhereTokens();
       if (tokens.size() > 0) {
           QueryToken last = tokens.get(tokens.size() - 1);
           if (last instanceof WhereToken || last instanceof CloseSubclauseToken) {
               isIntersect = true;

               tokens.add(IntersectMarkerToken.INSTANCE);
               return;
           }
       }

       throw new IllegalStateException("Cannot add INTERSECT at this point.");
   }
*/

func (q *AbstractDocumentQuery) _whereExists(fieldName string) {
	fieldName = q.ensureValidFieldName(fieldName, false)

	tokensRef := q.getCurrentWhereTokensRef()
	q.appendOperatorIfNeeded(tokensRef)
	q.negateIfNeeded(tokensRef, fieldName)

	tokens := *tokensRef
	tokens = append(tokens, WhereToken_create(WhereOperator_EXISTS, fieldName, ""))
	*tokensRef = tokens
}

/*
   @Override
     _containsAny(string fieldName, Collection<Object> values) {
       fieldName = ensureValidFieldName(fieldName, false);

       List<QueryToken> tokens = getCurrentWhereTokens();
       appendOperatorIfNeeded(tokens);
       negateIfNeeded(tokens, fieldName);

       Collection<Object> array = transformCollection(fieldName, unpackCollection(values));
       WhereToken whereToken = WhereToken.create(WhereOperator.IN, fieldName, addQueryParameter(array), new WhereToken.WhereOptions(false));
       tokens.add(whereToken);
   }

   @Override
     _containsAll(string fieldName, Collection<Object> values) {
       fieldName = ensureValidFieldName(fieldName, false);

       List<QueryToken> tokens = getCurrentWhereTokens();
       appendOperatorIfNeeded(tokens);
       negateIfNeeded(tokens, fieldName);

       Collection<Object> array = transformCollection(fieldName, unpackCollection(values));

       if (array.isEmpty()) {
           tokens.add(TrueToken.INSTANCE);
           return;
       }

       WhereToken whereToken = WhereToken.create(WhereOperator.ALL_IN, fieldName, addQueryParameter(array));
       tokens.add(whereToken);
   }

   @Override
     _addRootType(Class clazz) {
       rootTypes.add(clazz);


   @Override
*/

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

/*

    static Collection<Object> unpackCollection(Collection items) {
       List<Object> results = new ArrayList<>();

       for (Object item : items) {
           if (item instanceof Collection) {
               results.addAll(unpackCollection((Collection) item));
           } else {
               results.add(item);
           }
       }

       return results;
   }
*/

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
	return q.transformValue2(whereParams, false)
}

func (q *AbstractDocumentQuery) transformValue2(whereParams *WhereParams, forRange bool) Object {
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

/*
   protected  _withinRadiusOf(string fieldName, double radius, double latitude, double longitude, SpatialUnits radiusUnits, double distErrorPercent) {
       fieldName = ensureValidFieldName(fieldName, false);

       List<QueryToken> tokens = getCurrentWhereTokens();
       appendOperatorIfNeeded(tokens);
       negateIfNeeded(tokens, fieldName);

       WhereToken whereToken = WhereToken.create(WhereOperator.SPATIAL_WITHIN, fieldName, null, new WhereToken.WhereOptions(ShapeToken.circle(addQueryParameter(radius), addQueryParameter(latitude), addQueryParameter(longitude), radiusUnits), distErrorPercent));
       tokens.add(whereToken);
   }

   protected  _spatial(string fieldName, string shapeWkt, SpatialRelation relation, double distErrorPercent) {
       fieldName = ensureValidFieldName(fieldName, false);

       List<QueryToken> tokens = getCurrentWhereTokens();
       appendOperatorIfNeeded(tokens);
       negateIfNeeded(tokens, fieldName);

       ShapeToken wktToken = ShapeToken.wkt(addQueryParameter(shapeWkt));

       WhereOperator whereOperator;
       switch (relation) {
           case WITHIN:
               whereOperator = WhereOperator.SPATIAL_WITHIN;
               break;
           case CONTAINS:
               whereOperator = WhereOperator.SPATIAL_CONTAINS;
               break;
           case DISJOINT:
               whereOperator = WhereOperator.SPATIAL_DISJOINT;
               break;
           case INTERSECTS:
               whereOperator = WhereOperator.SPATIAL_INTERSECTS;
               break;
           default:
               throw new IllegalArgumentException();
       }

       tokens.add(WhereToken.create(whereOperator, fieldName, null, new WhereToken.WhereOptions(wktToken, distErrorPercent)));
   }

   @Override
     _spatial(DynamicSpatialField dynamicField, SpatialCriteria criteria) {
       List<QueryToken> tokens = getCurrentWhereTokens();
       appendOperatorIfNeeded(tokens);
       negateIfNeeded(tokens, null);

       tokens.add(criteria.toQueryToken(dynamicField.toField(this::ensureValidFieldName), this::addQueryParameter));
   }

   @Override
     _spatial(string fieldName, SpatialCriteria criteria) {
       fieldName = ensureValidFieldName(fieldName, false);

       List<QueryToken> tokens = getCurrentWhereTokens();
       appendOperatorIfNeeded(tokens);
       negateIfNeeded(tokens, fieldName);

       tokens.add(criteria.toQueryToken(fieldName, this::addQueryParameter));
   }

   @Override
     _orderByDistance(DynamicSpatialField field, double latitude, double longitude) {
       if (field == null) {
           throw new IllegalArgumentException("Field cannot be null");
       }
       _orderByDistance("'" + field.toField(this::ensureValidFieldName) + "'", latitude, longitude);
   }

   @Override
     _orderByDistance(string fieldName, double latitude, double longitude) {
       orderByTokens.add(OrderByToken.createDistanceAscending(fieldName, addQueryParameter(latitude), addQueryParameter(longitude)));
   }

   @Override
     _orderByDistance(DynamicSpatialField field, string shapeWkt) {
       if (field == null) {
           throw new IllegalArgumentException("Field cannot be null");
       }
       _orderByDistance("'" + field.toField(this::ensureValidFieldName) + "'", shapeWkt);
   }

   @Override
     _orderByDistance(string fieldName, string shapeWkt) {
       orderByTokens.add(OrderByToken.createDistanceAscending(fieldName, addQueryParameter(shapeWkt)));
   }

   @Override
     _orderByDistanceDescending(DynamicSpatialField field, double latitude, double longitude) {
       if (field == null) {
           throw new IllegalArgumentException("Field cannot be null");
       }
       _orderByDistanceDescending("'" + field.toField(this::ensureValidFieldName) + "'", latitude, longitude);
   }

   @Override
     _orderByDistanceDescending(string fieldName, double latitude, double longitude) {
       orderByTokens.add(OrderByToken.createDistanceDescending(fieldName, addQueryParameter(latitude), addQueryParameter(longitude)));
   }

   @Override
     _orderByDistanceDescending(DynamicSpatialField field, string shapeWkt) {
       if (field == null) {
           throw new IllegalArgumentException("Field cannot be null");
       }
       _orderByDistanceDescending("'" + field.toField(this::ensureValidFieldName) + "'", shapeWkt);
   }

   @Override
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

   @Override
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

   @Override
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
