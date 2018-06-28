package ravendb

// TODO: json-annotate?
type GenericQueryResult struct {
	QueryResultBase
	totalResults   int
	skippedResults int
	//TBD 4.1  map[string]map[string]List<String>>> highlightings
	durationInMs      int64
	scoreExplanations map[string]string
	timingsInMs       map[string]float64
	resultSize        int64
}

func (r *GenericQueryResult) getTotalResults() int {
	return r.totalResults
}

func (r *GenericQueryResult) setTotalResults(totalResults int) {
	r.totalResults = totalResults
}

func (r *GenericQueryResult) getSkippedResults() int {
	return r.skippedResults
}

func (r *GenericQueryResult) setSkippedResults(skippedResults int) {
	r.skippedResults = skippedResults
}

//TBD 4.1  map[string]map[string]List<String>>> getHighlightings()
//TBD 4.1   setHighlightings(map[string]map[string]List<String>>> highlightings) {

func (r *GenericQueryResult) getDurationInMs() int64 {
	return r.durationInMs
}

func (r *GenericQueryResult) setDurationInMs(durationInMs int64) {
	r.durationInMs = durationInMs
}

func (r *GenericQueryResult) getScoreExplanations() map[string]string {
	return r.scoreExplanations
}

func (r *GenericQueryResult) setScoreExplanations(scoreExplanations map[string]string) {
	r.scoreExplanations = scoreExplanations
}

func (r *GenericQueryResult) getTimingsInMs() map[string]float64 {
	return r.timingsInMs
}

func (r *GenericQueryResult) setTimingsInMs(timingsInMs map[string]float64) {
	r.timingsInMs = timingsInMs
}

func (r *GenericQueryResult) getResultSize() int64 {
	return r.resultSize
}

func (r *GenericQueryResult) setResultSize(resultSize int64) {
	r.resultSize = resultSize
}
