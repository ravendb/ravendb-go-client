package ravendb

var _ ILazyOperation = &LazyQueryOperation{}

// LazyQueryOperation describes server operation for lazy queries
type LazyQueryOperation struct {
	_conventions        *DocumentConventions
	_queryOperation     *queryOperation
	_afterQueryExecuted []func(*QueryResult)

	queryResult   *QueryResult
	requiresRetry bool
}

// newLazyQueryOperation returns new LazyQueryOperation
func newLazyQueryOperation(conventions *DocumentConventions, queryOperation *queryOperation, afterQueryExecuted []func(*QueryResult)) *LazyQueryOperation {
	return &LazyQueryOperation{
		_conventions:        conventions,
		_queryOperation:     queryOperation,
		_afterQueryExecuted: afterQueryExecuted,
	}
}

// needed for ILazyOperation
func (o *LazyQueryOperation) createRequest() *getRequest {
	return &getRequest{
		url:     "/queries",
		method:  "POST",
		query:   "?queryHash=" + o._queryOperation.indexQuery.GetQueryHash(),
		content: NewIndexQueryContent(o._conventions, o._queryOperation.indexQuery),
	}
}

// needed for ILazyOperation
func (o *LazyQueryOperation) getResult(result interface{}) error {
	return o._queryOperation.complete(result)
}

// needed for ILazyOperation
func (o *LazyQueryOperation) getQueryResult() *QueryResult {
	return o.queryResult
}

// needed for ILazyOperation
func (o *LazyQueryOperation) isRequiresRetry() bool {
	return o.requiresRetry
}

// needed for ILazyOperation
func (o *LazyQueryOperation) handleResponse(response *GetResponse) error {
	if response.IsForceRetry {
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

func (o *LazyQueryOperation) handleResponse2(queryResult *QueryResult) error {
	err := o._queryOperation.ensureIsAcceptableAndSaveResult(queryResult)
	if err != nil {
		return err
	}

	for _, e := range o._afterQueryExecuted {
		e(queryResult)
	}
	o.queryResult = queryResult
	return nil
}
