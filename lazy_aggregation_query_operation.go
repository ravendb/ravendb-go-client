package ravendb

var _ ILazyOperation = &LazyAggregationQueryOperation{}

// LazyAggregationQueryOperation represents lazy aggregation query operation
type LazyAggregationQueryOperation struct {
	conventions              *DocumentConventions
	indexQuery               *IndexQuery
	invokeAfterQueryExecuted func(*QueryResult)
	processResults           func(*QueryResult, *DocumentConventions) (map[string]*FacetResult, error)

	result        map[string]*FacetResult
	queryResult   *QueryResult
	requiresRetry bool
}

func newLazyAggregationQueryOperation(conventions *DocumentConventions, indexQuery *IndexQuery, invokeAfterQueryExecuted func(*QueryResult),
	processResults func(*QueryResult, *DocumentConventions) (map[string]*FacetResult, error)) *LazyAggregationQueryOperation {
	return &LazyAggregationQueryOperation{
		conventions:              conventions,
		indexQuery:               indexQuery,
		invokeAfterQueryExecuted: invokeAfterQueryExecuted,
		processResults:           processResults,
	}
}

// needed for ILazyOperation
func (o *LazyAggregationQueryOperation) createRequest() *getRequest {
	request := &getRequest{
		url:     "/queries",
		method:  "POST",
		query:   "?queryHash=" + o.indexQuery.GetQueryHash(),
		content: NewIndexQueryContent(o.conventions, o.indexQuery),
	}
	return request
}

// needed for ILazyOperation
func (o *LazyAggregationQueryOperation) getResult(results interface{}) error {
	return setInterfaceToValue(results, o.result)
}

// needed for ILazyOperation
func (o *LazyAggregationQueryOperation) getQueryResult() *QueryResult {
	return o.queryResult
}

// needed for ILazyOperation
func (o *LazyAggregationQueryOperation) isRequiresRetry() bool {
	return o.requiresRetry
}

// needed for ILazyOperation
func (o *LazyAggregationQueryOperation) handleResponse(response *GetResponse) error {
	if response.IsForceRetry {
		o.result = nil
		o.requiresRetry = true
		return nil
	}

	var queryResult *QueryResult
	err := jsonUnmarshal(response.Result, &queryResult)
	if err != nil {
		return err
	}
	return o.handleResponse2(queryResult)
}

func (o *LazyAggregationQueryOperation) handleResponse2(queryResult *QueryResult) error {
	var err error
	o.invokeAfterQueryExecuted(queryResult)
	o.result, err = o.processResults(queryResult, o.conventions)
	o.queryResult = queryResult
	return err
}
