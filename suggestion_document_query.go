package ravendb

import "reflect"

type ISuggestionDocumentQuery = SuggestionDocumentQuery

type SuggestionDocumentQuery struct {
	// from SuggestionQueryBase
	_session  *InMemoryDocumentSessionOperations
	_query    *IndexQuery
	_duration *Stopwatch

	_source *DocumentQuery
}

func NewSuggestionDocumentQuery(source *DocumentQuery) *SuggestionDocumentQuery {
	return &SuggestionDocumentQuery{
		_source:  source,
		_session: source.GetSession(),
	}
}

func (q *SuggestionDocumentQuery) Execute() (map[string]*SuggestionResult, error) {
	command := q.getCommand()

	q._duration = Stopwatch_createStarted()
	q._session.IncrementRequestCount()
	err := q._session.GetRequestExecutor().ExecuteCommand(command)
	if err != nil {
		return nil, err
	}

	return q.processResults(command.Result, q._session.GetConventions())
}

func (q *SuggestionDocumentQuery) processResults(queryResult *QueryResult, conventions *DocumentConventions) (map[string]*SuggestionResult, error) {
	q.InvokeAfterQueryExecuted(queryResult)

	results := map[string]*SuggestionResult{}

	jsResults := queryResult.getResults()

	for _, result := range jsResults {
		suggestionResult, err := treeToValue(reflect.TypeOf(&SuggestionResult{}), result)
		if err != nil {
			return nil, err
		}
		res := suggestionResult.(*SuggestionResult)
		results[res.Name] = res
	}

	QueryOperation_ensureIsAcceptable(queryResult, q._query.waitForNonStaleResults, q._duration, q._session)

	return results, nil
}

// onEval: v is map[string]*SuggestionResult
func (q *SuggestionDocumentQuery) ExecuteLazy(onEval func(v interface{})) *Lazy {
	q._query = q.getIndexQuery()
	clazz := reflect.TypeOf(map[string]*SuggestionResult{})
	afterFn := func(result *QueryResult) {
		q.InvokeAfterQueryExecuted(result)
	}
	processFn := func(queryResult *QueryResult, conventions *DocumentConventions) (map[string]*SuggestionResult, error) {
		return q.processResults(queryResult, conventions)
	}

	op := NewLazySuggestionQueryOperation(q._session.Conventions, q._query, afterFn, processFn)
	return q._session.session.addLazyOperation(clazz, op, onEval)
}

func (q *SuggestionDocumentQuery) InvokeAfterQueryExecuted(result *QueryResult) {
	q._source.InvokeAfterQueryExecuted(result)
}

func (q *SuggestionDocumentQuery) getIndexQuery() *IndexQuery {
	return q._source.GetIndexQuery()
}
func (q *SuggestionDocumentQuery) getCommand() *QueryCommand {
	q._query = q.getIndexQuery()

	return NewQueryCommand(q._session.GetConventions(), q._query, false, false)
}

func (q *SuggestionDocumentQuery) String() string {
	return q.getIndexQuery().String()
}
