package ravendb

// GetDocumentsResult is a result of GetDocument command
type GetDocumentsResult struct {
	Includes      ObjectNode `json:"Includes"`
	Results       ArrayNode  `json:"Results"`
	NextPageStart int        `json:"NextPageStart"`
}

func (r *GetDocumentsResult) GetIncludes() ObjectNode {
	return r.Includes
}

func (r *GetDocumentsResult) GetResults() ArrayNode {
	return r.Results
}

func (r *GetDocumentsResult) GetNextPageStart() int {
	return r.NextPageStart
}

/*
public void setIncludes(ObjectNode includes) {
	this.includes = includes;
}

public void setResults(ArrayNode results) {
	this.results = results;
}

public void setNextPageStart(int nextPageStart) {
	this.nextPageStart = nextPageStart;
}
*/
