package ravendb

import (
	"reflect"
	"time"
)

// Note: AggregationQueryBase also includes AggregationDocumentQuery
type aggregationQueryBase struct {
	session   *InMemoryDocumentSessionOperations
	query     *IndexQuery
	startTime time.Time

	// from AggregationDocumentQuery
	source *abstractDocumentQuery
	err    error
}

func newAggregationQueryBase(source *DocumentQuery) *aggregationQueryBase {
	res := &aggregationQueryBase{
		session: source.getSession(),
		source:  source.abstractDocumentQuery,
	}
	res.err = source.err
	return res
}

func newAggregationDocumentQuery(source *DocumentQuery) *AggregationDocumentQuery {
	return newAggregationQueryBase(source)
}

func (q *aggregationQueryBase) Execute() (map[string]*FacetResult, error) {
	if q.err != nil {
		return nil, q.err
	}
	command, err := q.GetCommand()
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

// arg to onEval is map[string]*FacetResult
// results is map[string]*FacetResult
func (q *aggregationQueryBase) ExecuteLazy(results map[string]*FacetResult, onEval func(interface{})) (*Lazy, error) {
	if q.err != nil {
		return nil, q.err
	}

	var err error
	q.query, err = q.GetIndexQuery()
	if err != nil {
		return nil, err
	}

	afterFn := func(result *QueryResult) {
		q.invokeAfterQueryExecuted(result)
	}

	processResultFn := func(queryResult *QueryResult, conventions *DocumentConventions) (map[string]*FacetResult, error) {
		res, err := q.processResults(queryResult, conventions)
		if err != nil {
			return nil, err
		}
		// processResult returns its own map, have to copy it to map provided by
		// caller as a result
		for k, v := range res {
			results[k] = v
		}
		return res, nil
	}
	op := newLazyAggregationQueryOperation(q.session.Conventions, q.query, afterFn, processResultFn)
	return q.session.session.addLazyOperation(results, op, onEval), nil
}

/*
// abstract
func (q *AggregationQueryBase) getIndexQuery() *IndexQuery {
	return nil
}

//	  abstract
func (q *AggregationQueryBase) invokeAfterQueryExecuted(result *QueryResult) {
}
*/

func (q *aggregationQueryBase) processResults(queryResult *QueryResult, conventions *DocumentConventions) (map[string]*FacetResult, error) {
	q.invokeAfterQueryExecuted(queryResult)

	results := map[string]*FacetResult{}
	for _, result := range queryResult.Results {
		res, err := convertValue(result, reflect.TypeOf(&FacetResult{}))
		if err != nil {
			return nil, err
		}
		facetResult := res.(*FacetResult)
		results[facetResult.Name] = facetResult
	}

	err := queryOperationEnsureIsAcceptable(queryResult, q.query.waitForNonStaleResults, q.startTime, q.session)
	if err != nil {
		return nil, err
	}
	return results, nil
}

func (q *aggregationQueryBase) GetCommand() (*QueryCommand, error) {
	if q.err != nil {
		return nil, q.err
	}
	var err error
	q.query, err = q.GetIndexQuery()
	if err != nil {
		return nil, err
	}

	return NewQueryCommand(q.session.GetConventions(), q.query, false, false)
}

func (q *aggregationQueryBase) string() (string, error) {
	if q.err != nil {
		return "", q.err
	}
	iq, err := q.GetIndexQuery()
	if err != nil {
		return "", err
	}
	return iq.String(), nil
}

// from AggregationDocumentQuery
func (q *AggregationDocumentQuery) AndAggregateByFacet(facet FacetBase) *AggregationDocumentQuery {
	if q.err != nil {
		return q
	}
	q.err = q.source.aggregateBy(facet)
	return q
}

func (q *AggregationDocumentQuery) GetIndexQuery() (*IndexQuery, error) {
	return q.source.GetIndexQuery()
}

func (q *AggregationDocumentQuery) invokeAfterQueryExecuted(result *QueryResult) {
	q.source.invokeAfterQueryExecuted(result)
}
