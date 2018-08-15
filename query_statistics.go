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

func cloneMapStringFloat(m map[string]float64) map[string]float64 {
	if m == nil {
		return nil
	}
	res := map[string]float64{}
	for k, v := range m {
		res[k] = v
	}
	return res
}

func cloneMapStringString(m map[string]string) map[string]string {
	if m == nil {
		return nil
	}
	res := map[string]string{}
	for k, v := range m {
		res[k] = v
	}
	return res
}


func (s *QueryStatistics) Clone() *QueryStatistics {
	res := *s
	res.timingsInMs = cloneMapStringFloat(s.timingsInMs)
	res.scoreExplanations = cloneMapStringString(s.scoreExplanations)
	return &res
}

func (s *QueryStatistics) IsStale() bool {
	return s._isStale
}

func (s *QueryStatistics) SetStale(stale bool) {
	s._isStale = stale
}

func (s *QueryStatistics) GetDurationInMs() int64 {
	return s.durationInMs
}

func (s *QueryStatistics) SetDurationInMs(durationInMs int64) {
	s.durationInMs = durationInMs
}

func (s *QueryStatistics) GetTotalResults() int {
	return s.totalResults
}

func (s *QueryStatistics) SetTotalResults(totalResults int) {
	s.totalResults = totalResults
}

func (s *QueryStatistics) GetSkippedResults() int {
	return s.skippedResults
}

func (s *QueryStatistics) SetSkippedResults(skippedResults int) {
	s.skippedResults = skippedResults
}

func (s *QueryStatistics) GetTimestamp() time.Time {
	return s.timestamp
}

func (s *QueryStatistics) SetTimestamp(timestamp time.Time) {
	s.timestamp = timestamp
}

func (s *QueryStatistics) GetIndexName() string {
	return s.indexName
}

func (s *QueryStatistics) SetIndexName(indexName string) {
	s.indexName = indexName
}

func (s *QueryStatistics) GetIndexTimestamp() time.Time {
	return s.indexTimestamp
}

func (s *QueryStatistics) SetIndexTimestamp(indexTimestamp time.Time) {
	s.indexTimestamp = indexTimestamp
}

func (s *QueryStatistics) GetLastQueryTime() time.Time {
	return s.lastQueryTime
}

func (s *QueryStatistics) SetLastQueryTime(lastQueryTime time.Time) {
	s.lastQueryTime = lastQueryTime
}

func (s *QueryStatistics) GetTimingsInMs() map[string]float64 {
	return s.timingsInMs
}

func (s *QueryStatistics) SetTimingsInMs(timingsInMs map[string]float64) {
	s.timingsInMs = timingsInMs
}

func (s *QueryStatistics) GetResultEtag() int64 {
	return s.resultEtag
}

func (s *QueryStatistics) SetResultEtag(resultEtag int64) {
	s.resultEtag = resultEtag
}

func (s *QueryStatistics) GetResultSize() int64 {
	return s.resultSize
}

func (s *QueryStatistics) SetResultSize(resultSize int64) {
	s.resultSize = resultSize
}

func (s *QueryStatistics) UpdateQueryStats(qr *QueryResult) {
	s._isStale = qr.isStale()
	s.durationInMs = qr.getDurationInMs()
	s.totalResults = qr.getTotalResults()
	s.skippedResults = qr.getSkippedResults()
	s.timestamp = qr.getIndexTimestamp().toTime()
	s.indexName = qr.getIndexName()
	s.indexTimestamp = qr.getIndexTimestamp().toTime()
	s.timingsInMs = qr.getTimingsInMs()
	s.lastQueryTime = qr.getLastQueryTime().toTime()
	s.resultSize = qr.getResultSize()
	s.resultEtag = qr.getResultEtag()
	s.scoreExplanations = qr.getScoreExplanations()
}

func (s *QueryStatistics) GetScoreExplanations() map[string]string {
	return s.scoreExplanations
}

func (s *QueryStatistics) SetScoreExplanations(scoreExplanations map[string]string) {
	s.scoreExplanations = scoreExplanations
}
