package ravendb

import (
	"reflect"
	"time"
)

// Note: ISuggestionDocumentQuery is SuggestionDocumentQuery

// SuggestionDocumentQuery represents "suggestion" query
type SuggestionDocumentQuery struct {
	// from SuggestionQueryBase
	session   *InMemoryDocumentSessionOperations
	query     *IndexQuery
	startTime time.Time

	source *DocumentQuery
	err    error
}

func newSuggestionDocumentQuery(source *DocumentQuery) *SuggestionDocumentQuery {
	res := &SuggestionDocumentQuery{
		source:  source,
		session: source.getSession(),
	}
	res.err = source.err
	return res
}

func (q *SuggestionDocumentQuery) Execute() (map[string]*SuggestionResult, error) {
	if q.err != nil {
		return nil, q.err
	}
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

	err := queryOperationEnsureIsAcceptable(queryResult, q.query.waitForNonStaleResults, q.startTime, q.session)
	if err != nil {
		return nil, err
	}

	return results, nil
}

// onEval: v is map[string]*SuggestionResult
func (q *SuggestionDocumentQuery) ExecuteLazy() (*Lazy, error) {
	if q.err != nil {
		return nil, q.err
	}
	var err error
	q.query, err = q.getIndexQuery()
	if err != nil {
		return nil, err
	}
	afterFn := func(result *QueryResult) {
		q.InvokeAfterQueryExecuted(result)
	}
	processFn := func(queryResult *QueryResult, conventions *DocumentConventions) (map[string]*SuggestionResult, error) {
		res, err := q.processResults(queryResult, conventions)
		if err != nil {
			return nil, err
		}
		return res, nil
	}

	op := newLazySuggestionQueryOperation(q.session.Conventions, q.query, afterFn, processFn)
	return q.session.session.addLazyOperation(op, nil, nil), nil
}

func (q *SuggestionDocumentQuery) InvokeAfterQueryExecuted(result *QueryResult) {
	q.source.invokeAfterQueryExecuted(result)
}

func (q *SuggestionDocumentQuery) getIndexQuery() (*IndexQuery, error) {
	if q.err != nil {
		return nil, q.err
	}
	return q.source.GetIndexQuery()
}

func (q *SuggestionDocumentQuery) getCommand() (*QueryCommand, error) {
	if q.err != nil {
		return nil, q.err
	}
	var err error
	q.query, err = q.getIndexQuery()
	if err != nil {
		return nil, err
	}

	return NewQueryCommand(q.session.GetConventions(), q.query, false, false)
}

func (q *SuggestionDocumentQuery) string() (string, error) {
	if q.err != nil {
		return "", q.err
	}
	iq, err := q.getIndexQuery()
	if err != nil {
		return "", err
	}

	return iq.String(), nil
}
