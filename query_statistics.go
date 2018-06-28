package ravendb

import "time"

// TODO: is time.Time here a *ServerTime?
// TODO: needs json annotations?
type QueryStatistics struct {
	_isStale          bool
	durationInMs      int64
	totalResults      int
	skippedResults    int
	timestamp         time.Time
	indexName         string
	indexTimestamp    time.Time
	lastQueryTime     time.Time
	timingsInMs       map[string]float64
	resultEtag        int64 // TODO: *int64 ?
	resultSize        int64
	scoreExplanations map[string]string
}

func NewQueryStatistics() *QueryStatistics {
	return &QueryStatistics{
		timingsInMs: make(map[string]float64),
	}
}

func (s *QueryStatistics) isStale() bool {
	return s._isStale
}

func (s *QueryStatistics) setStale(stale bool) {
	s._isStale = stale
}

func (s *QueryStatistics) getDurationInMs() int64 {
	return s.durationInMs
}

func (s *QueryStatistics) setDurationInMs(durationInMs int64) {
	s.durationInMs = durationInMs
}

func (s *QueryStatistics) getTotalResults() int {
	return s.totalResults
}

func (s *QueryStatistics) setTotalResults(totalResults int) {
	s.totalResults = totalResults
}

func (s *QueryStatistics) getSkippedResults() int {
	return s.skippedResults
}

func (s *QueryStatistics) setSkippedResults(skippedResults int) {
	s.skippedResults = skippedResults
}

func (s *QueryStatistics) getTimestamp() time.Time {
	return s.timestamp
}

func (s *QueryStatistics) setTimestamp(timestamp time.Time) {
	s.timestamp = timestamp
}

func (s *QueryStatistics) getIndexName() string {
	return s.indexName
}

func (s *QueryStatistics) setIndexName(indexName string) {
	s.indexName = indexName
}

func (s *QueryStatistics) getIndexTimestamp() time.Time {
	return s.indexTimestamp
}

func (s *QueryStatistics) setIndexTimestamp(indexTimestamp time.Time) {
	s.indexTimestamp = indexTimestamp
}

func (s *QueryStatistics) getLastQueryTime() time.Time {
	return s.lastQueryTime
}

func (s *QueryStatistics) setLastQueryTime(lastQueryTime time.Time) {
	s.lastQueryTime = lastQueryTime
}

func (s *QueryStatistics) getTimingsInMs() map[string]float64 {
	return s.timingsInMs
}

func (s *QueryStatistics) setTimingsInMs(timingsInMs map[string]float64) {
	s.timingsInMs = timingsInMs
}

func (s *QueryStatistics) getResultEtag() int64 {
	return s.resultEtag
}

func (s *QueryStatistics) setResultEtag(resultEtag int64) {
	s.resultEtag = resultEtag
}

func (s *QueryStatistics) getResultSize() int64 {
	return s.resultSize
}

func (s *QueryStatistics) setResultSize(resultSize int64) {
	s.resultSize = resultSize
}

func (s *QueryStatistics) updateQueryStats(qr *QueryResult) {
	panicIf(true, "NYI")
	/*
		        s._isStale = qr.isStale()
		        s.durationInMs = qr.getDurationInMs()
		        s.totalResults = qr.getTotalResults()
		        s.skippedResults = qr.getSkippedResults()
		        s.timestamp = qr.getIndexTimestamp()
		        s.indexName = qr.getIndexName()
		        s.indexTimestamp = qr.getIndexTimestamp()
		        s.timingsInMs = qr.getTimingsInMs()
		        s.lastQueryTime = qr.getLastQueryTime()
		        s.resultSize = qr.getResultSize()
		        s.resultEtag = qr.getResultEtag()
				s.scoreExplanations = qr.getScoreExplanations()
	*/
}

func (s *QueryStatistics) getScoreExplanations() map[string]string {
	return s.scoreExplanations
}

func (s *QueryStatistics) setScoreExplanations(scoreExplanations map[string]string) {
	s.scoreExplanations = scoreExplanations
}
