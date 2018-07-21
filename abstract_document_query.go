package ravendb

import (
	"reflect"
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
	//_addAfterQueryExecutedListener(this::updateStatsAndHighlightings);
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

/*
func (q *AbstractDocumentQuery)  initializeQueryOperation() *QueryOperation {
	indexQuery := q.getIndexQuery();

	return new QueryOperation(theSession, indexName, indexQuery, fieldsToFetchToken, disableEntitiesTracking, false, false);
}
*/

/*
public abstract class AbstractDocumentQuery<T, TSelf extends AbstractDocumentQuery<T, TSelf>> implements IAbstractDocumentQuery<T> {

    public IndexQuery getIndexQuery() {
        String query = toString();
        IndexQuery indexQuery = generateIndexQuery(query);
        invokeBeforeQueryExecuted(indexQuery);
        return indexQuery;
    }

    @Override
    public List<String> getProjectionFields() {
        return fieldsToFetchToken != null && fieldsToFetchToken.projections != null ? Arrays.asList(fieldsToFetchToken.projections) : Collections.emptyList();
    }

    @Override
    public void _randomOrdering() {
        assertNoRawQuery();
        orderByTokens.add(OrderByToken.random);
    }

    @Override
    public void _randomOrdering(String seed) {
        assertNoRawQuery();

        if (StringUtils.isBlank(seed)) {
            _randomOrdering();
            return;
        }

        orderByTokens.add(OrderByToken.createRandom(seed));
    }

    protected void addGroupByAlias(String fieldName, String projectedName) {
        _aliasToGroupByFieldName.put(projectedName, fieldName);
    }

    private void assertNoRawQuery() {
        if (queryRaw != null) {
            throw new IllegalStateException("RawQuery was called, cannot modify this query by calling on operations that would modify the query (such as Where, Select, OrderBy, GroupBy, etc)");
        }
    }

    public void _addParameter(String name, Object value) {
        name = StringUtils.stripStart(name, "$");
        if (queryParameters.containsKey(name)) {
            throw new IllegalStateException("The parameter " + name + " was already added");
        }

        queryParameters.put(name, value);
    }

    @Override
    public void _groupBy(String fieldName, String... fieldNames) {
        GroupBy[] mapping = Arrays.stream(fieldNames)
                .map(x -> GroupBy.field(x))
                .toArray(GroupBy[]::new);

        _groupBy(GroupBy.field(fieldName), mapping);
    }

    @Override
    public void _groupBy(GroupBy field, GroupBy... fields) {
        if (!fromToken.isDynamic()) {
            throw new IllegalStateException("groupBy only works with dynamic queries");
        }

        assertNoRawQuery();
        isGroupBy = true;

        String fieldName = ensureValidFieldName(field.getField(), false);

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
    public void _groupByKey(String fieldName) {
        _groupByKey(fieldName, null);
    }

    @Override
    public void _groupByKey(String fieldName, String projectedName) {
        assertNoRawQuery();
        isGroupBy = true;

        if (projectedName != null && _aliasToGroupByFieldName.containsKey(projectedName)) {
            String aliasedFieldName = _aliasToGroupByFieldName.get(projectedName);
            if (fieldName == null || fieldName.equalsIgnoreCase(projectedName)) {
                fieldName = aliasedFieldName;
            }
        } else if (fieldName != null && _aliasToGroupByFieldName.containsValue(fieldName)) {
            String aliasedFieldName = _aliasToGroupByFieldName.get(fieldName);
            fieldName = aliasedFieldName;
        }

        selectTokens.add(GroupByKeyToken.create(fieldName, projectedName));
    }

    @Override
    public void _groupBySum(String fieldName) {
        _groupBySum(fieldName, null);
    }

    @Override
    public void _groupBySum(String fieldName, String projectedName) {
        assertNoRawQuery();
        isGroupBy = true;

        fieldName = ensureValidFieldName(fieldName, false);
        selectTokens.add(GroupBySumToken.create(fieldName, projectedName));
    }

    @Override
    public void _groupByCount() {
        _groupByCount(null);
    }

    @Override
    public void _groupByCount(String projectedName) {
        assertNoRawQuery();
        isGroupBy = true;

        selectTokens.add(GroupByCountToken.create(projectedName));
    }

    @Override
    public void _whereTrue() {
        List<QueryToken> tokens = getCurrentWhereTokens();
        appendOperatorIfNeeded(tokens);
        negateIfNeeded(tokens, null);

        tokens.add(TrueToken.INSTANCE);
    }


    public MoreLikeThisScope _moreLikeThis() {
        appendOperatorIfNeeded(whereTokens);

        MoreLikeThisToken token = new MoreLikeThisToken();
        whereTokens.add(token);

        _isInMoreLikeThis = true;
        return new MoreLikeThisScope(token, this::addQueryParameter, () -> _isInMoreLikeThis = false);
    }

    @Override
    public void _include(String path) {
        includes.add(path);
    }

    @Override
    public void _take(int count) {
        pageSize = count;
    }

    @Override
    public void _skip(int count) {
        start = count;
    }

    public void _whereLucene(String fieldName, String whereClause, boolean exact) {
        fieldName = ensureValidFieldName(fieldName, false);

        List<QueryToken> tokens = getCurrentWhereTokens();
        appendOperatorIfNeeded(tokens);
        negateIfNeeded(tokens, fieldName);

        WhereToken.WhereOptions options = exact ? new WhereToken.WhereOptions(exact) : null;
        WhereToken whereToken = WhereToken.create(WhereOperator.LUCENE, fieldName, addQueryParameter(whereClause), options);
        tokens.add(whereToken);
    }

    @Override
    public void _openSubclause() {
        _currentClauseDepth++;

        List<QueryToken> tokens = getCurrentWhereTokens();
        appendOperatorIfNeeded(tokens);
        negateIfNeeded(tokens, null);

        tokens.add(OpenSubclauseToken.INSTANCE);
    }

    @Override
    public void _closeSubclause() {
        _currentClauseDepth--;

        List<QueryToken> tokens = getCurrentWhereTokens();
        tokens.add(CloseSubclauseToken.INSTANCE);
    }

    @Override
    public void _whereEquals(String fieldName, Object value) {
        _whereEquals(fieldName, value, false);
    }

    @Override
    public void _whereEquals(String fieldName, Object value, boolean exact) {
        WhereParams params = new WhereParams();
        params.setFieldName(fieldName);
        params.setValue(value);
        params.setExact(exact);
        _whereEquals(params);
    }

    @Override
    public void _whereEquals(String fieldName, MethodCall method) {
        _whereEquals(fieldName, method, false);
    }

    @Override
    public void _whereEquals(String fieldName, MethodCall method, boolean exact) {
        _whereEquals(fieldName, (Object) method, exact);
    }

    @SuppressWarnings("unchecked")
    public void _whereEquals(WhereParams whereParams) {
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
        String addQueryParameter = addQueryParameter(transformToEqualValue);
        WhereToken whereToken = WhereToken.create(WhereOperator.EQUALS, whereParams.getFieldName(), addQueryParameter, new WhereToken.WhereOptions(whereParams.isExact()));
        tokens.add(whereToken);
    }

    private boolean ifValueIsMethod(WhereOperator op, WhereParams whereParams, List<QueryToken> tokens) {
        if (whereParams.getValue() instanceof MethodCall) {
            MethodCall mc = (MethodCall) whereParams.getValue();

            String[] args = new String[mc.args.length];
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

    public void _whereNotEquals(String fieldName, Object value) {
        _whereNotEquals(fieldName, value, false);
    }

    public void _whereNotEquals(String fieldName, Object value, boolean exact) {
        WhereParams params = new WhereParams();
        params.setFieldName(fieldName);
        params.setValue(value);
        params.setExact(exact);

        _whereNotEquals(params);
    }

    @Override
    public void _whereNotEquals(String fieldName, MethodCall method) {
        _whereNotEquals(fieldName, (Object) method);
    }

    @Override
    public void _whereNotEquals(String fieldName, MethodCall method, boolean exact) {
        _whereNotEquals(fieldName, (Object) method, exact);
    }

    @SuppressWarnings("unchecked")
    public void _whereNotEquals(WhereParams whereParams) {
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

    public void negateNext() {
        negate = !negate;
    }

    @Override
    public void _whereIn(String fieldName, Collection<Object> values) {
        _whereIn(fieldName, values, false);
    }

    @Override
    public void _whereIn(String fieldName, Collection<Object> values, boolean exact) {
        fieldName = ensureValidFieldName(fieldName, false);

        List<QueryToken> tokens = getCurrentWhereTokens();
        appendOperatorIfNeeded(tokens);
        negateIfNeeded(tokens, fieldName);

        WhereToken whereToken = WhereToken.create(WhereOperator.IN, fieldName, addQueryParameter(transformCollection(fieldName, unpackCollection(values))));
        tokens.add(whereToken);
    }

    @SuppressWarnings("unchecked")
    public void _whereStartsWith(String fieldName, Object value) {
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
    public void _whereEndsWith(String fieldName, Object value) {
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
    public void _whereBetween(String fieldName, Object start, Object end) {
        _whereBetween(fieldName, start, end, false);
    }

    @Override
    public void _whereBetween(String fieldName, Object start, Object end, boolean exact) {
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

        String fromParameterName = addQueryParameter(start == null ? "*" : transformValue(startParams, true));
        String toParameterName = addQueryParameter(start == null ? "NULL" : transformValue(endParams, true));

        WhereToken whereToken = WhereToken.create(WhereOperator.BETWEEN, fieldName, null, new WhereToken.WhereOptions(exact, fromParameterName, toParameterName));
        tokens.add(whereToken);
    }

    public void _whereGreaterThan(String fieldName, Object value) {
        _whereGreaterThan(fieldName, value, false);
    }

    public void _whereGreaterThan(String fieldName, Object value, boolean exact) {
        fieldName = ensureValidFieldName(fieldName, false);

        List<QueryToken> tokens = getCurrentWhereTokens();
        appendOperatorIfNeeded(tokens);
        negateIfNeeded(tokens, fieldName);
        WhereParams whereParams = new WhereParams();
        whereParams.setValue(value);
        whereParams.setFieldName(fieldName);

        String parameter = addQueryParameter(value == null ? "*" : transformValue(whereParams, true));
        WhereToken whereToken = WhereToken.create(WhereOperator.GREATER_THAN, fieldName, parameter, new WhereToken.WhereOptions(exact));
        tokens.add(whereToken);
    }

    public void _whereGreaterThanOrEqual(String fieldName, Object value) {
        _whereGreaterThanOrEqual(fieldName, value, false);
    }

    public void _whereGreaterThanOrEqual(String fieldName, Object value, boolean exact) {
        fieldName = ensureValidFieldName(fieldName, false);

        List<QueryToken> tokens = getCurrentWhereTokens();
        appendOperatorIfNeeded(tokens);
        negateIfNeeded(tokens, fieldName);
        WhereParams whereParams = new WhereParams();
        whereParams.setValue(value);
        whereParams.setFieldName(fieldName);

        String parameter = addQueryParameter(value == null ? "*" : transformValue(whereParams, true));
        WhereToken whereToken = WhereToken.create(WhereOperator.GREATER_THAN_OR_EQUAL, fieldName, parameter, new WhereToken.WhereOptions(exact));
        tokens.add(whereToken);
    }

    public void _whereLessThan(String fieldName, Object value) {
        _whereLessThan(fieldName, value, false);
    }

    public void _whereLessThan(String fieldName, Object value, boolean exact) {
        fieldName = ensureValidFieldName(fieldName, false);

        List<QueryToken> tokens = getCurrentWhereTokens();
        appendOperatorIfNeeded(tokens);
        negateIfNeeded(tokens, fieldName);

        WhereParams whereParams = new WhereParams();
        whereParams.setValue(value);
        whereParams.setFieldName(fieldName);

        String parameter = addQueryParameter(value == null ? "NULL" : transformValue(whereParams, true));
        WhereToken whereToken = WhereToken.create(WhereOperator.LESS_THAN, fieldName, parameter, new WhereToken.WhereOptions(exact));
        tokens.add(whereToken);
    }

    public void _whereLessThanOrEqual(String fieldName, Object value) {
        _whereLessThanOrEqual(fieldName, value, false);
    }

    public void _whereLessThanOrEqual(String fieldName, Object value, boolean exact) {
        List<QueryToken> tokens = getCurrentWhereTokens();
        appendOperatorIfNeeded(tokens);
        negateIfNeeded(tokens, fieldName);

        WhereParams whereParams = new WhereParams();
        whereParams.setValue(value);
        whereParams.setFieldName(fieldName);

        String parameter = addQueryParameter(value == null ? "NULL" : transformValue(whereParams, true));
        WhereToken whereToken = WhereToken.create(WhereOperator.LESS_THAN_OR_EQUAL, fieldName, parameter, new WhereToken.WhereOptions(exact));
        tokens.add(whereToken);
    }

    @Override
    public void _whereRegex(String fieldName, String pattern) {
        List<QueryToken> tokens = getCurrentWhereTokens();
        appendOperatorIfNeeded(tokens);
        negateIfNeeded(tokens, fieldName);

        WhereParams whereParams = new WhereParams();
        whereParams.setValue(pattern);
        whereParams.setFieldName(fieldName);

        String parameter = addQueryParameter(transformValue(whereParams));

        WhereToken whereToken = WhereToken.create(WhereOperator.REGEX, fieldName, parameter);
        tokens.add(whereToken);
    }

    public void _andAlso() {
        List<QueryToken> tokens = getCurrentWhereTokens();
        if (tokens.isEmpty()) {
            return;
        }

        if (tokens.get(tokens.size() - 1) instanceof QueryOperatorToken) {
            throw new IllegalStateException("Cannot add AND, previous token was already an operator token.");
        }

        tokens.add(QueryOperatorToken.AND);
    }

    public void _orElse() {
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
    public void _boost(double boost) {
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
    public void _fuzzy(double fuzzy) {
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
    public void _proximity(int proximity) {
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

    public void _orderBy(String field) {
        _orderBy(field, OrderingType.STRING);
    }

    public void _orderBy(String field, OrderingType ordering) {
        assertNoRawQuery();
        String f = ensureValidFieldName(field, false);
        orderByTokens.add(OrderByToken.createAscending(f, ordering));
    }

    public void _orderByDescending(String field) {
        _orderByDescending(field, OrderingType.STRING);
    }

    public void _orderByDescending(String field, OrderingType ordering) {
        assertNoRawQuery();
        String f = ensureValidFieldName(field, false);
        orderByTokens.add(OrderByToken.createDescending(f, ordering));
    }

    public void _orderByScore() {
        assertNoRawQuery();

        orderByTokens.add(OrderByToken.scoreAscending);
    }

    public void _orderByScoreDescending() {
        assertNoRawQuery();
        orderByTokens.add(OrderByToken.scoreDescending);
    }

    public void _statistics(Reference<QueryStatistics> stats) {
        stats.value = queryStats;
    }

    public void invokeAfterQueryExecuted(QueryResult result) {
        EventHelper.invoke(afterQueryExecutedCallback, result);
    }

    public void invokeBeforeQueryExecuted(IndexQuery query) {
        EventHelper.invoke(beforeQueryExecutedCallback, query);
    }

    public void invokeAfterStreamExecuted(ObjectNode result) {
        EventHelper.invoke(afterStreamExecutedCallback, result);
    }

    protected IndexQuery generateIndexQuery(String query) {
        IndexQuery indexQuery = new IndexQuery();
        indexQuery.setQuery(query);
        indexQuery.setStart(start);
        indexQuery.setWaitForNonStaleResults(theWaitForNonStaleResults);
        indexQuery.setWaitForNonStaleResultsTimeout(timeout);
        indexQuery.setQueryParameters(queryParameters);
        indexQuery.setDisableCaching(disableCaching);

        if (pageSize != null) {
            indexQuery.setPageSize(pageSize);
        }
        return indexQuery;
    }

    @Override
    public void _search(String fieldName, String searchTerms) {
        _search(fieldName, searchTerms, SearchOperator.OR);
    }

    @Override
    public void _search(String fieldName, String searchTerms, SearchOperator operator) {
        List<QueryToken> tokens = getCurrentWhereTokens();
        appendOperatorIfNeeded(tokens);

        fieldName = ensureValidFieldName(fieldName, false);
        negateIfNeeded(tokens, fieldName);

        WhereToken whereToken = WhereToken.create(WhereOperator.SEARCH, fieldName, addQueryParameter(searchTerms), new WhereToken.WhereOptions(operator));
        tokens.add(whereToken);
    }

    @Override
    public String toString() {
        if (queryRaw != null) {
            return queryRaw;
        }

        if (_currentClauseDepth != 0) {
            throw new IllegalStateException("A clause was not closed correctly within this query, current clause depth = " + _currentClauseDepth);
        }

        StringBuilder queryText = new StringBuilder();
        buildDeclare(queryText);
        buildFrom(queryText);
        buildGroupBy(queryText);
        buildWhere(queryText);
        buildOrderBy(queryText);

        buildLoad(queryText);
        buildSelect(queryText);
        buildInclude(queryText);

        return queryText.toString();
    }

    private void buildInclude(StringBuilder queryText) {
        if (includes == null || includes.isEmpty()) {
            return;
        }

        queryText.append(" include ");
        boolean first = true;
        for (String include : includes) {
            if (!first) {
                queryText.append(",");
            }
            first = false;

            boolean requiredQuotes = false;

            for (int i = 0; i < include.length(); i++) {
                char ch = include.charAt(i);
                if (!Character.isLetterOrDigit(ch) && ch != '_' && ch != '.') {
                    requiredQuotes = true;
                    break;
                }
            }

            if (requiredQuotes) {
                queryText.append("'").append(include.replaceAll("'", "\\'")).append("'");
            } else {
                queryText.append(include);
            }
        }
    }

    @Override
    public void _intersect() {
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

    public void _whereExists(String fieldName) {
        fieldName = ensureValidFieldName(fieldName, false);

        List<QueryToken> tokens = getCurrentWhereTokens();
        appendOperatorIfNeeded(tokens);
        negateIfNeeded(tokens, fieldName);

        tokens.add(WhereToken.create(WhereOperator.EXISTS, fieldName, null));
    }

    @Override
    public void _containsAny(String fieldName, Collection<Object> values) {
        fieldName = ensureValidFieldName(fieldName, false);

        List<QueryToken> tokens = getCurrentWhereTokens();
        appendOperatorIfNeeded(tokens);
        negateIfNeeded(tokens, fieldName);

        Collection<Object> array = transformCollection(fieldName, unpackCollection(values));
        WhereToken whereToken = WhereToken.create(WhereOperator.IN, fieldName, addQueryParameter(array), new WhereToken.WhereOptions(false));
        tokens.add(whereToken);
    }

    @Override
    public void _containsAll(String fieldName, Collection<Object> values) {
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
    public void _addRootType(Class clazz) {
        rootTypes.add(clazz);
    }

    @Override
    public void _distinct() {
        if (isDistinct()) {
            throw new IllegalStateException("The is already a distinct query");
        }

        if (selectTokens.isEmpty()) {
            selectTokens.add(DistinctToken.INSTANCE);
        } else {
            selectTokens.add(0, DistinctToken.INSTANCE);
        }
    }

    private void updateStatsAndHighlightings(QueryResult queryResult) {
        queryStats.updateQueryStats(queryResult);
        //TBD 4.1 Highlightings.Update(queryResult);
    }

    private void buildSelect(StringBuilder writer) {
        if (selectTokens.isEmpty()) {
            return;
        }

        writer.append(" select ");
        if (selectTokens.size() == 1 && selectTokens.get(0) instanceof DistinctToken) {
            selectTokens.get(0).writeTo(writer);
            writer.append(" *");

            return;
        }

        for (int i = 0; i < selectTokens.size(); i++) {
            QueryToken token = selectTokens.get(i);
            if (i > 0 && !(selectTokens.get(i - 1) instanceof DistinctToken)) {
                writer.append(",");
            }

            DocumentQueryHelper.addSpaceIfNeeded(i > 0 ? selectTokens.get(i - 1) : null, token, writer);

            token.writeTo(writer);
        }
    }

    private void buildFrom(StringBuilder writer) {
        fromToken.writeTo(writer);
    }

    private void buildDeclare(StringBuilder writer) {
        if (declareToken != null) {
            declareToken.writeTo(writer);
        }
    }

    private void buildLoad(StringBuilder writer) {
        if (loadTokens == null || loadTokens.isEmpty()) {
            return;
        }

        writer.append(" load ");

        for (int i = 0; i < loadTokens.size(); i++) {
            if (i != 0) {
                writer.append(", ");
            }

            loadTokens.get(i).writeTo(writer);
        }
    }

    private void buildWhere(StringBuilder writer) {
        if (whereTokens.isEmpty()) {
            return;
        }

        writer
                .append(" where ");

        if (isIntersect) {
            writer
                    .append("intersect(");
        }

        for (int i = 0; i < whereTokens.size(); i++) {
            DocumentQueryHelper.addSpaceIfNeeded(i > 0 ? whereTokens.get(i - 1) : null, whereTokens.get(i), writer);
            whereTokens.get(i).writeTo(writer);
        }

        if (isIntersect) {
            writer.append(") ");
        }
    }

    private void buildGroupBy(StringBuilder writer) {
        if (groupByTokens.isEmpty()) {
            return;
        }

        writer
                .append(" group by ");

        boolean isFirst = true;

        for (QueryToken token : groupByTokens) {
            if (!isFirst) {
                writer.append(", ");
            }

            token.writeTo(writer);
            isFirst = false;
        }
    }

    private void buildOrderBy(StringBuilder writer) {
        if (orderByTokens.isEmpty()) {
            return;
        }

        writer
                .append(" order by ");

        boolean isFirst = true;

        for (QueryToken token : orderByTokens) {
            if (!isFirst) {
                writer.append(", ");
            }

            token.writeTo(writer);
            isFirst = false;
        }
    }

    private void appendOperatorIfNeeded(List<QueryToken> tokens) {
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
    private Collection<Object> transformCollection(String fieldName, Collection<Object> values) {
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

    private void negateIfNeeded(List<QueryToken> tokens, String fieldName) {
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

    private static Collection<Object> unpackCollection(Collection items) {
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

    private String ensureValidFieldName(String fieldName, boolean isNestedPath) {
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

    private Object transformValue(WhereParams whereParams) {
        return transformValue(whereParams, false);
    }

    private Object transformValue(WhereParams whereParams, boolean forRange) {
        if (whereParams.getValue() == null) {
            return null;
        }

        if ("".equals(whereParams.getValue())) {
            return "";
        }

        Reference<String> stringValueReference = new Reference<>();
        if (_conventions.tryConvertValueForQuery(whereParams.getFieldName(), whereParams.getValue(), forRange, stringValueReference)) {
            return stringValueReference.value;
        }

        Class<?> clazz = whereParams.getValue().getClass();
        if (Date.class.equals(clazz)) {
            return whereParams.getValue();
        }

        if (String.class.equals(clazz)) {
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

        if (String.class.equals(clazz)) {
            return whereParams.getValue();
        }

        if (Boolean.class.equals(clazz)) {
            return whereParams.getValue();
        }

        if (clazz.isEnum()) {
            return whereParams.getValue();
        }

        return whereParams.getValue();

    }

    private String addQueryParameter(Object value) {
        String parameterName = "p" + queryParameters.size();
        queryParameters.put(parameterName, value);
        return parameterName;
    }

    private List<QueryToken> getCurrentWhereTokens() {
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

    protected void updateFieldsToFetchToken(FieldsToFetchToken fieldsToFetch) {
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

    protected List<Consumer<IndexQuery>> beforeQueryExecutedCallback = new ArrayList<>();

    protected List<Consumer<QueryResult>> afterQueryExecutedCallback = new ArrayList<>();

    protected List<Consumer<ObjectNode>> afterStreamExecutedCallback = new ArrayList<>();

    protected QueryOperation queryOperation;

    public QueryOperation getQueryOperation() {
        return queryOperation;
    }

    public void _addBeforeQueryExecutedListener(Consumer<IndexQuery> action) {
        beforeQueryExecutedCallback.add(action);
    }

    public void _removeBeforeQueryExecutedListener(Consumer<IndexQuery> action) {
        beforeQueryExecutedCallback.remove(action);
    }

    public void _addAfterQueryExecutedListener(Consumer<QueryResult> action) {
        afterQueryExecutedCallback.add(action);
    }

    public void _removeAfterQueryExecutedListener(Consumer<QueryResult> action) {
        afterQueryExecutedCallback.remove(action);
    }

    public void _addAfterStreamExecutedListener(Consumer<ObjectNode> action) {
        afterStreamExecutedCallback.add(action);
    }

    public void _removeAfterStreamExecutedListener(Consumer<ObjectNode> action) {
        afterStreamExecutedCallback.remove(action);
    }

    public void _noTracking() {
        disableEntitiesTracking = true;
    }

    public void _noCaching() {
        disableCaching = true;
    }

    protected void _withinRadiusOf(String fieldName, double radius, double latitude, double longitude, SpatialUnits radiusUnits, double distErrorPercent) {
        fieldName = ensureValidFieldName(fieldName, false);

        List<QueryToken> tokens = getCurrentWhereTokens();
        appendOperatorIfNeeded(tokens);
        negateIfNeeded(tokens, fieldName);

        WhereToken whereToken = WhereToken.create(WhereOperator.SPATIAL_WITHIN, fieldName, null, new WhereToken.WhereOptions(ShapeToken.circle(addQueryParameter(radius), addQueryParameter(latitude), addQueryParameter(longitude), radiusUnits), distErrorPercent));
        tokens.add(whereToken);
    }

    protected void _spatial(String fieldName, String shapeWkt, SpatialRelation relation, double distErrorPercent) {
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
    public void _spatial(DynamicSpatialField dynamicField, SpatialCriteria criteria) {
        List<QueryToken> tokens = getCurrentWhereTokens();
        appendOperatorIfNeeded(tokens);
        negateIfNeeded(tokens, null);

        tokens.add(criteria.toQueryToken(dynamicField.toField(this::ensureValidFieldName), this::addQueryParameter));
    }

    @Override
    public void _spatial(String fieldName, SpatialCriteria criteria) {
        fieldName = ensureValidFieldName(fieldName, false);

        List<QueryToken> tokens = getCurrentWhereTokens();
        appendOperatorIfNeeded(tokens);
        negateIfNeeded(tokens, fieldName);

        tokens.add(criteria.toQueryToken(fieldName, this::addQueryParameter));
    }

    @Override
    public void _orderByDistance(DynamicSpatialField field, double latitude, double longitude) {
        if (field == null) {
            throw new IllegalArgumentException("Field cannot be null");
        }
        _orderByDistance("'" + field.toField(this::ensureValidFieldName) + "'", latitude, longitude);
    }

    @Override
    public void _orderByDistance(String fieldName, double latitude, double longitude) {
        orderByTokens.add(OrderByToken.createDistanceAscending(fieldName, addQueryParameter(latitude), addQueryParameter(longitude)));
    }

    @Override
    public void _orderByDistance(DynamicSpatialField field, String shapeWkt) {
        if (field == null) {
            throw new IllegalArgumentException("Field cannot be null");
        }
        _orderByDistance("'" + field.toField(this::ensureValidFieldName) + "'", shapeWkt);
    }

    @Override
    public void _orderByDistance(String fieldName, String shapeWkt) {
        orderByTokens.add(OrderByToken.createDistanceAscending(fieldName, addQueryParameter(shapeWkt)));
    }

    @Override
    public void _orderByDistanceDescending(DynamicSpatialField field, double latitude, double longitude) {
        if (field == null) {
            throw new IllegalArgumentException("Field cannot be null");
        }
        _orderByDistanceDescending("'" + field.toField(this::ensureValidFieldName) + "'", latitude, longitude);
    }

    @Override
    public void _orderByDistanceDescending(String fieldName, double latitude, double longitude) {
        orderByTokens.add(OrderByToken.createDistanceDescending(fieldName, addQueryParameter(latitude), addQueryParameter(longitude)));
    }

    @Override
    public void _orderByDistanceDescending(DynamicSpatialField field, String shapeWkt) {
        if (field == null) {
            throw new IllegalArgumentException("Field cannot be null");
        }
        _orderByDistanceDescending("'" + field.toField(this::ensureValidFieldName) + "'", shapeWkt);
    }

    @Override
    public void _orderByDistanceDescending(String fieldName, String shapeWkt) {
        orderByTokens.add(OrderByToken.createDistanceDescending(fieldName, addQueryParameter(shapeWkt)));
    }

    protected void initSync() {
        if (queryOperation != null) {
            return;
        }

        BeforeQueryEventArgs beforeQueryEventArgs = new BeforeQueryEventArgs(theSession, new DocumentQueryCustomizationDelegate(this));
        theSession.onBeforeQueryInvoke(beforeQueryEventArgs);

        queryOperation = initializeQueryOperation();
        executeActualQuery();
    }

    private void executeActualQuery() {
        try (CleanCloseable context = queryOperation.enterQueryContext()) {
            QueryCommand command = queryOperation.createRequest();
            theSession.getRequestExecutor().execute(command, theSession.sessionInfo);
            queryOperation.setResult(command.getResult());
        }
        invokeAfterQueryExecuted(queryOperation.getCurrentQueryResults());
    }

    @Override
    public Iterator<T> iterator() {
        return executeQueryOperation(null).iterator();
    }

    public List<T> toList() {
        return EnumerableUtils.toList(iterator());
    }

    public QueryResult getQueryResult() {
        initSync();

        return queryOperation.getCurrentQueryResults().createSnapshot();
    }

    public T first() {
        Collection<T> result = executeQueryOperation(1);
        return result.isEmpty() ? null : result.stream().findFirst().get();
    }

    public T firstOrDefault() {
        Collection<T> result = executeQueryOperation(1);
        return result.stream().findFirst().orElseGet(() -> Defaults.defaultValue(clazz));
    }

    public T single() {
        Collection<T> result = executeQueryOperation(2);
        if (result.size() > 1) {
            throw new IllegalStateException("Expected single result, got: " + result.size());
        }
        return result.stream().findFirst().orElse(null);
    }

    public T singleOrDefault() {
        Collection<T> result = executeQueryOperation(2);
        if (result.size() > 1) {
            throw new IllegalStateException("Expected single result, got: " + result.size());
        }
        if (result.isEmpty()) {
            return Defaults.defaultValue(clazz);
        }
        return result.stream().findFirst().get();
    }

    public int count() {
        _take(0);
        QueryResult queryResult = getQueryResult();
        return queryResult.getTotalResults();
    }

    public boolean any() {
        if (isDistinct()) {
            // for distinct it is cheaper to do count 1
            return executeQueryOperation(1).iterator().hasNext();
        }

        _take(0);
        QueryResult queryResult = getQueryResult();
        return queryResult.getTotalResults() > 0;
    }

    private Collection<T> executeQueryOperation(Integer take) {
        if (take != null && (pageSize == null || pageSize > take)) {
            _take(take);
        }

        initSync();

        return queryOperation.complete(clazz);
    }

    public void _aggregateBy(FacetBase facet) {
        for (QueryToken token : selectTokens) {
            if (token instanceof FacetToken) {
                continue;
            }

            throw new IllegalStateException("Aggregation query can select only facets while it got " + token.getClass().getSimpleName() + " token");
        }

        selectTokens.add(FacetToken.create(facet, this::addQueryParameter));
    }

    public void _aggregateUsing(String facetSetupDocumentId) {
        selectTokens.add(FacetToken.create(facetSetupDocumentId));
    }

    public Lazy<List<T>> lazily() {
        return lazily(null);
    }

    public Lazy<List<T>> lazily(Consumer<List<T>> onEval) {
        if (getQueryOperation() == null) {
            queryOperation = initializeQueryOperation();
        }

        LazyQueryOperation<T> lazyQueryOperation = new LazyQueryOperation<>(clazz, theSession.getConventions(), queryOperation, afterQueryExecutedCallback);
        return ((DocumentSession)theSession).addLazyOperation((Class<List<T>>) (Class<?>)List.class, lazyQueryOperation, onEval);
    }

    public Lazy<Integer> countLazily() {
        if (queryOperation == null) {
            _take(0);
            queryOperation = initializeQueryOperation();
        }

        LazyQueryOperation<T> lazyQueryOperation = new LazyQueryOperation<T>(clazz, theSession.getConventions(), queryOperation, afterQueryExecutedCallback);
        return ((DocumentSession)theSession).addLazyCountOperation(lazyQueryOperation);
    }

    @Override
    public void _suggestUsing(SuggestionBase suggestion) {
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

    private String getOptionsParameterName(SuggestionOptions options) {
        String optionsParameterName = null;
        if (options != null && options != SuggestionOptions.defaultOptions) {
            optionsParameterName = addQueryParameter(options);
        }

        return optionsParameterName;
    }

    private void assertCanSuggest() {
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
