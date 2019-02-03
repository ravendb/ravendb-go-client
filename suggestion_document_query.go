package ravendb

import (
	"reflect"
	"time"
)

type ISuggestionDocumentQuery = SuggestionDocumentQuery

type SuggestionDocumentQuery struct {
	// from SuggestionQueryBase
	session   *InMemoryDocumentSessionOperations
	query     *IndexQuery
	startTime time.Time

	source *DocumentQuery
}

func NewSuggestionDocumentQuery(source *DocumentQuery) *SuggestionDocumentQuery {
	return &SuggestionDocumentQuery{
		source:  source,
		session: source.getSession(),
	}
}

func (q *SuggestionDocumentQuery) Execute() (map[string]*SuggestionResult, error) {
	command, err := q.getCommand()
	if err != nil {
		return nil, err
	}

	q.startTime = time.Now()
	if err = q.session.incrementRequestCount(); err != nil {
		return nil, err
	}
	if err = q.session.GetRequestExecutor().ExecuteCommand(command, nil); err != nil {
		return nil, err
	}

	return q.processResults(command.Result, q.session.GetConventions())
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

	queryOperationEnsureIsAcceptable(queryResult, q.query.waitForNonStaleResults, q.startTime, q.session)

	return results, nil
}

// onEval: v is map[string]*SuggestionResult
func (q *SuggestionDocumentQuery) ExecuteLazy(results map[string]*SuggestionResult, onEval func(v interface{})) *Lazy {
	q.query = q.getIndexQuery()
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

	op := NewLazySuggestionQueryOperation(q.session.Conventions, q.query, afterFn, processFn)
	return q.session.session.addLazyOperation(results, op, onEval)
}

func (q *SuggestionDocumentQuery) InvokeAfterQueryExecuted(result *QueryResult) {
	q.source.invokeAfterQueryExecuted(result)
}

func (q *SuggestionDocumentQuery) getIndexQuery() *IndexQuery {
	return q.source.GetIndexQuery()
}
func (q *SuggestionDocumentQuery) getCommand() (*QueryCommand, error) {
	q.query = q.getIndexQuery()

	return NewQueryCommand(q.session.GetConventions(), q.query, false, false)
}

func (q *SuggestionDocumentQuery) String() string {
	return q.getIndexQuery().String()
}
