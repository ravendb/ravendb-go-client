package ravendb

type Explanations struct {
	_explanations map[string][]string
}

func (e *Explanations) getExplanations(key string) []string {
	results := e._explanations[key]
	return results
}

func (e *Explanations) update(queryResult *QueryResult) {
	e._explanations = queryResult.Explanations
}
