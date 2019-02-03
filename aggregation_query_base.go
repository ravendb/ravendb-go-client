package ravendb

import (
	"reflect"
	"time"
)

// Note: AggregationQueryBase also includes AggregationDocumentQuery
type AggregationQueryBase struct {
	session   *InMemoryDocumentSessionOperations
	query     *IndexQuery
	startTime time.Time

	// from AggregationDocumentQuery
	source *AbstractDocumentQuery
}

func NewAggregationQueryBase(source *DocumentQuery) *AggregationQueryBase {
	return &AggregationQueryBase{
		session: source.getSession(),
		source:  source.AbstractDocumentQuery,
	}

}
func NewAggregationDocumentQuery(source *DocumentQuery) *AggregationDocumentQuery {
	return NewAggregationQueryBase(source)
}

func (q *AggregationQueryBase) Execute() (map[string]*FacetResult, error) {
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
func (q *AggregationQueryBase) ExecuteLazy(results map[string]*FacetResult, onEval func(interface{})) *Lazy {
	q.query = q.GetIndexQuery()

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
	op := NewLazyAggregationQueryOperation(q.session.Conventions, q.query, afterFn, processResultFn)
	return q.session.session.addLazyOperation(results, op, onEval)
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

func (q *AggregationQueryBase) processResults(queryResult *QueryResult, conventions *DocumentConventions) (map[string]*FacetResult, error) {
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

func (q *AggregationQueryBase) GetCommand() (*QueryCommand, error) {
	q.query = q.GetIndexQuery()

	return NewQueryCommand(q.session.GetConventions(), q.query, false, false)
}

func (q *AggregationQueryBase) String() string {
	return q.GetIndexQuery().String()
}

// from AggregationDocumentQuery
func (q *AggregationDocumentQuery) AndAggregateBy(builder func(IFacetBuilder)) *AggregationDocumentQuery {
	f := NewFacetBuilder()
	builder(f)

	return q.AndAggregateByFacet(f.getFacet())
}

func (q *AggregationDocumentQuery) AndAggregateByFacet(facet FacetBase) *AggregationDocumentQuery {
	q.source.aggregateBy(facet)
	return q
}

func (q *AggregationDocumentQuery) GetIndexQuery() *IndexQuery {
	return q.source.GetIndexQuery()
}

func (q *AggregationDocumentQuery) invokeAfterQueryExecuted(result *QueryResult) {
	q.source.invokeAfterQueryExecuted(result)
}
