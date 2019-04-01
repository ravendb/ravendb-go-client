package ravendb

// QueryResults represents results of a query
type QueryResult struct {
	GenericQueryResult
}

func (r *QueryResult) createSnapshot() *QueryResult {
	queryResult := *r

	// TODO: deep copy Explanations and Timings ?
	//queryResult.Explanations = r.Explanations
	//queryResult.Timings = r.Timings

	return &queryResult
}
