package ravendb

var _ ILazyOperation = &LazyAggregationQueryOperation{}

// LazyAggregationQueryOperation represents lazy aggregation query operation
type LazyAggregationQueryOperation struct {
	_conventions              *DocumentConventions
	_indexQuery               *IndexQuery
	_invokeAfterQueryExecuted func(*QueryResult)
	_processResults           func(*QueryResult, *DocumentConventions) (map[string]*FacetResult, error)

	result        interface{}
	queryResult   *QueryResult
	requiresRetry bool
}

// NewLazyAggregationQueryOperation returns LazyAggregationQueryOperation
func NewLazyAggregationQueryOperation(conventions *DocumentConventions, indexQuery *IndexQuery, invokeAfterQueryExecuted func(*QueryResult),
	processResults func(*QueryResult, *DocumentConventions) (map[string]*FacetResult, error)) *LazyAggregationQueryOperation {
	return &LazyAggregationQueryOperation{
		_conventions:              conventions,
		_indexQuery:               indexQuery,
		_invokeAfterQueryExecuted: invokeAfterQueryExecuted,
		_processResults:           processResults,
	}
}

func (o *LazyAggregationQueryOperation) createRequest() *GetRequest {
	request := &GetRequest{
		url:     "/queries",
		method:  "POST",
		query:   "?queryHash=" + o._indexQuery.GetQueryHash(),
		content: NewIndexQueryContent(o._conventions, o._indexQuery),
	}
	return request
}

// needed for ILazyOperation
func (o *LazyAggregationQueryOperation) getResult() interface{} {
	return o.result
}

// needed for ILazyOperation
func (o *LazyAggregationQueryOperation) getQueryResult() *QueryResult {
	return o.queryResult
}

// needed for ILazyOperation
func (o *LazyAggregationQueryOperation) isRequiresRetry() bool {
	return o.requiresRetry
}

func (o *LazyAggregationQueryOperation) handleResponse(response *GetResponse) error {
	if response.isForceRetry {
		o.result = nil
		o.requiresRetry = true
		return nil
	}

	var queryResult *QueryResult
	err := jsonUnmarshal(response.result, &queryResult)
	if err != nil {
		return err
	}
	return o.handleResponse2(queryResult)
}

func (o *LazyAggregationQueryOperation) handleResponse2(queryResult *QueryResult) error {
	var err error
	o._invokeAfterQueryExecuted(queryResult)
	o.result, err = o._processResults(queryResult, o._conventions)
	o.queryResult = queryResult
	return err
}
