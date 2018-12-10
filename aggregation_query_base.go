package ravendb

import "reflect"

// Note: AggregationQueryBase also includes AggregationDocumentQuery
type AggregationQueryBase struct {
	_session  *InMemoryDocumentSessionOperations
	_query    *IndexQuery
	_duration *Stopwatch

	// from AggregationDocumentQuery
	_source *AbstractDocumentQuery
}

func NewAggregationQueryBase(source *DocumentQuery) *AggregationQueryBase {
	return &AggregationQueryBase{
		_session: source.GetSession(),
		_source:  source.AbstractDocumentQuery,
	}

}
func NewAggregationDocumentQuery(source *DocumentQuery) *AggregationDocumentQuery {
	return NewAggregationQueryBase(source)
}

func (q *AggregationQueryBase) Execute() (map[string]*FacetResult, error) {
	command := q.GetCommand()

	q._duration = Stopwatch_createStarted()

	q._session.IncrementRequestCount()
	err := q._session.GetRequestExecutor().ExecuteCommand(command)
	if err != nil {
		return nil, err
	}
	return q.processResults(command.Result, q._session.GetConventions())
}

// arg to onEval is map[string]*FacetResult
// returns Lazy<map[string]*FacetResult>
func (q *AggregationQueryBase) ExecuteLazy(onEval func(interface{})) *Lazy {
	q._query = q.GetIndexQuery()

	afterFn := func(result *QueryResult) {
		q.invokeAfterQueryExecuted(result)
	}

	processResultFn := func(queryResult *QueryResult, conventions *DocumentConventions) (map[string]*FacetResult, error) {
		return q.processResults(queryResult, conventions)
	}
	op := NewLazyAggregationQueryOperation(q._session.Conventions, q._query, afterFn, processResultFn)
	clazz := reflect.TypeOf(map[string]*FacetResult{})
	return q._session.session.addLazyOperation(clazz, op, onEval)
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

	err := QueryOperation_ensureIsAcceptable(queryResult, q._query.waitForNonStaleResults, q._duration, q._session)
	if err != nil {
		return nil, err
	}
	return results, nil
}

func (q *AggregationQueryBase) GetCommand() *QueryCommand {
	q._query = q.GetIndexQuery()

	return NewQueryCommand(q._session.GetConventions(), q._query, false, false)
}

func (q *AggregationQueryBase) String() string {
	return q.GetIndexQuery().String()
}

// from AggregationDocumentQuery
func (q *AggregationDocumentQuery) AndAggregateBy(builder func(IFacetBuilder)) *IAggregationDocumentQuery {
	f := NewFacetBuilder()
	builder(f)

	return q.AndAggregateByFacet(f.getFacet())
}

func (q *AggregationDocumentQuery) AndAggregateByFacet(facet FacetBase) *IAggregationDocumentQuery {
	q._source._aggregateBy(facet)
	return q
}

func (q *AggregationDocumentQuery) GetIndexQuery() *IndexQuery {
	return q._source.GetIndexQuery()
}

func (q *AggregationDocumentQuery) invokeAfterQueryExecuted(result *QueryResult) {
	q._source.InvokeAfterQueryExecuted(result)
}
