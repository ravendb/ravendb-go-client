package ravendb

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

func newLazySuggestionQueryOperation(conventions *DocumentConventions, indexQuery *IndexQuery, invokeAfterQueryExecuted func(*QueryResult),
	processResults func(*QueryResult, *DocumentConventions) (map[string]*SuggestionResult, error)) *LazySuggestionQueryOperation {
	return &LazySuggestionQueryOperation{
		_conventions:              conventions,
		_indexQuery:               indexQuery,
		_invokeAfterQueryExecuted: invokeAfterQueryExecuted,
		_processResults:           processResults,
	}
}

// needed for ILazyOperation
func (o *LazySuggestionQueryOperation) createRequest() *getRequest {
	return &getRequest{
		url:     "/queries",
		method:  "POST",
		query:   "?queryHash=" + o._indexQuery.GetQueryHash(),
		content: NewIndexQueryContent(o._conventions, o._indexQuery),
	}
}

// needed for ILazyOperation
func (o *LazySuggestionQueryOperation) getResult(result interface{}) error {
	return setInterfaceToValue(result, o.result)
}

func (o *LazySuggestionQueryOperation) getQueryResult() *QueryResult {
	return o.queryResult
}

// needed for ILazyOperation
func (o *LazySuggestionQueryOperation) isRequiresRetry() bool {
	return o.requiresRetry
}

// needed for ILazyOperation
func (o *LazySuggestionQueryOperation) handleResponse(response *GetResponse) error {
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
