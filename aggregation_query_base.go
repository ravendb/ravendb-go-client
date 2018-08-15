package ravendb

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

	q._session.incrementRequestCount()
	err := q._session.getRequestExecutor().ExecuteCommand(command)
	if err != nil {
		return nil, err
	}
	return q.processResults(command.Result, q._session.getConventions())
}

/* TODO:
func (q *AggregationQueryBase)      Lazy<map[string]*FacetResult> executeLazy() {
        return executeLazy(null)
    }

func (q *AggregationQueryBase)      Lazy<map[string]*FacetResult> executeLazy(Consumer<map[string]*FacetResult> onEval) {
        _query = getIndexQuery()
        return ((DocumentSession)_session).addLazyOperation((Class<map[string]*FacetResult>)(Class<?>)Map.class,
                new LazyAggregationQueryOperation( _session.getConventions(), _query, result -> invokeAfterQueryExecuted(result), this::processResults), onEval)
    }
*/

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
		res, err := convertValue(result, GetTypeOf(&FacetResult{}))
		if err != nil {
			return nil, err
		}
		facetResult := res.(*FacetResult)
		results[facetResult.GetName()] = facetResult
	}

	err := QueryOperation_ensureIsAcceptable(queryResult, q._query.isWaitForNonStaleResults(), q._duration, q._session)
	if err != nil {
		return nil, err
	}
	return results, nil
}

func (q *AggregationQueryBase) GetCommand() *QueryCommand {
	q._query = q.GetIndexQuery()

	return NewQueryCommand(q._session.getConventions(), q._query, false, false)
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
