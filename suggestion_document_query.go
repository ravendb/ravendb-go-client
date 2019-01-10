package ravendb

import "reflect"

type ISuggestionDocumentQuery = SuggestionDocumentQuery

type SuggestionDocumentQuery struct {
	// from SuggestionQueryBase
	_session  *InMemoryDocumentSessionOperations
	_query    *IndexQuery
	_duration *stopWatch

	_source *DocumentQuery
}

func NewSuggestionDocumentQuery(source *DocumentQuery) *SuggestionDocumentQuery {
	return &SuggestionDocumentQuery{
		_source:  source,
		_session: source.getSession(),
	}
}

func (q *SuggestionDocumentQuery) Execute() (map[string]*SuggestionResult, error) {
	command, err := q.getCommand()
	if err != nil {
		return nil, err
	}

	q._duration = newStopWatchStarted()
	if err = q._session.incrementRequestCount(); err != nil {
		return nil, err
	}
	if err = q._session.GetRequestExecutor().ExecuteCommand(command); err != nil {
		return nil, err
	}

	return q.processResults(command.Result, q._session.GetConventions())
}

func (q *SuggestionDocumentQuery) processResults(queryResult *QueryResult, conventions *DocumentConventions) (map[string]*SuggestionResult, error) {
	q.InvokeAfterQueryExecuted(queryResult)

	results := map[string]*SuggestionResult{}

	jsResults := queryResult.Results

	for _, result := range jsResults {
		suggestionResult, err := treeToValue(reflect.TypeOf(&SuggestionResult{}), result)
		if err != nil {
			return nil, err
		}
		res := suggestionResult.(*SuggestionResult)
		results[res.Name] = res
	}

	queryOperationEnsureIsAcceptable(queryResult, q._query.waitForNonStaleResults, q._duration, q._session)

	return results, nil
}

// onEval: v is map[string]*SuggestionResult
func (q *SuggestionDocumentQuery) ExecuteLazy(results map[string]*SuggestionResult, onEval func(v interface{})) *Lazy {
	q._query = q.getIndexQuery()
	afterFn := func(result *QueryResult) {
		q.InvokeAfterQueryExecuted(result)
	}
	processFn := func(queryResult *QueryResult, conventions *DocumentConventions) (map[string]*SuggestionResult, error) {
		res, err := q.processResults(queryResult, conventions)
		if err != nil {
			return nil, err
		}
		// processResult returns its own map, have to copy it to map provided by
		// caller as a result
		for k, v := range res {
			results[k] = v
		}
		return res, err
	}

	op := NewLazySuggestionQueryOperation(q._session.Conventions, q._query, afterFn, processFn)
	return q._session.session.addLazyOperation(results, op, onEval)
}

func (q *SuggestionDocumentQuery) InvokeAfterQueryExecuted(result *QueryResult) {
	q._source.invokeAfterQueryExecuted(result)
}

func (q *SuggestionDocumentQuery) getIndexQuery() *IndexQuery {
	return q._source.GetIndexQuery()
}
func (q *SuggestionDocumentQuery) getCommand() (*QueryCommand, error) {
	q._query = q.getIndexQuery()

	return NewQueryCommand(q._session.GetConventions(), q._query, false, false)
}

func (q *SuggestionDocumentQuery) String() string {
	return q.getIndexQuery().String()
}
