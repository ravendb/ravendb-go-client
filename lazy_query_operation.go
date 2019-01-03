package ravendb

var _ ILazyOperation = &LazyQueryOperation{}

// LazyQueryOperation describes server operation for lazy queries
type LazyQueryOperation struct {
	_conventions        *DocumentConventions
	_queryOperation     *QueryOperation
	_afterQueryExecuted []func(*QueryResult)

	result        interface{}
	queryResult   *QueryResult
	requiresRetry bool
}

// NewLazyQueryOperation returns new LazyQueryOperation
func NewLazyQueryOperation(result interface{}, conventions *DocumentConventions, queryOperation *QueryOperation, afterQueryExecuted []func(*QueryResult)) *LazyQueryOperation {
	return &LazyQueryOperation{
		result:              result,
		_conventions:        conventions,
		_queryOperation:     queryOperation,
		_afterQueryExecuted: afterQueryExecuted,
	}
}

func (o *LazyQueryOperation) createRequest() *GetRequest {
	return &GetRequest{
		url:     "/queries",
		method:  "POST",
		query:   "?queryHash=" + o._queryOperation.indexQuery.GetQueryHash(),
		content: NewIndexQueryContent(o._conventions, o._queryOperation.indexQuery),
	}
}

// needed for ILazyOperation
func (o *LazyQueryOperation) getResult() interface{} {
	return o.result
}

// needed for ILazyOperation
func (o *LazyQueryOperation) getQueryResult() *QueryResult {
	return o.queryResult
}

// needed for ILazyOperation
func (o *LazyQueryOperation) isRequiresRetry() bool {
	return o.requiresRetry
}

func (o *LazyQueryOperation) handleResponse(response *GetResponse) error {
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

func (o *LazyQueryOperation) handleResponse2(queryResult *QueryResult) error {
	o._queryOperation.ensureIsAcceptableAndSaveResult(queryResult)

	for _, e := range o._afterQueryExecuted {
		e(queryResult)
	}
	o.queryResult = queryResult
	err := o._queryOperation.complete(o.result)
	return err
}
