package ravendb

// JSONArrayResult describes server's JSON response to batch command
type JSONArrayResult struct {
	Results ArrayNode `json:"Results"`
}

func (r *JSONArrayResult) getResults() ArrayNode {
	return r.Results
}
