package ravendb

import "encoding/json"

var _ ILazyOperation = &LazySuggestionQueryOperation{}

type LazySuggestionQueryOperation struct {
	result        interface{}
	queryResult   *QueryResult
	requiresRetry bool

	_conventions              *DocumentConventions
	_indexQuery               *IndexQuery
	_invokeAfterQueryExecuted func(*QueryResult)
	_processResults           func(*QueryResult, *DocumentConventions) (map[string]*SuggestionResult, error)
}

func NewLazySuggestionQueryOperation(conventions *DocumentConventions, indexQuery *IndexQuery, invokeAfterQueryExecuted func(*QueryResult),
	processResults func(*QueryResult, *DocumentConventions) (map[string]*SuggestionResult, error)) *LazySuggestionQueryOperation {
	return &LazySuggestionQueryOperation{
		_conventions:              conventions,
		_indexQuery:               indexQuery,
		_invokeAfterQueryExecuted: invokeAfterQueryExecuted,
		_processResults:           processResults,
	}
}

func (o *LazySuggestionQueryOperation) createRequest() *GetRequest {
	return &GetRequest{
		url:     "/queries",
		method:  "POST",
		query:   "?queryHash=" + o._indexQuery.GetQueryHash(),
		content: NewIndexQueryContent(o._conventions, o._indexQuery),
	}
}

func (o *LazySuggestionQueryOperation) getResult() interface{} {
	return o.result
}

func (o *LazySuggestionQueryOperation) getQueryResult() *QueryResult {
	return o.queryResult
}

func (o *LazySuggestionQueryOperation) isRequiresRetry() bool {
	return o.requiresRetry
}

func (o *LazySuggestionQueryOperation) handleResponse(response *GetResponse) error {
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

func (o *LazySuggestionQueryOperation) handleResponse2(queryResult *QueryResult) error {
	if o._invokeAfterQueryExecuted != nil {
		o._invokeAfterQueryExecuted(queryResult)
	}

	var err error
	// TODO: is op._processResults always != nil ?
	o.result, err = o._processResults(queryResult, o._conventions)
	o.queryResult = queryResult
	return err
}
