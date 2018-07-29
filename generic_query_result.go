package ravendb

type GenericQueryResult struct {
	QueryResultBase
	TotalResults   int `json:"TotalResults"`
	SkippedResults int `json:"SkippedResults"`
	//TBD 4.1  map[string]map[string]List<String>>> highlightings
	DurationInMs int64 `json:"DurationInMs"`

	// TODO: json-annotate? don't seem to be present in json
	ScoreExplanations map[string]string
	TimingsInMs       map[string]float64
	ResultSize        int64
}

func (r *GenericQueryResult) getTotalResults() int {
	return r.TotalResults
}

func (r *GenericQueryResult) setTotalResults(totalResults int) {
	r.TotalResults = totalResults
}

func (r *GenericQueryResult) getSkippedResults() int {
	return r.SkippedResults
}

func (r *GenericQueryResult) setSkippedResults(skippedResults int) {
	r.SkippedResults = skippedResults
}

//TBD 4.1  map[string]map[string]List<String>>> getHighlightings()
//TBD 4.1   setHighlightings(map[string]map[string]List<String>>> highlightings) {

func (r *GenericQueryResult) getDurationInMs() int64 {
	return r.DurationInMs
}

func (r *GenericQueryResult) setDurationInMs(durationInMs int64) {
	r.DurationInMs = durationInMs
}

func (r *GenericQueryResult) getScoreExplanations() map[string]string {
	return r.ScoreExplanations
}

func (r *GenericQueryResult) setScoreExplanations(scoreExplanations map[string]string) {
	r.ScoreExplanations = scoreExplanations
}

func (r *GenericQueryResult) getTimingsInMs() map[string]float64 {
	return r.TimingsInMs
}

func (r *GenericQueryResult) setTimingsInMs(timingsInMs map[string]float64) {
	r.TimingsInMs = timingsInMs
}

func (r *GenericQueryResult) getResultSize() int64 {
	return r.ResultSize
}

func (r *GenericQueryResult) setResultSize(resultSize int64) {
	r.ResultSize = resultSize
}
