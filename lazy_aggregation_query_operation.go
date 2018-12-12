package ravendb

import "encoding/json"

var _ ILazyOperation = &LazyAggregationQueryOperation{}

type LazyAggregationQueryOperation struct {
	_conventions              *DocumentConventions
	_indexQuery               *IndexQuery
	_invokeAfterQueryExecuted func(*QueryResult)
	_processResults           func(*QueryResult, *DocumentConventions) (map[string]*FacetResult, error)

	result        interface{}
	queryResult   *QueryResult
	requiresRetry bool
}

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

func (o *LazyAggregationQueryOperation) getResult() interface{} {
	return o.result
}

func (o *LazyAggregationQueryOperation) setResult(result interface{}) {
	o.result = result
}

func (o *LazyAggregationQueryOperation) getQueryResult() *QueryResult {
	return o.queryResult
}

func (o *LazyAggregationQueryOperation) setQueryResult(queryResult *QueryResult) {
	o.queryResult = queryResult
}

func (o *LazyAggregationQueryOperation) isRequiresRetry() bool {
	return o.requiresRetry
}

func (o *LazyAggregationQueryOperation) setRequiresRetry(requiresRetry bool) {
	o.requiresRetry = requiresRetry
}

func (o *LazyAggregationQueryOperation) handleResponse(response *GetResponse) error {
	if response.isForceRetry {
		o.result = nil
		o.requiresRetry = true
		return nil
	}

	var queryResult *QueryResult
	err := json.Unmarshal(response.result, &queryResult)
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
