package ravendb

import (
	"reflect"
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
 abstract class AbstractDocumentQuery<T, TSelf extends AbstractDocumentQuery<T, TSelf>> implements IAbstractDocumentQuery<T> {

    @Override
     List<string> getProjectionFields() {
        return fieldsToFetchToken != null && fieldsToFetchToken.projections != null ? Arrays.asList(fieldsToFetchToken.projections) : Collections.emptyList();
    }

    @Override
      _randomOrdering() {
        assertNoRawQuery();
        orderByTokens.add(OrderByToken.random);
    }

    @Override
      _randomOrdering(string seed) {
        assertNoRawQuery();

        if (stringUtils.isBlank(seed)) {
            _randomOrdering();
            return;
        }

        orderByTokens.add(OrderByToken.createRandom(seed));
    }

    protected  addGroupByAlias(string fieldName, string projectedName) {
        _aliasToGroupByFieldName.put(projectedName, fieldName);
    }
*/

func (q *AbstractDocumentQuery) assertNoRawQuery() {
	panicIf(q.queryRaw != "", "RawQuery was called, cannot modify this query by calling on operations that would modify the query (such as Where, Select, OrderBy, GroupBy, etc)")
}

/*
     _addParameter(string name, Object value) {
       name = stringUtils.stripStart(name, "$");
       if (queryParameters.containsKey(name)) {
           throw new IllegalStateException("The parameter " + name + " was already added");
       }

       queryParameters.put(name, value);
   }

   @Override
     _groupBy(string fieldName, string... fieldNames) {
       GroupBy[] mapping = Arrays.stream(fieldNames)
               .map(x -> GroupBy.field(x))
               .toArray(GroupBy[]::new);

       _groupBy(GroupBy.field(fieldName), mapping);
   }

   @Override
     _groupBy(GroupBy field, GroupBy... fields) {
       if (!fromToken.isDynamic()) {
           throw new IllegalStateException("groupBy only works with dynamic queries");
       }

       assertNoRawQuery();
       isGroupBy = true;

       string fieldName = ensureValidFieldName(field.getField(), false);

       groupByTokens.add(GroupByToken.create(fieldName, field.getMethod()));

       if (fields == null || fields.length <= 0) {
           return;
       }

       for (GroupBy item : fields) {
           fieldName = ensureValidFieldName(item.getField(), false);
           groupByTokens.add(GroupByToken.create(fieldName, item.getMethod()));
       }
   }

   @Override
     _groupByKey(string fieldName) {
       _groupByKey(fieldName, null);
   }

   @Override
     _groupByKey(string fieldName, string projectedName) {
       assertNoRawQuery();
       isGroupBy = true;

       if (projectedName != null && _aliasToGroupByFieldName.containsKey(projectedName)) {
           string aliasedFieldName = _aliasToGroupByFieldName.get(projectedName);
           if (fieldName == null || fieldName.equalsIgnoreCase(projectedName)) {
               fieldName = aliasedFieldName;
           }
       } else if (fieldName != null && _aliasToGroupByFieldName.containsValue(fieldName)) {
           string aliasedFieldName = _aliasToGroupByFieldName.get(fieldName);
           fieldName = aliasedFieldName;
       }

       selectTokens.add(GroupByKeyToken.create(fieldName, projectedName));
   }

   @Override
     _groupBySum(string fieldName) {
       _groupBySum(fieldName, null);
   }

   @Override
     _groupBySum(string fieldName, string projectedName) {
       assertNoRawQuery();
       isGroupBy = true;

       fieldName = ensureValidFieldName(fieldName, false);
       selectTokens.add(GroupBySumToken.create(fieldName, projectedName));
   }

   @Override
     _groupByCount() {
       _groupByCount(null);
   }

   @Override
     _groupByCount(string projectedName) {
       assertNoRawQuery();
       isGroupBy = true;

       selectTokens.add(GroupByCountToken.create(projectedName));
   }

   @Override
     _whereTrue() {
       List<QueryToken> tokens = getCurrentWhereTokens();
       appendOperatorIfNeeded(tokens);
       negateIfNeeded(tokens, null);

       tokens.add(TrueToken.INSTANCE);
   }


    MoreLikeThisScope _moreLikeThis() {
       appendOperatorIfNeeded(whereTokens);

       MoreLikeThisToken token = new MoreLikeThisToken();
       whereTokens.add(token);

       _isInMoreLikeThis = true;
       return new MoreLikeThisScope(token, this::addQueryParameter, () -> _isInMoreLikeThis = false);
   }

   @Override
     _include(string path) {
       includes.add(path);
   }

   @Override
     _take(int count) {
       pageSize = count;
   }

   @Override
     _skip(int count) {
       start = count;
   }

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

     _andAlso() {
       List<QueryToken> tokens = getCurrentWhereTokens();
       if (tokens.isEmpty()) {
           return;
       }

       if (tokens.get(tokens.size() - 1) instanceof QueryOperatorToken) {
           throw new IllegalStateException("Cannot add AND, previous token was already an operator token.");
       }

       tokens.add(QueryOperatorToken.AND);
   }

     _orElse() {
       List<QueryToken> tokens = getCurrentWhereTokens();
       if (tokens.isEmpty()) {
           return;
       }

       if (tokens.get(tokens.size() - 1) instanceof QueryOperatorToken) {
           throw new IllegalStateException("Cannot add OR, previous token was already an operator token.");
       }

       tokens.add(QueryOperatorToken.OR);
   }

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

func Character_isLetterOrDigit(r rune) bool {
	if r >= 'a' && r <= 'z' {
		return true
	}
	if r >= 'A' && r <= 'Z' {
		return true
	}
	if r >= '0' && r <= '9' {
		return true
	}
	return false
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

     _whereExists(string fieldName) {
       fieldName = ensureValidFieldName(fieldName, false);

       List<QueryToken> tokens = getCurrentWhereTokens();
       appendOperatorIfNeeded(tokens);
       negateIfNeeded(tokens, fieldName);

       tokens.add(WhereToken.create(WhereOperator.EXISTS, fieldName, null));
   }

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

/*
     appendOperatorIfNeeded(List<QueryToken> tokens) {
       assertNoRawQuery();

       if (tokens.isEmpty()) {
           return;
       }

       QueryToken lastToken = tokens.get(tokens.size() - 1);
       if (!(lastToken instanceof WhereToken) && !(lastToken instanceof CloseSubclauseToken)) {
           return;
       }

       WhereToken lastWhere = null;

       for (int i = tokens.size() - 1; i >= 0; i--) {
           if (tokens.get(i) instanceof WhereToken) {
               lastWhere = (WhereToken) tokens.get(i);
               break;
           }
       }

       QueryOperatorToken token = defaultOperator == QueryOperator.AND ? QueryOperatorToken.AND : QueryOperatorToken.OR;

       if (lastWhere != null && lastWhere.getOptions().getSearchOperator() != null) {
           token = QueryOperatorToken.OR; // default to OR operator after search if AND was not specified explicitly
       }

       tokens.add(token);
   }

   @SuppressWarnings("unchecked")
    Collection<Object> transformCollection(string fieldName, Collection<Object> values) {
       List<Object> result = new ArrayList<>();
       for (Object value : values) {
           if (value instanceof Collection) {
               result.addAll(transformCollection(fieldName, (Collection) value));
           } else {
               WhereParams nestedWhereParams = new WhereParams();
               nestedWhereParams.setAllowWildcards(true);
               nestedWhereParams.setFieldName(fieldName);
               nestedWhereParams.setValue(value);

               result.add(transformValue(nestedWhereParams));
           }
       }
       return result;
   }

     negateIfNeeded(List<QueryToken> tokens, string fieldName) {
       if (!negate) {
           return;
       }

       negate = false;

       if (tokens.isEmpty() || tokens.get(tokens.size() - 1) instanceof OpenSubclauseToken) {
           if (fieldName != null) {
               _whereExists(fieldName);
           } else {
               _whereTrue();
           }
           _andAlso();
       }

       tokens.add(NegateToken.INSTANCE);
   }

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

    string ensureValidFieldName(string fieldName, bool isNestedPath) {
       if (theSession == null || theSession.getConventions() == null || isNestedPath || isGroupBy) {
           return QueryFieldUtil.escapeIfNecessary(fieldName);
       }

       for (Class rootType : rootTypes) {
           Field identityProperty = theSession.getConventions().getIdentityProperty(rootType);
           if (identityProperty != null && identityProperty.getName().equals(fieldName)) {
               return Constants.Documents.Indexing.Fields.DOCUMENT_ID_FIELD_NAME;
           }
       }

       return QueryFieldUtil.escapeIfNecessary(fieldName);
   }

    Object transformValue(WhereParams whereParams) {
       return transformValue(whereParams, false);
   }

    Object transformValue(WhereParams whereParams, bool forRange) {
       if (whereParams.getValue() == null) {
           return null;
       }

       if ("".equals(whereParams.getValue())) {
           return "";
       }

       Reference<string> stringValueReference = new Reference<>();
       if (_conventions.tryConvertValueForQuery(whereParams.getFieldName(), whereParams.getValue(), forRange, stringValueReference)) {
           return stringValueReference.value;
       }

       Class<?> clazz = whereParams.getValue().getClass();
       if (Date.class.equals(clazz)) {
           return whereParams.getValue();
       }

       if (string.class.equals(clazz)) {
           return whereParams.getValue();
       }

       if (Integer.class.equals(clazz)) {
           return whereParams.getValue();
       }

       if (Long.class.equals(clazz)) {
           return whereParams.getValue();
       }

       if (Float.class.equals(clazz)) {
           return whereParams.getValue();
       }

       if (Double.class.equals(clazz)) {
           return whereParams.getValue();
       }

       if (Duration.class.equals(clazz)) {
           return ((Duration) whereParams.getValue()).toNanos() / 100;
       }

       if (string.class.equals(clazz)) {
           return whereParams.getValue();
       }

       if (bool.class.equals(clazz)) {
           return whereParams.getValue();
       }

       if (clazz.isEnum()) {
           return whereParams.getValue();
       }

       return whereParams.getValue();

   }

    string addQueryParameter(Object value) {
       string parameterName = "p" + queryParameters.size();
       queryParameters.put(parameterName, value);
       return parameterName;
   }

    List<QueryToken> getCurrentWhereTokens() {
       if (!_isInMoreLikeThis) {
           return whereTokens;
       }

       if (whereTokens.isEmpty()) {
           throw new IllegalStateException("Cannot get MoreLikeThisToken because there are no where token specified.");
       }

       QueryToken lastToken = whereTokens.get(whereTokens.size() - 1);

       if (lastToken instanceof MoreLikeThisToken) {
           MoreLikeThisToken moreLikeThisToken = (MoreLikeThisToken) lastToken;
           return moreLikeThisToken.whereTokens;
       } else {
           throw new IllegalStateException("Last token is not MoreLikeThisToken");
       }
   }

   protected  updateFieldsToFetchToken(FieldsToFetchToken fieldsToFetch) {
       this.fieldsToFetchToken = fieldsToFetch;

       if (selectTokens.isEmpty()) {
           selectTokens.add(fieldsToFetch);
       } else {
           Optional<QueryToken> fetchToken = selectTokens.stream()
                   .filter(x -> x instanceof FieldsToFetchToken)
                   .findFirst();

           if (fetchToken.isPresent()) {
               int idx = selectTokens.indexOf(fetchToken.get());
               selectTokens.set(idx, fieldsToFetch);
           } else {
               selectTokens.add(fieldsToFetch);
           }
       }
   }
*/

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

     string getOptionsParameterName(SuggestionOptions options) {
        string optionsParameterName = null;
        if (options != null && options != SuggestionOptions.defaultOptions) {
            optionsParameterName = addQueryParameter(options);
        }

        return optionsParameterName;
    }

      assertCanSuggest() {
        if (!whereTokens.isEmpty()) {
            throw new IllegalStateException("Cannot add suggest when WHERE statements are present.");
        }

        if (!selectTokens.isEmpty()) {
            throw new IllegalStateException("Cannot add suggest when SELECT statements are present.");
        }

        if (!orderByTokens.isEmpty()) {
            throw new IllegalStateException("Cannot add suggest when ORDER BY statements are present.");
        }
    }
}
*/
